// Package firestore supports randomizer storage in Google Cloud Firestore.
package firestore

import (
	"context"
	"errors"

	"cloud.google.com/go/firestore"
)

type Store struct {
	client *firestore.Client
}

func New(client *firestore.Client) Store {
	return Store{client}
}

func (f Store) List(ctx context.Context) ([]string, error) {
	return nil, errors.ErrUnsupported
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
