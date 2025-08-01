package queue_manager

// QueueManager manages multiple message queues. 

import (
	"log"
	"sync"

	"github.com/Weile-Zheng/simplyQ/internal/queue"
)

type QueueManager struct {
	Queues map[string]*queue.QueueIO
	Lock   sync.RWMutex
}

type QueueManagerConfig struct {
	Name string
}

// NewQueueManager creates a new QueueManager instance.
func NewQueueManager(config QueueManagerConfig) QueueManager {
	return QueueManager{
		Queues: make(map[string]*queue.QueueIO),
		Lock:   sync.RWMutex{},
	}
}

// CreateQueue creates a new queue with the given configuration.
func (qm *QueueManager) CreateQueue(config queue.QueueConfig) queue.Code {
	qm.Lock.Lock()
	defer qm.Lock.Unlock()
	id := config.Name

	// if queue with the same name exists, return
	if _, exists := qm.Queues[id]; exists {
		return queue.QUEUE_ALREADY_EXISTS
	}

	// Assuming queueIO has a Name field to identify the queue
	qm.Queues[id] = queue.MakeQueue(id, config)
	return queue.OK
}

// DeleteQueue removes a queue by its ID.
func (qm *QueueManager) DeleteQueue(id string) queue.Code {
	qm.Lock.Lock()
	defer qm.Lock.Unlock()

	if q, exists := qm.Queues[id]; exists {
		q.Close()
		delete(qm.Queues, id)
		return queue.OK
	}
	return queue.QUEUE_NOT_FOUND
}

// We use R lock for send, peek, delete message to allow parallel operations on different queues
// Concurrent operations within the same queues are handled next level down by the queue itself.

// SendMessage sends a message to the specified queue.
func (qm *QueueManager) SendMessage(queueID string, message queue.Message) queue.Response {
	qm.Lock.RLock()
	defer qm.Lock.RUnlock()

	if q, exists := qm.Queues[queueID]; exists {
		return q.InsertQueue(message)
	}
	return queue.Response{Code: queue.QUEUE_NOT_FOUND, Message: queue.Message{}}
}

// PeekMessage retrieves the next message from the specified queue without removing it.
func (qm *QueueManager) PeekMessage(queueID string) queue.Response {
	qm.Lock.RLock()
	defer qm.Lock.RUnlock()

	if q, exists := qm.Queues[queueID]; exists {
		return q.PeekQueue()
	}
	return queue.Response{Code: queue.QUEUE_NOT_FOUND, Message: queue.Message{}}
}

// PopMessage removes a message from the specified queue.
func (qm *QueueManager) PopMessage(queueID string) queue.Response {
	qm.Lock.RLock()
	defer qm.Lock.RUnlock()

	if q, exists := qm.Queues[queueID]; exists {
		return q.RemoveQueue()
	}
	return queue.Response{Code: queue.QUEUE_NOT_FOUND, Message: queue.Message{}}
}

// ViewAllMessages printout all messages from the specified queue (does not remove them or call peek/receive).
func (qm *QueueManager) ViewAllMessages(queueID string) []queue.Message {
	qm.Lock.RLock()
	defer qm.Lock.RUnlock()

	if q, exists := qm.Queues[queueID]; exists {
		return q.SnapshotQueue().Messages
	}
	log.Printf("Queue %s not found", queueID)
	return nil
}

// ViewAllQueues returns a snapshot of all queues and their messages.
func (qm *QueueManager) ViewAllQueues() map[string]queue.Queue {
	qm.Lock.Lock() // write lock for taking a snapshot of the all queues.
	defer qm.Lock.Unlock()

	queuesSnapshot := make(map[string]queue.Queue)
	for id, q := range qm.Queues {
		queuesSnapshot[id] = q.SnapshotQueue()
	}
	return queuesSnapshot
}

func (qm *QueueManager) RestoreAllQueues(queues map[string]queue.Queue) {
	qm.Lock.Lock()
	defer qm.Lock.Unlock()

	restored_queues := make(map[string]*queue.QueueIO)

	for id := range queues {
		restored_queues[id] = queue.MakeQueue(id, queues[id].Config)
		for _, msg := range queues[id].Messages {
			restored_queues[id].InsertQueue(msg)
		}
	}
}
