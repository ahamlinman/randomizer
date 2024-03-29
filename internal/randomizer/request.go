package randomizer

import (
	"context"
	"fmt"
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
	Context   context.Context
	Operation operation
	Operand   string
	Args      []string
}

func (a App) newRequest(ctx context.Context, args []string) (req request, err error) {
	req.Context = ctx
	req.Operation, req.Operand, req.Args, err = parseArgs(args)
	return
}

func parseArgs(args []string) (op operation, operand string, opargs []string, err error) {
	// We accept the standard flag syntax for help, but strongly expect that users
	// won't know that syntax in advance. Logic elsewhere in the randomizer
	// prevents the use of "help" as a group name to avoid conflicts with this
	// special case.
	if len(args) == 0 || args[0] == "/help" || len(args) == 1 && args[0] == "help" {
		return showHelp, "", args, nil
	}

	switch args[0] {
	// Arguments without an explicitly known flag always trigger a randomization,
	// even if the first argument starts with a slash, simply because it's less
	// work to implement and unlikely to cause big problems in practice. Logic
	// elsewhere in the randomizer prevents the use of flag-like group names, so
	// that new flags can't make existing groups inaccessible.
	default:
		return makeSelection, "", args, nil

	// Listing groups requires no arguments...
	case "/list":
		return listGroups, "", args, nil

	// ...and everything else needs the name of a group to operate on, which we
	// validate and extract out from the rest of the arguments for convenience. We
	// make no assumptions about how each operation uses the rest of the available
	// arguments.
	case "/show":
		op = showGroup
	case "/save":
		op = saveGroup
	case "/delete":
		op = deleteGroup
	}

	if len(args) < 2 {
		return op, "", nil, Error{
			cause:    fmt.Errorf("%q flag requires an argument", args[0]),
			helpText: fmt.Sprintf("Whoops, %q requires an argument!", args[0]),
		}
	}

	return op, args[1], args[2:], nil
}
