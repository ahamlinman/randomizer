// Package store and its sub-packages provide real-world [randomizer.Store]
// implementations.
package store

import (
	"context"
	"os"

	"github.com/ahamlinman/randomizer/internal/randomizer"
	"github.com/ahamlinman/randomizer/internal/store/bbolt"
	"github.com/ahamlinman/randomizer/internal/store/dynamodb"
	"github.com/ahamlinman/randomizer/internal/store/firestore"
)

// Factory represents a type for functions that produce a store for the
// randomizer to use for a given "partition" (e.g. Slack channel).
//
// A Factory may panic if it requires a non-empty partition and no partition is
// given.
type Factory func(partition string) randomizer.Store

// FactoryFromEnv constructs and returns a Factory based on available
// environment variables. If a known DynamoDB environment variable is set, it
// will return a DynamoDB store. Otherwise, it will return a bbolt store.
func FactoryFromEnv(ctx context.Context) (func(string) randomizer.Store, error) {
	if envHasAny("DYNAMODB", "DYNAMODB_TABLE", "DYNAMODB_ENDPOINT") {
		return dynamodb.FactoryFromEnv(ctx)
	}
	if envHasAny("FIRESTORE_PROJECT_ID", "FIRESTORE_DATABASE_ID") {
		return firestore.FactoryFromEnv()
	}
	return bbolt.FactoryFromEnv()
}

func envHasAny(names ...string) bool {
	for _, name := range names {
		if _, ok := os.LookupEnv(name); ok {
			return true
		}
	}
	return false
}
