// Package registry facilitates environment-based store setup without requiring
// every store to be linked into a randomizer binary.
package registry

import (
	"context"

	"github.com/ahamlinman/randomizer/internal/randomizer"
)

// Registry provides environment keys and factory constructors for the store
// backends linked into this binary.
var Registry = map[string]Entry{}

// Entry represents a single store backend.
type Entry struct {
	// EnvironmentKeys is the list of environment variables that this store's
	// factory checks for configuration. If any one of these keys is set in the
	// environment (and no conflicting keys are set), this store will be selected
	// as the backend for this randomizer instance.
	EnvironmentKeys []string

	// FactoryFromEnv creates a factory for this backend based on its environment
	// variables.
	FactoryFromEnv func(context.Context) (func(partition string) randomizer.Store, error)
}
