package server

import (
	"encoding/json"
	"net/http"

	"github.com/Weile-Zheng/simplyQ/internal/queue"
)

// Ping
func (s *QueueServer) pingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Pong"))
}

// createQueueHandler handles the creation of a new queue.
func (s *QueueServer) createQueueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var config queue.QueueConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	code := s.QueueManager.CreateQueue(config)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": code,
	})
}

// sendMessageHandler handles sending a message to a queue.
func (s *QueueServer) sendMessageHandler(w http.ResponseWriter, r *http.Request) {
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
	response := s.QueueManager.SendMessage(queueID, req.Message)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// peekMessageHandler handles peeking a message from a queue.
func (s *QueueServer) peekMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	queueID := r.URL.Query().Get("queueID")
	response := s.QueueManager.PeekMessage(queueID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// deleteMessageHandler handles deleting a message from a queue.
func (s *QueueServer) deleteMessageHandler(w http.ResponseWriter, r *http.Request) {
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
	response := s.QueueManager.PopMessage(queueID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *QueueServer) viewQueueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	queueID := r.URL.Query().Get("queueID")
	response := s.QueueManager.ViewAllMessages(queueID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
