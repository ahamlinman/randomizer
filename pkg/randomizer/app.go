package randomizer

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/pkg/errors"
)

// Store represents an object that provides persistence for "groups" of
// options.
type Store interface {
	// List should return a list of all available groups. If no groups have been
	// saved, an empty list should be returned.
	List() (groups []string, err error)

	// Get should return the list of options in the named group. If the group
	// does not exist, an empty list should be returned.
	Get(group string) (options []string, err error)

	// Put should save the provided options as a named group, completely
	// overwriting any previous group of that name.
	Put(group string, options []string) error

	// Delete should ensure that the named group no longer exists, returning true
	// or false to indicate whether the group existed prior to this deletion
	// attempt.
	Delete(group string) (bool, error)
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
		name:  name,
		store: store,

		shuffle: func(options []string) {
			source := rand.NewSource(time.Now().UnixNano())
			rand.New(source).Shuffle(len(options), func(i, j int) {
				options[i], options[j] = options[j], options[i]
			})
		},
	}
}

type appHandler func(App, request) (Result, error)

// appHandlers maps every possible mode of operation for the randomizer to a
// handler method, which processes a request from the user and returns an
// appropriate result.
var appHandlers = map[operation]appHandler{
	makeSelection: App.makeSelection,
	showHelp:      App.showHelp,
	listGroups:    App.listGroups,
	showGroup:     App.showGroup,
	saveGroup:     App.saveGroup,
	deleteGroup:   App.deleteGroup,
}

// Main is the entrypoint to the randomizer tool.
//
// Note that all errors returned from this function will be of this package's
// Error type. This provides the HelpText method for user-friendly output
// formatting.
func (a App) Main(args []string) (result Result, err error) {
	request, err := a.newRequestFromArgs(args)
	if err != nil {
		return Result{}, err // Comes from this package, no re-wrapping needed
	}

	handler := appHandlers[request.Operation]
	if handler == nil {
		panic(errors.Errorf("invalid request: %+v", request))
	}
	return handler(a, request)
}

func (a App) listGroups(_ request) (Result, error) {
	groups, err := a.store.List()
	if err != nil {
		return Result{}, Error{
			cause:    err,
			helpText: "Whoops, I had trouble getting this channel's groups. Please try again later!",
		}
	}

	if len(groups) == 0 {
		return Result{
			resultType: ListedGroups,
			message:    "Whoops, no groups are available in this channel. (Use the /save flag to create one!)",
		}, nil
	}

	sort.Strings(groups)

	return Result{
		resultType: ListedGroups,
		message: fmt.Sprintf(
			"The following groups are available in this channel:\n%s",
			bulletize(groups),
		),
	}, nil
}

func (a App) showGroup(request request) (Result, error) {
	name := request.Operand

	group, err := a.store.Get(name)
	if err != nil {
		return Result{}, Error{
			cause:    err,
			helpText: "Whoops, I had trouble getting that group. Please try again later!",
		}
	}

	if len(group) == 0 {
		return Result{}, Error{
			cause:    errors.New("group does not exist"),
			helpText: "Whoops, I can't find that group in this channel. (Use the /save flag to create it!)",
		}
	}

	sort.Strings(group)

	return Result{
		resultType: ShowedGroup,
		message: fmt.Sprintf(
			"The %q group has the following options:\n%s",
			name,
			bulletize(group),
		),
	}, nil
}

func (a App) saveGroup(request request) (Result, error) {
	name := request.Operand
	options, err := a.expandArgs(request.Args)
	if err != nil {
		return Result{}, err
	}

	if len(options) < 2 {
		return Result{}, Error{
			cause:    errors.New("too few options to save"),
			helpText: "Whoops, I need at least two options to save a group!",
		}
	}

	if err := a.store.Put(name, options); err != nil {
		return Result{}, Error{
			cause:    err,
			helpText: "Whoops, I had trouble saving that group. Please try again later!",
		}
	}

	sort.Strings(options)

	return Result{
		resultType: SavedGroup,
		message: fmt.Sprintf(
			"Done! The %q group was saved in this channel with the following options:\n%s",
			name,
			bulletize(options),
		),
	}, nil
}

func (a App) deleteGroup(request request) (Result, error) {
	name := request.Operand
	existed, err := a.store.Delete(name)
	if err != nil {
		return Result{}, Error{
			cause:    err,
			helpText: "Whoops, I had trouble deleting that group. Please try again later!",
		}
	}

	if !existed {
		return Result{}, Error{
			cause:    errors.New("group does not exist"),
			helpText: "Whoops, I can't find that group in this channel!",
		}
	}

	return Result{
		resultType: DeletedGroup,
		message:    fmt.Sprintf("Done! The %q group was deleted.", name),
	}, nil
}
