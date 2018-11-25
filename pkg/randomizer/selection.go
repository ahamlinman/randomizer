package randomizer

import (
	"fmt"
	"regexp"

	"github.com/pkg/errors"
)

func (a App) makeSelection(request request) (Result, error) {
	options, err := a.expandArgs(request.Args)
	if err != nil {
		return Result{}, err
	}

	if len(options) < 2 {
		return Result{}, Error{
			cause:    errors.New("too few options"),
			helpText: "Whoops, I need at least two options to pick from!",
		}
	}

	a.shuffle(options)

	return Result{
		resultType: Selection,
		message:    fmt.Sprintf("I choose *%s*!", options[0]),
	}, nil
}

var expandArgModifiers = regexp.MustCompile("^[+-]")

func (a App) expandArgs(args []string) ([]string, error) {
	result := make([]string, 0, len(args))

	for _, arg := range args {
		var err error

		switch expandArgModifiers.FindString(arg) {
		case "":
			// No modifier; simply add this as a possible option
			result = append(result, arg)

		case "+":
			// Modifier for a group name; add all elements from the group to the set
			// of options
			group := arg[1:]
			result, err = a.appendGroup(result, group)
			if err != nil {
				return nil, err
			}

		case "-":
			// Modifier for a removal; remove the first instance of this arg from
			// the option set
			option := arg[1:]
			result, err = remove(result, option)
			if err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

func (a App) appendGroup(options []string, group string) ([]string, error) {
	expansion, err := a.store.Get(group)
	if err != nil {
		return options, Error{
			cause: err,
			helpText: fmt.Sprintf(
				"Whoops, I had trouble getting the %q group. Please try again later!",
				group,
			),
		}
	}

	if len(expansion) == 0 {
		return options, Error{
			cause:    err,
			helpText: fmt.Sprintf("Whoops, I couldn't find the %q group in this channel!", group),
		}
	}

	return append(options, expansion...), nil
}

func remove(options []string, option string) ([]string, error) {
	for i, item := range options {
		if item == option {
			return append(options[:i], options[i+1:]...), nil
		}
	}

	return options, Error{
		cause:    errors.Errorf("option %q not found for removal", option),
		helpText: fmt.Sprintf("Whoops, %q wasn't available for me to remove!", option),
	}
}
