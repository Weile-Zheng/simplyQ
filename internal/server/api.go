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

	json.NewEncoder(w).Encode(map[string]interface{}{
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
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    code,
		"message": message,
	})

}

// func (s *QueueServer) peekMessageHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodGet {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	queueID := r.URL.Query().Get("queue")
// 	if queueID == "" {
// 		http.Error(w, "Missing queue ID", http.StatusBadRequest)
// 		return
// 	}

// 	// Create command
// 	command := queue_manager.Command{
// 		Type:    queue_manager.PEEK_MESSAGE,
// 		QueueID: queueID,
// 	}

// 	// Forward to leader if not the leader
// 	if !s.RaftNode.IsLeader() {
// 		http.Error(w, fmt.Sprintf("Not the leader. Current leader: %s", s.RaftNode.GetLeader()), http.StatusTemporaryRedirect)
// 		return
// 	}

// 	// Apply command via Raft
// 	commandBytes, err := json.Marshal(command)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Failed to marshal command: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	result := s.RaftNode.ApplyCommand(commandBytes, 5*time.Second)
// 	if result != nil {
// 		http.Error(w, fmt.Sprintf("Failed to apply command: %v", result), http.StatusInternalServerError)
// 		return
// 	}

// 	// Directly get the result since PEEK doesn't modify state
// 	response := s.QueueManager.PeekMessage(queueID)

// 	if response.Code != queue.OK {
// 		http.Error(w, fmt.Sprintf("Failed to peek message: %v", response.Message), http.StatusInternalServerError)
// 		return
// 	}

// 	json.NewEncoder(w).Encode(response.Message)
// }

// func (s *QueueServer) deleteMessageHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodDelete {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	queueID := r.URL.Query().Get("queue")
// 	if queueID == "" {
// 		http.Error(w, "Missing queue ID", http.StatusBadRequest)
// 		return
// 	}

// 	// Create command
// 	command := queue_manager.Command{
// 		Type:    queue_manager.POP_MESSAGE,
// 		QueueID: queueID,
// 	}

// 	// Forward to leader if not the leader
// 	if !s.RaftNode.IsLeader() {
// 		http.Error(w, fmt.Sprintf("Not the leader. Current leader: %s", s.RaftNode.GetLeader()), http.StatusTemporaryRedirect)
// 		return
// 	}

// 	// Apply command via Raft
// 	commandBytes, err := json.Marshal(command)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Failed to marshal command: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	err = s.RaftNode.ApplyCommand(commandBytes, 5*time.Second)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Failed to apply command: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	fmt.Fprintf(w, "Message deleted successfully\n")
// }

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

	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": code,
	})
}
