package lib

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	//"database/sql"  // useed for sql.NullInt64
	"math/big"
	"time"
)

const (
	CALL_STATUS_ACTIVE  = iota
	CALL_STATUS_EXPIRED = iota
)

type User struct {
	ID                  int64
	Username            string // valid email address
	Name                string // First and Last Name
	Password			string
	Profilepicture      []byte
	ProfilepictureId    string
	LastSeen            time.Time
	Status              int
	StatusText          string
	UpdateToken         string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Phonenumber         string
	PhonenumberToken    string
	PhonenumberVerified bool
	BlockedUsers        []User `gorm:"foreignkey:user_id;associationforeignkey:blocked_user_id;many2many:user_blocked;"`
	RepeatPingDisabled  bool   `gorm:"update"`
	PreKeyBundles       []PreKeyBundle
	Email				string
	Website				string
	Title				string
	Organization		string
	Alias				string
}

// Currently unused models
type Device struct {
	ID 					int64
	UserID    			int64
	DeviceToken 		string			// software/hw generated identifier for a device
	DeviceName			string          // user's name for the given device
	X         			string 			// Public Key X
	Y         			string 			// Public Key Y
	APNSToken           string
	LastSeen  			time.Time
	CreatedAt 			time.Time
	UpdatedAt 			time.Time
	// just sketching here
	// type ECDSASignature?
	Signature 			string
}

/*
type StoredMessage struct {
	ID                      int64
	User                    User
	UserID                  sql.NullInt64
	Message                 []byte
	MessageId               string
	ImmediateOnly           bool
	ValidUntil              string
	Notification            bool
	RepeatNotification      bool
	RepeatNotificationCount int64
	RepeatNotificationNext  time.Time
}
*/
type ContactLink struct {
	ID						int64
	UserID					int64		// From ID
	ContactID				int64		// Link To ID
	LinkType				int         // A LINK_TYPE value
}

type StoredMessage struct {
	ID                      int64
	Message                 []byte
	MessageToken            string
	UserID					int64				// Sender UserID
	DeviceID				int64				// Sender DeviceID
	ValidUntil              string
	ReadyToSend				bool
}

type Recipient struct {
	ID                      int64
	MessageID				int64
	User                    User
	UserID                  int64         //sql.NullInt64
	DeviceID				int64
	Notification            bool
	RepeatNotification      bool
	RepeatNotificationCount int64
	RepeatNotificationNext  time.Time
}

func (u User) GenerateUpdateToken() {
	u.UpdateToken = GetRandomHash()
}

func (device Device) GetPublicKey() ecdsa.PublicKey {
	Xi := big.NewInt(0)
	Xi.SetString(device.X, 10)
	Yi := big.NewInt(0)
	Yi.SetString(device.Y, 10)
	pk := ecdsa.PublicKey{Curve: elliptic.P521(), X: Xi, Y: Yi}

	return pk
}

func (device Device) VerifySignature(hash, r, s string) bool {
	ri := big.NewInt(0)
	si := big.NewInt(0)
	ri.SetString(r, 10)
	si.SetString(s, 10)
	publicKey := device.GetPublicKey()
	if ecdsa.Verify(&publicKey, []byte(hash), ri, si) {
		return true
	} else {
		return false
	}
}

func (u User) IsIn(users []User) bool {
	for _, user := range users {
		if user.ID == u.ID {
			return true
		}
	}
	return false
}


type Channel struct {
	Id                             int64
	Creator                        User
	CreatorId                      int64
	Name                           string
	Description                    string
	Picture                        []byte
	Users                          []User `gorm:"many2many:channel_users;"`
	NotificationsDisabled          []User `gorm:"many2many:channel_notifications_disabled;"`
	HighlightNotificationsDisabled []User `gorm:"many2many:channel_highlight_notifications_disabled"`
}

type ChannelUser struct {
	ChannelId      					int64
	UserId      					int64	
}

type Groupchat struct {
	Id                             int64
	Identifier                     string
	Creator                        User
	CreatorId                      int64
	Name                           string
	Picture                        []byte
	Users                          []GroupchatToUser
	NotificationsDisabled          []User `gorm:"many2many:groupchat_notifications_disabled;"`
	HighlightNotificationsDisabled []User `gorm:"many2many:groupchat_highlight_notifications_disabled"`
}

type GroupchatToUser struct {
	Id          int64
	Groupchat   Groupchat
	GroupchatId int64
	User        User
	UserId      int64
}

type PreKeyBundle struct {
	Id          			int64
	DeviceID				int64
	PreKeyBundleID			int64
	Bundle                  string `sql:"size:2048"`
	User        			User
	UserId      			int64
}

type Session struct {
	Id   int64
	User User
	//Challenge string
	Token     string // Random token per session to avoid replay attacks.
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Call struct {
	Id         int64
	CallToken  string
	Status     int
	Creator    User
	CreatorId  int64
	LastUpkeep time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
