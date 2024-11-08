package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var(
	WebsocketUpgrader = websocket.Upgrader{
		ReadBufferSize: 1024,
		WriteBufferSize: 1024,
	}

)

	type Manager struct{

	}

	func NewManager() *Manager{
		return &Manager{}
	}

	func(m *Manager) ServeWS(w http.ResponseWriter, r *http.Request){
		conn, err := WebsocketUpgrader.Upgrade(w, r, nil)
		if err != nil{
			log.Println(err)
			return
		}
		conn.Close()

	}