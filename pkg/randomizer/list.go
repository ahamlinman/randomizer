package randomizer

import (
	"fmt"
	"regexp"

	"github.com/pkg/errors"
)

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
