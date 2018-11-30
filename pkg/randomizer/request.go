package randomizer

import (
	"fmt"

	"github.com/pkg/errors"
)

type operation int

const (
	makeSelection operation = iota
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
	Operand   string
	Args      []string
}

func (a App) newRequestFromArgs(args []string) (request, error) {
	if len(args) < 1 {
		return request{}, nil
	}

	r := request{}
	err := r.parseArgs(args)
	return r, err
}

func (r *request) parseArgs(args []string) error {
	if isFlag(args[0]) {
		var err error
		args, err = r.consumeFlag(args)
		if err != nil {
			return err
		}
	}

	r.Args = args
	return nil
}

func isFlag(flag string) bool {
	_, ok := flagHandlers[flag]
	return ok
}

func (r *request) consumeFlag(args []string) ([]string, error) {
	handler := flagHandlers[args[0]]
	consumed, err := handler(r, args)
	if err != nil {
		return args, err
	}

	return args[consumed:], nil
}

// flagHandler is a type for functions that can parse a flag and its value(s)
// from an argument list into a request struct.
//
// The argument slice provided to the handler starts at the argument containing
// the flag itself. If the returned error is nil, the returned int is the total
// number of arguments (1 or more) consumed by parsing this flag and its
// value(s).
type flagHandler func(*request, []string) (int, error)

var flagHandlers = map[string]flagHandler{
	// As a special case, show the help message if "help" is the only argument
	// provided by the user (in case they don't yet know the flag syntax)
	"help":  (*request).parseHelp,
	"/help": (*request).parseHelp,

	"/list":   (*request).parseList,
	"/show":   (*request).parseShow,
	"/save":   (*request).parseSave,
	"/delete": (*request).parseDelete,
}

func (r *request) parseHelp(args []string) (int, error) {
	if args[0] == "help" && len(args) > 1 {
		// If "help" isn't the only argument given, treat it as a normal option to
		// be randomized
		return 0, nil
	}

	r.Operation = showHelp
	return 1, nil
}

func (r *request) parseList(_ []string) (int, error) {
	r.Operation = listGroups
	return 1, nil
}

func (r *request) parseShow(args []string) (int, error) {
	r.Operation = showGroup
	return 2, r.parseOperand(args)
}

func (r *request) parseSave(args []string) (int, error) {
	r.Operation = saveGroup
	return 2, r.parseOperand(args)
}

func (r *request) parseDelete(args []string) (int, error) {
	r.Operation = deleteGroup
	return 2, r.parseOperand(args)
}

func (r *request) parseOperand(args []string) (err error) {
	r.Operand, err = parseFlagValue(args)
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
