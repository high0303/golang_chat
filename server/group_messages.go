//
//  group_messages.go
//  Sechat Server
//
//  Created by AX on 2016-02-02.
//  Copyright 2016 JW Technologies. All rights reserved.
//

package main

import (
	"fmt"
	//"log"
	"github.com/jinzhu/gorm"
	"sechat-server/lib"
	"sechat-server/scutil"
)

type ClientEndToEndGroupChatMessage struct {
	lib.ClientMessage
	Identifier          string
	Encrypted           []byte
	Signature           string
	Highlight           []string
	Notification        int
	ImmediateOnly       bool
	Priority            int
	MessageToken        string
	RepeatNotifications bool
}

type ClientCreateGroupchatMessage struct {
	lib.ClientMessage
	Name  string
	Users []string
}

type ServerCreateGroupchatMessage struct {
	lib.ServerMessage
	Successful bool
	Error      string
	Identifier string
}

type ServerGroupchatInvitationMessage struct {
	lib.ServerMessage
	Name       string
	Creator    string
	Users      []string
	Identifier string
}

type ServerEndToEndGroupChatMessage struct {
	lib.ServerMessage
	Identifier string
	Encrypted  []byte
	Signature  string
	// TODO can we hide the sender somehow?
	//      maybe use keyIDs? difficult i guess.
	Sender   string
	Priority int
}

// Groupchat messages
// TODO: Reject groupchat invitation?
type ClientCreateGroupChatMessage struct {
	lib.ClientMessage
	Name  string
	Users []string
}

type ServerCreateGroupChatResponseMessage struct {
	lib.ServerMessage
	Successful bool
	Error      string
	Name       string
	Identifier string
}

type ServerGroupChatInvitationMessage struct {
	lib.ServerMessage
	Name       string
	Creator    string
	Users      []string
	Identifier string
}

type ClientGroupChatUpdateMessage struct {
	lib.ClientMessage
	Users      []string
	Identifier string
}

type ClientGroupChatLeaveMessage struct {
	lib.ClientMessage
	Identifier 	string			// Unique identifier for group
	Name  		string			// Name of Chatgroup to remove
}
type ServerGroupChatUpdateMessage struct {
	lib.ServerMessage
	Creator    string
	Name       string
	Users      []string
	Identifier string
}

type ClientGroupChatBlockMessage struct {
	lib.ClientMessage
	Block      bool
	Identifier string
}

func handleClientCreateGroupChatMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	//scutil.DebugLog("handleClientCreateGroupChatMessage")
	
	var msg lib.ClientCreateGroupchatMessage
	var groupChat lib.Groupchat
	var groupChatIdentifier string
	var groupChatToUsers []lib.GroupchatToUser
	var bOkay bool = true
	
	lib.MpUnmarshal(message, &msg)

	groupChatName := msg.Name

	var groupChatCheck lib.Groupchat
	//err := db.Where("Name = ?", msg.Name).First(&groupChatCheck).Error
	err := db.Where("Name = ? and creator_id = ?", msg.Name, conn.User.ID).First(&groupChatCheck).Error
	    
    if (err == nil) {
        //scutil.DebugLog("Could find Group Chat %s with ID %s", groupChatCheck.Name, groupChatCheck.Identifier)
		response := ServerCreateGroupChatResponseMessage{
			Successful: false,
			Error:      fmt.Sprintf("Group %s already exists for this user", groupChatCheck.Name)}
		setMessageType(&response)
		sendAnswer(conn, response, "ServerCreateGroupChatResponseMessage")
		bOkay = false
        scutil.DebugLog("Group %s already exists for this user", groupChatCheck.Name)
    } else {
        scutil.DebugLog("Could not find Group Chat %s", msg.Name)
	}
	
	if (bOkay) {
        //scutil.DebugLog("Assemble group chat user list")
		// + 1 because we also need to store ourselves.
		groupChatToUsers = make([]lib.GroupchatToUser, len(msg.Users)+1)
		groupChatToUsers[0] = lib.GroupchatToUser{User: conn.User}
	
		for index, username := range msg.Users {
			user := getUser(db, username)
			if user == nil {
				response := ServerCreateGroupChatResponseMessage{
					Successful: false,
					Error:      fmt.Sprintf("Could not find user %s.", username)}
				setMessageType(&response)
				sendAnswer(conn, response, "ServerCreateGroupChatResponseMessage")
				bOkay = false
				break
			}
	
			groupChatToUsers[index+1] = lib.GroupchatToUser{User: *user}
		}
	}

	if (bOkay) {
        //scutil.DebugLog("Create group chat")
		groupChatIdentifier = lib.GetRandomHash()
	
		groupChat = lib.Groupchat{
			Identifier: groupChatIdentifier,
			Creator:    conn.User,
			Name:       msg.Name,
			Users:      groupChatToUsers,
		}
	
		err := db.Save(&groupChat).Error
		if err != nil {
			scutil.SLog("Saving to database failed.")
			response := ServerCreateGroupChatResponseMessage{
				Successful: false,
				Error:      "Server error."}
			setMessageType(&response)
			sendAnswer(conn, response, "ServerCreateGroupChatResponseMessage:false")
			bOkay = false
		}
	}

	if (bOkay) {
        //scutil.DebugLog("Send group chat invites")
		response := ServerCreateGroupChatResponseMessage{
			Successful: true,
			Name:       groupChat.Name,
			Identifier: groupChatIdentifier,
		}
		setMessageType(&response)
		sendAnswer(conn, response, "ServerCreateGroupChatResponseMessage: true")
	
		for _, groupChatToUser := range groupChatToUsers {
			if groupChatToUser.UserId == conn.User.ID {
				continue
			}
	
			invite := ServerGroupChatInvitationMessage{
				Name:       groupChatName,
				Creator:    conn.User.Username,
				Identifier: groupChat.Identifier,
				Users:      msg.Users,
			}
			setMessageType(&invite)
	
			sendToUser(conn, &groupChatToUser.User, invite, "ServerGroupChatInvitationMessage")
		}
	}

}

/*
func handleClientCreateGroupchatMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")

	var create_groupchat_message lib.ClientCreateGroupchatMessage
	
	err := lib.MpUnmarshal(message, &create_groupchat_message)

	if err != nil {
		panic("handleClientCreateGroupchatMessage: Failed to unmarshal message")
	}

	groupchatName := create_groupchat_message.Name

	groupchatToUsers := make([]lib.GroupchatToUser, len(create_groupchat_message.Users)+1)
	groupchatToUsers[0] = lib.GroupchatToUser{User: conn.User}
	for index, username := range create_groupchat_message.Users {
		var user lib.User
		err = db.Where("username = ?", username).First(&user).Error
		if err != nil {
			// TODO add username into error
			create_groupchat_response := lib.ServerCreateGroupchatMessage{ServerMessage: lib.ServerMessage{MessageType: "ServerCreateGroupchatMessage"}, Successful: false, Error: "User not found."}
			sendAnswer(conn, create_groupchat_response, "ServerCreateGroupchatMessage:false")
			return
		}

		groupchatToUsers[index+1] = lib.GroupchatToUser{User: user}
	}

	groupchatIdentifier := lib.GetRandomHash()

	var groupchat = lib.Groupchat{Identifier: groupchatIdentifier, Creator: conn.User, Name: groupchatName, Users: groupchatToUsers}

	err = db.Save(&groupchat).Error
	if err != nil {
		scutil.SLog("Critical Error: Saving to database failed!")
		create_groupchat_response := lib.ServerCreateGroupchatMessage{ServerMessage: lib.ServerMessage{MessageType: "ServerCreateGroupchatMessage"}, Successful: false, Error: "Server-side error."}
		sendAnswer(conn, create_groupchat_response, "ServerCreateGroupchatMessage: false")
		return
	}

	scutil.SLog("Groupchat %s created", groupchat.Name)
	create_groupchat_response := lib.ServerCreateGroupchatMessage{ServerMessage: lib.ServerMessage{MessageType: "ServerCreateGroupchatMessage"}, Successful: true, Identifier: groupchatIdentifier}
	sendAnswer(conn, create_groupchat_response, "ServerCreateGroupchatMessage: true")

	for _, groupchatToUser := range groupchatToUsers {
		if groupchatToUser.UserId == conn.User.ID {
			continue
		}
		groupchat_invite_json := lib.MpMarshal(lib.ServerGroupchatInvitationMessage{ServerMessage: lib.ServerMessage{MessageType: "ServerGroupchatInvitationMessage"}, Name: groupchatName, Creator: conn.User.Username, Identifier: groupchatIdentifier, Users: create_groupchat_message.Users})
		log.Print(string(groupchat_invite_json[:]))
		groupchat_invite := lib.CreateFrame(groupchat_invite_json)
		ic := lib.InterconnectContainer{User: groupchatToUser.User, Frame: groupchat_invite}
		conn.LocalMessageHandler <- &ic
	}
}
*/

func handleClientGroupChatLeaveMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.DebugLog("Begin handleClientGroupChatLeaveMessage")
	var msg ClientGroupChatLeaveMessage
	lib.MpUnmarshal(message, &msg)

	scutil.DebugLog("Chat Name:%s", msg.Name)
    
	var groupchat lib.Groupchat
	err := db.Where("identifier = ?", msg.Identifier).First(&groupchat).Error
    
    if (err != nil) {
        scutil.DebugLog("Could not find Group Chat %s with identifier %s", msg.Name, msg.Identifier)
    } else {
        var groupchatToUsers []lib.GroupchatToUser
        var groupchatToUsers2 lib.GroupchatToUser
        var nGroupChatCount int64
        var nIndex int64 = 0
        var creatorUserName string
 		db.Model(&groupchatToUsers2).Where("groupchat_id = ?", groupchat.Id).Count(&nGroupChatCount)
        db.Model(&groupchat).Related(&groupchatToUsers)

        if (nGroupChatCount > 0) {
           scutil.DebugLog("User group count = %d", nGroupChatCount)
           groupUserList := make([]string, nGroupChatCount-1)
           
           // assemble update list of users still in group
           for _, groupchatToUser := range groupchatToUsers {
                if (groupchatToUser.UserId == conn.User.ID) {
                    scutil.DebugLog("Delete User %d", groupchatToUser.UserId)
                    db.Delete(&groupchatToUser)
                } else {
                    user := getUserFromID(db, groupchatToUser.UserId)
                    scutil.DebugLog("New User List %s", user.Username)
                    groupUserList[nIndex] = user.Username
                    nIndex++
                    
                    if groupchat.CreatorId != user.ID {
                        creatorUserName = user.Username
                    }
                    
                }
            }
 
            var groupchatToUsersUpdated []lib.GroupchatToUser
            var nGroupChatCountUpdated int64
            db.Model(&groupchatToUsers2).Where("groupchat_id = ?", groupchat.Id).Count(&nGroupChatCountUpdated)
            db.Model(&groupchat).Related(&groupchatToUsersUpdated)
            scutil.DebugLog("Updated User group count = %d", nGroupChatCountUpdated)
        
            if (nGroupChatCountUpdated > 0) {
                // Still users in the group so notify all of them of the updated
                // list of users in the group
                for _, newGroupChatToUser := range groupchatToUsersUpdated {
                    updatedUser := getUserFromID(db, newGroupChatToUser.UserId)
                    scutil.DebugLog("Send update to %d, %s", newGroupChatToUser.UserId, updatedUser.Username)
                    // Update is sent to everyone, even ourselves!
                    update := ServerGroupChatUpdateMessage{
                    Identifier: msg.Identifier,
                    Users:      groupUserList,
                    Name:       groupchat.Name,
                    Creator:    creatorUserName,
                    }
                    setMessageType(&update)
                    sendToUser(conn, updatedUser, update, "ServerGroupChatUpdateMessage")
                   
                 }
            } else {
                // Nobody left in the group so delete
                scutil.DebugLog("Nobody left in Group %s, delete", groupchat.Name)
                db.Delete(&groupchat)
            }
        }

        
    }
}

func handleClientGroupChatUpdateMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.DebugLog("Begin handleClientGroupChatUpdateMessage")
	var msg ClientGroupChatUpdateMessage
	lib.MpUnmarshal(message, &msg)

	var groupchat lib.Groupchat
    var allUsers []lib.User
	var groupchatToUsersOld []lib.GroupchatToUser
	var userIndex int = 0
	var bOkay bool = true
	
	db.Where("identifier = ?", msg.Identifier).First(&groupchat)

	if groupchat.CreatorId != conn.User.ID {
		panic("Update groupchat from wrong user!")
	}

	db.Model(&groupchat).Related(&groupchatToUsersOld)

	allUsers = make([]lib.User, len(msg.Users)+len(groupchatToUsersOld))
	
	for _, groupchatToUserOld := range groupchatToUsersOld {
		user := getUserFromID(db, groupchatToUserOld.UserId)
		allUsers[userIndex] = *user
		userIndex++
	    scutil.DebugLog("Older chat to user %d (%s) ", groupchatToUserOld.UserId, user.Username)
		db.Delete(&groupchatToUserOld)
	}

	newGroupChatToUsers := make([]lib.GroupchatToUser, len(msg.Users)+1)
	newGroupChatToUsers[0] = lib.GroupchatToUser{User: conn.User}

	scutil.DebugLog("newGroupChatToUser assembly msg.Users len = %d", len(msg.Users))
	for index, username := range msg.Users {
	   scutil.DebugLog("start loop username %s", username)
		user := getUser(db, username)
		if user == nil {
			response := ServerCreateGroupChatResponseMessage{
				Successful: false,
				Error:      fmt.Sprintf("Could not find user %s.", username)}
			setMessageType(&response)
			sendAnswer(conn, response, "ServerCreateGroupChatResponseMessage: false")
			bOkay = false
			break
		} else {
			newGroupChatToUser := lib.GroupchatToUser{User: *user}
			newGroupChatToUsers[index+1] = newGroupChatToUser
			scutil.DebugLog("newGroupChatToUser.User.ID = %d", newGroupChatToUser.User.ID)
			var bMatch bool = false
			for _, groupchatToUserOld := range groupchatToUsersOld {
				scutil.DebugLog("newGroupChatToUser.User.ID = %d and groupchatToUserOld.UserId = %d",
								newGroupChatToUser.User.ID, groupchatToUserOld.UserId)
				if (newGroupChatToUser.User.ID == groupchatToUserOld.UserId) {
					bMatch = true
					scutil.DebugLog("match")
					break
				}
			}
			if (!bMatch) {
		        scutil.DebugLog("add newGroupChatToUser.User.ID", newGroupChatToUser.User.ID)
				allUsers[userIndex] = *user
				userIndex++
			}
			
		}
		scutil.DebugLog("end loop")
	}
	
	scutil.DebugLog("userIndex = %d", userIndex)
	
	if (bOkay) {
		groupchat.Users = newGroupChatToUsers
		db.Save(&groupchat)
		
		for i := 0; i < userIndex; i++ {
			user := allUsers[i]
			scutil.DebugLog("i = %d (%s)", i, user.Username)
			update := ServerGroupChatUpdateMessage{
				Identifier: msg.Identifier,
				Users:      msg.Users,
				Name:       groupchat.Name,
				Creator:    conn.User.Username,
			}
			setMessageType(&update)
	
			sendToUser(conn, &user, update, "ServerGroupChatUpdateMessage")
		}		
		
	}
	
}

func handleClientGroupChatUpdateMessageBak(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.DebugLog("Begin handleClientGroupChatUpdateMessage")
	var msg ClientGroupChatUpdateMessage
	lib.MpUnmarshal(message, &msg)

	var groupchat lib.Groupchat
	db.Where("identifier = ?", msg.Identifier).First(&groupchat)

	if groupchat.CreatorId != conn.User.ID {
		panic("Update groupchat from wrong user!")
	}

	var groupchatToUsers []lib.GroupchatToUser
	db.Model(&groupchat).Related(&groupchatToUsers)

	for _, groupchatToUser := range groupchatToUsers {
		db.Delete(&groupchatToUser)
	}

	newGroupChatToUsers := make([]lib.GroupchatToUser, len(msg.Users)+1)
	newGroupChatToUsers[0] = lib.GroupchatToUser{User: conn.User}

	for index, username := range msg.Users {
		user := getUser(db, username)
		if user == nil {
			response := ServerCreateGroupChatResponseMessage{
				Successful: false,
				Error:      fmt.Sprintf("Could not find user %s.", username)}
			setMessageType(&response)
			sendAnswer(conn, response, "ServerCreateGroupChatResponseMessage: false")
			return
		}

		newGroupChatToUsers[index+1] = lib.GroupchatToUser{User: *user}
	}

	groupchat.Users = newGroupChatToUsers
	db.Save(&groupchat)
	for _, newGroupChatToUser := range newGroupChatToUsers {
		// Update is sent to everyone, even ourselves!
		update := ServerGroupChatUpdateMessage{
			Identifier: msg.Identifier,
			Users:      msg.Users,
			Name:       groupchat.Name,
			Creator:    conn.User.Username,
		}
		setMessageType(&update)

		sendToUser(conn, &newGroupChatToUser.User, update, "ServerGroupChatUpdateMessage")
	}
}

func handleClientGroupChatBlockMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")
	var msg ClientGroupChatBlockMessage
	lib.MpUnmarshal(message, &msg)

	var groupchat lib.Groupchat
	db.Where("identifier = ?", msg.Identifier).First(&groupchat)

	var groupchatToUsers []lib.GroupchatToUser
	db.Model(&groupchat).Related(&groupchatToUsers)

	isInGroup := false
	for _, groupchatToUser := range groupchatToUsers {
		if groupchatToUser.UserId == conn.User.ID {
			isInGroup = true
		}
	}
	if !isInGroup {
		scutil.SLog("User %s not in group!", conn.User.Username)
		// TODO send error to client
		panic("handleGroupChatBlockMessage: Message to groupchat of which user is not a member.")
	}

	if msg.Block {
		scutil.SLog("Disable notifications")
		db.Model(&groupchat).Association("NotificationsDisabled").Append(conn.User)
	} else {
		scutil.SLog("Enable notifications")
		db.Model(&groupchat).Association("NotificationsDisabled").Delete(conn.User)
	}
}

func handleClientGroupChatBlockHighlightsMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")
	var msg ClientGroupChatBlockMessage
	lib.MpUnmarshal(message, &msg)

	var groupchat lib.Groupchat
	db.Where("identifier = ?", msg.Identifier).First(&groupchat)

	var groupchatToUsers []lib.GroupchatToUser
	db.Model(&groupchat).Related(&groupchatToUsers)

	isInGroup := false
	for _, groupchatToUser := range groupchatToUsers {
		if groupchatToUser.UserId == conn.User.ID {
			isInGroup = true
		}
	}
	if !isInGroup {
		scutil.SLog("User %s not in group!", conn.User.Username)
		// TODO send error to client
		panic("handleGroupChatBlockMessage: Message to groupchat of which user is not a member.")
	}

	if msg.Block {
		scutil.SLog("Disable notifications")
		db.Model(&groupchat).Association("HighlightNotificationsDisabled").Append(conn.User)
	} else {
		scutil.SLog("Enable notifications")
		db.Model(&groupchat).Association("HighlightNotificationsDisabled").Delete(conn.User)
	}
}

func handleClientEndToEndGroupChatMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")
	var msg ClientEndToEndGroupChatMessage
	lib.MpUnmarshal(message, &msg)

	var groupchat lib.Groupchat
	db.Where("identifier = ?", msg.Identifier).First(&groupchat)

	var groupchatToUsers []lib.GroupchatToUser
	db.Model(&groupchat).Related(&groupchatToUsers)

	var notificationsDisabled []lib.User
	db.Model(&groupchat).Related(&notificationsDisabled, "NotificationsDisabled")
	for _, notificationDisabledUser := range notificationsDisabled {
		scutil.SLog("User: %d %s", notificationDisabledUser.ID, notificationDisabledUser.Username)
	}

	var highlightNotificationsDisabled []lib.User
	db.Model(&groupchat).Related(&highlightNotificationsDisabled, "HighlightNotificationsDisabled")

	isInGroup := false

	for _, groupchatToUser := range groupchatToUsers {
		scutil.SLog("Groupchat has user %d %s (Sent by %d)", groupchatToUser.UserId, groupchatToUser.User.Username, conn.User.ID)
		if groupchatToUser.UserId == conn.User.ID {
			isInGroup = true
		}
	}

	if !isInGroup {
		// TODO send error to client...
		panic("handleGroupchatTextMessage: Message to groupchat of which user is not a member.")
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

	outMsg := ServerEndToEndGroupChatMessage{
		Identifier: groupchat.Identifier,
		Encrypted:  msg.Encrypted,
		Signature:  msg.Signature,
		Sender:     conn.User.Username,
		Priority:   10,
	}
	setMessageType(&outMsg)

    var messageID int64
 	if (!msg.ImmediateOnly) {
        // [1] Prepare stored message
        messageID = prepareStoredMessage(db, conn, outMsg, msg.MessageToken, "")
    }	
	
	for _, groupchatToUser := range groupchatToUsers {
		var user lib.User
		err := db.First(&user, groupchatToUser.UserId).Error
		if err != nil {
			scutil.SLog("Error: User not found: %s", user.Username)
			continue
		}
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
			scutil.SLog("Sending notification to %s", user.Username)
			go sendAPNS(db, user, msg.Notification)
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
