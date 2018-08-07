package bbolt

import (
	"bytes"
	"encoding/gob"

	bolt "github.com/coreos/bbolt"
	"github.com/pkg/errors"
)

// DefaultBucket is the default bucket in which groups are stored in the bbolt
// database.
const DefaultBucket = "groups"

// Store is a store backed by a bbolt database.
type Store struct {
	db     *bolt.DB
	bucket string
}

// Option represents a type for options that can be applied to a Store.
type Option func(*Store)

// New creates a new store backed by the provided (pre-opened) bbolt database.
func New(db *bolt.DB, options ...Option) *Store {
	store := &Store{
		db:     db,
		bucket: DefaultBucket,
	}

	for _, opt := range options {
		opt(store)
	}

	return store
}

// WithBucketName creates a Store that stores groups in the named bucket in the
// bbolt database, rather than the default "groups" bucket.
func WithBucketName(name string) Option {
	return func(b *Store) {
		b.bucket = name
	}
}

// List obtains the set of stored groups.
func (b *Store) List() (groups []string, err error) {
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
func (b *Store) Get(name string) (options []string, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(b.bucket))
		if bucket == nil {
			return errors.Errorf("group %q does not exist", name)
		}

		result := bucket.Get([]byte(name))
		if result == nil {
			return errors.Errorf("group %q does not exist", name)
		}

		decoder := gob.NewDecoder(bytes.NewReader(result))
		return errors.Wrapf(decoder.Decode(&options), "decoding group %q", name)
	})
	return
}

// Put saves the provided options into a named group.
func (b *Store) Put(name string, options []string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(b.bucket))
		if err != nil {
			return errors.Wrap(err, "creating bucket")
		}

		var result bytes.Buffer
		encoder := gob.NewEncoder(&result)
		if err := encoder.Encode(&options); err != nil {
			return errors.Wrapf(err, "encoding group %q (%v)", name, options)
		}

		return errors.Wrapf(
			bucket.Put([]byte(name), result.Bytes()),
			"writing group %q",
			name,
		)
	})
}

// Delete removes the named group from the store.
func (b *Store) Delete(name string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(b.bucket))
		if bucket == nil {
			return errors.Errorf("group %q does not exist", name)
		}

		return errors.Wrapf(bucket.Delete([]byte(name)), "deleting group %q", name)
	})
}