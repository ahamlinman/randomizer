// Package firestore supports randomizer storage in Google Cloud Firestore.
package firestore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
)

type Store struct {
	client    *firestore.Client
	partition string
}

type optionsDoc struct {
	Options []string `firestore:"options"`
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
	ref := f.client.Collection(f.partition).Doc(group)
	doc, err := ref.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting document: %w", err)
	}

	var result optionsDoc
	err = doc.DataTo(&result)
	if err != nil {
		return nil, fmt.Errorf("decoding document: %w", err)
	}

	return result.Options, nil
}

func (f Store) Put(ctx context.Context, group string, options []string) error {
	ref := f.client.Collection(f.partition).Doc(group)
	_, err := ref.Set(ctx, optionsDoc{options})
	return err
}

func (f Store) Delete(ctx context.Context, group string) (bool, error) {
	ref := f.client.Collection(f.partition).Doc(group)
	_, err := ref.Delete(ctx)
	return (err == nil), err // TODO: Support the first return value.
}
