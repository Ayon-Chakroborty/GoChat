package main

import (
	"encoding/json"
	"time"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type EventHandler func(event Event, c *Client) error

const (
	EventSendMessage    = "send_message"
	EventNewMessage     = "new_message"
	EventChangeChatRoom = "change_room"
)

type SendMessageEvent struct {
	Message  string `json:"message"`
	From     string `json:"from"`
	Email    string `json:"email"`
	Chatroom string `json:"chatroom"`
}

type NewMessageEvent struct {
	SendMessageEvent
	Sent time.Time `json:"sent"`
}

type ChangeRoomEvent struct {
	Name string `json:"name"`
}
