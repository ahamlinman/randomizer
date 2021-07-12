package store

import (
	"context"
	"os"

	"go.alexhamlin.co/randomizer/internal/randomizer"
	"go.alexhamlin.co/randomizer/internal/store/bbolt"
	"go.alexhamlin.co/randomizer/internal/store/dynamodb"
)

// Factory represents a type for functions that produce a store for the
// randomizer to use for a given "partition" (e.g. Slack channel).
//
// A Factory may panic if it requires a non-empty partition and no partition is
// given.
type Factory func(partition string) randomizer.Store

// FactoryFromEnv constructs and returns a Factory based on available
// environment variables. It delegates to more specific store implementations
// as appropriate.
func FactoryFromEnv(ctx context.Context) (func(string) randomizer.Store, error) {
	if envHasAny("DYNAMODB", "DYNAMODB_TABLE", "DYNAMODB_ENDPOINT") {
		return dynamodb.FactoryFromEnv(ctx)
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
