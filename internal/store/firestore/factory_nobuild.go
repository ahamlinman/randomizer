//go:build !randomizer.firestore

package firestore

import (
	"errors"

	"go.alexhamlin.co/randomizer/internal/randomizer"
)

// FactoryFromEnv always fails, as this build of the randomizer does not include
// Google Cloud Firestore support.
func FactoryFromEnv() (func(string) randomizer.Store, error) {
	return nil, errors.ErrUnsupported
}
