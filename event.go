package main

import (
	"encoding/json"
	"time"
)

type Event struct {
	Type    string 			`json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type SendMessageEvent struct {
	Message string `json:"message"`
	From	string `json:"from"`
}

type NewMessageEvent struct {
	SendMessageEvent
	Sent time.Time `json:"sent"`
}

var (
	EventSendMessage = "send_message"
	EventNewMessage = "new_message"
)

type EventHandler func(event Event, c *Client) error