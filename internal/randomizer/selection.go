package randomizer

import (
	"fmt"

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
		message:    fmt.Sprintf("I randomized and got: %s.", listify(embolden(options))),
	}, nil
}

func (a App) expandArgs(args []string) ([]string, error) {
	if len(args) == 1 {
		return a.expandGroup(args[0])
	}

	return args, nil
}

func (a App) expandGroup(group string) ([]string, error) {
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
			cause: errors.New("group not found"),
			helpText: fmt.Sprintf(
				`Whoops, I couldn't find the %q group in this channel. (Type "%s help" to learn more about groups!)`,
				group,
				a.name,
			),
		}
	}

	return expansion, nil
}
