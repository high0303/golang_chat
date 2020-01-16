//
//  server_common.go
//  Sechat Server
//
//  Copyright 2017 JW Technologies. All rights reserved.
//

package main

import (
	"fmt"
	"runtime"
	"log"
	"github.com/jinzhu/gorm"
	"sechat-server/lib"
    "reflect"
    "sechat-server/scutil"
    "github.com/anachronistic/apns"
    twilio "github.com/carlosdp/twiliogo"
	"time"
)

func setMessageType(a interface{}) {
	newServerMessage := lib.ServerMessage{MessageType: reflect.ValueOf(a).Elem().Type().Name()}
	serverMessage := reflect.ValueOf(a).Elem().FieldByName("ServerMessage")
	serverMessage.Set(reflect.ValueOf(newServerMessage))
}

func getUser(db gorm.DB, username string) *lib.User {
	var user lib.User
	err := db.Where("username = ?", username).First(&user).Error
	if err != nil {
		scutil.SLog("Couldn't find a user named '%s'", username)
		return nil
	}
	return &user
}

func getUserFromID(db gorm.DB, userID int64) *lib.User {
	var user lib.User
	err := db.Where("ID = ?", userID).First(&user).Error
	if err != nil {
		scutil.SLog("Couldn't find a user ID '%d'", userID)
		return nil
	}
	return &user
}

func getUserDevice(db gorm.DB, userID int64, deviceToken string) *lib.Device {
    var device lib.Device
    err := db.Where("user_id = ? and device_token = ?", userID, deviceToken).First(&device).Error
    if err != nil {
        return nil
    }
    return &device
}

func sendCustomAPNSWithSound(db gorm.DB, device lib.Device, message string, sound string) {

	if len(device.APNSToken) == 0 {
		scutil.APNSLog("No APNSToken for User Device: %d", device.ID) //user.Username)
		return
	}

	var recipients []lib.Recipient
	scutil.APNSLog("notification = true and user_id = %d", device.ID)
	db.Where("notification = true and device_id = ?", device.ID).Find(&recipients)

	payload := apns.NewPayload()
	payload.Alert = message
	payload.Sound = sound
	payload.Badge = len(recipients) + 1

	payload.ContentAvailable = 1
	pn := apns.NewPushNotification()
	pn.DeviceToken = device.APNSToken
	pn.AddPayload(payload)

	scutil.APNSLog("APNSToken for User Device: %d is %s", device.ID, device.APNSToken) //user.Username)
	client := apns.NewClient("gateway.sandbox.push.apple.com:2195", "apns.pem", "apns_private.pem")
	resp := client.Send(pn)

	alert, _ := pn.PayloadString()
	scutil.APNSLog("  Alert: ", alert)
	scutil.APNSLog("Success: ", resp.Success)
	scutil.APNSLog("  Error: ", resp.Error)
}

func sendCustomAPNS(db gorm.DB, device lib.Device, message string) {
	sendCustomAPNSWithSound(db, device, message, "bingbong.aiff")
}

func sendAPNS(db gorm.DB, recipient lib.User, notificationType int) {
	scutil.SLog("Sending notification of type: ", notificationType)
	var notificationText string
    var bSendWithSound bool = false
    var bSendAPNS bool = true
	if notificationType == lib.NOTIFICATION_TYPE_MESSAGE {
		notificationText = "New message"
	} else if notificationType == lib.NOTIFICATION_TYPE_CALL {
		notificationText = "Incoming call"
        bSendWithSound = true
		//sendCustomAPNSWithSound(db, device, notificationText, "old_telephone.mp3")
		//return
	} else if notificationType == lib.NOTIFICATION_TYPE_PING {
		notificationText = "Someone requests your attention"
	} else if notificationType == lib.NOTIFICATION_TYPE_HIGHLIGHT {
		notificationText = "You were highlighted"
	} else {
        bSendAPNS = false
		//return
	}
    
    if (bSendAPNS) {
        // Send APNS to all the user's devices
        var devices []lib.Device
        db.Where("user_id = ?", recipient.ID).Find(&devices)

        for _, device := range devices {
            if (bSendWithSound) {
                sendCustomAPNSWithSound(db, device, notificationText, "old_telephone.mp3")
            } else {
                sendCustomAPNS(db, device, notificationText)
            }
        }         
    } 
}

func sendSMS(number string, messageText string) bool {
	scutil.SLog("Begin")
	client := twilio.NewClient("")

	message, err := twilio.NewMessage(client, "+88", number, twilio.Body(messageText))

	if err != nil {
		scutil.SLog("Error sending SMS: %s", err)
		return false
	} else {
		scutil.SLog("SMS sent: %s", message.Status)
		return true
	}
}

func sendAnswer(conn *lib.Connection, v interface{}, messageType string) {
	scutil.TopLog("[%s] Sending: %s", conn.User.Username, messageType)
	
	json_obj := lib.MpMarshal(v)

	frame := lib.CreateFrame(json_obj)
	conn.Conn.Write(frame)
}

func sendAnswerWithData(conn *lib.Connection, v interface{}, data []byte) {
	json_obj := lib.MpMarshal(v)

	frame := lib.CreateFrameWithData(json_obj, data)
	conn.Conn.Write(frame)
}

func sendToUser(conn *lib.Connection, user *lib.User, v interface{}, messageType string) {
	scutil.TopLog("[%s=>%s] Sending: %s", conn.User.Username, user.Username, messageType)
	//sendToUserDetailed(conn, user, v, false, "", "", false, false)
	sendImmediateMessage(conn, user, v)
}


func sendImmediateMessage(conn *lib.Connection,
                                 user *lib.User,
                                 v interface{}) {
                                 //immediateOnly bool,
                                 //validUntil string,
                                 //messageId string,
                                 //notification bool,
                                 //repeatNotifications bool)
	mp := lib.MpMarshal(v)
	var msg lib.ClientEndToEndMessage
	lib.MpUnmarshal(mp, &msg)

	frame := lib.CreateFrame(mp)

	//if user.RepeatPingDisabled {
	//	repeatNotifications = false
	//}
	ic := lib.InterconnectMessageContainer{User: *user, Frame: frame} //ImmediateOnly: immediateOnly, ValidUntil: validUntil, MessageId: messageId, Notification: notification, RepeatNotifications: repeatNotifications}
	conn.ImmediateMessageHandler <- &ic
}


func prepareStoredMessage(db gorm.DB,
                                 conn *lib.Connection,
                                 v interface{},
                                 messageToken string,
                                 validUntil string,
                                 ) int64 {
	//var storedMessage *lib.StoredMessageNew
	//storedMessage = nil
    var nStoredMessageID int64 = 0
    
	mp := lib.MpMarshal(v)
	var msg lib.ClientEndToEndMessage
	lib.MpUnmarshal(mp, &msg)

	frame := lib.CreateFrame(mp)

    
    storedMessage := lib.StoredMessage{
        Message:                 frame,
        MessageToken:            messageToken,
        UserID:                  conn.User.ID,
        DeviceID:                conn.Device.ID,
        ValidUntil:              validUntil,
        ReadyToSend:             false}
    db.Save(&storedMessage)

    nStoredMessageID = storedMessage.ID

    return nStoredMessageID
}

//addReceipientsForMessage(db, messageID, conn.User.ID, conn.Device.ID)
func addMessageRecipient(db gorm.DB,
                              messageID int64,
                              recipient *lib.User,
                              excludeDeviceID int64,
                              notification bool,
                              repeatNotifications bool) {
    scutil.StoredMessageLog("addMessageRecipient start")
    var devices []lib.Device
    db.Where("user_id = ?", recipient.ID).Find(&devices)

	if recipient.RepeatPingDisabled {
		repeatNotifications = false
	}
    
    for _, device := range devices {
        if (device.ID == excludeDeviceID) {
            scutil.StoredMessageLog("Skip storing recipient record for device %d", device.ID)
        } else {
             scutil.StoredMessageLog("Store recipient record for device %d", device.ID)
             recipient := lib.Recipient {
                MessageID:      messageID,
                UserID:         recipient.ID,
                DeviceID:       device.ID,
                //Notification:            notification,
                RepeatNotification:      repeatNotifications,
                RepeatNotificationCount: 1,
                RepeatNotificationNext:  time.Now().UTC().Add(time.Second * 10),
            }
            db.Save(&recipient)
        }
    }
    scutil.StoredMessageLog("addMessageRecipient end")
}

func sendStoredMessage(db gorm.DB,
                 conn *lib.Connection,
                 messageID int64) {
	var storedMessage lib.StoredMessage
	err := db.Where("ID = ?", messageID).First(&storedMessage).Error

    // trigger send in the DB
	if (err == nil) {
        storedMessage.ReadyToSend = true
        db.Save(&storedMessage)
    }
    // trigger our channel to send message to current connections
    conn.StoredMessageHandler <- messageID
}

func ErrorMessage(conn *lib.Connection,
				  errorCode int,
				  token string,
				  format string, a ...interface{}) {
	// Get name of calling function
	pc, _, _, _ := runtime.Caller(1)
	caller := runtime.FuncForPC(pc).Name()
	logmessage := fmt.Sprintf(format, a...)
	log.Printf("ERROR MESSAGE [%s]: %s", caller, logmessage)
	
	errorMessage := lib.ServerErrorMessage {
		ErrorCode: errorCode,
		Description : logmessage,
		Message: caller,
		Token: token,
	}
	
    setMessageType(&errorMessage)
        sendAnswer(conn, errorMessage, "ServerErrorMessage") 	
	
}
