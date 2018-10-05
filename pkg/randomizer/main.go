package randomizer // import "go.alexhamlin.co/randomizer/pkg/randomizer"

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"sort"
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
	// ShowedGroup indicates that the options of a single group were successfully obtained.
	ShowedGroup
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
func NewApp(name string, store Store) App {
	return App{
		name:  name,
		store: store,
	}
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

	fs := buildFlagSet()
	err = fs.Parse(args)
	if err != nil || (len(args) == 1 && args[0] == "help") {
		return Result{}, Error{
			cause:    errors.New("help requested"),
			helpText: a.buildUsage(),
		}
	}

	if fs.listGroups {
		return a.listGroups()
	}

	if fs.showGroup != "" {
		return a.showGroup(fs.showGroup)
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

	if fs.count < 1 {
		return Result{}, Error{
			cause:    errors.New("count too small"),
			helpText: "Whoops, I can't pick less than one option!",
		}
	}

	if fs.count > len(options) {
		return Result{}, Error{
			cause:    errors.New("count too large"),
			helpText: "Whoops, I can't pick more options than I was given!",
		}
	}

	source := rand.NewSource(time.Now().UnixNano())
	rand.New(source).Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})

	var choices []string
	if fs.all {
		choices = options
	} else {
		choices = options[:fs.count]
	}

	for i, choice := range choices {
		choices[i] = "*" + choice + "*"
	}

	return Result{
		resultType: Selection,
		message:    fmt.Sprintf("I choose %s!", strings.Join(choices, " and ")),
	}, nil
}

type flagSet struct {
	*flag.FlagSet

	count int
	all   bool

	listGroups  bool
	showGroup   string
	saveGroup   string
	deleteGroup string
}

func buildFlagSet() *flagSet {
	fs := &flagSet{
		FlagSet: flag.NewFlagSet("randomizer", flag.ContinueOnError),
	}
	fs.SetOutput(ioutil.Discard)

	fs.IntVar(&fs.count, "n", 1, "number of items to pick")
	fs.BoolVar(&fs.all, "all", false, "pick all items in a random order")

	fs.BoolVar(&fs.listGroups, "list", false, "list all known groups")
	fs.StringVar(&fs.showGroup, "show", "", "show the options in the specified group")
	fs.StringVar(&fs.saveGroup, "save", "", "save options into the specified group")
	fs.StringVar(&fs.deleteGroup, "delete", "", "delete the specified group")

	return fs
}

var usageTmpl = template.Must(template.New("").Parse(
	`{{.Name}} helps you pick options randomly out of a list.

*Example:* {{.Name}} one two three
> I choose *three*!

You can choose more than one option at a time. The selected options will be given back in a random order.

*Example:* {{.Name}} -n 2 one two three
> I choose *two* and *one*!

*Example:* {{.Name}} -all one two three
> I choose *two* and *three* and *one*!

You can also create *groups* for the current channel or DM.

*Save a group:* {{.Name}} -save first3 one two three
*Randomize from a group:* {{.Name}} +first3
*Combine groups with other options:* {{.Name}} -n 3 +first3 +next3 seven eight
*List groups:* {{.Name}} -list
*Show options in a group:* {{.Name}} -show first3
*Delete a group:* {{.Name}} -delete first3

Note that the selection is weighted. An option is more likely to be picked if it is given multiple times. This also applies when multiple groups are given, and an option is in more than one of them.`))

func (a App) buildUsage() string {
	var buf bytes.Buffer
	usageTmpl.Execute(&buf, struct{ Name string }{a.name})
	return buf.String()
}

func (a App) listGroups() (Result, error) {
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
			message:    "No groups are available in this channel. (Use the -save option to create one!)",
		}, nil
	}

	result := bytes.NewBufferString("The following groups are available in this channel:\n")
	a.formatList(result, groups)

	return Result{
		resultType: ListedGroups,
		message:    result.String()[:result.Len()-1],
	}, nil
}

func (a App) showGroup(name string) (Result, error) {
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

func (a App) deleteGroup(name string) (Result, error) {
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

func (a App) expandGroups(argOpts []string) ([]string, error) {
	options := make([]string, 0, len(argOpts))

	for _, opt := range argOpts {
		if !strings.HasPrefix(opt, "+") {
			options = append(options, opt)
			continue
		}

		groupName := opt[1:]
		groupOpts, err := a.store.Get(groupName)
		if err != nil {
			return nil, Error{
				cause:    err,
				helpText: fmt.Sprintf("Whoops, I couldn't find the %q group in this channel!", groupName),
			}
		}

		for _, opt := range groupOpts {
			options = append(options, opt)
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
