package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/FkLalita/anonymous-chat-app/handlers"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, nil)
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/ws", handlers.HandleConnections)

	log.Println("Server is running on :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
