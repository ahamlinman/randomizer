package randomizer

import (
	"context"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

type mockStore map[string][]string

type validator func(*testing.T, Result, error)

// testCases defines, through various examples, the expected behavior of the
// randomizer.
//
// All cases require a human-readable description, an argument list to give to
// the randomizer as input, and a "check" on the output. For checks, consider
// using the isResult and isError helper functions, each of which returns an
// appropriate validator.
//
// In the tests, "randomization" is performed by sorting the input items rather
// than shuffling them. This allows for consistent assertions on output across
// test runs.
//
// The provided store will be used to build the randomizer app instance. If an
// expectedStore is defined, the store will be compared against it after the
// randomizer finishes. Nil stores return an error on every operation.
var testCases = []struct {
	description   string
	store         mockStore
	args          []string
	check         validator
	expectedStore mockStore
}{
	// Basic functionality

	{
		description: "providing no options",
		args:        []string{},
		check:       isResult(ShowedHelp),
	},

	{
		description: "randomizing a set of options",
		args:        []string{"three", "two", "one"},
		check:       isResult(Selection, "*one*", "*three*", "*two*"),
	},

	// Selecting from groups

	{
		description: "randomizing a group",
		store:       mockStore{"test": {"three", "two", "one"}},
		args:        []string{"test"},
		check:       isResult(Selection, "*one*", "*three*", "*two*"),
	},

	{
		description: "randomizing a group that does not exist",
		store:       mockStore{},
		args:        []string{"test"},
		check:       isError(`couldn't find the "test" group`),
	},

	{
		description: "error while getting a group",
		store:       nil,
		args:        []string{"test"},
		check:       isError(`had trouble getting the "test" group`),
	},

	// Group CRUD operations

	{
		description: "listing groups",
		store:       mockStore{"first": {"one"}, "second": {"two"}},
		args:        []string{"/list"},
		check:       isResult(ListedGroups, "• first", "• second"),
	},

	{
		description: "listing groups when there are none",
		store:       mockStore{},
		args:        []string{"/list"},
		check:       isResult(ListedGroups, "no groups are available"),
	},

	{
		description: "unable to list groups",
		store:       nil,
		args:        []string{"/list"},
		check:       isError("trouble getting this channel's groups"),
	},

	{
		description: "showing a group",
		store:       mockStore{"test": {"one", "two", "three"}},
		args:        []string{"/show", "test"},
		check:       isResult(ShowedGroup, "• one", "• three", "• two"),
	},

	{
		description: "showing a group that does not exist",
		store:       mockStore{},
		args:        []string{"/show", "test"},
		check:       isError("can't find that group"),
	},

	{
		description: "unable to show a group",
		store:       nil,
		args:        []string{"/show", "test"},
		check:       isError("had trouble getting that group"),
	},

	{
		description: "no group provided to show",
		store:       mockStore{},
		args:        []string{"/show"},
		check:       isError("requires an argument"),
	},

	{
		description:   "saving a group",
		store:         mockStore{},
		args:          []string{"/save", "test", "one", "two"},
		check:         isResult(SavedGroup, `The "test" group was saved`, "• one", "• two"),
		expectedStore: mockStore{"test": {"one", "two"}},
	},

	{
		description: "unable to save a group",
		store:       nil,
		args:        []string{"/save", "test", "one", "two"},
		check:       isError("trouble saving that group"),
	},

	{
		description: "saving a group with a flag name",
		store:       mockStore{},
		args:        []string{"/save", "/delete", "one", "two"},
		check:       isError("has a special meaning"),
	},

	{
		description: "saving a group with a potential flag name",
		store:       mockStore{},
		args:        []string{"/save", "/futureflag", "one", "two"},
		check:       isError("has a special meaning"),
	},

	{
		description: `saving a group named "help"`,
		store:       mockStore{},
		args:        []string{"/save", "help", "one", "two"},
		check:       isError("has a special meaning"),
	},

	{
		description: "not enough options provided to save",
		store:       mockStore{},
		args:        []string{"/save", "test", "one"},
		check:       isError("need at least two options"),
	},

	{
		description:   "deleting a group",
		store:         mockStore{"test": {"one", "two"}},
		args:          []string{"/delete", "test"},
		check:         isResult(DeletedGroup, `The "test" group was deleted`),
		expectedStore: mockStore{},
	},

	{
		description: "deleting a group that does not exist",
		store:       mockStore{},
		args:        []string{"/delete", "test"},
		check:       isError("can't find that group"),
	},

	{
		description: "unable to delete a group",
		store:       nil,
		args:        []string{"/delete", "test"},
		check:       isError("trouble deleting that group"),
	},

	{
		description: "no group provided to delete",
		store:       mockStore{},
		args:        []string{"/delete"},
		check:       isError("requires an argument"),
	},

	// Requesting help

	{
		description: "help as a flag",
		args:        []string{"/help"},
		check:       isResult(ShowedHelp),
	},

	{
		description: "help as a standalone argument",
		args:        []string{"help"},
		check:       isResult(ShowedHelp),
	},

	{
		description: "help as an option to be randomized",
		args:        []string{"help", "me"},
		check:       isResult(Selection, "*help*"),
	},
}

func TestMain(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			app := NewApp("randomizer", tc.store)
			app.shuffle = func(options []string) {
				sort.Strings(options)
			}

			res, err := app.Main(context.TODO(), tc.args)
			tc.check(t, res, err)

			if tc.expectedStore != nil && !reflect.DeepEqual(tc.store, tc.expectedStore) {
				t.Errorf("unexpected store state\ngot:  %v\nwant: %v", tc.store, tc.expectedStore)
			}
		})
	}
}

func isResult(expectedType ResultType, contains ...string) validator {
	return func(t *testing.T, res Result, err error) {
		if err != nil {
			t.Fatalf("unexpected error %v", err)
		}

		if res.Type() != expectedType {
			t.Errorf("got result type %v, want %v", res.Type(), expectedType)
		}

		// Ensure that expected substrings appear *in order* in the response
		message := res.Message()
		for _, c := range contains {
			i := strings.Index(message, c)

			if i < 0 {
				t.Errorf("result missing %q in expected position\n%v", c, res.Message())
				continue
			}

			message = message[i+len(c):]
		}
	}
}

func isError(contains string) validator {
	return func(t *testing.T, res Result, err error) {
		if err == nil {
			t.Fatalf("unexpected result %v", res)
		}

		if _, ok := err.(Error); !ok {
			t.Fatalf("unexpected error type %T", err)
		}

		rerr := err.(Error)

		if !strings.Contains(rerr.HelpText(), contains) {
			t.Errorf("error help text missing substring %q", contains)
		}
	}
}

func (ms mockStore) List(_ context.Context) ([]string, error) {
	if ms == nil {
		return nil, errors.New("mock store list error")
	}

	keys := make([]string, 0, len(ms))
	for k := range ms {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}

func (ms mockStore) Get(_ context.Context, name string) ([]string, error) {
	if ms == nil {
		return nil, errors.New("mock store get error")
	}

	return ms[name], nil
}

func (ms mockStore) Put(_ context.Context, name string, options []string) error {
	if ms == nil {
		return errors.New("mock store put error")
	}

	copied := make([]string, len(options))
	copy(copied, options)
	sort.Strings(copied)
	ms[name] = copied
	return nil
}

func (ms mockStore) Delete(_ context.Context, name string) (existed bool, err error) {
	if ms == nil {
		return false, errors.New("mock store delete error")
	}

	_, existed = ms[name]
	delete(ms, name)
	return
}
