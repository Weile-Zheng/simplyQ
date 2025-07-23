package queue

import (
	"encoding/json"

	"github.com/weilezheng/simplyQ/internal/storage"
	bolt "go.etcd.io/bbolt"
)

// Write QueueConfig to Bolt
func WriteConfigToBolt(config QueueConfig) error {
	db, err := storage.NewBoltDB("queue_config.db")
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("QueueConfig"))
		if err != nil {
			return err
		}

		data, err := json.Marshal(config)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(config.Name), data)
	})
}

// Write QueueManagerConfig to Bolt
func WriteManagerConfigToBolt(config QueueManagerConfig) error {
	db, err := storage.NewBoltDB("queue_config.db")
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("QueueConfig"))
		if err != nil {
			return err
		}

		data, err := json.Marshal(config)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(config.Name), data)
	})
}
