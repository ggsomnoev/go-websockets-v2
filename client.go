package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	connection *websocket.Conn
	manager    *Manager

	// to avoid concurrent writes on the connection
	egress chan Event
}

var (
	pongWait     = 10 * time.Second
	pingInterval = (9 * pongWait) / 10
)

func NewClient(connection *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: connection,
		manager:    manager,
		egress:     make(chan Event),
	}
}

func (c *Client) readMessages() {
	defer func() {
		c.manager.removeClient(c)
	}()

	if err := c.connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		fmt.Printf("Error trying to set read deadline: %v\n", err)
		return
	}

	c.connection.SetReadLimit(512) // to avoid large messages

	c.connection.SetPongHandler(c.pongHandler)

	for {
		_, payload, err := c.connection.ReadMessage()

		if err != nil {
			fmt.Printf("Error trying to read messages: %v\n", err)
			break
		}

		var req Event

		if err := json.Unmarshal(payload, &req); err != nil {
			fmt.Printf("Coudln't unmarshal message req: %v\n", err)
		}

		fmt.Printf("Message received: %v\n", req)

		c.manager.routeMessage(req, c)
	}
}

func (c *Client) writeMessages() {
	defer func() {
		c.manager.removeClient(c)
	}()

	ticker := time.NewTicker(pingInterval)

	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					fmt.Printf("Coudln't send message, closed channel: %v\n", err)
				}
				return
			}

			data, err := json.Marshal(message)

			if err != nil {
				fmt.Printf("Error trying to marshal the message: %v\n", err)
				return
			}

			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				fmt.Printf("Coudln't send message: %v\n", err)
				return
			}
		case <-ticker.C:
			if err := c.connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				fmt.Printf("Coudln't send ping message, closed channel: %v\n", err)
				return
			}
		}
	}
}

func (c *Client) pongHandler(message string) error {
	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}
