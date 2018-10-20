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

		if !strings.HasPrefix(rerr.HelpText(), "Whoops") {
			t.Error(`error help text missing standard prefix "Whoops"`)
		}

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
	{
		description: "choosing one of a set of options",
		args:        []string{"three", "two", "one"},
		validIf:     isResult(Selection, "*one*"),
	},

	{
		description: "providing no options",
		args:        []string{},
		validIf:     isError("need at least two options"),
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
