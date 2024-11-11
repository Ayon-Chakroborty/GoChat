package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

var (
	pongWait = 10 * time.Second

	// need to be lower than pong wait
	pingIntegerval = (pongWait * 9) / 10
)

type ClientList map[*Client]bool

type Client struct {
	connection *websocket.Conn
	manager    *Manager

	// egress is used to aboid concurrent writes on the websocket connection
	egress chan Event
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
		egress:     make(chan Event),
	}
}

func (c *Client) readMessages() {
	defer func() {
		// cleanup connection
		c.manager.removeClient(c)
	}()
	
	// configer wait time for pong response
	if err := c.connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil{
		log.Println("Read Msg Err:", err)
		return
	}

	// maximum size per message is 512 bytes
	c.connection.SetReadLimit(512)

	// reset wait time once pong response is received
	c.connection.SetPongHandler(c.pongHandler)

	for {
		_, payload, err := c.connection.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			break
		}

		var request Event

		if err := json.Unmarshal(payload, &request); err != nil {
			log.Printf("error marshalling event: %v", err)
			break
		}

		if err := c.manager.routeEvent(request, c); err != nil {
			log.Println("error handeling message: ", err)
		}
	}
}

func (c *Client) writeMessages() {
	defer func() {
		c.manager.removeClient(c)
	}()

	ticker := time.NewTicker(pingIntegerval)

	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("connection closed: ", err)
				}
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Println(err)
				return
			}

			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("failed to send message: %v\n", err)
			}
			log.Printf("Message sent: %s\n", message)
		case <- ticker.C:
			log.Println("ping")

			// Send a Ping to the Client
			if err := c.connection.WriteMessage(websocket.PingMessage, []byte{}); err != nil{
				log.Println("Write Msg Err: ", err)
				return
			}
		}
	}
}

func (c *Client) pongHandler(pongMsg string) error{
	log.Println("pong")
	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}
