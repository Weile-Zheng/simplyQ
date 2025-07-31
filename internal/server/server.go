package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Weile-Zheng/simplyQ/internal/queue_manager"
)

type QueueServer struct {
	QueueManager queue_manager.QueueManager
}

func StartNewServer() {
	managerConfig := queue_manager.QueueManagerConfig{
		Name: "SimplyQManager",
	}

	server := QueueServer{
		QueueManager: queue_manager.NewQueueManager(managerConfig),
	}

	http.HandleFunc("/ping", server.pingHandler)
	http.HandleFunc("/createQueue", server.createQueueHandler)
	http.HandleFunc("/sendMessage", server.sendMessageHandler)
	http.HandleFunc("/peekMessage", server.peekMessageHandler)
	http.HandleFunc("/deleteMessage", server.deleteMessageHandler)
	http.HandleFunc("/viewAllMessages", server.viewQueueHandler)

	port := 8080
	log.Printf("Starting SimplyQ server on port %d...\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
