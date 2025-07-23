package storage

import (
	bolt "go.etcd.io/bbolt"
)

// Create a new BoltDB instance
func NewBoltDB(path string) (*bolt.DB, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	return db, nil
}

