package queue

import (
	"math/rand"
	"time"
)

func generateRandomID() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return "queue-" + string(r.Intn(1000000))
}

var DEFAULT_QUEUE_CONFIG = QueueConfig{
	Name:              generateRandomID(),
	Type:              "fifo",
	RetentionPeriod:   24 * time.Hour,
	VisibilityTimeout: 30 * time.Second,
	MaxReceiveCount:   3,
	MaxMessageSize:    1024 * 1024,
}
