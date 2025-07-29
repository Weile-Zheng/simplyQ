package queue

import "sync"

type QueueManager struct {
	Queues map[string]QueueIO
	Lock   sync.RWMutex
}

type QueueManagerConfig struct {
	Name string
}

// NewQueueManager creates a new QueueManager instance.
func NewQueueManager(config QueueManagerConfig) QueueManager {
	return QueueManager{
		Queues: make(map[string]QueueIO),
		Lock:   sync.RWMutex{},
	}
}

// CreateQueue creates a new queue with the given configuration.
func (qm *QueueManager) CreateQueue(config QueueConfig) Code {
	qm.Lock.Lock()
	defer qm.Lock.Unlock()
	id := "queue-" + config.Name

	// if queue with the same name exists, return
	if _, exists := qm.Queues[id]; exists {
		return QUEUE_ALREADY_EXISTS
	}

	// Assuming queueIO has a Name field to identify the queue
	qm.Queues[id] = MakeQueue(id, config)
	return OK
}

// DeleteQueue removes a queue by its ID.
func (qm *QueueManager) DeleteQueue(id string) Code {
	qm.Lock.Lock()
	defer qm.Lock.Unlock()

	if queue, exists := qm.Queues[id]; exists {
		queue.Close()
		delete(qm.Queues, id)
		return OK
	}
	return QUEUE_NOT_FOUND
}

// We use R lock for send, peek, delete message to allow concurrent operations on different queues
// Concurrent operations within the same queues are handled next level down by the queue itself.

// SendMessage sends a message to the specified queue.
func (qm *QueueManager) SendMessage(queueID string, message Message) Response {
	qm.Lock.RLock()
	defer qm.Lock.RUnlock()

	if queue, exists := qm.Queues[queueID]; exists {
		return queue.InsertQueue(message)
	}
	return Response{Code: QUEUE_NOT_FOUND, Message: Message{}}
}

// PeekMessage retrieves the next message from the specified queue without removing it.
func (qm *QueueManager) PeekMessage(queueID string) Response {
	qm.Lock.RLock()
	defer qm.Lock.RUnlock()

	if queue, exists := qm.Queues[queueID]; exists {
		return queue.PeekQueue()
	}
	return Response{Code: QUEUE_NOT_FOUND, Message: Message{}}
}

// DeleteMessage removes a message from the specified queue.
func (qm *QueueManager) DeleteMessage(queueID string, req Request) Response {
	qm.Lock.RLock()
	defer qm.Lock.RUnlock()

	if queue, exists := qm.Queues[queueID]; exists {
		return queue.RemoveQueue()
	}
	return Response{Code: QUEUE_NOT_FOUND, Message: Message{}}
}
