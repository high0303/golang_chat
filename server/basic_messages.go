//
//  basic_messages.go
//  Sechat Server
//
//  Created by AX on 2017-02-26.
//  Copyright 2017 JW Technologies. All rights reserved.
//

package main

import (
	"strings"
	"github.com/jinzhu/gorm"
	"sechat-server/lib"
	"sechat-server/scutil"
	//"time"
)

type ClientAcknowledgeMessage struct {
	lib.ClientMessage
	MessageId int
}

func handleClientEndToEndMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")
	var msg lib.ClientEndToEndMessage
	lib.MpUnmarshal(message, &msg)
	var bOkay bool = true

	//scutil.EndToEndMessageLog("handleClientEndToEndMessage start")
	receiver := getUser(db, msg.Receiver)
	if receiver == nil {
		scutil.SLog("Receiver not found.")
		// TODO send error to client
		bOkay = false
	}

	if (bOkay) {
		// Get the blocked users of the receiver
		var blockedUsers []lib.User
		db.Model(&receiver).Related(&blockedUsers, "BlockedUsers")

		// Check if the sender is in the blocked users of the receiver
		if conn.User.IsIn(blockedUsers) {
			scutil.EndToEndMessageLog("User %s is blocked by %s - ignoring message.", conn.User.Username, receiver.Username)
			bOkay = false
		}
	}

	if (bOkay) {
		go sendAPNS(db, *receiver, msg.Notification)
	
		outMsg := lib.ServerEndToEndMessage{
			Encrypted: msg.Encrypted,
			Signature: msg.Signature,
			Sender:    conn.User.Username,
			Priority:  msg.Priority,
		}
		setMessageType(&outMsg)
	
		if (strings.Contains(msg.MessageToken, "Debug")) {
			if (!strings.Contains(msg.MessageToken, "sendInChatStatus")) {
				scutil.TopLog("[%s] handleClientMessage: MessageType: ClientEndToEndMessage", conn.User.Username)
				scutil.EndToEndMessageLog("[%s=>%s]ServerEndToEndMessage: %s", conn.User.Username, msg.Receiver, msg.MessageToken)
			}
		}
		
		if len(msg.MessageToken) > 0 {
			if (strings.Contains(msg.MessageToken, "Debug")) {
				if (!strings.Contains(msg.MessageToken, "sendInChatStatus")) {
					scutil.TopLog("[%s] handleClientMessage: MessageType: ClientEndToEndMessage", conn.User.Username)
					scutil.EndToEndMessageLog("[%s=>%s]ServerEndToEndMessage: %s", conn.User.Username, msg.Receiver, msg.MessageToken)
				}
			} else {
				scutil.DebugLog("Send e2e response %s", msg.MessageToken)
				e2eResponse := lib.ServerEndToEndResponseMessage{MessageId: msg.MessageToken}
				setMessageType(&e2eResponse)
				sendAnswer(conn, e2eResponse, "ServerEndToEndResponseMessage")
			}
		} else {
			scutil.EndToEndMessageLog("[%s] handleClientEndToEndMessage Unknown", conn.User.Username)
		}
	
		//sendToUserDetailed(conn, receiver, outMsg, msg.ImmediateOnly, "", msg.MessageId, true, msg.RepeatNotifications)
		if (msg.ImmediateOnly) {
			//scutil.EndToEndMessageLog("sendImmediateMessage")
			sendImmediateMessage(conn, receiver, outMsg)
		} else {
			// [1] Prepare stored message
			//messageID := prepareMessageStorageForRecipients(db, conn, receiver, outMsg, msg.MessageId, "", true, msg.RepeatNotifications)
			scutil.EndToEndMessageLog("prepareStoredMessage")
            messageID := prepareStoredMessage(db, conn, outMsg, msg.MessageToken, "")
			
			// [2] Prepare list of recipients
			// mirror message to user's other devices
			addMessageRecipient(db,
                             messageID,                 // db message id
                             &conn.User,                 // user
                             conn.Device.ID,            // don't sent to self
                             true,                      // notification
                             msg.RepeatNotifications)   // repeat notifications
			
            // Send to the other recipients
			addMessageRecipient(db,
                             messageID,                 // db message id
                             receiver,                  // user
                             0,                         // don't sent to self
                             true,                      // notification
                             msg.RepeatNotifications)   // repeat notifications
			
			// [3] Trigger channel to send the stored message to all the recipients
			sendStoredMessage(db, conn, messageID)
			
		}
	}
	//scutil.EndToEndMessageLog("handleClientEndToEndMessage finish")
}

func handleClientAcknowledgeMessageBak(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")
	var cam ClientAcknowledgeMessage
	lib.MpUnmarshal(message, &cam)

	var storedMessage lib.StoredMessage
	err := db.Where("id = ?", cam.MessageId).First(&storedMessage).Error
	if err != nil {
		scutil.SLog("No message found to invalidate")
		return
	} 

	db.Delete(&storedMessage)
}

func handleClientAcknowledgeMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")
	var cam ClientAcknowledgeMessage
	lib.MpUnmarshal(message, &cam)

	var recipient lib.Recipient
    scutil.EndToEndMessageLog("Received ACK for message %d for user %s from device %d (%s)", cam.MessageId, conn.User.Username, conn.Device.ID, conn.Device.DeviceName)
	err := db.Where("message_id = ? and user_id = ? and device_id = ?", cam.MessageId, conn.User.ID, conn.Device.ID).First(&recipient).Error
	if err == nil {
		scutil.EndToEndMessageLog("Acknowledge message %d for  recipient %d:%d", cam.MessageId, conn.User.ID, conn.Device.ID)
        db.Delete(&recipient)
       
       	var count int64
	    var tempRecipient lib.Recipient
		db.Model(&tempRecipient).Where("message_id = ?", cam.MessageId).Count(&count)
        if (count == 0) {
            var storedMessage lib.StoredMessage
            err := db.Where("id = ?", cam.MessageId).First(&storedMessage).Error
            if err == nil {
                db.Delete(&storedMessage)
                scutil.EndToEndMessageLog("Deleted stored message %d", cam.MessageId)
            } 
        } else {
            scutil.EndToEndMessageLog("Recipients Left = %d", count)
        }
 	} else {
		scutil.EndToEndMessageLog("No message found to Acknowledge")
        
    }
}



func handleClientBlockUserMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")
	var msg lib.ClientBlockUserMessage
	lib.MpUnmarshal(message, &msg)

	var userToBlock = getUser(db, msg.Username)
	// Check if the user exists...
	if userToBlock == nil {
		// TODO Send error
		scutil.SLog("Error: User %s not found.", msg.Username)
		return
	}
	// We can't block ourselves...
	if userToBlock.ID == conn.User.ID {
		// TODO Send error
		scutil.SLog("Error: Can't block ourselves")
		return
	}

	if msg.Block {
		scutil.SLog("Blocking user %s for user %s", msg.Username, conn.User.Username)
		db.Model(&conn.User).Association("BlockedUsers").Append(userToBlock)
	} else {
		scutil.SLog("Unblocking user %s for user %s", msg.Username, conn.User.Username)
		db.Model(&conn.User).Association("BlockedUsers").Delete(userToBlock)
	}

	// TODO send an answer
}

func handleClientRecallEndToEndMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")
	var msg lib.ClientRecallEndToEndMessage
	lib.MpUnmarshal(message, &msg)

	var response lib.ServerRecallEndToEndResponseMessage
	setMessageType(&response)
	response.MessageId = msg.MessageId
	response.Successful = false

	var storedMessage lib.StoredMessage
	err := db.Where("message_id = ?", msg.MessageId).First(&storedMessage).Error
	if err != nil {
		scutil.SLog("No message found to recall")
		// TODO add ownership to storedmessage to make sure only the original sender can recall messages
		/*} else if storedMessage.UserId != conn.User.Id {
		scutil.SLog("Message not owned by recaller!")*/
	} else {
		// TODO we probably need to check the return type here
		db.Delete(&storedMessage)
		response.Successful = true
	}
	sendAnswer(conn, response, "ServerRecallEndToEndResponseMessage")
}

func handleClientEndToEndGenericMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.DebugLog("handleClientEndToEndGenericMessage")
	var msg lib.ClientEndToEndGenericMessage
	lib.MpUnmarshal(message, &msg)
	var bOkay bool = true

	//scutil.EndToEndMessageLog("handleClientEndToEndMessage start")
	receiver := getUser(db, msg.Receiver)
	if receiver == nil {
		scutil.DebugLog("Receiver not found.")
		// TODO send error to client
		bOkay = false
	}

	if (bOkay) {
		// Get the blocked users of the receiver
		var blockedUsers []lib.User
		db.Model(&receiver).Related(&blockedUsers, "BlockedUsers")

		// Check if the sender is in the blocked users of the receiver
		if conn.User.IsIn(blockedUsers) {
			scutil.EndToEndMessageLog("User %s is blocked by %s - ignoring message.", conn.User.Username, receiver.Username)
			bOkay = false
		}
	}

	if (bOkay) {
		//go sendAPNS(db, *receiver, msg.Notification)
	
		outMsg := lib.ServerEndToEndGenericMessage{
			Data:     msg.Data,
			EncryptionType: msg.EncryptionType,
			Sender:    conn.User.Username,
			Priority:  msg.Priority,
		}
		setMessageType(&outMsg)
	
		scutil.DebugLog("[%s=>%s]ServerEndToEndGenericMessage: %s", conn.User.Username, msg.Receiver, msg.MessageToken)

		scutil.DebugLog("Send e2e response %s", msg.MessageToken)
		e2eResponse := lib.ServerEndToEndResponseMessage{MessageId: msg.MessageToken}
		setMessageType(&e2eResponse)
		sendAnswer(conn, e2eResponse, "ServerEndToEndResponseMessage")
	
		//sendToUserDetailed(conn, receiver, outMsg, msg.ImmediateOnly, "", msg.MessageId, true, msg.RepeatNotifications)
		if (msg.ImmediateOnly) {
			//scutil.EndToEndMessageLog("sendImmediateMessage")
			sendImmediateMessage(conn, receiver, outMsg)
		} else {
			// [1] Prepare stored message
			//messageID := prepareMessageStorageForRecipients(db, conn, receiver, outMsg, msg.MessageId, "", true, msg.RepeatNotifications)
			scutil.DebugLog("ServerEndToEndGenericMessage prepareStoredMessage")
            messageID := prepareStoredMessage(db, conn, outMsg, msg.MessageToken, "")
			
			// [2] Prepare list of recipients
			// mirror message to user's other devices
			addMessageRecipient(db,
                             messageID,                 // db message id
                             &conn.User,                 // user
                             conn.Device.ID,            // don't sent to self
                             true,                      // notification
                             msg.RepeatNotifications)   // repeat notifications
			
            // Send to the other recipients
			addMessageRecipient(db,
                             messageID,                 // db message id
                             receiver,                  // user
                             0,                         // don't sent to self
                             true,                      // notification
                             msg.RepeatNotifications)   // repeat notifications
			
			// [3] Trigger channel to send the stored message to all the recipients
			sendStoredMessage(db, conn, messageID)
			
		}
	}
	//scutil.EndToEndMessageLog("handleClientEndToEndMessage finish")
}

/*
func handleClientEndToEndMessageBak(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")
	var msg lib.ClientEndToEndMessage
	lib.MpUnmarshal(message, &msg)

	receiver := getUser(db, msg.Receiver)
	if receiver == nil {
		scutil.SLog("Receiver not found.")
		// TODO send error to client
		return
	}

	// Get the blocked users of the receiver
	var blockedUsers []lib.User
	db.Model(&receiver).Related(&blockedUsers, "BlockedUsers")

	// Check if the sender is in the blocked users of the receiver
	if conn.User.IsIn(blockedUsers) {
		scutil.SLog("User %s is blocked by %s - ignoring message.", conn.User.Username, receiver.Username)
		return
	}

	go sendAPNS(db, *receiver, msg.Notification)

	outMsg := lib.ServerEndToEndMessage{
		Encrypted: msg.Encrypted,
		Signature: msg.Signature,
		Sender:    conn.User.Username,
		Priority:  msg.Priority,
	}
	setMessageType(&outMsg)

	if (strings.Contains(msg.MessageId, "Debug")) {
		if (!strings.Contains(msg.MessageId, "sendInChatStatus")) {
			scutil.TopLog("[%s] handleClientMessage: MessageType: ClientEndToEndMessage", conn.User.Username)
			scutil.DebugLog("[%s=>%s]ServerEndToEndMessage: %s", conn.User.Username, msg.Receiver, msg.MessageId)
		}
	}
	
	if len(msg.MessageId) > 0 {
		if (strings.Contains(msg.MessageId, "Debug")) {
			if (!strings.Contains(msg.MessageId, "sendInChatStatus")) {
				scutil.TopLog("[%s] handleClientMessage: MessageType: ClientEndToEndMessage", conn.User.Username)
				scutil.DebugLog("[%s=>%s]ServerEndToEndMessage: %s", conn.User.Username, msg.Receiver, msg.MessageId)
			}
		} else {
			scutil.DebugLog("Send e2e response %s", msg.MessageId)
			e2eResponse := lib.ServerEndToEndResponseMessage{MessageId: msg.MessageId}
			setMessageType(&e2eResponse)
			sendAnswer(conn, e2eResponse, "ServerEndToEndResponseMessage")
		}
	} else {
		scutil.DebugLog("[%s] handleClientEndToEndMessage Unknown", conn.User.Username)
	}

	sendToUserDetailed(conn, receiver, outMsg, msg.ImmediateOnly, "", msg.MessageId, true, msg.RepeatNotifications)
}
*/