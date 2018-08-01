package randomizer

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"
)

// ResultType represents the type of successful result returned by the
// randomizer.
type ResultType int

const (
	// Selection indicates that the randomizer made a random selection from input
	// options.
	Selection ResultType = iota
	// ListedGroups indicates that a group list was successfully obtained.
	ListedGroups
	// SavedGroup indicates that a group was successfully saved.
	SavedGroup
	// DeletedGroup indicates that a group was successfully deleted.
	DeletedGroup
)

// Result represents a successful randomizer operation.
type Result struct {
	resultType ResultType
	message    string
}

// Type returns the type of this result.
func (r Result) Type() ResultType {
	return r.resultType
}

// Message returns the user-friendly output associated with this result.
func (r Result) Message() string {
	return r.message
}

// Error represents an error encountered by the randomizer. It includes
// friendly help messages that can be displayed directly to users when errors
// occur, along with an underlying developer-friendly error that may be useful
// for debugging.
type Error struct {
	cause    error
	helpText string
}

func (e Error) Error() string {
	return e.cause.Error()
}

// Cause returns the underlying developer-friendly error that represents this
// usage error.
func (e Error) Cause() error {
	return e.cause
}

// HelpText returns user-friendly help text associated with this error. While
// the underlying error is more suitable for developer use, the help text may
// be displayed directly to a user.
func (e Error) HelpText() string {
	if e.helpText != "" {
		return e.helpText
	}

	return fmt.Sprintf("Whoops, I had a problem… %v", e.cause)
}

// App represents a randomizer app that can accept commands.
type App struct {
	name  string
	store Store
}

// Store represents an object that provides persistence for "groups" of
// options.
type Store interface {
	List() ([]string, error)
	Get(name string) ([]string, error)
	Put(name string, options []string) error
	Delete(name string) error
}

// NewApp returns an App.
func NewApp(name string, store Store) *App {
	return &App{
		name:  name,
		store: store,
	}
}

// Main is the entrypoint to the randomizer tool.
//
// Note that all errors returned from this function will be of this package's
// Error type. This provides the HelpText method for user-friendly output
// formatting.
func (a *App) Main(args []string) (result Result, err error) {
	defer func() {
		if err != nil {
			if _, ok := err.(Error); !ok {
				// Just in case…
				err = Error{cause: err}
			}
		}
	}()

	fs := buildFlagSet()
	err = fs.Parse(args)
	if err != nil {
		return Result{}, Error{
			cause:    err,
			helpText: a.buildUsage(),
		}
	}

	if fs.listGroups {
		return a.listGroups()
	}

	if fs.deleteGroup != "" {
		return a.deleteGroup(fs.deleteGroup)
	}

	options, err := a.expandGroups(fs.Args())
	if err != nil {
		return Result{}, err
	}

	if len(options) < 2 {
		return Result{}, Error{
			cause:    errors.New("too few options"),
			helpText: "Whoops, I need at least two options to work with!",
		}
	}

	if fs.saveGroup != "" {
		return a.saveGroup(fs.saveGroup, options)
	}

	source := rand.NewSource(time.Now().UnixNano())
	rander := rand.New(source)
	selector := Selector(rander.Intn)
	return Result{
		resultType: Selection,
		message:    fmt.Sprintf("I choose… *%s*", selector.PickString(options)),
	}, nil
}

type flagSet struct {
	*flag.FlagSet

	listGroups  bool
	saveGroup   string
	deleteGroup string
}

func buildFlagSet() *flagSet {
	fs := &flagSet{
		FlagSet: flag.NewFlagSet("randomizer", flag.ContinueOnError),
	}
	fs.SetOutput(ioutil.Discard)

	fs.BoolVar(&fs.listGroups, "list", false, "list all known groups")
	fs.StringVar(&fs.saveGroup, "save", "", "save options into the specified group")
	fs.StringVar(&fs.deleteGroup, "delete", "", "delete the specified group")

	return fs
}

var usageTmpl = template.Must(template.New("").Parse(
	`{{.Name}} is a simple command that picks a random option from a list.

*Example:* {{.Name}} one two three
> I choose… *three*

You can also save *groups* for later use.

*Saving a group:* {{.Name}} -save first3 one two three
*Randomizing from a group:* {{.Name}} +first3
*Combining groups with other options:* {{.Name}} +first3 +next3 seven eight
*Listing groups:* {{.Name}} -list
*Deleting a group:* {{.Name}} -delete first3
`))

func (a *App) buildUsage() string {
	var buf bytes.Buffer
	usageTmpl.Execute(&buf, struct{ Name string }{a.name})
	return buf.String()
}

func (a *App) listGroups() (Result, error) {
	groups, err := a.store.List()
	if err != nil {
		return Result{}, Error{
			cause:    err,
			helpText: "Whoops, I had trouble getting your groups. Please try again later!",
		}
	}

	if len(groups) == 0 {
		return Result{
			resultType: ListedGroups,
			message:    "No groups are available. (Use the -save option to create one!)",
		}, nil
	}

	result := bytes.NewBufferString("The following groups are available:\n")
	for _, g := range groups {
		result.Write([]byte(fmt.Sprintf("• %s\n", g)))
	}

	return Result{
		resultType: ListedGroups,
		message:    result.String()[:result.Len()-1],
	}, nil
}

func (a *App) deleteGroup(name string) (Result, error) {
	if err := a.store.Delete(name); err != nil {
		return Result{}, Error{
			cause:    err,
			helpText: "Whoops, I had trouble deleting that group. Please try again later!",
		}
	}

	return Result{
		resultType: DeletedGroup,
		message:    fmt.Sprintf("Done! Group %q was deleted.", name),
	}, nil
}

func (a *App) expandGroups(argOpts []string) ([]string, error) {
	options := make([]string, 0, len(argOpts))

	for _, opt := range argOpts {
		if !strings.HasPrefix(opt, "+") {
			options = append(options, opt)
			continue
		}

		groupOpts, err := a.store.Get(opt[1:])
		if err != nil {
			return nil, Error{
				cause:    err,
				helpText: fmt.Sprintf("Whoops, I couldn't find the %q group!", opt),
			}
		}

		for _, opt := range groupOpts {
			options = append(options, opt)
		}
	}

	return options, nil
}

func (a *App) saveGroup(name string, options []string) (Result, error) {
	if err := a.store.Put(name, options); err != nil {
		return Result{}, Error{
			cause:    err,
			helpText: "Whoops, I had trouble saving that group. Please try again later!",
		}
	}

	return Result{
		resultType: SavedGroup,
		message:    fmt.Sprintf("Done! Group %q was saved with %d options.", name, len(options)),
	}, nil
}
