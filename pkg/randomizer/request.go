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
	r := request{}
	err := r.populateFromArgs(args)
	return r, err
}

func (r *request) populateFromArgs(args []string) error {
	rest, err := r.consumeFlagIfPresent(args)
	r.Args = rest
	return err
}

func (r *request) consumeFlagIfPresent(args []string) (rest []string, err error) {
	if len(args) >= 1 && isFlag(args[0]) {
		return r.consumeFlag(args)
	}

	return args, nil
}

func isFlag(flag string) bool {
	_, ok := flagHandlers[flag]
	return ok
}

func (r *request) consumeFlag(args []string) (rest []string, err error) {
	flag := args[0]
	handler := flagHandlers[flag]

	consumed, err := handler(r, args)
	if err != nil {
		return args, err
	}

	return args[consumed:], nil
}

type flagHandler func(r *request, args []string) (argsConsumed int, err error)

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
