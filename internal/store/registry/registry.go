// Package registry facilitates environment-based store setup without requiring
// every store to be linked into a randomizer binary.
package registry

import (
	"context"
	"fmt"

	"github.com/featherbread/randomizer/internal/randomizer"
)

// Registry provides environment keys and factory constructors for the store
// backends linked into this binary.
var Registry = map[string]Entry{}

// Entry represents a single store backend.
type Entry struct {
	// EnvironmentKeys lists the names of environment variables that this store's
	// factory checks for configuration. If the environment contains any of these
	// keys, and contains no keys for other factories, the randomizer selects this
	// store as its backend.
	EnvironmentKeys []string

	// FactoryFromEnv creates a factory for this backend based on its environment
	// variables.
	FactoryFromEnv func(context.Context) (func(partition string) randomizer.Store, error)
}

// Provide registers a new store backend, or panics if a backend is already
// registered under this name.
func Provide(
	name string,
	factoryFromEnv func(context.Context) (func(partition string) randomizer.Store, error),
	environmentKeys ...string,
) {
	if _, ok := Registry[name]; ok {
		panic(fmt.Errorf("%s already registered as a store backend", name))
	}
	Registry[name] = Entry{
		EnvironmentKeys: environmentKeys,
		FactoryFromEnv:  factoryFromEnv,
	}
}
