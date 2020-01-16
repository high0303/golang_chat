//
//  encryption_messages.go
//  Sechat Server
//
//  Created by AX on 2017-02-24.
//  Copyright 2017 JW Technologies. All rights reserved.
//

package main

import (
	"github.com/jinzhu/gorm"
	"sechat-server/lib"
	"sechat-server/scutil"
)

type ClientPublishPreKeyBundlesMessage struct {
	lib.ClientMessage
	Username   				string
	FirstSignedPreKeyID		int
	PreKeyBundles			[]string
}

type ClientRequestPreKeyBundle struct {
	lib.ClientMessage
	Username   				string
}

type ServerPreKeyBundleMessage struct {
	lib.ServerMessage
	Successful 				bool
	Error      			    string
	Username   				string
	DeviceID				int
	SignedPreKeyID			int
	PreKeyBundle			string
}

func handleClientPublishPreKeyBundlesMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	//scutil.DebugLog("Begin handleClientPublishPreKeyBundlesMessage")
	var spm lib.ClientPublishPreKeyBundlesMessage
	err := lib.MpUnmarshal(message, &spm)
	if err != nil {
		scutil.DebugLog("handleClientPublishPreKeyBundlesMessage: Failed to unmarshal message.")
		panic("handleClientPublishPreKeyBundlesMessage: Failed to unmarshal message.")
	}

	//scutil.DebugLog("Username: %s", spm.Username)
	//scutil.DebugLog("FirstSignedPreKeyID: %d", spm.FirstSignedPreKeyID)
	var  FirstSignedPreKeyID int64 = int64(spm.FirstSignedPreKeyID)

	for _, PreKeyBundle := range spm.PreKeyBundles {
		//scutil.DebugLog("PreKeyBundle: %s", PreKeyBundle)
		//var  theDeviceID int64 = int64(spm.DeviceID)
		//theUserID = ( int64)spm.DeviceID
		var newPreKeyBundle = lib.PreKeyBundle{DeviceID: conn.Device.ID, PreKeyBundleID: FirstSignedPreKeyID, Bundle: PreKeyBundle, UserId: conn.User.ID}
		FirstSignedPreKeyID += 2

		err = db.Save(&newPreKeyBundle).Error
		if err != nil {
			scutil.DebugLog("Critical Error: Saving newPreKeyBundle to database failed!")
			return
		}
		//var user lib.User
	}	
}
 
func handleClientRequestPreKeyBundle(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	//scutil.DebugLog("Begin handleClientRequestPreKeyBundle")
	var spm lib.ClientRequestPreKeyBundle
	err := lib.MpUnmarshal(message, &spm)
	if err != nil {
		scutil.DebugLog("handleClientRequestPreKeyBundle: Failed to unmarshal message.")
		panic("handleClientRequestPreKeyBundle: Failed to unmarshal message.")
	} else {
		    scutil.DebugLog("handleClientRequestPreKeyBundle: Look for user %s", spm.Username)
		user := getUser(db, spm.Username)
		var preKeyBundle lib.PreKeyBundle
		    //scutil.DebugLog("handleClientRequestPreKeyBundle: user %s has id %d", spm.Username, user.ID)
		dbErr := db.Where("user_id = ?", user.ID).First(&preKeyBundle).Error	
		if (dbErr == nil) {
		    //scutil.DebugLog("handleClientRequestPreKeyBundle: Find prekeybundle for %s", spm.Username)
			var count int64
			var tempPreKeyBundle lib.PreKeyBundle
			db.Model(&tempPreKeyBundle).Where("user_id = ?", user.ID).Count(&count)
			//scutil.DebugLog("Prekeybundles for user id = %d, Count is %d", user.ID, count);
			
			preKeyBundleMessage := lib.ServerPreKeyBundleMessage{ServerMessage: lib.ServerMessage{MessageType: "ServerPreKeyBundleMessage"},
																 Successful: true, Error: "", Username: user.Username,
																 DeviceID: int(preKeyBundle.DeviceID), SignedPreKeyID: int(preKeyBundle.PreKeyBundleID),
																 PreKeyBundle: preKeyBundle.Bundle}
			sendAnswer(conn, preKeyBundleMessage, "ServerPreKeyBundleMessage")
			// remove the pre key bundle from the db
			db.Delete(&preKeyBundle)
			db.Model(&tempPreKeyBundle).Where("user_id = ?", user.ID).Count(&count)
			//scutil.DebugLog("Prekeybundles for user id = %d, Count is %d", user.ID, count)
		} else {
			//scutil.DebugLog("Prekeybundle for user id = %d not found", user.ID)
			preKeyBundleMessage := lib.ServerPreKeyBundleMessage{ServerMessage: lib.ServerMessage{MessageType: "ServerPreKeyBundleMessage"},
																 Successful: false, Error: "No PrekeyBundles found"}
			
			sendAnswer(conn, preKeyBundleMessage, "ServerPreKeyBundleMessage")		
		}
	}


}
