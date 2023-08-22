package main

import (
	"fmt"
	"context"
	"net/http"
)


func main() {
	handleEndpoints()

	fmt.Println("https server listening on port 8080")
	http.ListenAndServeTLS("localhost:8080", "server.crt", "server.key", nil)
}

func handleEndpoints() {
	manager := NewManager(context.Background())

	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/ws", manager.upgradeConnection)
	http.HandleFunc("/login", manager.loginHandler)
}