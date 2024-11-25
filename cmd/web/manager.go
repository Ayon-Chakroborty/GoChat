package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	WebsocketUpgrader = websocket.Upgrader{
		CheckOrigin:     checkOrigin,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type Manager struct {
	clients ClientList
	sync.RWMutex

	handlers map[string]EventHandler
}

func (app *application) NewManager() *Manager {
	m := &Manager{
		clients:  make(ClientList),
		handlers: make(map[string]EventHandler),
	}
	return m
}

func (app *application) setupEventHandlers() {
	app.wsManager.handlers[EventSendMessage] = app.SendMessage
	app.wsManager.handlers[EventChangeChatRoom] = ChatRoomHandler
}

func ChatRoomHandler(event Event, c *Client) error {
	var changeRoomEvent ChangeRoomEvent

	if err := json.Unmarshal(event.Payload, &changeRoomEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}
	changeRoomEvent.Name = strings.TrimSpace(changeRoomEvent.Name)
	if changeRoomEvent.Name == "" {
		return nil
	}

	c.chatroom = changeRoomEvent.Name

	return nil
}

func (app *application) SendMessage(event Event, c *Client) error {
	var chatEvent SendMessageEvent

	if err := json.Unmarshal(event.Payload, &chatEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	var broadMessage NewMessageEvent

	broadMessage.Sent = time.Now()
	broadMessage.Message = chatEvent.Message
	broadMessage.From = chatEvent.From
	broadMessage.Email = chatEvent.Email
	broadMessage.Chatroom = chatEvent.Chatroom

	data, err := json.Marshal(broadMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message : %v", err)
	}

	if broadMessage.Message != "" {
		c.chatroom = broadMessage.Chatroom
		err = app.chatModel.Insert(broadMessage.Chatroom, broadMessage.Email, false, broadMessage.Message, broadMessage.From)
		if err != nil {
			return fmt.Errorf("failed to save broadcast message : %v", err)
		}
	}

	outgoingEvent := Event{
		Payload: data,
		Type:    EventNewMessage,
	}

	for client := range c.manager.clients {
		log.Println(client)
		if client.chatroom == c.chatroom {
			client.egress <- outgoingEvent
		}
	}

	return nil
}

func (m *Manager) routeEvent(event Event, c *Client) error {
	if handler, ok := m.handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("there is no such event type")
	}
}

func (app *application) ServeWS(w http.ResponseWriter, r *http.Request) {
	log.Println("new connection")

	// upgrade http connection to websocket
	conn, err := WebsocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := app.NewClient(r, conn, app.wsManager)

	app.addClient(client)

	// Start client process
	go app.readMessages(client)
	go app.writeMessages(client)
}

func (app *application) addClient(client *Client) {
	app.wsManager.Lock()
	defer app.wsManager.Unlock()

	app.wsManager.clients[client] = true
}

func (app *application) removeClient(client *Client) {
	app.wsManager.Lock()
	defer app.wsManager.Unlock()

	if _, ok := app.wsManager.clients[client]; ok {
		client.connection.Close()
		delete(app.wsManager.clients, client)
	}
}

func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	switch origin {
	case "https://localhost:4000":
		return true
	default:
		return false
	}
}
