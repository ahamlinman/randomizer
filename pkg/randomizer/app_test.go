package randomizer

import (
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
		check:       isError("need at least two options"),
	},

	{
		description: "choosing one of a set of options",
		args:        []string{"three", "two", "one"},
		check:       isResult(Selection, "*one*"),
	},

	// Selecting from groups

	{
		description: "choosing one option from a group",
		store:       mockStore{"test": {"three", "two", "one"}},
		args:        []string{"+test"},
		check:       isResult(Selection, "*one*"),
	},

	{
		description: "combining groups with literal options",
		store:       mockStore{"test": {"three", "two", "one"}},
		args:        []string{"+test", "four"},
		check:       isResult(Selection, "*four*"),
	},

	{
		description: "combining multiple groups",
		store: mockStore{
			"first":  {"one", "two", "three"},
			"second": {"four", "five", "six"},
		},
		args:  []string{"+first", "+second"},
		check: isResult(Selection, "*five*"),
	},

	{
		description: "choosing from a group that does not exist",
		store:       mockStore{},
		args:        []string{"+test"},
		check:       isError(`couldn't find the "test" group`),
	},

	{
		description: "error while getting a group",
		store:       nil,
		args:        []string{"+test"},
		check:       isError(`had trouble getting the "test" group`),
	},

	{
		description: "removing an option from consideration",
		store:       mockStore{"test": {"three", "two", "one"}},
		args:        []string{"+test", "-one"},
		check:       isResult(Selection, "*three*"),
	},

	{
		description: "removing an option that does not exist",
		store:       mockStore{"test": {"three", "two", "one"}},
		args:        []string{"+test", "-four"},
		check:       isError(`"four" wasn't available for me to remove`),
	},

	// Multiple selections

	{
		description: "choosing multiple options (prefixed flag)",
		args:        []string{"/n", "2", "one", "two", "three", "four"},
		check:       isResult(Selection, "*four*", "*one*"),
	},

	{
		description: "choosing multiple options (postfixed flag)",
		args:        []string{"one", "two", "three", "four", "/n", "2"},
		check:       isResult(Selection, "*four*", "*one*"),
	},

	{
		description: "choosing all options",
		args:        []string{"/n", "all", "one", "two", "three", "four"},
		check:       isResult(Selection, "*four*", "*one*", "*three*", "*two*"),
	},

	{
		description: "choosing too few options",
		args:        []string{"/n", "0", "one", "two"},
		check:       isError("can't pick less than one option"),
	},

	{
		description: "choosing too many options",
		args:        []string{"/n", "3", "one", "two"},
		check:       isError("can't pick more options than I was given"),
	},

	{
		description: "non-integer options count",
		args:        []string{"/n", "2.1", "one", "two"},
		check:       isError("isn't a valid count"),
	},

	{
		description: "invalid options count",
		args:        []string{"/n", "wat", "one", "two"},
		check:       isError("isn't a valid count"),
	},

	{
		description: "no options count provided",
		args:        []string{"one", "two", "/n"},
		check:       isError("requires an argument"),
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
		description: "saving a group with expanded arguments",
		store:       mockStore{"test": {"one", "two"}},
		args:        []string{"/save", "new", "+test", "three"},
		check:       isResult(SavedGroup, `The "new" group was saved`, "• one", "• three", "• two"),
		expectedStore: mockStore{
			"test": {"one", "two"},
			"new":  {"one", "two", "three"},
		},
	},

	{
		description: "unable to save a group",
		store:       nil,
		args:        []string{"/save", "test", "one", "two"},
		check:       isError("trouble saving that group"),
	},

	{
		description: "unable to expand arguments when saving a group",
		store:       mockStore{},
		args:        []string{"/save", "new", "+test", "three"},
		check:       isError(`couldn't find the "test" group`),
	},

	{
		description: "no options provided to save",
		store:       mockStore{},
		args:        []string{"/save", "test"},
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
		description: "help in a category",
		args:        []string{"/help", "groups"},
		check:       isResult(ShowedHelp, "lets you save *groups*"),
	},

	{
		description: "help as an option to be randomized",
		args:        []string{"help", "me"},
		check:       isResult(Selection, "*help*"),
	},

	// Invalid operations

	{
		description: "using an invalid operation flag",
		args:        []string{"/svae", "group", "one", "two"},
		check:       isError("isn't a valid flag"),
	},
}

func TestMain(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			app := NewApp("randomizer", tc.store)
			app.shuffle = func(options []string) {
				sort.Strings(options)
			}

			res, err := app.Main(tc.args)
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

func (ms mockStore) List() ([]string, error) {
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

func (ms mockStore) Get(name string) ([]string, error) {
	if ms == nil {
		return nil, errors.New("mock store get error")
	}

	return ms[name], nil
}

func (ms mockStore) Put(name string, options []string) error {
	if ms == nil {
		return errors.New("mock store put error")
	}

	copied := make([]string, len(options))
	copy(copied, options)
	ms[name] = copied
	return nil
}

func (ms mockStore) Delete(name string) (existed bool, err error) {
	if ms == nil {
		return false, errors.New("mock store delete error")
	}

	_, existed = ms[name]
	delete(ms, name)
	return
}
