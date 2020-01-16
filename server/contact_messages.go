//
// contact_messages.go
//  Sechat Server
//
//  Copyright 2017 JW Technologies. All rights reserved.
//

package main

import (
	"github.com/jinzhu/gorm"
	"sechat-server/lib"
	"sechat-server/scutil"
	"math/rand"
	"fmt"
    //"sechat-server/sechat_common"
)

/*
type ServerContactRequest struct {
	lib.ServerMessage
	Username            string // valid user
    RequestToken        string
}
*/
type ClientContactAccept struct {
	lib.ClientMessage
	Username            string // valid user
    RequestToken        string
	Successful          bool
}

type ServerContactAccept struct {
	lib.ServerMessage
	Username            string
    RequestToken        string
	Successful          bool
	Error               string
    Name                string
	Alias				string
	PhoneNumber			string
	Email				string
	Website				string
	Title				string
	Organization		string		
}

type ClientAddContactMessage struct {
	lib.ClientMessage
	Name string
}

type ServerAddContactMessage struct {
	lib.ServerMessage
	Successful bool
	Error      string
	Username   string
	//UserX      string
	//UserY      string
	
	Alias				string
	PhoneNumber			string
	Email				string
	Website				string
	Title				string
	Organization		string		
}

type ServerGetProfileResponseMessage struct {
	lib.ServerMessage
	Successful bool
	Error      string
	Username   string
	//UserX      string
	//UserY      string
	// TODO: Devices & device keys
	Profilepicture []byte // Depends on whether the GetProfileMessage had Profilepicture set
}

type ClientContactRequestMessage struct {
	lib.ClientMessage
    Token               string
	Contact             string // valid user
    LinkQuery           int
}


func handleClientContactRequestMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.DebugLog("handleClientContactRequestMessage")
    
    var bContinue bool = true
 	var clientContactRequest ClientContactRequestMessage
	err := lib.MpUnmarshal(message, &clientContactRequest)
    
    if (err != nil) {
        ErrorMessage(conn,
                     lib.ERROR_CODE_INVALID_MESSAGE,
                     clientContactRequest.Token,
                     "Invalid Message ClientContactRequestMessage")
        bContinue = false
    } 
    
    if (bContinue) {
        contactUser := getUser(db, clientContactRequest.Contact)
        if (contactUser != nil) {
            if (clientContactRequest.LinkQuery == lib.LINK_TYPE_REQUEST) {
                scutil.DebugLog("LINK_TYPE_REQUEST")
                var userContactLink lib.ContactLink
                err1 := db.Debug().Where("user_id = ? and contact_id = ?", conn.User.ID, contactUser.ID).First(&userContactLink).Error
                if (err1 != nil) {
                    scutil.DebugLog("make new contact link with request flag")
                     newUserContactLink := lib.ContactLink {
                        UserID: conn.User.ID,
                        ContactID: contactUser.ID,
                        LinkType: lib.LINK_TYPE_REQUEST,
                    }
                    db.Save(&newUserContactLink)
                }
            } else {
                scutil.DebugLog("LINK_TYPE_ACCEPT")
                 var contactUserLink lib.ContactLink
                  err2 := db.Debug().Where("user_id = ? and contact_id = ?", contactUser.ID, conn.User.ID).First(&contactUserLink).Error
                 if (err2 != nil) {
                    scutil.DebugLog("make new contact link with accept flag")
                   newContactUserLink := lib.ContactLink {
                        UserID: contactUser.ID,
                        ContactID: conn.User.ID,
                        LinkType: clientContactRequest.LinkQuery,
                    }
                    db.Save(&newContactUserLink)
                 } else {
                   scutil.DebugLog("update contact link with accept flag")
                    contactUserLink.LinkType = clientContactRequest.LinkQuery
                    db.Save(&contactUserLink)
                 }
               
                
            }
        } else {
                  scutil.DebugLog("Error message")
            ErrorMessage(conn,
                         lib.ERROR_CODE_INVALID_PARAMETER,
                         clientContactRequest.Token,
                         "Invalid Contact %s", clientContactRequest.Contact)
        }        
    }

   
    
   
}

func handleClientAddContactMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")

	var add_contact_message lib.ClientAddContactMessage
	err := lib.MpUnmarshal(message, &add_contact_message)
	if err != nil {
		panic("handleClientAddContactMessage: Failed to unmarshal message")
	}

	var user lib.User
	err = db.Where("username = ?", add_contact_message.Name).First(&user).Error
	
	if (err == nil) {
		scutil.SLog("User \"%s\" found.", add_contact_message.Name)
		add_contact_response := lib.ServerAddContactMessage {
			ServerMessage: lib.ServerMessage{MessageType: "ServerAddContactMessage"},
			Successful: true,
			Error: "",
			Username: user.Username,
			//UserX: user.X,
			//UserY: user.Y,
			Alias: user.Alias,
			PhoneNumber: user.Phonenumber,
			Email: user.Email,
			Website: user.Website,
			Title: user.Title,
			Organization: user.Organization,
		}
		sendAnswer(conn, add_contact_response, "ServerAddContactMessage")

		if user.Profilepicture != nil {
			getprofileresponsemessage := ServerGetProfileResponseMessage{
				Successful:     true,
				Username:       user.Username,
				//UserX:          user.X,
				//UserY:          user.Y,
				Profilepicture: user.Profilepicture,
			}
			setMessageType(&getprofileresponsemessage)
			scutil.SLog("Sending profile!")
			sendAnswer(conn, getprofileresponsemessage, "ServerGetProfileResponseMessage")
		}
	} else {
		scutil.SLog("Error: User \"%s\" not found.", add_contact_message.Name)
		add_contact_response := lib.ServerAddContactMessage {
			ServerMessage: lib.ServerMessage{
			MessageType: "ServerAddContactMessage"},
			Successful: false,
			Error: "User not found.",
		}
		sendAnswer(conn, add_contact_response, "ServerAddContactMessage")		
	}
}


func handleClientRegisterPhonenumberMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")
	var rpm lib.ClientRegisterPhonenumberMessage
	lib.MpUnmarshal(message, &rpm)

	verificationCode := rand.Intn(1000000)
	verificationString := fmt.Sprintf("%06d", verificationCode)
	verificationMessage := fmt.Sprintf("Your verification code is: %s", verificationString)
	conn.User.PhonenumberToken = verificationString
	conn.User.Phonenumber = rpm.Phonenumber
	conn.User.PhonenumberVerified = false
	db.Save(conn.User)
	var vpr lib.ServerRegisterPhonenumberResponseMessage
	if !sendSMS(rpm.Phonenumber, verificationMessage) {
		vpr = lib.ServerRegisterPhonenumberResponseMessage{
			Successful: false,
			Error:      "Invalid number.",
		}
	} else {
		vpr = lib.ServerRegisterPhonenumberResponseMessage{
			Successful: true,
		}
	}
	setMessageType(&vpr)
	sendAnswer(conn, vpr, "ServerRegisterPhonenumberResponseMessage")
}

func handleClientVerifyPhonenumberMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")
	var vpm lib.ClientVerifyPhonenumberMessage
	lib.MpUnmarshal(message, &vpm)

	var vpr lib.ServerVerifyPhonenumberResponseMessage
	if conn.User.PhonenumberToken == vpm.VerificationCode {
		scutil.SLog("Phonenumber verified!")
		conn.User.PhonenumberVerified = true
		db.Save(conn.User)
		vpr = lib.ServerVerifyPhonenumberResponseMessage{Successful: true}
	} else {
		scutil.SLog("Wrong verification code!")
		vpr = lib.ServerVerifyPhonenumberResponseMessage{Successful: false}
	}
	setMessageType(&vpr)
	sendAnswer(conn, vpr, "ServerVerifyPhonenumberResponseMessage")
}

func handleClientSyncContactsMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")
	var msg lib.ClientSyncContactsMessage
	lib.MpUnmarshal(message, &msg)

	var response lib.ServerSyncContactsResponseMessage
	setMessageType(&response)
	// TODO Check if this doesn't take up too much memory (but shouldn't)
	response.Users = make([]lib.UsernamePhoneNumberPair, 0)
	for _, phoneNumber := range msg.PhoneNumbers {
		scutil.SLog("Checking phonenumber: %s", phoneNumber)
		var user lib.User
		err := db.Where("phonenumber = ?", phoneNumber).First(&user).Error
		if err != nil {
			scutil.SLog("No user found for number: %s", phoneNumber)
			continue
		}
		response.Users = append(response.Users, lib.UsernamePhoneNumberPair{Username: user.Username, PhoneNumber: phoneNumber})
	}

	sendAnswer(conn, response, "ServerSyncContactsResponseMessage")
}

