package queue

/**
 * Queue package provides a simple implementation of thread safe message queue.

 * 
 */

import (
	"time"
)

type Message struct {
	ID        string
	Body      string
	TimeStamp time.Time
}

type Queue struct {
	ID              string
	Config          QueueConfig
	Messages        []Message
	DeadLetterQueue []Message
}

type QueueConfig struct {
	Name              string
	Type              QueueType
	RetentionPeriod   time.Duration
	VisibilityTimeout time.Duration
	MaxReceiveCount   uint16
	MaxMessageSize    uint32
}

type Request struct {
	Type    opType
	Message Message
	Result  chan Response
}

type Response struct {
	Message Message
	Code    Code
}

type QueueIO struct {
	SendChan chan<- Request
	Snapshot <-chan Queue
	End      chan<- any
}

// InsertQueue sends an insert request to the queue and waits for a response.
func (q *QueueIO) InsertQueue(message Message) Response {
	response := make(chan Response)
	q.SendChan <- Request{
		Type:    INSERT,
		Message: message,
		Result:  response,
	}
	return <-response
}

// PeekQueue sends a peek request to the queue and waits for a response.
func (q *QueueIO) PeekQueue() Response {
	response := make(chan Response)
	q.SendChan <- Request{
		Type:   PEEK,
		Result: response,
	}
	return <-response
}

// RemoveQueue sends a delete request to the queue and waits for a response.
func (q *QueueIO) RemoveQueue() Response {
	response := make(chan Response)
	q.SendChan <- Request{
		Type:   DELETE,
		Result: response,
	}
	return <-response
}

// Requeue move a message from the dead letter queue back to the main queue.
func (q *QueueIO) Requeue() Response {
	response := make(chan Response)
	q.SendChan <- Request{
		Type:   REQUEUE,
		Result: response,
	}
	return <-response
}

// SnapshotQueue sends a print request to the queue and returns a list of messages.
func (q *QueueIO) SnapshotQueue() Queue {
	return <-q.Snapshot
}

func (q *QueueIO) Close() {
	close(q.End)
}

func MakeQueue(id string, config QueueConfig) *QueueIO {
	send, snapshot, end := make(chan Request), make(chan Queue), make(chan any)
	queueIO := QueueIO{
		SendChan: send,
		Snapshot: snapshot,
		End:      end,
	}

	queue := Queue{
		Config:          config,
		ID:              id,
		Messages:        []Message{},
		DeadLetterQueue: []Message{},
	}

	queueHeadReceiveCount := uint16(0) // the number of times the head message has been received.

	go func() {
		for {
			select {
			case req := <-send:
				switch req.Type {
				case INSERT:
					queue.Messages = append(queue.Messages, req.Message)
					req.Result <- Response{
						Message: req.Message,
						Code:    OK,
					}
				case PEEK:
					if len(queue.Messages) > 0 {
						queueHeadReceiveCount++
						message := queue.Messages[0]
						if queueHeadReceiveCount == queue.Config.MaxReceiveCount {
							// Move the message to the dead letter queue
							queue.DeadLetterQueue = append(queue.DeadLetterQueue, queue.Messages[0])
							queue.Messages = queue.Messages[1:]
							queueHeadReceiveCount = 0 // Reset the receive count after moving to dead letter queue
						}
						req.Result <- Response{
							Message: message,
							Code:    OK,
						}
					} else {
						req.Result <- Response{
							Message: Message{},
							Code:    EMPTY_QUEUE,
						}
					}
				case DELETE:
					if len(queue.Messages) > 0 {
						queue.Messages = queue.Messages[1:]
						queueHeadReceiveCount = 0 // Reset the receive count after deletion
						req.Result <- Response{
							Message: Message{},
							Code:    OK,
						}
					} else {
						req.Result <- Response{
							Message: Message{},
							Code:    EMPTY_QUEUE,
						}
					}
				case REQUEUE:
					if len(queue.DeadLetterQueue) > 0 {
						message := queue.DeadLetterQueue[0]
						queue.DeadLetterQueue = queue.DeadLetterQueue[1:]
						queue.Messages = append(queue.Messages, message)
						req.Result <- Response{
							Message: message,
							Code:    OK,
						}
					} else {
						req.Result <- Response{
							Message: Message{},
							Code:    EMPTY_DEAD_LETTER_QUEUE,
						}
					}
				}
			case snapshot <- Queue{
				ID:              queue.ID,
				Config:          queue.Config,
				Messages:        queue.Messages,
				DeadLetterQueue: queue.DeadLetterQueue}:
			case <-end:
				return
			}
		}
	}()
	return &queueIO
}
