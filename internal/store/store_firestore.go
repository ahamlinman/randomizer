//go:build randomizer.firestore

package store

import _ "github.com/ahamlinman/randomizer/internal/store/firestore"

func init() {
	hasNonBoltStoreBackend = true
}
