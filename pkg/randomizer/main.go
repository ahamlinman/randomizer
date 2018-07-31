package randomizer

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// ErrTooFewOptions is returned when, after parsing arguments, there are fewer
// than two options to choose from for randomization.
var ErrTooFewOptions = errors.New("need at least two options")

// App represents a randomizer app that can accept commands.
type App struct {
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
func NewApp(store Store) *App {
	return &App{
		store: store,
	}
}

// Main is the entrypoint to the randomizer tool.
func (a *App) Main(args []string) (string, error) {
	fs := buildFlagSet()
	err := fs.Parse(args)
	if err != nil {
		return "", err
	}

	if fs.listGroups {
		return a.listGroups()
	}

	if fs.deleteGroup != "" {
		return a.deleteGroup(fs.deleteGroup)
	}

	options, err := a.expandGroups(fs.Args())
	if err != nil {
		return "", err
	}

	if len(options) < 2 {
		return "", ErrTooFewOptions
	}

	if fs.saveGroup != "" {
		return a.saveGroup(fs.saveGroup, options)
	}

	source := rand.NewSource(time.Now().UnixNano())
	rander := rand.New(source)
	selector := Selector(rander.Intn)
	return selector.PickString(options), nil
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

func (a *App) listGroups() (string, error) {
	groups, err := a.store.List()
	if err != nil {
		return "", err
	}

	if len(groups) == 0 {
		return "No groups are saved", nil
	}

	result := bytes.NewBufferString("The following groups are saved:\n")
	for _, g := range groups {
		result.Write([]byte(fmt.Sprintf("• %s\n", g)))
	}

	return result.String()[:result.Len()-1], nil
}

func (a *App) deleteGroup(name string) (string, error) {
	if err := a.store.Delete(name); err != nil {
		return "", err
	}

	return fmt.Sprintf("Group %q was deleted", name), nil
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
			return nil, err
		}

		for _, opt := range groupOpts {
			options = append(options, opt)
		}
	}

	return options, nil
}

func (a *App) saveGroup(name string, options []string) (string, error) {
	if err := a.store.Put(name, options); err != nil {
		return "", err
	}

	return fmt.Sprintf("Saved group %q with %d options", name, len(options)), nil
}
