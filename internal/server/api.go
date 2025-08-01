package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Weile-Zheng/simplyQ/internal/queue"
	"github.com/Weile-Zheng/simplyQ/internal/queue_manager"
	"github.com/hashicorp/raft"
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

// sendMessageHandler handles sending a message to a queue.
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

	command := queue_manager.Command{
		Type:    queue_manager.SEND_MESSAGE,
		QueueID: queueID,
		Message: message,
	}

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

	command := queue_manager.Command{
		Type:    queue_manager.PEEK_MESSAGE,
		QueueID: queueID,
	}

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

// popMessageHandler removes the first(current) message from a queue
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

	command := queue_manager.Command{
		Type:    queue_manager.POP_MESSAGE,
		QueueID: queueID,
	}

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
	json.NewEncoder(w).Encode(map[string]any{
		"code": response,
	})
}

// viewQueueHandler retrieves all messages in a queue with no side effects
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

// raftStatusHandler provides information about the Raft cluster status
func (s *QueueServer) raftStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	isLeader := s.RaftNode.IsLeader()
	leader := string(s.RaftNode.GetLeader())

	fmt.Fprintf(w, "Raft Status:\n")
	fmt.Fprintf(w, "  Is Leader: %v\n", isLeader)
	fmt.Fprintf(w, "  Current Leader: %s\n", leader)
}

type joinRequest struct {
	ID      string `json:"id"`
	Address string `json:"address"`
}

// raftJoinHandler allows a new node to join the Raft cluster
func (s *QueueServer) raftJoinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req joinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.ID == "" || req.Address == "" {
		http.Error(w, "Missing node ID or address", http.StatusBadRequest)
		return
	}

	future := s.RaftNode.Raft.AddVoter(raft.ServerID(req.ID), raft.ServerAddress(req.Address), 0, 0)
	if err := future.Error(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to add voter: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Node %s at %s successfully joined the cluster", req.ID, req.Address)
}
