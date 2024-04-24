// Package bbolt supports randomizer storage in a local bbolt database file.
package bbolt

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

// Store is a store backed by a bbolt database.
type Store struct {
	db     *bolt.DB
	bucket string
}

// New creates a new store backed by the provided (pre-opened) bbolt database.
func New(db *bolt.DB, bucket string) (Store, error) {
	if db == nil {
		return Store{}, errors.New("bolt.DB instance is required")
	}
	if bucket == "" {
		return Store{}, errors.New("bucket is required")
	}

	return Store{db: db, bucket: bucket}, nil
}

// List obtains the set of stored groups.
func (b Store) List(_ context.Context) (groups []string, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(b.bucket))
		if bucket == nil {
			return nil
		}

		return bucket.ForEach(func(k, _ []byte) error {
			groups = append(groups, string(k))
			return nil
		})
	})
	return
}

// Get obtains the options in a single named group.
func (b Store) Get(_ context.Context, name string) (options []string, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(b.bucket))
		if bucket == nil {
			return nil
		}

		result := bucket.Get([]byte(name))
		if result == nil {
			return nil
		}

		decoder := gob.NewDecoder(bytes.NewReader(result))
		err := decoder.Decode(&options)
		if err != nil {
			return fmt.Errorf("decoding group %q: %w", name, err)
		}

		return nil
	})
	return
}

// Put saves the provided options into a named group.
func (b Store) Put(_ context.Context, name string, options []string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(b.bucket))
		if err != nil {
			return fmt.Errorf("creating bucket: %w", err)
		}

		var result bytes.Buffer
		encoder := gob.NewEncoder(&result)
		err = encoder.Encode(&options)
		if err != nil {
			return fmt.Errorf("encoding group %q (%v): %w", name, options, err)
		}

		err = bucket.Put([]byte(name), result.Bytes())
		if err != nil {
			return fmt.Errorf("writing group %q: %w", name, err)
		}

		return nil
	})
}

// Delete removes the named group from the store.
func (b Store) Delete(_ context.Context, name string) (existed bool, err error) {
	err = b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(b.bucket))
		if bucket == nil {
			return nil
		}
		if bucket.Get([]byte(name)) == nil {
			return nil
		}

		existed = true
		err := bucket.Delete([]byte(name))
		if err != nil {
			return fmt.Errorf("deleting group %q: %w", name, err)
		}

		return nil
	})
	return
}
