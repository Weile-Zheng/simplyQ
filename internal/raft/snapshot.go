package raft

import (
	"encoding/json"
	"fmt"

	"github.com/Weile-Zheng/simplyQ/internal/queue"
	"github.com/hashicorp/raft"
)

type Snapshot struct {
	Queues map[string][]queue.Message
}

// Persist saves the snapshot to the provided sink.
func (s *Snapshot) Persist(sink raft.SnapshotSink) error {
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
func (s *Snapshot) Release() {
	fmt.Printf("Releasing snapshot with %d queues\n", len(s.Queues))
}
