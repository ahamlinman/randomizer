package randomizer

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

type mockStore map[string][]string

func (ms mockStore) List() ([]string, error) {
	keys := make([]string, 0, len(ms))
	for k := range ms {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}

func (ms mockStore) Get(name string) ([]string, error) {
	options, ok := ms[name]
	if !ok {
		return nil, errors.Errorf("group %q not found", name)
	}
	return options, nil
}

func (ms mockStore) Put(name string, options []string) error {
	ms[name] = options
	return nil
}

func (ms mockStore) Delete(name string) error {
	delete(ms, name)
	return nil
}

type validator func(*testing.T, Result, error)

func isResult(expectedType ResultType, contains ...string) validator {
	return func(t *testing.T, res Result, err error) {
		if err != nil {
			t.Fatalf("unexpected error %v", err)
		}

		if res.Type() != expectedType {
			t.Errorf("got result type %v, want %v", res.Type(), expectedType)
		}

		for _, c := range contains {
			if !strings.Contains(res.Message(), c) {
				t.Errorf("result missing substring %q", c)
			}
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

var testCases = []struct {
	description   string
	store         mockStore
	args          []string
	validIf       validator
	expectedStore mockStore
}{
	// Basic functionality

	{
		description: "providing no options",
		args:        []string{},
		validIf:     isError("need at least two options"),
	},

	{
		description: "choosing one of a set of options",
		args:        []string{"three", "two", "one"},
		validIf:     isResult(Selection, "*one*"),
	},

	// Selecting from groups

	{
		description: "choosing one option from a group",
		store:       mockStore{"test": {"three", "two", "one"}},
		args:        []string{"+test"},
		validIf:     isResult(Selection, "*one*"),
	},

	{
		description: "combining groups with literal options",
		store:       mockStore{"test": {"three", "two", "one"}},
		args:        []string{"+test", "four"},
		validIf:     isResult(Selection, "*four*"),
	},

	{
		description: "combining multiple groups",
		store: mockStore{
			"first":  {"one", "two", "three"},
			"second": {"four", "five", "six"},
		},
		args:    []string{"+first", "+second"},
		validIf: isResult(Selection, "*five*"),
	},

	{
		description: "choosing from a group that does not exist",
		args:        []string{"+test"},
		validIf:     isError(`couldn't find the "test" group`),
	},

	{
		description: "removing an option from consideration",
		store:       mockStore{"test": {"three", "two", "one"}},
		args:        []string{"+test", "-one"},
		validIf:     isResult(Selection, "*three*"),
	},

	{
		description: "removing an option that does not exist",
		store:       mockStore{"test": {"three", "two", "one"}},
		args:        []string{"+test", "-four"},
		validIf:     isError(`"four" wasn't available for me to remove`),
	},

	// Multiple selections

	{
		description: "choosing multiple options",
		args:        []string{"-n", "2", "one", "two", "three", "four"},
		validIf:     isResult(Selection, "*four*", "*one*"),
	},

	{
		description: "choosing all options",
		args:        []string{"-n", "all", "one", "two", "three", "four"},
		validIf:     isResult(Selection, "*four*", "*one*", "*three*", "*two*"),
	},

	{
		description: "choosing too few options",
		args:        []string{"-n", "0", "one", "two"},
		validIf:     isError("can't pick less than one option"),
	},

	{
		description: "choosing too many options",
		args:        []string{"-n", "3", "one", "two"},
		validIf:     isError("can't pick more options than I was given"),
	},

	{
		description: "non-integer options count",
		args:        []string{"-n", "2.1", "one", "two"},
		validIf:     isError("helps you pick options randomly"), // usage message
	},

	{
		description: "invalid options count",
		args:        []string{"-n", "wat", "one", "two"},
		validIf:     isError("helps you pick options randomly"), // usage message
	},

	// Group CRUD operations

	{
		description: "listing groups",
		store:       mockStore{"first": {"one"}, "second": {"two"}},
		args:        []string{"-list"},
		validIf:     isResult(ListedGroups, "• first", "• second"),
	},

	{
		description: "listing groups when there are none",
		args:        []string{"-list"},
		validIf:     isResult(ListedGroups, "No groups are available"),
	},

	{
		description: "showing a group",
		store:       mockStore{"test": {"one", "two", "three"}},
		args:        []string{"-show", "test"},
		validIf:     isResult(ShowedGroup, "• one", "• two", "• three"),
	},

	{
		description: "showing a group that does not exist",
		args:        []string{"-show", "test"},
		validIf:     isError("couldn't find that group"),
	},

	{
		description:   "saving a group",
		store:         mockStore{},
		args:          []string{"-save", "test", "one", "two"},
		validIf:       isResult(SavedGroup, `The "test" group was saved`, "• one", "• two"),
		expectedStore: mockStore{"test": {"one", "two"}},
	},

	{
		description:   "deleting a group",
		store:         mockStore{"test": {"one", "two"}},
		args:          []string{"-delete", "test"},
		validIf:       isResult(DeletedGroup, `The "test" group was deleted`),
		expectedStore: mockStore{},
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
			tc.validIf(t, res, err)

			if tc.expectedStore != nil && !reflect.DeepEqual(tc.store, tc.expectedStore) {
				t.Errorf("unexpected store state\n  got %v\n  want %v", tc.store, tc.expectedStore)
			}
		})
	}
}
