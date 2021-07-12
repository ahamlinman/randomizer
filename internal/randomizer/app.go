package randomizer

import (
	"context"
	"math/rand"
	"time"
)

// Store represents an object that provides persistence for "groups" of
// options.
type Store interface {
	// List should return a list of all available groups. If no groups have been
	// saved, an empty list should be returned.
	List(ctx context.Context) (groups []string, err error)

	// Get should return the list of options in the named group. If the group
	// does not exist, an empty list should be returned.
	Get(ctx context.Context, group string) (options []string, err error)

	// Put should save the provided options as a named group, completely
	// overwriting any previous group of that name.
	Put(ctx context.Context, group string, options []string) error

	// Delete should ensure that the named group no longer exists, returning true
	// or false to indicate whether the group existed prior to this deletion
	// attempt.
	Delete(ctx context.Context, group string) (bool, error)
}

// App represents a randomizer app that can accept commands.
type App struct {
	name  string
	store Store

	// May be overridden in unit tests for more predictable behavior
	shuffle func([]string)
}

// NewApp returns an App.
func NewApp(name string, store Store) App {
	return App{
		name:    name,
		store:   store,
		shuffle: shuffle,
	}
}

func shuffle(options []string) {
	source := rand.NewSource(time.Now().UnixNano())
	rand.New(source).Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})
}

// Main is the entrypoint to the randomizer tool.
//
// Note that all errors returned from this function will be of this package's
// Error type. This provides the HelpText method for user-friendly output
// formatting.
func (a App) Main(ctx context.Context, args []string) (Result, error) {
	request, err := a.newRequest(ctx, args)
	if err != nil {
		return Result{}, err
	}

	handler := appHandlers[request.Operation]
	return handler(a, request)
}

type appHandler func(App, request) (Result, error)

// appHandlers maps every possible mode of operation for the randomizer to a
// handler method, which processes a request from the user and returns an
// appropriate result.
var appHandlers = map[operation]appHandler{
	// selection.go
	makeSelection: App.makeSelection,

	// help.go
	showHelp: App.showHelp,

	// groups.go
	listGroups:  App.listGroups,
	showGroup:   App.showGroup,
	saveGroup:   App.saveGroup,
	deleteGroup: App.deleteGroup,
}
