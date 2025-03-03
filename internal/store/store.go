// Package store and its sub-packages provide real-world [randomizer.Store]
// implementations.
package store

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"

	"github.com/ahamlinman/randomizer/internal/randomizer"
	"github.com/ahamlinman/randomizer/internal/store/registry"

	_ "github.com/ahamlinman/randomizer/internal/store/bbolt"
	_ "github.com/ahamlinman/randomizer/internal/store/dynamodb"
	_ "github.com/ahamlinman/randomizer/internal/store/firestore"
)

// Factory represents a type for functions that produce a store for the
// randomizer to use for a given "partition" (e.g. Slack channel). Factories
// may panic if a non-empty partition is required and not given.
//
// Factory is provided for documentation purposes. Do not import the store
// package just to use this alias; this will link support for all possible
// store backends into the final program, even if this was not intended.
type Factory = func(partition string) randomizer.Store

// FactoryFromEnv constructs and returns a [Factory] based on both runtime
// environment variables and build tags.
//
// Each store backend defines a set of environment variables for configuration.
// On startup, the randomizer selects one store backend based on the presence
// of its environment variables, or defaults to the bbolt backend if this build
// includes it. FactoryFromEnv fails if it cannot select a single backend based
// on these rules.
func FactoryFromEnv(ctx context.Context) (Factory, error) {
	if len(registry.Registry) == 0 {
		return nil, errors.New("no store backends available in this build")
	}

	chosen := make(map[string]registry.Entry)
	for name, entry := range registry.Registry {
		if envHasAny(entry.EnvironmentKeys...) {
			chosen[name] = entry
		}
	}
	if len(chosen) == 0 {
		if bbolt, ok := registry.Registry["bbolt"]; ok {
			return bbolt.FactoryFromEnv(ctx)
		} else {
			available := slices.Collect(maps.Keys(registry.Registry))
			return nil, fmt.Errorf(
				"can't find environment settings to select between store backends: %v", available)
		}
	}
	if len(chosen) == 1 {
		for _, entry := range chosen {
			return entry.FactoryFromEnv(ctx)
		}
	}

	available := slices.Collect(maps.Keys(chosen))
	return nil, fmt.Errorf("environment settings match multiple store backends: %v", available)
}

func envHasAny(names ...string) bool {
	for _, name := range names {
		if _, ok := os.LookupEnv(name); ok {
			return true
		}
	}
	return false
}
