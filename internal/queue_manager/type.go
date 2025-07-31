package queue_manager

type CommandType int

const (
	CREATE_QUEUE CommandType = iota
	DELETE_QUEUE

	SEND_MESSAGE
	PEEK_MESSAGE
	POP_MESSAGE
)

type Command struct {
	Type    CommandType `json:"type"`
	QueueID string      `json:"queue_id,omitempty"`
	Message string      `json:"message,omitempty"`
}
