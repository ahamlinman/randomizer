// Package rndtest provides helpers for tests that invoke a randomizer.
package rndtest

import (
	"context"
	"errors"
	"slices"

	"golang.org/x/exp/maps"
)

// Store implements randomizer.Store by mapping group names to sorted lists of
// strings. A nil Store will return errors for every operation.
type Store map[string][]string

// List implements randomizer.Store.
func (s Store) List(_ context.Context) ([]string, error) {
	if s == nil {
		return nil, errors.New("store list error")
	}
	keys := maps.Keys(s)
	slices.Sort(keys)
	return keys, nil
}

// Get implements randomizer.Store.
func (s Store) Get(_ context.Context, name string) ([]string, error) {
	if s == nil {
		return nil, errors.New("store get error")
	}
	return s[name], nil
}

// Put implements randomizer.Store.
func (s Store) Put(_ context.Context, name string, options []string) error {
	if s == nil {
		return errors.New("store put error")
	}
	copied := slices.Clone(options)
	slices.Sort(copied)
	s[name] = copied
	return nil
}

// Delete implements randomizer.Store.
func (s Store) Delete(_ context.Context, name string) (existed bool, err error) {
	if s == nil {
		return false, errors.New("store delete error")
	}
	_, existed = s[name]
	delete(s, name)
	return
}
