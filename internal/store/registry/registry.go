// Package registry facilitates environment-based store setup without requiring
// every store to be linked into a randomizer binary.
package registry

import (
	"context"

	"github.com/ahamlinman/randomizer/internal/randomizer"
)

// Registry defines the environment keys for all possible store backends, and
// provides access to factory constructors for the backends linked into this
// binary.
//
// Note that the environment keys must be pre-defined in the registry, and must
// be kept in sync with the store's implementation. At the risk of introducing
// programming errors (by keeping the keys out of sync), this approach ensures
// that attempts to configure an unavailable backend result in a hard error to
// users, rather than silently defaulting to an incorrect backend.
var Registry = map[string]*Entry{
	"bbolt":     {EnvironmentKeys: []string{"DB_PATH"}},
	"dynamodb":  {EnvironmentKeys: []string{"DYNAMODB", "DYNAMODB_TABLE", "DYNAMODB_ENDPOINT"}},
	"firestore": {EnvironmentKeys: []string{"FIRESTORE_PROJECT_ID", "FIRESTORE_DATABASE_ID"}},
}

// Entry represents a single store backend.
type Entry struct {
	// EnvironmentKeys is the list of environment variables that this store's
	// factory checks for configuration. If any one of these keys is set in the
	// environment (and no conflicting keys are set), this store will be selected
	// as the backend for this randomizer instance.
	EnvironmentKeys []string

	// FactoryFromEnv creates a factory for this backend based on its environment
	// variables. It is nil if this backend is not available in this build.
	FactoryFromEnv func(context.Context) (func(partition string) randomizer.Store, error)
}
