package randomizer

import (
	"context"
	"math/rand"
	"time"
)

// Store represents an object that provides persistence for "groups" of
// options.
type Store interface {
	// List returns the names of all available groups. If no groups have been
	// saved, it returns an empty list with a nil error.
	List(ctx context.Context) (groups []string, err error)

	// Get returns the list of options in the named group. If the group does not
	// exist, it returns an empty list with a nil error.
	Get(ctx context.Context, group string) (options []string, err error)

	// Put saves the provided options as a named group, overwriting any previous
	// group with that name.
	Put(ctx context.Context, group string, options []string) error

	// Delete ensures that the named group no longer exists, returning true or
	// false to indicate whether the group existed prior to this deletion
	// attempt.
	Delete(ctx context.Context, group string) (bool, error)
}

// App represents a randomizer app that can accept commands.
type App struct {
	name    string
	store   Store
	shuffle func([]string) // Overridden in tests for predictable behavior
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

var appHandlers = map[operation]appHandler{
	showHelp:      App.showHelp,
	makeSelection: App.makeSelection,
	listGroups:    App.listGroups,
	showGroup:     App.showGroup,
	saveGroup:     App.saveGroup,
	deleteGroup:   App.deleteGroup,
}
