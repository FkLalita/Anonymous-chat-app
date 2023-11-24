// websocket.go

package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Message represents a WebSocket message
type Message struct {
	Type    string   `json:"type"`
	Content string   `json:"content"`
	Users   []string `json:"users"`
}

// User struct to store user information
type User struct {
	Conn     *websocket.Conn
	Username string
}

var (
	clients     = make(map[*websocket.Conn]*User)
	clientsLock sync.Mutex
	userCounter int
)

// HandleConnections handles WebSocket connections
func HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	// Prompt user for a username
	err = conn.WriteMessage(websocket.TextMessage, []byte("Please enter your username"))
	if err != nil {
		log.Println(err)
		return
	}

	// Receive username from client
	_, username, err := conn.ReadMessage()
	if err != nil {
		log.Println(err)
		return
	}

	client := &User{
		Conn:     conn,
		Username: string(username),
	}

	clientsLock.Lock()
	clients[conn] = client
	userCounter++
	clientsLock.Unlock()

	broadcastUserList()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		if messageType == websocket.TextMessage {
			message := string(p)
			handleMessage(client, message)
		}
	}

	clientsLock.Lock()
	delete(clients, conn)
	userCounter--
	clientsLock.Unlock()

	broadcastUserList()
}

func broadcastUserList() {
	userList := []string{}

	clientsLock.Lock()
	for _, client := range clients {
		userList = append(userList, client.Username)
	}
	clientsLock.Unlock()

	message := Message{
		Type:  "userList",
		Users: userList,
	}

	broadcast(message)
}

func broadcast(msg Message) {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	for client := range clients {
		err := client.WriteJSON(msg)
		if err != nil {
			log.Println(err)
			client.Close()
			delete(clients, client)
			userCounter--
		}
	}
}

func handleMessage(sender *User, message string) {
	msg := Message{
		Type:    "mainChat",
		Content: fmt.Sprintf("%s: %s", sender.Username, message),
	}

	broadcast(msg)
}
