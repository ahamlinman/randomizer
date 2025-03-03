//go:build !randomizer.bbolt && !randomizer.dynamodb && !randomizer.firestore

package store

import (
	_ "github.com/ahamlinman/randomizer/internal/store/bbolt"
	_ "github.com/ahamlinman/randomizer/internal/store/dynamodb"
	_ "github.com/ahamlinman/randomizer/internal/store/firestore"
)

func init() {
	hasAllStoreBackends = true
	hasNonBoltStoreBackend = true
}
