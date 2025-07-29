package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/weilezheng/simplyQ/internal/queue"
)

var queueManager queue.QueueManager

func main() {
	// Initialize the queue manager
	managerConfig := queue.QueueManagerConfig{
		Name: "SimplyQManager",
	}
	queueManager = queue.NewQueueManager(managerConfig)

	// Set up HTTP routes
	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/createQueue", createQueueHandler)
	http.HandleFunc("/sendMessage", sendMessageHandler)
	http.HandleFunc("/peekMessage", peekMessageHandler)
	http.HandleFunc("/deleteMessage", deleteMessageHandler)

	// Start the server
	port := 8080
	fmt.Printf("Starting SimplyQ server on port %d...\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

// Ping
func pingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Pong"))
}

// createQueueHandler handles the creation of a new queue.
func createQueueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var config queue.QueueConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	code := queueManager.CreateQueue(config)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": code,
	})
}

// sendMessageHandler handles sending a message to a queue.
func sendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req queue.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	queueID := r.URL.Query().Get("queueID")
	response := queueManager.SendMessage(queueID, req.Message)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// peekMessageHandler handles peeking a message from a queue.
func peekMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	queueID := r.URL.Query().Get("queueID")
	response := queueManager.PeekMessage(queueID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// deleteMessageHandler handles deleting a message from a queue.
func deleteMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req queue.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	queueID := r.URL.Query().Get("queueID")
	response := queueManager.DeleteMessage(queueID, req)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
