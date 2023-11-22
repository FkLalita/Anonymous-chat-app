package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/FkLalita/anonymous-chat-app/handlers"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Render HTML template
	tmpl, err := template.New("index").Parse(`
	<html>
		<head>
			<title>WebSocket Chat</title>
		</head>
		<body>
			<h1>WebSocket Chat</h1>
			<input type="text" id="messageInput" placeholder="Type a message...">
			<button onclick="sendMessage()">Send</button>
			<div id="messageArea"></div>
			
			<script>
				const socket = new WebSocket("ws://localhost:8080/ws");
				
				socket.addEventListener("message", event => {
					document.getElementById("messageArea").innerHTML += "<p>" + event.data + "</p>";
				});
				
				function sendMessage() {
					const messageInput = document.getElementById("messageInput");
					const message = messageInput.value;
					socket.send(message);
					messageInput.value = "";
				}
			</script>
		</body>
	</html>
	`)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, nil)
}

func main() {
	http.HandleFunc("/ws", handlers.WebsocketHandler)
	http.HandleFunc("/", homeHandler)

	log.Println("Server is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
