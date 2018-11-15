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

	var choices []string
	switch {
	case request.All:
		choices = options

	case request.Count < 1:
		return Result{}, Error{
			cause:    errors.New("count too small"),
			helpText: "Whoops, I can't pick less than one option!",
		}

	case request.Count > len(options):
		return Result{}, Error{
			cause:    errors.New("count too large"),
			helpText: "Whoops, I can't pick more options than I was given!",
		}

	default:
		choices = options[:request.Count]
	}

	choices = embolden(choices)

	return Result{
		resultType: Selection,
		message:    fmt.Sprintf("I choose %s!", listify(choices)),
	}, nil
}

var expandArgModifiers = regexp.MustCompile("^[+-]")

func (a App) expandArgs(args []string) ([]string, error) {
	result := make([]string, 0, len(args))

	for _, arg := range args {
		switch expandArgModifiers.FindString(arg) {
		case "":
			// No modifier; simply add this as a possible option
			result = append(result, arg)

		case "+":
			// Modifier for a group name; add all elements from the group to the set
			// of options
			group := arg[1:]
			expansion, err := a.store.Get(group)
			if err != nil {
				return nil, Error{
					cause: err,
					helpText: fmt.Sprintf(
						"Whoops, I had trouble getting the %q group. Please try again later!",
						group,
					),
				}
			}

			if len(expansion) == 0 {
				return nil, Error{
					cause:    err,
					helpText: fmt.Sprintf("Whoops, I couldn't find the %q group in this channel!", group),
				}
			}

			result = append(result, expansion...)

		case "-":
			// Modifier for a removal; remove the first instance of this arg from
			// the option set
			option := arg[1:]
			var ok bool
			result, ok = remove(result, option)
			if !ok {
				return nil, Error{
					cause:    errors.Errorf("option %q not found for removal", option),
					helpText: fmt.Sprintf("Whoops, %q wasn't available for me to remove!", option),
				}
			}
		}
	}

	return result, nil
}

// remove attempts to remove the first instance of the provided string in the
// provided slice, modifying it in place. It returns an updated slice, along
// with a boolean indicating whether the provided string was found.
func remove(items []string, itemToRemove string) ([]string, bool) {
	for i, item := range items {
		if item == itemToRemove {
			return append(items[:i], items[i+1:]...), true
		}
	}

	return items, false
}
