//
//  messagehandlers.go
//  Sechat Server
//
//  Copyright 2017 JW Technologies. All rights reserved.
//

package main

import (
	"github.com/jinzhu/gorm"
	"sechat-server/lib"
	"sechat-server/scutil"
	"time"
	//"strings"
)

type MessageHandlerFunction func(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte)

type MessageHandler struct {
	Function        MessageHandlerFunction
	MessageTypeName string
}

// TODO/FIXME: unauthenticatedMessageHandlers are needed!

var messageHandlers map[string]MessageHandlerFunction = map[string]MessageHandlerFunction{
	
	// Register Messages
	"ClientRegisterMessage":           					handleClientRegisterMessage,
	
	// Login Messages
	"ClientAppLoginMessage":							handleClientAppLoginMessage,
	"ClientCheckinMessage":       						handleClientCheckinMessage,
	"ClientCheckinResponseMessage":      				handleClientCheckinResponseMessage,
	"ClientPublishPreKeyBundlesMessage": 				handleClientPublishPreKeyBundlesMessage,
	"ClientRequestPreKeyBundleMessage":	   				handleClientRequestPreKeyBundle,
	"ClientAppLogoutMessage":							handleClientAppLogoutMessage,
	
	// Group Messages
	//"ClientCreateGroupchatMessage":    handleClientCreateGroupchatMessage,
	"ClientGroupChatUpdateMessage":    					handleClientGroupChatUpdateMessage,
	"ClientGroupChatLeaveMessage":     					handleClientGroupChatLeaveMessage,
	"ClientEndToEndGroupChatMessage":  					handleClientEndToEndGroupChatMessage,
	"ClientCreateGroupChatMessage":    					handleClientCreateGroupChatMessage,
	"ClientGroupChatBlockMessage":           			handleClientGroupChatBlockMessage,
	"ClientGroupChatBlockHighlightsMessage": 			handleClientGroupChatBlockHighlightsMessage,
	
	// Profile, Status Messages
	"ClientSetProfilepictureMessage":  					handleClientSetProfilepictureMessage,
	"ClientSetProfileSettingsMessage": 					handleClientSetProfileSettingsMessage,
	"ClientSetStatusMessage":          					handleClientSetStatusMessage,
	"ClientSubscribeStatusMessage":    					handleClientSubscribeStatusMessage,
	"ClientUpdateProfileMessage":      					handleClientUpdateProfileMessage,
	
	// Call Messages
	"ClientCallCreateMessage":    						handleClientCallCreateMessage,
	"ClientCallGetStatusMessage": 						handleClientCallGetStatusMessage,
	"ClientCallUpkeepMessage":    						handleClientCallUpkeepMessage,

	// Contact
	"ClientAddContactMessage":         					handleClientAddContactMessage,
    "ClientContactRequestMessage":                      handleClientContactRequestMessage,
	
	// Basic Messages
	"ClientEndToEndMessage":           					handleClientEndToEndMessage,
	"ClientAcknowledgeMessage":         				handleClientAcknowledgeMessage,
	"ClientEndToEndGenericMessage":						handleClientEndToEndGenericMessage,

	
	// Misc Messages
	"ClientRegisterPhonenumberMessage": 				handleClientRegisterPhonenumberMessage,
	"ClientVerifyPhonenumberMessage":   				handleClientVerifyPhonenumberMessage,
	"ClientRecallEndToEndMessage": 						handleClientRecallEndToEndMessage,
	"ClientSyncContactsMessage":             			handleClientSyncContactsMessage,
	"ClientBlockUserMessage": 							handleClientBlockUserMessage,

	// Channel Messages
	"ClientEndToEndChannelMessage":    					handleClientEndToEndChannelMessage,
	"ClientCreateChannelMessage": 						handleClientCreateChannelMessage,
	"ClientJoinChannelMessage":   						handleClientJoinChannelMessage,
	"ClientGetChannelsMessage":   						handleClientGetChannelsMessage,
}

func handleClientMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	defer func() {
		if r := recover(); r != nil {
			scutil.TopLog("handleClientMessage: Recovering from ", r)
		}
	}()
	// TODO: This is ineffective as hell on big messages, it's
	// unpacking them twice...
	var data map[string]interface{}
	if err := lib.MpUnmarshal(message, &data); err != nil {
		panic(err)
	}

	conn.LastActivity = time.Now().UTC()

		
	//scutil.TopLog("MessageId: %q", string(data["MessageId"].([]uint8)) )
	str := string(data["MessageType"].([]uint8))
	
	// Filter out the real busy messages for debugging
	if (str != "ClientEndToEndMessage" &&
		str != "ClientPingMessage" &&
		str != "ClientEndToEndGroupChatMessage" &&
		str != "ClientEndToEndChannelMessage") {
		scutil.TopLog("[%s] handleClientMessage: MessageType: %q", conn.User.Username, data["MessageType"])
	}
	/*if !ok {
		log.Printf("%T\n", data["MessageType"])
		panic("handleClientMessage: MessageType is not a string")
	}*/
	if val, ok := messageHandlers[str]; ok {
		val(conn, db, message, message_data)
	} else {
		// TODO: analyze unhandled messages
		//log.Print("handleClientMessage: Ignoring unknown message type %s", str)
		//log.Print(data)
	}
	//messageHandlers[str](conn, db, message)
}



