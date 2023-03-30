package utils

import "time"

// Message is primary record to pass data from/to bots.
type Message struct {
	ID         int
	From       User
	SenderChat SenderChat `json:"sender_chat,omitempty"`
	// Chat is the conversation the message belongs to
	Chat     *Chat `json:"chat"`
	ChatID   int64
	Sent     time.Time
	HTML     string    `json:",omitempty"`
	Text     string    `json:",omitempty"`
	Entities *[]Entity `json:",omitempty"`
	Image    *Image    `json:",omitempty"`
	ReplyTo  struct {
		From       User
		Text       string `json:",omitempty"`
		Sent       time.Time
		SenderChat SenderChat `json:"sender_chat,omitempty"`
	} `json:",omitempty"`
}

// Response describes bot's answer on particular message.
type Response struct {
	Text        string
	ReadyToSend bool  // if true, then response will be sent to user
	Pin         bool  // enable pin
	Unpin       bool  // enable unpin
	Preview     bool  // enable web preview
	User        User  // user to ban
	ChannelID   int64 // channel to ban, if set then User and BanInterval are ignored
	ReplyTo     int   // message to reply to, if 0 then no reply but common message
}

// SenderChat is the sender of the message, sent on behalf of a chat. The
// channel itself for channel messages. The supergroup itself for messages
// from anonymous group administrators. The linked channel for messages
// automatically forwarded to the discussion group
type SenderChat struct {
	// ID is a unique identifier for this chat
	ID int64 `json:"id"`
	// field below used only for logging purposes
	// UserName for private chats, supergroups and channels if available, optional
	UserName string `json:"username,omitempty"`
	// Type of chat, can be either “private”, “group”, “supergroup” or “channel”
	Type string `json:"type"`
}

// Chat represents a chat.
type Chat struct {
	// ID is a unique identifier for this chat
	ID int64 `json:"id"`
	// Type of chat, can be either “private”, “group”, “supergroup” or “channel”
	Type string `json:"type"`
}

// Entity represents one special entity in a text message.
// For example, hashtags, usernames, URLs, etc.
type Entity struct {
	Type   string
	Offset int
	Length int
	URL    string `json:",omitempty"` // For “text_link” only, url that will be opened after user taps on the text
	User   *User  `json:",omitempty"` // For “text_mention” only, the mentioned user
}

// Image represents image.
type Image struct {
	// FileID corresponds to Telegram file_id
	FileID   string
	Width    int
	Height   int
	Caption  string    `json:",omitempty"`
	Entities *[]Entity `json:",omitempty"`
}

// User defines user info of the Message.
type User struct {
	ID          int64
	Username    string
	DisplayName string
}
