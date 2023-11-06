package firestore

import (
	"context"
	"errors"
	"os"

	"cloud.google.com/go/firestore"
	"go.alexhamlin.co/randomizer/internal/randomizer"
)

// FactoryFromEnv returns a store.Factory whose stores are backed by a Google
// Cloud Firestore database.
func FactoryFromEnv() (func(string) randomizer.Store, error) {
	projectID, ok := os.LookupEnv("FIRESTORE_PROJECT_ID")
	if !ok {
		return nil, errors.New("missing FIRESTORE_PROJECT_ID in environment")
	}

	databaseID, ok := os.LookupEnv("FIRESTORE_DATABASE_ID")
	if !ok {
		return nil, errors.New("missing FIRESTORE_DATABASE_ID in environment")
	}

	return func(partition string) randomizer.Store {
		client, err := firestore.NewClientWithDatabase(context.TODO(), projectID, databaseID)
		if err != nil {
			panic(err)
		}
		return New(client, partition)
	}, nil
}
