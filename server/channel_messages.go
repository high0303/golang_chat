//
//  channel_messages.go
//  Sechat Server
//
//  Created by AX on 2017-02-25.
//  Copyright 2017 JW Technologies. All rights reserved.
//

package main

import (
	"github.com/jinzhu/gorm"
	"sechat-server/lib"
	"sechat-server/scutil"
	//"time"
)

type ClientEndToEndChannelMessage struct {
	lib.ClientMessage
	Channel             string
	Encrypted           []byte
	Signature           string
	Highlight           []string
	Notification        int
	ImmediateOnly       bool
	Priority            int
	MessageToken        string
	RepeatNotifications bool
}

type ServerEndToEndChannelMessage struct {
	lib.ServerMessage
	Channel   string
	Encrypted []byte
	Signature string
	// TODO can we hide the sender somehow?
	//      maybe use keyIDs? difficult i guess.
	Sender   string
	Priority int
}

type ClientGetChannelsMessage struct {
	lib.ClientMessage
}

type ServerGetChannelsResponseMessage struct {
	lib.ServerMessage
	Channels []string
}

type ClientCreateChannelMessage struct {
	lib.ClientMessage
	Channel     string
	Description string
}

type ServerCreateChannelResponseMessage struct {
	lib.ServerMessage
	Channel    string
	Successful bool
}

type ClientJoinChannelMessage struct {
	lib.ClientMessage
	Channel string
}

type ClientLeaveChannelMessage struct {
	lib.ClientMessage
	Channel string
}

type ServerJoinChannelResponseMessage struct {
	lib.ServerMessage
	Channel    string
	Successful bool
}

type ServerChannelInfoMessage struct {
	lib.ServerMessage
	Channel string
	Joined  string
	Left    string
	Users   []string
}

func handleClientEndToEndChannelMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")
	var msg ClientEndToEndChannelMessage
	lib.MpUnmarshal(message, &msg)

	var channel lib.Channel
	db.Where("name = ?", msg.Channel).First(&channel)

	var users []lib.User
	db.Model(&channel).Related(&users, "Users")

	var notificationsDisabled []lib.User
	db.Model(&channel).Related(&notificationsDisabled, "NotificationsDisabled")

	var highlightNotificationsDisabled []lib.User
	db.Model(&channel).Related(&highlightNotificationsDisabled, "HighlightNotificationsDisabled")

	isInChannel := false

	for _, user := range users {
		scutil.SLog("User: %s", user.Name)
		if user.ID == conn.User.ID {
			isInChannel = true
		}
	}

	if !isInChannel {
		// TODO send error to client...
		panic("handleClientEndToEndChannelMessage: Message to channel of which user is not a member.")
	}

	if len(msg.MessageToken) > 0 {
		scutil.SLog("Send e2e response")
		e2eResponse := lib.ServerEndToEndResponseMessage{MessageId: msg.MessageToken}
		setMessageType(&e2eResponse)
		sendAnswer(conn, e2eResponse, "ServerEndToEndResponseMessage")
	}

	highlight := make([]lib.User, len(msg.Highlight))
	for i, highlightedUsername := range msg.Highlight {
		scutil.SLog("Highlights user: %s", highlightedUsername)
		highlightedUser := getUser(db, highlightedUsername)
		if highlightedUser == nil {
			scutil.SLog("Error: User %s not found.", highlightedUsername)
			continue
		}

		highlight[i] = *highlightedUser
	}

	outMsg := ServerEndToEndChannelMessage{
		Channel:   channel.Name,
		Encrypted: msg.Encrypted,
		Signature: msg.Signature,
		Sender:    conn.User.Username,
		Priority:  10,
	}
	setMessageType(&outMsg)

    var messageID int64
 	if (!msg.ImmediateOnly) {
        // [1] Prepare stored message
        messageID = prepareStoredMessage(db, conn, outMsg, msg.MessageToken, "")
    }
    
	for _, user := range users {
		// Ignore ourselves..
		if user.ID == conn.User.ID {
			continue
		}

		notification := true

		if user.IsIn(highlight) {
			if !user.IsIn(highlightNotificationsDisabled) {
				scutil.SLog("Sending highlight notification to %s", user.Username)
				go sendAPNS(db, user, lib.NOTIFICATION_TYPE_HIGHLIGHT)
			}
		} else if !user.IsIn(notificationsDisabled) {
			scutil.SLog("NOT Sending notification to %s", user.Username)
			//go sendAPNS(db, user, msg.Notification)
		} else {
			scutil.SLog("Notifications blocked for %s", user.Username)
			notification = false
		}

		if msg.Notification == lib.NOTIFICATION_TYPE_NONE {
			notification = false
		}

		scutil.SLog("Sending to user %s from %s", user.Username, conn.User.Username)
        
		//sendToUserDetailed(conn, &user, outMsg, msg.ImmediateOnly, "", msg.MessageId, notification, msg.RepeatNotifications)
        if (msg.ImmediateOnly) {
 			sendImmediateMessage(conn, &user, outMsg)
        } else {
            addMessageRecipient(db,
                         messageID,              // db message id
                         &user,                      // user
                         0,                         // device to exclude
                         notification,                      // notification
                         msg.RepeatNotifications)   // repeat notifications
       }
        
         
	}
    
 	if (!msg.ImmediateOnly) {
        // [3] Trigger channel to send the stored message to all the recipients
        sendStoredMessage(db, conn, messageID)
    }
    
}

func handleClientGetChannelsMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")

	var msg ClientGetChannelsMessage
	lib.MpUnmarshal(message, &msg)

	scutil.SLog("Getting channels")
	var channels []lib.Channel
	db.Find(&channels)

	var response ServerGetChannelsResponseMessage
	setMessageType(&response)

	response.Channels = make([]string, 0, len(channels))
	for _, channel := range channels {
		response.Channels = append(response.Channels, channel.Name)
	}
	sendAnswer(conn, response, "ServerGetChannelsResponseMessage")
}

func handleClientCreateChannelMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")

	var msg ClientCreateChannelMessage
	lib.MpUnmarshal(message, &msg)

	var channelExists lib.Channel
	db.Where("name = ?", msg.Channel).First(&channelExists)

	if channelExists.Id != 0 {
		var response ServerCreateChannelResponseMessage
		setMessageType(&response)

		response.Channel = msg.Channel
		response.Successful = false
		sendAnswer(conn, response, "ServerCreateChannelResponseMessage:false")
		return
	}
	channel := lib.Channel{
		Name:        msg.Channel,
		Description: msg.Description,
		Creator:     conn.User,
	}
	db.Save(&channel)

	var response ServerCreateChannelResponseMessage
	setMessageType(&response)

	response.Channel = msg.Channel
	response.Successful = true
	sendAnswer(conn, response, "ServerCreateChannelResponseMessage:true")
}

func handleClientJoinChannelMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.DebugLog("handleClientJoinChannelMessage")

	var msg ClientJoinChannelMessage
	lib.MpUnmarshal(message, &msg)

	var channel lib.Channel
	db.Where("name = ?", msg.Channel).First(&channel)

	if channel.Id == 0 {
		var response ServerJoinChannelResponseMessage
		setMessageType(&response)

		response.Channel = msg.Channel
		response.Successful = false
		sendAnswer(conn, response, "ServerJoinChannelResponseMessage:false")
		scutil.DebugLog("ServerJoinChannelResponseMessage false")
		//return
	} else {
		//db.Model(channel).Association("Users").Append(conn.User)
		// Association is not working with the GORM do a raw insert into the join table (workaround)
		var channelUser = lib.ChannelUser{ChannelId: channel.Id, UserId: conn.User.ID}
		var response ServerJoinChannelResponseMessage
		err := db.Save(&channelUser).Error
		if (err != nil) {
			scutil.DebugLog("JoinChannel add channel user failed")
			response.Successful = false
		} else {
			response.Successful = true
		}
	
		setMessageType(&response)

		response.Channel = msg.Channel
		sendAnswer(conn, response, "ServerJoinChannelResponseMessage:true")
		sendChannelInfo(channel, conn.User.Username, "", conn, db)
	}

}

func handleClientLeaveChannelMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")

	var msg ClientLeaveChannelMessage
	lib.MpUnmarshal(message, &msg)

	var channel lib.Channel
	db.Where("name = ?", msg.Channel).First(&channel)

	db.Model(channel).Association("Users").Delete(conn.User)

	// TODO: Send a response here?

	sendChannelInfo(channel, "", conn.User.Username, conn, db)
}

func sendChannelInfo(channel lib.Channel, joined string, left string, conn *lib.Connection, db gorm.DB) {
	scutil.SLog("Begin")
	var users []lib.User
	db.Model(&channel).Related(&users, "Users")

	infoMessage := ServerChannelInfoMessage{
		Channel: channel.Name,
		Users:   make([]string, 0, len(users)),
		Joined:  joined,
		Left:    left,
	}
	setMessageType(&infoMessage)
	for _, user := range users {
		infoMessage.Users = append(infoMessage.Users, user.Username)
	}

	for _, user := range users {
		sendToUser(conn, &user, infoMessage, "ServerChannelInfoMessage")
	}
}
