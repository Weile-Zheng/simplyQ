package raft_fsm

import (
	"os"
	"path/filepath"
	"time"

	"github.com/Weile-Zheng/simplyQ/internal/queue_manager"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

type RaftNode struct {
	Raft *raft.Raft
}

func NewRaftNode(dataDir string, nodeID string, bindAddr string, peers []string, queueManager *queue_manager.QueueManager) (*RaftNode, error) {
	// Raft configuration
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(nodeID)

	// Log store
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft-log.db"))
	if err != nil {
		return nil, err
	}

	// Stable store
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft-stable.db"))
	if err != nil {
		return nil, err
	}

	// Snapshot store
	snapshotStore, err := raft.NewFileSnapshotStore(dataDir, 1, os.Stdout)
	if err != nil {
		return nil, err
	}

	// Transport
	transport, err := raft.NewTCPTransport(bindAddr, nil, 3, 10*time.Second, os.Stdout)
	if err != nil {
		return nil, err
	}

	fsm := &FSM{QueueManager: queueManager}

	// Raft system
	raftNode, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return nil, err
	}

	// Bootstrap cluster if necessary
	if len(peers) == 0 {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		raftNode.BootstrapCluster(configuration)
	}

	return &RaftNode{Raft: raftNode}, nil
}

// ApplyCommand applies a command to the Raft log.
func (rn *RaftNode) ApplyCommand(command []byte, timeout time.Duration) (interface{}, error) {
	future := rn.Raft.Apply(command, timeout)
	if err := future.Error(); err != nil {
		return nil, err
	}
	return future.Response(), nil
}

// IsLeader checks if the current node is the leader.
func (rn *RaftNode) IsLeader() bool {
	return rn.Raft.State() == raft.Leader
}

// GetLeader returns the current leader of the Raft cluster.
func (rn *RaftNode) GetLeader() raft.ServerAddress {
	return rn.Raft.Leader()
}
