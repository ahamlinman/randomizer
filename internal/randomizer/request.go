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
	if isHelpRequest(args) {
		r.Operation = showHelp
		return nil
	}

	rest, err := r.consumeFlagIfPresent(args)
	r.Args = rest
	return err
}

func isHelpRequest(args []string) bool {
	// Show the help message if...
	switch {
	// ...the user doesn't know how to ask for help...
	case len(args) == 0:
		return true

	// ...or doesn't yet know the flag syntax ("/" prefix)...
	case len(args) == 1 && args[0] == "help":
		return true

	// ...or actually asks for it directly.
	case args[0] == "/help":
		return true

	default:
		return false
	}
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
	"/list":   (*request).parseList,
	"/show":   (*request).parseShow,
	"/save":   (*request).parseSave,
	"/delete": (*request).parseDelete,
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
