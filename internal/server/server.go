package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Weile-Zheng/simplyQ/internal/queue_manager"
	raftnode "github.com/Weile-Zheng/simplyQ/internal/raft"
	"github.com/hashicorp/raft"
)

type QueueServer struct {
	RaftNode *raftnode.RaftNode
}

func StartNewServer(dataDir, nodeID, bindAddr string, raftPort string, httpPort string, peers []string) {
	// Initialize QueueManager
	managerConfig := queue_manager.QueueManagerConfig{
		Name: "SimplyQManager",
	}
	queueManager := queue_manager.NewQueueManager(managerConfig)

	// Initialize RaftNode
	raftAddr := fmt.Sprintf("%s:%s", bindAddr, raftPort)
	raftNode, err := raftnode.NewRaftNode(dataDir, nodeID, raftAddr, peers, &queueManager)
	if err != nil {
		log.Fatalf("Failed to initialize RaftNode: %v", err)
	}

	server := QueueServer{
		RaftNode: raftNode,
	}

	http.HandleFunc("/ping", server.pingHandler)
	http.HandleFunc("/createQueue", server.createQueueHandler)
	http.HandleFunc("/sendMessage", server.sendMessageHandler)
	// http.HandleFunc("/peekMessage", server.peekMessageHandler)
	// http.HandleFunc("/deleteMessage", server.deleteMessageHandler)
	http.HandleFunc("/viewAllMessages", server.viewQueueHandler)
	http.HandleFunc("/raft/status", server.raftStatusHandler)
	http.HandleFunc("/raft/join", server.raftJoinHandler)

	log.Printf("Starting SimplyQ server on port %s with Raft on %s...\n", httpPort, raftAddr)
	log.Fatal(http.ListenAndServe(bindAddr+":"+httpPort, nil))
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

// raftJoinHandler handles requests to join the Raft cluster
func (s *QueueServer) raftJoinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !s.RaftNode.IsLeader() {
		http.Error(w, "Not the leader", http.StatusBadRequest)
		return
	}

	nodeID := r.URL.Query().Get("id")
	address := r.URL.Query().Get("address")

	if nodeID == "" || address == "" {
		http.Error(w, "Missing node ID or address", http.StatusBadRequest)
		return
	}

	// Add the node to the Raft configuration
	future := s.RaftNode.Raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(address), 0, 0)
	if err := future.Error(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to add voter: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Node %s at %s successfully joined the cluster", nodeID, address)
}
