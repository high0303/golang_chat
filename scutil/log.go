package scutil

import (
	"fmt"
	"runtime"
	//"errors"
	"log"
	"github.com/jinzhu/gorm"

)

// Package Scope Variables
var bTopLevel bool
var bFramingLog bool
var bSLog bool
var bDBLog bool
var bDebugLog bool
var bRegisterLoginLog bool
var bEndToEndMessageLog bool
var bStoredMessageLog bool
var bAPNSLog bool

func init() {
	log.Printf("Calling log init")
	bTopLevel = true			// Basic top level activity: always leave on
	bFramingLog = false			// Debug framing of messages: verbose leave off!!
	bSLog = false				// Legacy messages
	bDBLog = true				// Debug Database access
	bDebugLog = true			// Generic Debug statement (for code development)
	bRegisterLoginLog = true    // Debug Register, Login and Checkin messages
	bEndToEndMessageLog = false  // Basic messaging
	bStoredMessageLog = false    // Storage of messges before sending
	bAPNSLog = true              // Apple Push Notification Service
}

func TopLog(format string, a ...interface{}) {
	if (bTopLevel) {
		logmessage := fmt.Sprintf(format, a...)
		log.Printf("%s", logmessage)
	}
}

// Current global log function
func SLog(format string, a ...interface{}) {
	if (bSLog) {
		// Get name of calling function
		pc, _, _, _ := runtime.Caller(1)
		caller := runtime.FuncForPC(pc).Name()
		logmessage := fmt.Sprintf(format, a...)
		log.Printf("%s: %s", caller, logmessage)
	}
}

func ErrorLog(format string, a ...interface{}) {
	// Get name of calling function
	pc, _, _, _ := runtime.Caller(1)
	caller := runtime.FuncForPC(pc).Name()
	logmessage := fmt.Sprintf(format, a...)
	log.Printf("ERROR [%s]: %s", caller, logmessage)
}

func FramingLog(format string, a ...interface{}) {
	if (bFramingLog) {
		logmessage := fmt.Sprintf(format, a...)
		log.Printf("Framing: %s", logmessage)
	}
}

func EnableDBLog(db gorm.DB) {
	if (bDBLog) {
		db.LogMode(true)
	} else {
		db.LogMode(false)
	}
}

func DebugLog(format string, a ...interface{}) {
	if (bDebugLog) {
		logmessage := fmt.Sprintf(format, a...)
		log.Printf("%s", logmessage)
	}
}

func RegisterLoginLog(format string, a ...interface{}) {
	if (bRegisterLoginLog) {
		logmessage := fmt.Sprintf(format, a...)
		log.Printf("REG/LOG: %s", logmessage)
	}
}

func EndToEndMessageLog(format string, a ...interface{}) {
	if (bEndToEndMessageLog) {
		logmessage := fmt.Sprintf(format, a...)
		log.Printf("E2EMSG: %s", logmessage)
	}
}

func StoredMessageLog(format string, a ...interface{}) {
	if (bStoredMessageLog) {
		logmessage := fmt.Sprintf(format, a...)
		log.Printf("STRMSG: %s", logmessage)
	}
}

func APNSLog(format string, a ...interface{}) {
	if (bAPNSLog) {
		logmessage := fmt.Sprintf(format, a...)
		log.Printf("APNS: %s", logmessage)
	}
}
