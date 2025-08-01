package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Weile-Zheng/simplyQ/internal/queue"
	"github.com/Weile-Zheng/simplyQ/internal/queue_manager"
)

func (s *QueueServer) pingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Pong\n")
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

	command := queue_manager.Command{
		Type:        queue_manager.CREATE_QUEUE,
		QueueConfig: config,
	}

	if !s.RaftNode.IsLeader() {
		http.Error(w, fmt.Sprintf("Not the leader. Current leader: %s", s.RaftNode.GetLeader()), http.StatusTemporaryRedirect)
		return
	}

	// Apply command via Raft
	commandBytes, err := json.Marshal(command)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal command: %v", err), http.StatusInternalServerError)
		return
	}

	code, err := s.RaftNode.ApplyCommand(commandBytes, 5*time.Second)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to apply command: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"code":         code,
		"queue_config": config,
	})
}

func (s *QueueServer) sendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	queueID := r.URL.Query().Get("queueID")
	if queueID == "" {
		http.Error(w, "Missing queue ID", http.StatusBadRequest)
		return
	}

	var message queue.Message
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		http.Error(w, "Invalid message format", http.StatusBadRequest)
		return
	}

	// Create command
	command := queue_manager.Command{
		Type:    queue_manager.SEND_MESSAGE,
		QueueID: queueID,
		Message: message,
	}

	// Forward to leader if not the leader
	if !s.RaftNode.IsLeader() {
		http.Error(w, fmt.Sprintf("Not the leader. Current leader: %s", s.RaftNode.GetLeader()), http.StatusTemporaryRedirect)
		return
	}

	// Apply command via Raft
	commandBytes, err := json.Marshal(command)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal command: %v", err), http.StatusInternalServerError)
		return
	}

	code, err := s.RaftNode.ApplyCommand(commandBytes, 5*time.Second)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to apply command: %v", err), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]any{
		"code":    code,
		"message": message,
	})

}

// peekMessageHandler gets the next message from a queue without removing it
func (s *QueueServer) peekMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	queueID := r.URL.Query().Get("queueID")
	if queueID == "" {
		http.Error(w, "Missing queue ID", http.StatusBadRequest)
		return
	}

	// Create command
	command := queue_manager.Command{
		Type:    queue_manager.PEEK_MESSAGE,
		QueueID: queueID,
	}

	// Forward to leader if not the leader
	if !s.RaftNode.IsLeader() {
		http.Error(w, fmt.Sprintf("Not the leader. Current leader: %s", s.RaftNode.GetLeader()), http.StatusTemporaryRedirect)
		return
	}

	// Apply command via Raft
	commandBytes, err := json.Marshal(command)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal command: %v", err), http.StatusInternalServerError)
		return
	}

	response, err := s.RaftNode.ApplyCommand(commandBytes, 5*time.Second)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to apply command: %v", err), http.StatusInternalServerError)
		return
	}

	// Handle the response
	// If it's a queue.PeekResponse, extract the message and code
	if peekResponse, ok := response.(queue.Response); ok {
		result := map[string]interface{}{
			"code": peekResponse.Code,
		}

		// Only include message if one was found
		if peekResponse.Code == queue.OK {
			result["message"] = peekResponse.Message
		}

		json.NewEncoder(w).Encode(result)
		return
	}

	// If it's just a code, return that
	if code, ok := response.(queue.Code); ok {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": code,
		})
		return
	}

	// Unexpected response type
	http.Error(w, "Unexpected response type from queue manager", http.StatusInternalServerError)
}

// popMessageHandler removes a message from a queue
func (s *QueueServer) popMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	queueID := r.URL.Query().Get("queueID")
	if queueID == "" {
		http.Error(w, "Missing queue ID", http.StatusBadRequest)
		return
	}

	messageID := r.URL.Query().Get("messageID")
	if messageID == "" {
		http.Error(w, "Missing message ID", http.StatusBadRequest)
		return
	}

	// Create command
	command := queue_manager.Command{
		Type:    queue_manager.POP_MESSAGE,
		QueueID: queueID,
	}

	// Forward to leader if not the leader
	if !s.RaftNode.IsLeader() {
		http.Error(w, fmt.Sprintf("Not the leader. Current leader: %s", s.RaftNode.GetLeader()), http.StatusTemporaryRedirect)
		return
	}

	// Apply command via Raft
	commandBytes, err := json.Marshal(command)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal command: %v", err), http.StatusInternalServerError)
		return
	}

	response, err := s.RaftNode.ApplyCommand(commandBytes, 5*time.Second)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to apply command: %v", err), http.StatusInternalServerError)
		return
	}

	// Return the response code
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": response,
	})
}

func (s *QueueServer) viewQueueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	queueID := r.URL.Query().Get("queueID")
	if queueID == "" {
		http.Error(w, "Missing queue ID", http.StatusBadRequest)
		return
	}

	command := queue_manager.Command{
		Type:    queue_manager.VIEW_QUEUE,
		QueueID: queueID,
	}

	// Forward to leader if not the leader
	if !s.RaftNode.IsLeader() {
		http.Error(w, fmt.Sprintf("Not the leader. Current leader: %s", s.RaftNode.GetLeader()), http.StatusTemporaryRedirect)
		return
	}

	// Apply command via Raft
	commandBytes, err := json.Marshal(command)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal command: %v", err), http.StatusInternalServerError)
		return
	}

	code, err := s.RaftNode.ApplyCommand(commandBytes, 5*time.Second)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to apply command: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"code": code,
	})
}
