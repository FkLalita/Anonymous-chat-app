package handlers

import (
	"encoding/json"
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

	requestUsername(conn)

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		if messageType == websocket.TextMessage {
			message := string(p)
			handleMessage(conn, message)
		}
	}

	clientsLock.Lock()
	delete(clients, conn)
	clientsLock.Unlock()

	broadcastUserList()
}

func requestUsername(conn *websocket.Conn) {
	message := Message{
		Type: "requestUsername",
	}

	err := conn.WriteJSON(message)
	if err != nil {
		log.Println(err)
		return
	}

	_, p, err := conn.ReadMessage()
	if err != nil {
		log.Println(err)
		return
	}

	var usernameMessage Message
	err = json.Unmarshal(p, &usernameMessage)
	if err != nil {
		log.Println("Error decoding username message:", err)
		return
	}

	if usernameMessage.Type == "username" {
		client := &User{
			Conn:     conn,
			Username: usernameMessage.Content,
		}

		clientsLock.Lock()
		clients[conn] = client
		clientsLock.Unlock()

		broadcastUserList()
	}
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
		if msg.Type == "mainChat" {
			// Send the content directly without wrapping it in another JSON object
			err := client.WriteJSON(msg.Content)
			if err != nil {
				log.Println(err)
				client.Close()
				delete(clients, client)
				userCounter--
			}
		} else {
			// For other message types, send the entire message as before
			err := client.WriteJSON(msg)
			if err != nil {
				log.Println(err)
				client.Close()
				delete(clients, client)
				userCounter--
			}
		}
	}
}

func handleMessage(conn *websocket.Conn, message string) {
	clientsLock.Lock()
	sender, ok := clients[conn]
	clientsLock.Unlock()

	if !ok {
		log.Println("Sender not found")
		return
	}

	msg := Message{
		Type:    "mainChat",
		Content: fmt.Sprintf("%s: %s", sender.Username, message),
	}

	broadcast(msg)
}
