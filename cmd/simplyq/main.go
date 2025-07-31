package main

import (
	"log"
	"os"
	"strings"

	"github.com/Weile-Zheng/simplyQ/internal/server"
)

func main() {
	dataDir := os.Getenv("DATA_DIR")

	if dataDir == "" {
		dataDir = "./data"
	}

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		log.Fatal("NODE_ID environment variable must be set")
	}

	bindAddr := os.Getenv("BIND_ADDR")
	if bindAddr == "" {
		bindAddr = "127.0.0.1"
	}

	raftPort := os.Getenv("RAFT_PORT")
	if raftPort == "" {
		raftPort = "10000"
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	// Parse peers from environment variable
	var peers []string
	peersEnv := os.Getenv("PEERS")
	if peersEnv != "" {
		peers = strings.Split(peersEnv, ",")
	}

	server.StartNewServer(dataDir, nodeID, bindAddr, raftPort, httpPort, peers)
}
