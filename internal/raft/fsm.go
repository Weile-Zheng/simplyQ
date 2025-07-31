package raft

import (
	"encoding/json"
	"io"

	"github.com/Weile-Zheng/simplyQ/internal/queue"
	"github.com/Weile-Zheng/simplyQ/internal/queue_manager"
	"github.com/hashicorp/raft"
)

type FSM struct {
	QueueManager *queue_manager.QueueManager
}

func (f *FSM) Apply(log *raft.Log) interface{} {
	var command queue_manager.Command
	if err := json.Unmarshal(log.Data, &command); err != nil {
		return err
	}

	switch command.Type {
	case queue_manager.CREATE_QUEUE:
		return f.QueueManager.CreateQueue(command.QueueConfig)
	case queue_manager.DELETE_QUEUE:
		return f.QueueManager.DeleteQueue(command.QueueID)
	case queue_manager.SEND_MESSAGE:
		return f.QueueManager.SendMessage(command.QueueID, command.Message)
	case queue_manager.PEEK_MESSAGE:
		return f.QueueManager.PeekMessage(command.QueueID)
	case queue_manager.POP_MESSAGE:
		return f.QueueManager.PopMessage(command.QueueID)
	case queue_manager.VIEW_QUEUE:
		return f.QueueManager.ViewAllMessages(command.QueueID)
	}
	return nil
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	snapshot := Snapshot{f.QueueManager.ViewAllQueues()}
	return &snapshot, nil
}

func (f *FSM) Restore(rc io.ReadCloser) error {
	var queues map[string]queue.QueueIO
	if err := json.NewDecoder(rc).Decode(&queues); err != nil {
		return err
	}
	f.QueueManager.Queues = queues
	return nil
}
