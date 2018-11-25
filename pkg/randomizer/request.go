package randomizer

import (
	"fmt"
	"strings"

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

func isOperation(flag string) bool {
	_, ok := operationFlagHandlers[flag]
	return ok
}

func (a App) newRequestFromArgs(args []string) (request, error) {
	request := request{}

	if len(args) < 1 {
		return request, nil
	}

	consumeFlag := func(handler flagHandler) error {
		consumed, err := handler(&request, args)
		if err != nil {
			return err
		}

		args = args[consumed:]
		return nil
	}

	if flag := args[0]; isOperation(flag) {
		// Consume an operation flag
		handler := operationFlagHandlers[flag]
		if err := consumeFlag(handler); err != nil {
			return request, err
		}
	} else if strings.HasPrefix(flag, "/") {
		// The user may have mistyped a flag; let them know about the error.
		return request, Error{
			cause: errors.Errorf("unknown flag %q", args[0]),
			helpText: fmt.Sprintf(
				`Whoops, %q isn't a valid flag. (Try "%s /help" to learn more about flags!)`,
				args[0],
				a.name,
			),
		}
	}

	request.Args = append(request.Args, args...)
	return request, nil
}

func (r *request) parseHelp(args []string) (int, error) {
	if args[0] == "help" && len(args) > 1 {
		// If "help" isn't the only argument given, treat it as a normal option to
		// be randomized
		return 0, nil
	}

	r.Operation = showHelp

	// Consume a help category if one was provided
	if err := r.parseOperand(args); err == nil {
		return 2, nil
	}

	// Otherwise, show the default help
	r.Operand = ""
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
