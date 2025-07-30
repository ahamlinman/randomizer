//go:build !randomizer.bbolt && !randomizer.dynamodb && !randomizer.firestore

package store

import (
	_ "github.com/featherbread/randomizer/internal/store/bbolt"
	_ "github.com/featherbread/randomizer/internal/store/dynamodb"
	_ "github.com/featherbread/randomizer/internal/store/firestore"
)

func init() {
	haveAllStoreBackends = true
}
