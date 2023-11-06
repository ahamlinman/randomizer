// Package firestore supports randomizer storage in Google Cloud Firestore.
package firestore

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
)

type Store struct {
	client    *firestore.Client
	partition string
}

func New(client *firestore.Client, partition string) Store {
	return Store{client, partition}
}

func (f Store) List(ctx context.Context) ([]string, error) {
	refs := f.client.Collection(f.partition).DocumentRefs(ctx)
	allRefs, err := refs.GetAll()
	if err != nil {
		return nil, fmt.Errorf("listing collection: %w", err)
	}

	result := make([]string, len(allRefs))
	for i, ref := range allRefs {
		result[i] = ref.ID
	}
	return result, nil
}

func (f Store) Get(ctx context.Context, group string) ([]string, error) {
	return nil, errors.ErrUnsupported
}

func (f Store) Put(ctx context.Context, group string, options []string) error {
	return errors.ErrUnsupported
}

func (f Store) Delete(ctx context.Context, group string) (bool, error) {
	return false, errors.ErrUnsupported
}
