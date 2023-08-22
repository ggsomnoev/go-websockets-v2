package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var WebsockerUpdrager = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

type Manager struct {
	clients map[*websocket.Conn]*Client
	sync.RWMutex

	otps     RetentionMap
	handlers map[string]EventHandler
}

type userLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type response struct {
	OTP string `json:"otp"`
}

func NewManager(ctx context.Context) *Manager {
	m := &Manager{
		clients:  make(map[*websocket.Conn]*Client),
		handlers: make(map[string]EventHandler),
		otps:     NewRetentionMap(ctx, 5*time.Second),
	}

	m.setupEventHandlers()

	return m
}

func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendMessage] = SendMessage
}

func SendMessage(event Event, c *Client) error {
	var chatEvent SendMessageEvent

	if err := json.Unmarshal(event.Payload, &chatEvent); err != nil {
		return fmt.Errorf("Error trying to unmarshal message payload: %v\n", err)
	}

	broadCastMessage := NewMessageEvent {
		SendMessageEvent: chatEvent,
		Sent: time.Now(),
	}

	data, err := json.Marshal(broadCastMessage)

	if err != nil {
		return fmt.Errorf("Error trying to marshal broadcast message: %v\n", err)
	}

	outGoingData := Event {
		Payload: data,
		Type: EventNewMessage,
	}

	for _, client := range c.manager.clients {
		client.egress <- outGoingData
	}
	return nil
}

func (m *Manager) routeMessage(event Event, c *Client) {
	if handler, ok := m.handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			fmt.Printf("Error occured trying to handle the message: %v\n", err)
		}
	}
}

func (m *Manager) upgradeConnection(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Client connected: %v\n", r.RemoteAddr)

	otp := r.URL.Query().Get("otp")

	if !m.otps.verifyOTP(otp) {
		http.Error(w, "Not a valid OTP token", http.StatusUnauthorized)
		return
	}

	conn, err := WebsockerUpdrager.Upgrade(w, r, nil)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error occured during connection upgrade: %v\n", err), http.StatusUnauthorized)
		return
	}

	client := NewClient(conn, m)

	m.addClient(client)

	go client.readMessages()
	go client.writeMessages()
}

func (m *Manager) loginHandler(w http.ResponseWriter, r *http.Request) {
	var req userLoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Couldn't decode request body: %v\n", err), http.StatusBadRequest)
		return
	}

	if req.Username != "test" && req.Password != "123" {
		http.Error(w, "Wrong Credentials", http.StatusUnauthorized)
		return
	}

	otp := m.otps.generateOTP()

	res := response{
		OTP: otp.Key,
	}

	data, err := json.Marshal(res)

	if err != nil {
		http.Error(w, "Error marshaling the generated token data", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (m *Manager) addClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	m.clients[client.connection] = client
}

func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.clients[client.connection]; ok {
		fmt.Printf("Connection Closed for: %v\n", client.connection.RemoteAddr())
		if err := client.connection.Close(); err != nil {
			fmt.Printf("Error occured trying to close the connection: %v\n", err)
		}
		delete(m.clients, client.connection)
	}
}

func checkOrigin(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Origin"), "localhost:8080")
}
