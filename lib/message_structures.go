package lib

const (
	NOTIFICATION_TYPE_NONE      = 0
	NOTIFICATION_TYPE_MESSAGE   = 1
	NOTIFICATION_TYPE_CALL      = 2
	NOTIFICATION_TYPE_PING      = 3
	NOTIFICATION_TYPE_HIGHLIGHT = 4
)

// TODO:
// - Suffix everything with Message or not?
// - Should clientmessages have the possibility to include a message token to
//   assign responses to the according requests?

// Login & registration messages
/*
type ClientRegisterMessage struct {
	ClientMessage
	Username   string
	UserX      string // user public key X
	UserY      string // user public key Y
	Devicename string // currently unused.. should still be sent
	DeviceX    string // device public key X
	DeviceY    string // device public key Y
}

type ServerRegisterResponseMessage struct {
	ServerMessage
	Successful bool
	Error      string
}

type ClientLoginInitialMessage struct {
	ClientMessage
	Username   string
	Devicename string
}

type ServerLoginChallengeMessage struct {
	ServerMessage
	Successful bool
	Error      string
	Challenge  string
}

type ClientLoginChallengeResponseMessage struct {
	ClientMessage
	R         string // Signature part R
	S         string // Signature part S
	APNSToken string // Apple push notification token - can be ommitted.
}
*/

type ServerEncapsulatingMessage struct {
	Frame []byte
	Id    int64
}

// Messaging functions
type ServerErrorMessage struct {
	ServerMessage
	ErrorCode			int
	Description			string
	Message				string
	Token				string
}

type ClientEndToEndMessage struct {
	ClientMessage
	Receiver            string
	Encrypted           []byte
	Signature           string
	Notification        int
	ImmediateOnly       bool
	Priority            int
	MessageToken        string
	RepeatNotifications bool
}

type ClientEndToEndGenericMessage struct {
	ClientMessage
	Receiver            string
	Data                []byte
	EncryptionType      string
	Notification        int
	ImmediateOnly       bool
	Priority            int
	MessageToken        string
	RepeatNotifications bool
}

type ServerEndToEndResponseMessage struct {
	ServerMessage
	MessageId string
}

type ServerEndToEndMessage struct {
	ServerMessage
	Encrypted []byte
	Signature string
	// TODO can we hide the sender somehow?
	//      maybe use keyIDs? difficult i guess.
	Sender   string
	Priority int
}

type ServerEndToEndGenericMessage struct {
	ServerMessage
	Data     					[]byte
	EncryptionType 			    string
	Sender   					string
	Priority 					int
}

type ClientBlockUserMessage struct {
	ClientMessage
	Block    bool
	Username string
}

/*
type ClientUpdateGroupChatMessage struct {
	ClientMessage
	Name string
	Identifier string
	Users []string
}

type ServerUpdateGroupChatResponseMessage struct {
	ServerMessage
	Identifier string
	Successful bool
	Error string
}
*/



type ClientRegisterPhonenumberMessage struct {
	ClientMessage
	Phonenumber string
}

type ServerRegisterPhonenumberResponseMessage struct {
	ServerMessage
	Successful bool
	Error      string
}

type ClientVerifyPhonenumberMessage struct {
	ClientMessage
	VerificationCode string
}

type ServerVerifyPhonenumberResponseMessage struct {
	ServerMessage
	Successful bool
}

type ClientRecallEndToEndMessage struct {
	ClientMessage
	MessageId string
}

type ServerRecallEndToEndResponseMessage struct {
	ServerMessage
	MessageId  string
	Successful bool
}

type ClientSyncContactsMessage struct {
	ClientMessage
	PhoneNumbers []string
}

// TODO this should probably not be here
type UsernamePhoneNumberPair struct {
	Username    string
	PhoneNumber string
}

type ServerSyncContactsResponseMessage struct {
	ServerMessage
	Users []UsernamePhoneNumberPair
}

