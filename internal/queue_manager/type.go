package queue_manager

import "github.com/Weile-Zheng/simplyQ/internal/queue"

type CommandType int

const (
	CREATE_QUEUE CommandType = iota
	DELETE_QUEUE

	SEND_MESSAGE
	PEEK_MESSAGE
	POP_MESSAGE

	VIEW_QUEUE
)

type Command struct {
	Type        CommandType       `json:"type"`
	QueueID     string            `json:"queue_id,omitempty"`
	Message     queue.Message     `json:"message,omitempty"`
	QueueConfig queue.QueueConfig `json:"queue_config,omitempty"`
}
