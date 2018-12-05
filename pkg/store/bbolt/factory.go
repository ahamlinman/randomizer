package bbolt

import (
	"os"

	bolt "go.etcd.io/bbolt"

	"go.alexhamlin.co/randomizer/pkg/randomizer"
)

// FactoryFromEnv returns a function that can return individual randomizer
// stores backed by a local Bolt database. See the "store" package
// documentation for details.
func FactoryFromEnv() (func(string) randomizer.Store, error) {
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
