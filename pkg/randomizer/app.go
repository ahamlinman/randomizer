package randomizer

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Store represents an object that provides persistence for "groups" of
// options.
type Store interface {
	List() (groups []string, err error)

	Get(group string) (options []string, err error)

	Put(group string, options []string) error

	Delete(group string) error
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

var appHandlers = map[operation]appHandler{
	showHelp:    App.showHelp,
	listGroups:  App.listGroups,
	showGroup:   App.showGroup,
	deleteGroup: App.deleteGroup,
}

// Main is the entrypoint to the randomizer tool.
//
// Note that all errors returned from this function will be of this package's
// Error type. This provides the HelpText method for user-friendly output
// formatting.
func (a App) Main(args []string) (result Result, err error) {
	defer func() {
		if err != nil {
			if _, ok := err.(Error); !ok {
				// Just in case…
				err = Error{cause: err}
			}
		}
	}()

	request, err := newRequestFromArgs(args)
	if err != nil {
		return Result{}, err // Comes from this package, no re-wrapping needed
	}

	if handler := appHandlers[request.Operation]; handler != nil {
		return handler(a, request)
	}

	options, err := a.expandOptions(request.Args)
	if err != nil {
		return Result{}, err
	}

	if len(options) < 2 {
		return Result{}, Error{
			cause:    errors.New("too few options"),
			helpText: "Whoops, I need at least two options to work with!",
		}
	}

	if request.Operation == saveGroup {
		return a.saveGroup(request.GroupName, options)
	}

	a.shuffle(options)

	var choices []string
	switch {
	case request.All:
		choices = options

	case request.Count < 1:
		return Result{}, Error{
			cause:    errors.New("count too small"),
			helpText: "Whoops, I can't pick less than one option!",
		}

	case request.Count > len(options):
		return Result{}, Error{
			cause:    errors.New("count too large"),
			helpText: "Whoops, I can't pick more options than I was given!",
		}

	default:
		choices = options[:request.Count]
	}

	for i, choice := range choices {
		choices[i] = "*" + choice + "*"
	}

	return Result{
		resultType: Selection,
		message:    fmt.Sprintf("I choose %s!", strings.Join(choices, " and ")), // TODO: More "proper" formatting
	}, nil
}

func (a App) showHelp(_ request) (Result, error) {
	return Result{
		resultType: ShowedHelp,
		message:    buildHelpMessage(a.name),
	}, nil
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
			message:    "No groups are available in this channel. (Use the /save flag to create one!)",
		}, nil
	}

	result := bytes.NewBufferString("The following groups are available in this channel:\n")
	a.formatList(result, groups)

	return Result{
		resultType: ListedGroups,
		message:    result.String()[:result.Len()-1],
	}, nil
}

func (a App) showGroup(request request) (Result, error) {
	name := request.GroupName

	group, err := a.store.Get(name)
	if err != nil {
		return Result{}, Error{
			cause:    err,
			helpText: "Whoops, I couldn't find that group in this channel!",
		}
	}

	result := bytes.NewBufferString(fmt.Sprintf("The %q group has the following options:\n", name))
	a.formatList(result, group)

	return Result{
		resultType: ShowedGroup,
		message:    result.String()[:result.Len()-1],
	}, nil
}

func (a App) deleteGroup(request request) (Result, error) {
	name := request.GroupName

	if err := a.store.Delete(name); err != nil {
		return Result{}, Error{
			cause:    err,
			helpText: "Whoops, I had trouble deleting that group. Please try again later!",
		}
	}

	return Result{
		resultType: DeletedGroup,
		message:    fmt.Sprintf("Done! The %q group was deleted.", name),
	}, nil
}

var optModifierRegexp = regexp.MustCompile("^[+-]")

func (a App) expandOptions(argOpts []string) ([]string, error) {
	options := make([]string, 0, len(argOpts))

argsLoop:
	for _, opt := range argOpts {
		switch optModifierRegexp.FindString(opt) {
		case "":
			// No modifier; simply add this as a possible option
			options = append(options, opt)

		case "+":
			// Modifier for a group name; add all elements from the group to the set
			// of options
			groupName := opt[1:]
			groupOpts, err := a.store.Get(groupName)
			if err != nil {
				return nil, Error{
					cause:    err,
					helpText: fmt.Sprintf("Whoops, I couldn't find the %q group in this channel!", groupName),
				}
			}
			options = append(options, groupOpts...)

		case "-":
			// Modifier for a removal; remove the first instance of this item from
			// the option set
			removeItem := opt[1:]
			for i, opt := range options {
				if opt == removeItem {
					options = append(options[:i], options[i+1:]...)
					continue argsLoop
				}
			}

			return nil, Error{
				cause:    errors.Errorf("option %q not found for removal", removeItem),
				helpText: fmt.Sprintf("Whoops, %q wasn't available for me to remove!", removeItem),
			}
		}
	}

	return options, nil
}

func (a App) saveGroup(name string, options []string) (Result, error) {
	if err := a.store.Put(name, options); err != nil {
		return Result{}, Error{
			cause:    err,
			helpText: "Whoops, I had trouble saving that group. Please try again later!",
		}
	}

	result := bytes.NewBufferString(fmt.Sprintf("Done! The %q group was saved in this channel with the following options:\n", name))
	a.formatList(result, options)

	return Result{
		resultType: SavedGroup,
		message:    result.String()[:result.Len()-1],
	}, nil
}

func (App) formatList(w io.Writer, items []string) {
	sorted := make([]string, len(items))
	copy(sorted, items)
	sort.Strings(sorted)

	for _, g := range sorted {
		w.Write([]byte(fmt.Sprintf("• %s\n", g)))
	}
}
