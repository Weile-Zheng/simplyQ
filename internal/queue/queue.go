package queue

import "time"

type Message struct {
	ID   string
	Body string
}

type Queue struct {
	QueueConfig     QueueConfig
	Messages        []Message
	DeadLetterQueue []Message
}

type QueueConfig struct {
	ID                string
	Name              string
	Type              QueueType
	RetentionPeriod   time.Duration
	VisibilityTimeout time.Duration
	MaxReceiveCount   int
	MaxMessageSize    int
}

type Request struct {
	Type    OpType
	Message Message
	Result  chan Response
}

type Response struct {
	Message Message
	Code    Code
}

type QueueIO struct {
	SendChan chan<- Request
	End      chan<- any
}

// InsertQueue sends an insert request to the queue and waits for a response.
func (q *QueueIO) InsertQueue(req Request) Response {
	response := make(chan Response)
	q.SendChan <- Request{
		Type:    INSERT,
		Message: req.Message,
		Result:  response,
	}
	return <-response
}

// PeekQueue sends a peek request to the queue and waits for a response.
func (q *QueueIO) PeekQueue(req Request) Response {
	response := make(chan Response)
	q.SendChan <- Request{
		Type:    PEEK,
		Message: req.Message,
		Result:  response,
	}
	return <-response
}

// DeleteQueue sends a delete request to the queue and waits for a response.
func (q *QueueIO) DeleteQueue(req Request) Response {
	response := make(chan Response)
	q.SendChan <- Request{
		Type:    DELETE,
		Message: req.Message,
		Result:  response,
	}
	return <-response
}

// RequeueQueue move a message from the dead letter queue back to the main queue.
func (q *QueueIO) RequeueQueue(req Request) Response {
	response := make(chan Response)
	q.SendChan <- Request{
		Type:    REQUEUE,
		Message: req.Message,
		Result:  response,
	}
	return <-response
}

func (q *QueueIO) Close() {
	close(q.End)
}

func MakeQueue(config QueueConfig) QueueIO {
	send, end := make(chan Request), make(chan any)
	queueIO := QueueIO{
		SendChan: send,
		End:      end,
	}

	queue := Queue{
		QueueConfig:     config,
		Messages:        []Message{},
		DeadLetterQueue: []Message{},
	}

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
						req.Result <- Response{
							Message: queue.Messages[0],
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
						queue.Messages = queue.Messages[1:] // O(1) in Go
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
			case <-end:
				return
			}
		}
	}()
	return queueIO
}
