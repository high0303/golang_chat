//
//  profile_messages.go
//  Sechat Server
//
//  Copyright 2017 JW Technologies. All rights reserved.
//

package main

import (
	"sechat-server/lib"
	"github.com/jinzhu/gorm"
	"sechat-server/scutil"
	"time"
)

type ClientSetProfilepictureMessage struct {
	lib.ClientMessage
	NotifyUsers    []string // TODO this will be changed in a future version.. just a hack to get it working
	ProfilePicture []byte
}

// TODO this is just temporary to get profile pictures working.. will be removed in a future version
type ServerProfilepictureUpdateMessage struct {
	lib.ServerMessage
	Sender         string
	ProfilePicture []byte
}

type ServerStatusNotificationMessage struct {
	lib.ServerMessage
	User       string
	StatusText string
	LastSeen   int64
	Status     int // see STATUS_ constants at the beginning of this file.
}

// Status and Profile messages //////////////////////////
/////////////////////////////////////////////

type ClientSubscribeStatusMessage struct {
	lib.ClientMessage
	SubscribeTo []string
}

type ClientSetStatusMessage struct {
	lib.ClientMessage
	StatusText string
}

type ClientUpdateProfileMessage struct {
    lib.ClientMessage
	Alias				string
	Email				string
	Website				string
	Title				string
	Organization		string	
}

type ClientSetProfileSettingsMessage struct {
	lib.ClientMessage
	RepeatPingDisabled bool
}


/*
// Profile messages
// TODO: Should we limit who can access profiledata to protect against scraping etc?

type ClientSetProfilepictureMessage struct {
	ClientMessage
	Profilepicture []byte
}

type ClientGetProfileMessage struct {
	ClientMessage
	Name string
	Profilepicture bool // Whether the server should send the profilepicture too
}
*/

func handleClientSetProfilepictureMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")
	var spm lib.ClientSetProfilepictureMessage
	err := lib.MpUnmarshal(message, &spm)
	if err != nil {
		panic("handleClientSetProfilepictureMessage: Failed to unmarshal message.")
	}

	conn.User.Profilepicture = spm.ProfilePicture
	conn.User.GenerateUpdateToken()
	db.Save(conn.User)
	scutil.SLog("Stored profilepicture with size %d for user %s", len(message_data), conn.User.Username)
	//ServerProfilepictureUpdateMessage
	var spum lib.ServerProfilepictureUpdateMessage
	spum.MessageType = "ServerProfilepictureUpdateMessage"
	spum.Sender = conn.User.Username
	spum.ProfilePicture = spm.ProfilePicture
	spum_json := lib.MpMarshal(spum)
	spum_frame := lib.CreateFrame(spum_json)

	for _, username := range spm.NotifyUsers {
		scutil.SLog("Notifying user: %s", username)
		var user lib.User
		err = db.Where("username = ?", username).First(&user).Error
		if err != nil {
			scutil.SLog("User not found: %s", username)
			continue
		}
		//ic := lib.InterconnectContainer{User: user, Frame: spum_frame}
		ic := lib.InterconnectMessageContainer{User: user, Frame: spum_frame}
		conn.ImmediateMessageHandler <- &ic
	}
}

func handleClientSetProfileSettingsMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")
	var msg ClientSetProfileSettingsMessage
	lib.MpUnmarshal(message, &msg)

	conn.User.RepeatPingDisabled = msg.RepeatPingDisabled
	db.Save(&conn.User)
}

func handleClientSetStatusMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")
	var msg ClientSetStatusMessage
	lib.MpUnmarshal(message, &msg)

	conn.User.StatusText = msg.StatusText
	if !conn.Invisible {
		conn.User.LastSeen = time.Now().UTC()
	}
	db.Save(&conn.User)

	conn.StatusNotificationHandler <- &conn.User
	//notifyStatuses(conn.User)
}

func handleClientSubscribeStatusMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")
	var msg ClientSubscribeStatusMessage
	lib.MpUnmarshal(message, &msg)

	conn.SubscribedTo = msg.SubscribeTo
	sendStatuses(conn, db)
}

func sendStatuses(conn *lib.Connection, db gorm.DB) {
	for _, username := range conn.SubscribedTo {
		user := getUser(db, username)
		if user == nil {
			// TODO does this need to be handled somehow?
			continue
		}

		// Get the blockedUsers of the user that we want to send the status of
		var blockedUsers []lib.User
		db.Model(&user).Related(&blockedUsers, "BlockedUsers")

		if conn.User.IsIn(blockedUsers) {
			scutil.SLog("User %s is blocked by %s - not sending status.", conn.User.Username, user.Username)
			continue
		}

		statusMessage := lib.ServerStatusNotificationMessage{
			User:       user.Username,
			StatusText: user.StatusText,
			LastSeen:   user.LastSeen.Unix(),
			Status:     user.Status,
		}
		setMessageType(&statusMessage)
		sendAnswer(conn, statusMessage, "ServerStatusNotificationMessage")
	}
}

func handleClientUpdateProfileMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
    scutil.DebugLog("handleClientUpdateProfileMessage")
	var msg ClientUpdateProfileMessage
	lib.MpUnmarshal(message, &msg)

    conn.User.Alias = msg.Alias
    conn.User.Email = msg.Email
    conn.User.Website = msg.Website
    conn.User.Title = msg.Title
    conn.User.Organization = msg.Organization
    
	db.Save(&conn.User)

	conn.ProfileNotificationHandler <- &conn.User

}

