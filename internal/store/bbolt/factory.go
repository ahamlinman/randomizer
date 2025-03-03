package bbolt

import (
	"context"
	"os"

	bolt "go.etcd.io/bbolt"

	"github.com/ahamlinman/randomizer/internal/randomizer"
	"github.com/ahamlinman/randomizer/internal/store/registry"
)

func init() {
	registry.Registry["bbolt"].FactoryFromEnv = FactoryFromEnv
}

// FactoryFromEnv returns a store.Factory whose stores are backed by a local
// Bolt database (using the CoreOS "bbolt" fork).
func FactoryFromEnv(_ context.Context) (func(string) randomizer.Store, error) {
	path := pathFromEnv()

	db, err := bolt.Open(path, os.ModePerm&0644, nil)
	if err != nil {
		return nil, err
	}

	return func(partition string) randomizer.Store {
		store, err := New(db, partition)
		if err != nil {
			panic(err)
		}
		return store
	}, nil
}

func pathFromEnv() string {
	if path := os.Getenv("DB_PATH"); path != "" {
		return path
	}
	return "randomizer.db"
}
