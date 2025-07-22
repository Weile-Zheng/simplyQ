package queue

type QueueManager struct {
	queues map[string]Queue
}

type QueueManagerConfig struct {
	Name string
}
