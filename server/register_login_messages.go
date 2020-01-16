//
//  register_login_messages.go
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
)

type ServerRegisterMessage struct {
	lib.ServerMessage
	Successful bool
	Error      string
}

type ClientAppLoginMessage struct {
	lib.ClientMessage
	Username   			string 
	DeviceToken			string // Unique identifier for the device
	DeviceName 			string // User string for the device
	Password   			string
	DeviceX      		string // device public key X
	DeviceY      		string // device public key Y
}

type ClientAppLogoutMessage struct {
	lib.ClientMessage
	Username   			string
	DeviceToken			string
}

type ClientCheckinMessage struct {
	lib.ClientMessage
	Username   			string
	DeviceToken 		string
}

type ServerCheckinChallengeMessage struct {
	lib.ServerMessage
	Successful bool
	Error      string
	Challenge  string
}

type ClientCheckinResponseMessage struct {
	lib.ClientMessage
	R           string // Signature part R
	S           string // Signature part S
	APNSToken   string
	Invisible   bool
	SubscribeTo []string
}

type ServerCheckinResultMessage struct {
	lib.ServerMessage
	Successful 			bool
	Error      			string
	NumKeyBundles		int  // number of Axolotl prekey bundles available
}

func handleClientRegisterMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	var registerMessage lib.ClientRegisterMessage
	lib.MpUnmarshal(message, &registerMessage)
	
	user := getUser(db, registerMessage.Username)
	
	// Username already taken?
	if (user == nil) {
		newUser := lib.User{Username: registerMessage.Username,
							Password: registerMessage.Password}
		newUser.GenerateUpdateToken()
		db.Save(&newUser)
		scutil.RegisterLoginLog("Registered user %s with ID = %d", newUser.Username, newUser.ID)
		
		registerAnswer := lib.ServerRegisterMessage{Successful: true, Error: ""}
		setMessageType(&registerAnswer)
		sendAnswer(conn, registerAnswer, "ServerRegisterMessage:true")
	} else {
		scutil.RegisterLoginLog("Username %s already taken.", registerMessage.Username)
		registerAnswer := lib.ServerRegisterMessage{Successful: false, Error: "Username already taken."}
		setMessageType(&registerAnswer)
		sendAnswer(conn, registerAnswer, "ServerRegisterMessage:false")		
	}
}

func handleClientAppLoginMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {

	var clim lib.ClientAppLoginMessage
    var bOkay bool = true
    
	lib.MpUnmarshal(message, &clim)

	user := getUser(db, clim.Username)

	// Check if user exists
	if user == nil {
		scutil.RegisterLoginLog("User not found.")
		sendAnswer(conn, lib.ServerCheckinChallengeMessage{ServerMessage: lib.ServerMessage{MessageType: "ServerCheckinChallengeMessage"}, Successful: false, Error: "User not found"},
				   "ServerCheckinChallengeMessage:false")
		bOkay = false
	}

	// check password
    if (bOkay) {
        if (clim.Password != user.Password) {
            scutil.RegisterLoginLog("Invalid Password")
            sendAnswer(conn, lib.ServerCheckinChallengeMessage{ServerMessage: lib.ServerMessage{MessageType: "ServerCheckinChallengeMessage"}, Successful: false, Error: "Invalid Password"},
                     "ServerCheckinChallengeMessage:false")
             scutil.RegisterLoginLog("Password incorrect")
            bOkay = false
        } else {
            scutil.RegisterLoginLog("Password checks out")
        }
    }
    
	// Update with device logged in from
	if (bOkay) {
		device := getUserDevice(db, user.ID, clim.DeviceToken)
		
		if (device == nil) {
 			newDevice := lib.Device{
				UserID: user.ID,
				DeviceToken: clim.DeviceToken,
				DeviceName:  clim.DeviceName,
				X: clim.DeviceX,
				Y: clim.DeviceY,
			}
			db.Save(&newDevice)
			scutil.RegisterLoginLog("Added Device %s(%s) for user %s", clim.DeviceToken, clim.DeviceName, user.Username)
			//sendCheckinChallengeMessage(conn, user, &newDevice)
			sendCheckinChallengeMessage(conn, user, &newDevice, lib.EConnectionStatus_Challenge_With_DeviceToken)
		} else {
			scutil.RegisterLoginLog("Retrieved Device %s(%s) for user %s", clim.DeviceToken, clim.DeviceName, user.Username)
			
			if (device.DeviceName != clim.DeviceName ||
				device.X != clim.DeviceX ||
				device.Y != clim.DeviceY) {
				scutil.RegisterLoginLog("Device Credentials were updated")
				device.DeviceName = clim.DeviceName
				device.X = clim.DeviceX
				device.Y = clim.DeviceY
				db.Save(&device)	
			}
			
			//sendCheckinChallengeMessage(conn, user, device)
			sendCheckinChallengeMessage(conn, user, device, lib.EConnectionStatus_Challenge_With_DeviceToken)
		}
		
	}
}

func sendCheckinChallengeMessage (conn *lib.Connection,
								  user *lib.User,
								  device *lib.Device,
								  newStatus lib.EConnectionStatus) {
        conn.User = *user
		conn.Status = newStatus
		if (device != nil) {
			conn.Device = *device
		}
        challenge := lib.GetRandomHash()
        conn.Challenge = challenge
        checkin_challenge_message := lib.ServerCheckinChallengeMessage{Successful: true, Error: "", Challenge: challenge}
        setMessageType(&checkin_challenge_message)
        //scutil.SLog("Type: %s", checkin_challenge_message.ServerMessage.MessageType)
        sendAnswer(conn, checkin_challenge_message,
                       "ServerCheckinChallengeMessage:true") 	
}

func handleClientCheckinMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	//scutil.SLog("Begin")
	scutil.DebugLog("handleClientCheckinMessage")

	var clim lib.ClientCheckinMessage
	lib.MpUnmarshal(message, &clim)

	user := getUser(db, clim.Username)
	
	if user == nil {
		scutil.RegisterLoginLog("User %s not found", clim.Username)
		sendAnswer(conn, lib.ServerCheckinChallengeMessage{ServerMessage: lib.ServerMessage{MessageType: "ServerCheckinChallengeMessage"},
																	Successful: false,
																	Error: "User not found"},
				   "ServerCheckinChallengeMessage:false")
	} else {
		device := getUserDevice(db, user.ID, clim.DeviceToken)
		if (device == nil) {
			scutil.RegisterLoginLog("For user %s, Device %s NOT found", user.Username, clim.DeviceToken)
			conn.DeviceToken = clim.DeviceToken
			//conn.Status = lib.EConnectionStatus_Challenge_No_DeviceToken
			//sendAnswer(conn, lib.ServerCheckinChallengeMessage{ServerMessage: lib.ServerMessage{MessageType: "ServerCheckinChallengeMessage"},
			//															Successful: false,
			//															Error: "Device not found"},
			//		   "ServerCheckinChallengeMessage:false")
			sendCheckinChallengeMessage(conn, user, device, lib.EConnectionStatus_Challenge_No_DeviceToken)
		} else {
			scutil.RegisterLoginLog("For user %s, Device %s found", user.Username, clim.DeviceToken)
			//conn.Status = lib.EConnectionStatus_Challenge_With_DeviceToken
			sendCheckinChallengeMessage(conn, user, device, lib.EConnectionStatus_Challenge_With_DeviceToken)
		}

	}

}

func handleClientAppLogoutMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {

	var clim lib.ClientAppLogoutMessage
	lib.MpUnmarshal(message, &clim)
    
	if !conn.Invisible && conn.Status == lib.EConnectionStatus_Authenticated { // conn.Authenticated {
		conn.User.Status = lib.STATUS_OFFLINE
		conn.User.LastSeen = time.Now().UTC()
		db.Save(&conn.User)
		conn.StatusNotificationHandler <- &conn.User
	}
    
}

func handleClientCheckinResponseMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	var clrm lib.ClientCheckinResponseMessage
	lib.MpUnmarshal(message, &clrm)
	
	var bAuthenticated bool = false

	if (conn.Status == lib.EConnectionStatus_Challenge_With_DeviceToken) {
		if (conn.Device.VerifySignature(conn.Challenge, clrm.R, clrm.S)) {
			bAuthenticated = true
		}
	} else if (conn.Status == lib.EConnectionStatus_Challenge_No_DeviceToken) {
		var devices []lib.Device
		db.Where("user_id = ?", conn.User.ID).Find(&devices)
		for _, device := range devices {
			if (device.VerifySignature(conn.Challenge, clrm.R, clrm.S)) {
				bAuthenticated = true
				scutil.RegisterLoginLog("Found right device and Verified Signature. Update device token")
				scutil.RegisterLoginLog("Replace token %s with %s", device.DeviceToken, conn.DeviceToken)
				device.DeviceToken = conn.DeviceToken
				conn.Device = device
				//db.Save(&device)
				break
			}
		}
		
	}
	
	if (bAuthenticated) {
		scutil.RegisterLoginLog("Verified Signature: Ok")
		scutil.RegisterLoginLog("Setup Connection for User %s with Device %s(%s)", conn.User.Username, conn.Device.DeviceToken, conn.Device.DeviceName)
				//conn.Authenticated = true
		conn.Status = lib.EConnectionStatus_Authenticated
		conn.Device.LastSeen = time.Now().UTC()
		//db.Save(&conn.Device)
		conn.RegisterConnection <- conn
		
		
		
		if conn.Device.APNSToken != clrm.APNSToken {
			scutil.RegisterLoginLog("save APNSToken %s", clrm.APNSToken)
			conn.Device.APNSToken = clrm.APNSToken
		}
	
		conn.Invisible = clrm.Invisible
		if !clrm.Invisible {
			conn.User.Status = lib.STATUS_ONLINE
			conn.User.LastSeen = time.Now().UTC()
		}
	
		//if (conn.Status == lib.EConnectionStatus_Authenticated) { //conn.Authenticated) {
			//scutil.DebugLog("handleClientLoginResponseMessage authenticate");
		var count int64
		var newPreKeyBundle lib.PreKeyBundle
		//db.Where("user_id = ?", conn.User.ID).Find(&newPreKeyBundle)
		db.Model(&newPreKeyBundle).Where("user_id = ? and device_id = ?", conn.User.ID, conn.Device.ID).Count(&count)
		//scutil.DebugLog("Prekeybundle for user id = %d is %d", conn.User.ID, count);
		
		// Notify the client of the number of prekey bundles
		loginResultMessage := lib.ServerCheckinResultMessage{Successful: true, Error: "", NumKeyBundles: int(count)}
		setMessageType(&loginResultMessage)
		sendAnswer(conn, loginResultMessage, "ServerCheckinResultMessage")
		//}
		
		db.Save(&conn.Device)
		db.Save(&conn.User)
		conn.StatusNotificationHandler <- &conn.User
		
	} else {
		scutil.RegisterLoginLog("Verified Signature: Failed!!")
		loginResultMessage := lib.ServerCheckinResultMessage{Successful: false, Error: "Checkin Failed!", NumKeyBundles: 0}
		setMessageType(&loginResultMessage)
		sendAnswer(conn, loginResultMessage, "ServerCheckinResultMessage")
		//panic("handleClientCheckinResponseMessage: Invalid signature")
	}
}
