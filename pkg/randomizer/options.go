package randomizer

import (
	"bytes"
	"fmt"
	"strconv"
	"text/template"

	"github.com/pkg/errors"
)

type operation int

const (
	noOp operation = iota
	listGroups
	showGroup
	saveGroup
	deleteGroup
)

type options struct {
	Name      string
	Operation operation
	Operand   string
	Args      []string
	Count     int
}

var parseHandlers = map[string]func(*options, []string) (int, error){
	"/list":   (*options).parseList,
	"/show":   (*options).parseShow,
	"/save":   (*options).parseSave,
	"/delete": (*options).parseDelete,
	"/n":      (*options).parseN,
	"/help":   (*options).parseHelp,
}

func parseArgs(name string, args []string) (options, error) {
	opts := options{
		Name:  name, // TODO: Weird that options has to know about this
		Count: 1,
	}

	// TODO: This is such a terrible special case.
	if len(args) == 1 && args[0] == "help" {
		_, err := opts.parseHelp(args)
		return opts, err
	}

	for len(args) > 0 {
		if handler := parseHandlers[args[0]]; handler != nil {
			consumed, err := handler(&opts, args)
			if err != nil {
				return opts, err
			}

			args = args[consumed:]
			continue
		}

		var nonFlagArgs []string
		nonFlagArgs, args = splitArgsAtNextFlag(args)
		opts.Args = append(opts.Args, nonFlagArgs...)
	}

	return opts, nil
}

func splitArgsAtNextFlag(args []string) (nonFlags []string, rest []string) {
	// Start by assuming that none of the remaining arguments have flags.
	nonFlags = args

	// Run through the array and check this assumption.
	for i, arg := range args {
		if _, ok := parseHandlers[arg]; !ok {
			continue
		}

		// It's wrong, so overwrite the result based on the flag we've identified.
		nonFlags = args[:i]
		rest = args[i:]
		break
	}

	return
}

func (opts *options) parseList(args []string) (int, error) {
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
		opts.Count = -5000 // TODO: Magic number
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

func (opts *options) parseHelp(_ []string) (int, error) {
	var buf bytes.Buffer
	usageTmpl.Execute(&buf, struct{ Name string }{opts.Name})
	return 1, Error{
		cause:    errors.New("help requested"),
		helpText: buf.String(),
	}
}

var usageTmpl = template.Must(template.New("").Parse(
	`{{.Name}} helps you pick options randomly out of a list.

*Example:* {{.Name}} one two three
> I choose *three*!

You can choose more than one option at a time. The selected options will be given back in a random order.

*Example:* {{.Name}} /n 2 one two three
> I choose *two* and *one*!

*Example:* {{.Name}} /n all one two three
> I choose *two* and *three* and *one*!

You can also create *groups* for the current channel or DM.

*Save a group:* {{.Name}} /save first3 one two three
*Randomize from a group:* {{.Name}} +first3
*Combine groups with other options:* {{.Name}} /n 3 +first3 +next3 seven eight
*Remove some options from consideration:* {{.Name}} +first3 +next3 -two -five
*List groups:* {{.Name}} /list
*Show options in a group:* {{.Name}} /show first3
*Delete a group:* {{.Name}} /delete first3

Note that the selection is weighted. An option is more likely to be picked if it is given multiple times. This also applies when multiple groups are given, and an option is in more than one of them.`))
