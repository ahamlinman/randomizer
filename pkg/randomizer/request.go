package randomizer

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

type operation int

const (
	selection operation = iota
	showHelp
	listGroups
	showGroup
	saveGroup
	deleteGroup
)

// request represents a single user request to a randomizer instance, created
// from raw user input.
type request struct {
	Operation operation
	GroupName string

	Args  []string
	All   bool
	Count int
}

// flagHandler is a type for functions that can parse a flag and its value(s)
// from an argument list into a request struct.
//
// The argument slice provided to the handler starts at the argument containing
// the flag itself. If the returned error is nil, the returned int is the total
// number of arguments (1 or more) consumed by parsing this flag and its
// value(s).
type flagHandler func(*request, []string) (int, error)

// operationFlagHandlers represents "operation" flags, which must appear at the
// start of the argument list. Operations are alternate modes of behavior that
// do not involve randomly selecting from lists of items.
var operationFlagHandlers = map[string]flagHandler{
	// As a special case, show the help message if "help" is the only argument
	// provided by the user (in case they don't yet know the flag syntax)
	"help":  (*request).parseHelp,
	"/help": (*request).parseHelp,

	"/list":   (*request).parseList,
	"/show":   (*request).parseShow,
	"/save":   (*request).parseSave,
	"/delete": (*request).parseDelete,
}

// modifierFlagHandlers represents "modifier" flags, which may appear anywhere
// in the argument list. Modifiers affect the behavior of an operation,
// particularly the normal operation of selecting randomly from a list.
var modifierFlagHandlers = map[string]flagHandler{
	"/n": (*request).parseN,
}

func newRequestFromArgs(args []string) (request, error) {
	opts := request{
		Count: 1,
	}

	if len(args) < 1 {
		return opts, nil
	}

	consumeFlag := func(handler flagHandler) error {
		consumed, err := handler(&opts, args)
		if err != nil {
			return err
		}

		args = args[consumed:]
		return nil
	}

	// Consume an operation flag, if we have one.
	if handler := operationFlagHandlers[args[0]]; handler != nil {
		if err := consumeFlag(handler); err != nil {
			return opts, err
		}
	}

	// Process all remaining arguments, consuming modifiers as they appear.
	for len(args) > 0 {
		if handler := modifierFlagHandlers[args[0]]; handler != nil {
			if err := consumeFlag(handler); err != nil {
				return opts, err
			}
			continue
		}

		var nonFlags []string
		nonFlags, args = splitArgsAtNextModifier(args)
		opts.Args = append(opts.Args, nonFlags...)
	}

	return opts, nil
}

func splitArgsAtNextModifier(args []string) (nonFlags []string, rest []string) {
	// Look for modifier flags in the argument list. If we find one, split the
	// list so it becomes the first item in rest.
	for i, arg := range args {
		if _, ok := modifierFlagHandlers[arg]; ok {
			nonFlags, rest = args[:i], args[i:]
			return
		}
	}

	// Otherwise, all arguments are non-flag arguments.
	nonFlags, rest = args, nil
	return
}

func (opts *request) parseHelp(args []string) (int, error) {
	if args[0] == "help" && len(args) > 1 {
		// If "help" isn't the only argument given, treat it as a normal option to
		// be randomized
		return 0, nil
	}

	opts.Operation = showHelp
	return 1, nil
}

func (opts *request) parseList(_ []string) (int, error) {
	opts.Operation = listGroups
	return 1, nil
}

func (opts *request) parseShow(args []string) (int, error) {
	opts.Operation = showGroup
	return 2, opts.parseGroupName(args)
}

func (opts *request) parseSave(args []string) (int, error) {
	opts.Operation = saveGroup
	return 2, opts.parseGroupName(args)
}

func (opts *request) parseDelete(args []string) (int, error) {
	opts.Operation = deleteGroup
	return 2, opts.parseGroupName(args)
}

func (opts *request) parseN(args []string) (consumed int, err error) {
	consumed = 2

	value, err := parseFlagValue(args)
	if err != nil {
		return
	}

	if value == "all" {
		opts.All = true
		return
	}

	opts.Count, err = strconv.Atoi(value)
	if err != nil {
		err = Error{
			cause:    err,
			helpText: fmt.Sprintf("Whoops, %q isn't a valid count!", value),
		}
	}
	return
}

func (opts *request) parseGroupName(args []string) (err error) {
	opts.GroupName, err = parseFlagValue(args)
	return
}

func parseFlagValue(args []string) (string, error) {
	if len(args) < 2 {
		return "", Error{
			cause:    errors.Errorf("%q flag requires an argument", args[0]),
			helpText: fmt.Sprintf("Whoops, %q requires an argument!", args[0]),
		}
	}

	return args[1], nil
}
