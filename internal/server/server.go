package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Weile-Zheng/simplyQ/internal/queue_manager"
	raftnode "github.com/Weile-Zheng/simplyQ/internal/raft_fsm"
)

type QueueServer struct {
	RaftNode *raftnode.RaftNode
}

var managerConfig = queue_manager.QueueManagerConfig{
	Name: "SimplyQManager",
}

func StartNewServer(dataDir, nodeID, bindAddr string, raftPort string, httpPort string, peers []string) {
	queueManager := queue_manager.NewQueueManager(managerConfig)

	raftAddr := fmt.Sprintf("%s:%s", bindAddr, raftPort)

	raftNode, err := raftnode.NewRaftNode(dataDir, nodeID, raftAddr, peers, &queueManager)

	if err != nil {
		log.Fatalf("Failed to initialize RaftNode: %v", err)
	}

	server := QueueServer{
		RaftNode: raftNode,
	}

	mux := http.NewServeMux()
	registerSystemRoutes(mux, &server)
	registerQueueRoutes(mux, &server)

	log.Printf("Starting SimplyQ server on port %s with Raft on %s...\n", httpPort, raftAddr)
	log.Fatal(http.ListenAndServe(bindAddr+":"+httpPort, mux))
}

func registerSystemRoutes(mux *http.ServeMux, server *QueueServer) {
	mux.HandleFunc("/ping", server.pingHandler)
	mux.HandleFunc("/raft/status", server.raftStatusHandler)
	mux.Handle("/raft/join", server.LeaderRedirectMiddleWare(http.HandlerFunc(server.raftJoinHandler)))
}

func registerQueueRoutes(mux *http.ServeMux, server *QueueServer) {
	mux.Handle("/createQueue", server.LeaderRedirectMiddleWare(http.HandlerFunc(server.createQueueHandler)))
	mux.Handle("/sendMessage", server.LeaderRedirectMiddleWare(http.HandlerFunc(server.sendMessageHandler)))
	mux.Handle("/peekMessage", server.LeaderRedirectMiddleWare(http.HandlerFunc(server.peekMessageHandler)))
	mux.Handle("/popMessage", server.LeaderRedirectMiddleWare(http.HandlerFunc(server.popMessageHandler)))
	mux.Handle("/viewAllMessages", server.LeaderRedirectMiddleWare(http.HandlerFunc(server.viewQueueHandler)))
}
