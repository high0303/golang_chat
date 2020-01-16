package lib

import (
	"net"
	"time"
)

const (
	STATUS_OFFLINE   = 1
	STATUS_CONNECTED = 2
	STATUS_ONLINE    = 3
)

type EConnectionStatus int

const (
	EConnectionStatus_Uninitialized EConnectionStatus = 1 + iota
	EConnectionStatus_Challenge_With_DeviceToken
	EConnectionStatus_Challenge_No_DeviceToken
	EConnectionStatus_Authenticated
)

const (
    LINK_TYPE_REQUEST   = 1
	LINK_TYPE_ACCEPT   = 2
	LINK_TYPE_DENY	   = 3
	LINK_TYPE_BLOCK	   = 4
)

const (
	ERROR_CODE_INVALID_PARAMETER = 1 + iota
	ERROR_CODE_INVALID_MESSAGE
)
type Connection struct {
	Conn                      net.Conn
	User                      User						// User
	Device					  Device					// Device used by the user for this connection
	Challenge                 string
	//Authenticated             bool
	Status					  EConnectionStatus		  
	RegisterConnection        chan<- *Connection
	DeregisterConnection      chan<- *Connection
	ImmediateMessageHandler   chan<- *InterconnectMessageContainer
	StoredMessageHandler      chan<- int64
	StatusNotificationHandler chan<- *User
	ProfileNotificationHandler chan<- *User
	SendDataChannel           chan []byte
	SubscribedTo              []string
	Invisible                 bool
	LastActivity              time.Time
	DeviceToken				  string
}

type InterconnectMessageContainer struct {
	User                User
	Frame               []byte
	//ImmediateOnly       bool
	//ValidUntil          string
	//MessageId           string
	//Notification        bool
	//RepeatNotifications bool
}

type ClientMessage struct {
	MessageType string
}

type ServerMessage struct {
	MessageType string
}

func (cm ClientMessage) SetMessageType(messagetype string) {
	cm.MessageType = messagetype
}

func (sm ServerMessage) SetMessageType(messagetype string) {
	sm.MessageType = messagetype
}

type ClientRegisterMessage struct {
	ClientMessage
	Username   			string
	Password   			string
}

type ServerRegisterMessage struct {
	ServerMessage
	Successful bool
	Error      string
}

type ClientAppLoginMessage struct {
	ClientMessage
	Username   			string 
	DeviceToken			string // Unique identifier for the device
	DeviceName 			string // User string for the device
	Password   			string
	DeviceX      		string // device public key X
	DeviceY      		string // device public key Y
}

type ClientAppLogoutMessage struct {
	ClientMessage
	Username   			string
	DeviceToken			string
}

type ClientCheckinMessage struct {
	ClientMessage
	Username   			string
	DeviceToken 		string
}

type ServerCheckinChallengeMessage struct {
	ServerMessage
	Successful bool
	Error      string
	Challenge  string
}

type ClientCheckinResponseMessage struct {
	ClientMessage
	R           string // Signature part R
	S           string // Signature part S
	APNSToken   string
	Invisible   bool
	SubscribeTo []string
}

type ServerCheckinResultMessage struct {
	ServerMessage
	Successful 			bool
	Error      			string
	NumKeyBundles		int  // number of Axolotl prekey bundles available
}

type ClientPublishPreKeyBundlesMessage struct {
	ClientMessage
	Username   				string
	FirstSignedPreKeyID		int
	PreKeyBundles			[]string
}

type ClientRequestPreKeyBundle struct {
	ClientMessage
	Username   				string
}

type ServerPreKeyBundleMessage struct {
	ServerMessage
	Successful 				bool
	Error      			    string
	Username   				string
	DeviceID				int
	SignedPreKeyID			int
	PreKeyBundle			string
}

type ClientSetProfilepictureMessage struct {
	ClientMessage
	NotifyUsers    []string // TODO this will be changed in a future version.. just a hack to get it working
	ProfilePicture []byte
}

// TODO this is just temporary to get profile pictures working.. will be removed in a future version
type ServerProfilepictureUpdateMessage struct {
	ServerMessage
	Sender         string
	ProfilePicture []byte
}

type ServerStatusNotificationMessage struct {
	ServerMessage
	User       string
	StatusText string
	LastSeen   int64
	Status     int // see STATUS_ constants at the beginning of this file.
}

type ClientAddContactMessage struct {
	ClientMessage
	Name string
}

type ServerAddContactMessage struct {
	ServerMessage
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

type ClientCreateGroupchatMessage struct {
	ClientMessage
	Name  string
	Users []string
}

type ServerCreateGroupchatMessage struct {
	ServerMessage
	Successful bool
	Error      string
	Identifier string
}

type ServerGroupchatInvitationMessage struct {
	ServerMessage
	Name       string
	Creator    string
	Users      []string
	Identifier string
}
