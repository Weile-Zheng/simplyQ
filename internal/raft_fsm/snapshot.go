package raft_fsm

import (
	"encoding/json"

	"github.com/Weile-Zheng/simplyQ/internal/queue"
	"github.com/hashicorp/raft"
)

type RaftSnapshot struct {
	Queues map[string]queue.Queue
}

func (s *RaftSnapshot) Persist(sink raft.SnapshotSink) error {
	data, err := json.Marshal(s.Queues)
	if err != nil {
		sink.Cancel()
		return err
	}

	if _, err := sink.Write(data); err != nil {
		sink.Cancel()
		return err
	}

	return sink.Close()
}

// Release releases any resources associated with the snapshot.
func (s *RaftSnapshot) Release() {
	// No additional resources allocated during persist
}
