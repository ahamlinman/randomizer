package main

import (
	"bytes"
	"encoding/gob"

	bolt "github.com/coreos/bbolt"
	"github.com/pkg/errors"
)

const groupsBucket = "groups"

type boltStore struct {
	db *bolt.DB
}

func (b *boltStore) List() (groups []string, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(groupsBucket))
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

func (b *boltStore) Get(name string) (options []string, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(groupsBucket))
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

func (b *boltStore) Put(name string, options []string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(groupsBucket))
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

func (b *boltStore) Delete(name string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(groupsBucket))
		if bucket == nil {
			return errors.Errorf("group %q does not exist", name)
		}

		return errors.Wrapf(bucket.Delete([]byte(name)), "deleting group %q", name)
	})
}
