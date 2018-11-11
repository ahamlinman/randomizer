package randomizer

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

type operation int

const (
	noOp operation = iota
	showHelp
	listGroups
	showGroup
	saveGroup
	deleteGroup
)

type options struct {
	Operation operation
	Operand   string

	Args  []string
	All   bool
	Count int
}

type flagHandler func(*options, []string) (int, error)

var operationFlagHandlers = map[string]flagHandler{
	"/help":   (*options).parseHelp,
	"/list":   (*options).parseList,
	"/show":   (*options).parseShow,
	"/save":   (*options).parseSave,
	"/delete": (*options).parseDelete,
}

var modifierFlagHandlers = map[string]flagHandler{
	"/n": (*options).parseN,
}

func parseArgs(args []string) (options, error) {
	opts := options{
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

	if handler := operationFlagHandlers[args[0]]; handler != nil {
		if err := consumeFlag(handler); err != nil {
			return opts, err
		}
	}

	for len(args) > 0 {
		if handler := modifierFlagHandlers[args[0]]; handler != nil {
			if err := consumeFlag(handler); err != nil {
				return opts, err
			}
			continue
		}

		var nonFlagArgs []string
		nonFlagArgs, args = splitArgsAtNextModifier(args)
		opts.Args = append(opts.Args, nonFlagArgs...)
	}

	return opts, nil
}

func splitArgsAtNextModifier(args []string) (nonFlags []string, rest []string) {
	// Start by assuming that none of the remaining arguments have flags.
	nonFlags = args
	// rest = nil

	// Run through the array and check this assumption.
	for i, arg := range args {
		if _, ok := modifierFlagHandlers[arg]; !ok {
			continue
		}

		// It's wrong, so overwrite the result based on the flag we've identified.
		nonFlags = args[:i]
		rest = args[i:]
		break
	}

	return
}

func (opts *options) parseHelp(_ []string) (int, error) {
	opts.Operation = showHelp
	return 1, nil
}

func (opts *options) parseList(_ []string) (int, error) {
	opts.Operation = listGroups
	return 1, nil
}

func (options) parseFlagValue(args []string) (string, error) {
	if len(args) < 2 {
		return "", Error{
			cause:    errors.Errorf("%q option requires an argument", args[0]),
			helpText: fmt.Sprintf("Whoops, %q requires an argument!", args[0]),
		}
	}

	return args[1], nil
}

func (opts *options) parseOperation(op operation, args []string) (consumed int, err error) {
	consumed = 2

	value, err := opts.parseFlagValue(args)
	if err != nil {
		return
	}

	opts.Operation = op
	opts.Operand = value
	return
}

func (opts *options) parseShow(args []string) (int, error) {
	return opts.parseOperation(showGroup, args)
}

func (opts *options) parseSave(args []string) (int, error) {
	return opts.parseOperation(saveGroup, args)
}

func (opts *options) parseDelete(args []string) (int, error) {
	return opts.parseOperation(deleteGroup, args)
}

func (opts *options) parseN(args []string) (consumed int, err error) {
	consumed = 2

	value, err := opts.parseFlagValue(args)
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
