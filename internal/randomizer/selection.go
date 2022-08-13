package randomizer

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

func (a App) makeSelection(request request) (Result, error) {
	options, err := a.expandArgs(request.Context, request.Args)
	if err != nil {
		return Result{}, err
	}

	a.shuffle(options)

	return Result{
		resultType: Selection,
		message:    fmt.Sprintf("I randomized and got: %s.", inlinelist(options)),
	}, nil
}

func (a App) expandArgs(ctx context.Context, args []string) ([]string, error) {
	if len(args) == 1 {
		return a.expandGroup(ctx, args[0])
	}

	return args, nil
}

func (a App) expandGroup(ctx context.Context, group string) ([]string, error) {
	expansion, err := a.store.Get(ctx, group)
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
			cause: errors.Errorf("group %q not found", group),
			helpText: fmt.Sprintf(
				`Whoops, I couldn't find the %q group in this channel. (Type "%s help" to learn more about groups!)`,
				group, a.name,
			),
		}
	}

	return expansion, nil
}
