//
//  call_messages.go
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
	"time"
)

// Call messages /////////////////////
//////////////////////////////////////

type ClientCallCreateMessage struct {
	lib.ClientMessage
}

type ServerCallCreateResponseMessage struct {
	lib.ServerMessage
	CallToken string
}

type ClientCallGetStatusMessage struct {
	lib.ClientMessage
	CallToken string
}

type ServerCallGetStatusResponseMessage struct {
	lib.ServerMessage
	CallToken string
	Status    int
	CreatedAt string
}

type ClientCallUpkeepMessage struct {
	lib.ClientMessage
	CallToken string
}

type ClientCallInvitationMessage struct {
	lib.ClientMessage
	Receiver   string
	PublicIP   string
	PublicPort int
	Token      string
	Users      []string
}

type ServerCallInvitationMessage struct {
	lib.ServerMessage
	Sender     string
	PublicIP   string
	PublicPort int
	Token      string
	Users      []string
}

type ClientCallInvitationResponseMessage struct {
	lib.ClientMessage
	Receiver   string
	Accept     bool
	PublicIP   string
	PublicPort int
	Token      string
}

type ServerCallInvitationResponseMessage struct {
	lib.ServerMessage
	Sender     string
	Accept     bool
	PublicIP   string
	PublicPort int
	Token      string
}

type ClientCallHangupMessage struct {
	lib.ClientMessage
	Receiver string
	Token    string
}

type ServerCallHangupMessage struct {
	lib.ServerMessage
	Sender string
	Token  string
}

func handleClientCallCreateMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")
	var ccm ClientCallCreateMessage
	lib.MpUnmarshal(message, &ccm)

	callIdentifier := lib.GetRandomHash()

	call := lib.Call{
		CallToken:  callIdentifier,
		Status:     lib.CALL_STATUS_ACTIVE,
		Creator:    conn.User,
		LastUpkeep: time.Now().UTC(),
	}

	err := db.Save(&call).Error
	if err != nil {
		scutil.ErrorLog("handleClientCallCreateMessage: Failed to create call!")
		panic("handleClientCallCreateMessage: Failed to create call!")
	}

	ccrm := ServerCallCreateResponseMessage{
		CallToken: callIdentifier,
	}
	setMessageType(&ccrm)
	sendAnswer(conn, ccrm, "ServerCallCreateResponseMessage")
	scutil.DebugLog("Created call")
}

func handleClientCallGetStatusMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")
	var cgsm ClientCallGetStatusMessage
	lib.MpUnmarshal(message, &cgsm)

	var call lib.Call
	err := db.Where("call_token = ?", cgsm.CallToken).First(&call).Error
	if err != nil {
		panic("handleClientCallGetStatusMessage: Call not found: " + cgsm.CallToken)
	}

	upkeep := call.LastUpkeep.Add(30 * time.Second)
	if time.Now().UTC().After(upkeep) {
		scutil.SLog("Call is expired!")
		call.Status = lib.CALL_STATUS_EXPIRED
		db.Save(call)
	}

	cgsrm := ServerCallGetStatusResponseMessage{
		CallToken: call.CallToken,
		Status:    call.Status,
		CreatedAt: call.CreatedAt.Format(time.RFC3339),
	}

	setMessageType(&cgsrm)
	sendAnswer(conn, cgsrm, "ServerCallGetStatusResponseMessage")
}

func handleClientCallUpkeepMessage(conn *lib.Connection, db gorm.DB, message []byte, message_data []byte) {
	scutil.SLog("Begin")
	var cum ClientCallUpkeepMessage
	lib.MpUnmarshal(message, &cum)

	var call lib.Call
	err := db.Where("call_token = ?", cum.CallToken).First(&call).Error
	if err != nil {
		panic("handleClientCallUpkeepMessage: Call not found: " + cum.CallToken)
	}
	if call.Status == lib.CALL_STATUS_EXPIRED {
		scutil.SLog("Call was already expired")
		return
		// TODO Notify client
	}
	if call.CreatorId == conn.User.ID {
		scutil.SLog("Connection is owner of call")
		call.LastUpkeep = time.Now().UTC()
		db.Save(&call)
	} else {
		scutil.SLog("Connection is not owner of call")
	}
}
