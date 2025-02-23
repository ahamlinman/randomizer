package randomizer

import (
	"context"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/ahamlinman/randomizer/internal/randomizer/rndtest"
)

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
	store         rndtest.Store
	args          []string
	check         validator
	expectedStore rndtest.Store
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
		store:       rndtest.Store{"test": {"three", "two", "one"}},
		args:        []string{"test"},
		check:       isResult(Selection, "*one*", "*three*", "*two*"),
	},

	{
		description: "randomizing a group that does not exist",
		store:       rndtest.Store{},
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
		store:       rndtest.Store{"first": {"one"}, "second": {"two"}},
		args:        []string{"/list"},
		check:       isResult(ListedGroups, "• first", "• second"),
	},

	{
		description: "listing groups when there are none",
		store:       rndtest.Store{},
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
		store:       rndtest.Store{"test": {"one", "two", "three"}},
		args:        []string{"/show", "test"},
		check:       isResult(ShowedGroup, "• one", "• three", "• two"),
	},

	{
		description: "showing a group that does not exist",
		store:       rndtest.Store{},
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
		store:       rndtest.Store{},
		args:        []string{"/show"},
		check:       isError("requires an argument"),
	},

	{
		description:   "saving a group",
		store:         rndtest.Store{},
		args:          []string{"/save", "test", "one", "two"},
		check:         isResult(SavedGroup, `The "test" group was saved`, "• one", "• two"),
		expectedStore: rndtest.Store{"test": {"one", "two"}},
	},

	{
		description: "unable to save a group",
		store:       nil,
		args:        []string{"/save", "test", "one", "two"},
		check:       isError("trouble saving that group"),
	},

	{
		description: "saving a group with a flag name",
		store:       rndtest.Store{},
		args:        []string{"/save", "/delete", "one", "two"},
		check:       isError("has a special meaning"),
	},

	{
		description: "saving a group with a potential flag name",
		store:       rndtest.Store{},
		args:        []string{"/save", "/futureflag", "one", "two"},
		check:       isError("has a special meaning"),
	},

	{
		description: `saving a group named "help"`,
		store:       rndtest.Store{},
		args:        []string{"/save", "help", "one", "two"},
		check:       isError("has a special meaning"),
	},

	{
		description: "no group name provided to save",
		store:       rndtest.Store{},
		args:        []string{"/save"},
		check:       isError("requires an argument"),
	},

	{
		description: "no options provided to save",
		store:       rndtest.Store{},
		args:        []string{"/save", "test"},
		check:       isError("need at least two options"),
	},

	{
		description: "only one option provided to save",
		store:       rndtest.Store{},
		args:        []string{"/save", "test", "one"},
		check:       isError("need at least two options"),
	},

	{
		description:   "deleting a group",
		store:         rndtest.Store{"test": {"one", "two"}},
		args:          []string{"/delete", "test"},
		check:         isResult(DeletedGroup, `The "test" group was deleted`),
		expectedStore: rndtest.Store{},
	},

	{
		description: "deleting a group that does not exist",
		store:       rndtest.Store{},
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
		store:       rndtest.Store{},
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
		check:       isResult(Selection, "*help*", "*me*"),
	},
}

func TestMain(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			store := tc.store.Clone()
			app := NewApp("randomizer", store)
			app.shuffle = slices.Sort

			res, err := app.Main(context.Background(), tc.args)
			tc.check(t, res, err)

			if tc.expectedStore != nil && !reflect.DeepEqual(store, tc.expectedStore) {
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
