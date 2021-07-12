package randomizer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

func (a App) listGroups(request request) (Result, error) {
	var (
		ctx = request.Context
	)

	groups, err := a.store.List(ctx)
	if err != nil {
		return Result{}, Error{
			cause:    err,
			helpText: "Whoops, I had trouble getting this channel's groups. Please try again later!",
		}
	}

	if len(groups) == 0 {
		return Result{
			resultType: ListedGroups,
			message:    "Whoops, no groups are available in this channel. (Use the /save flag to create one!)",
		}, nil
	}

	sort.Strings(groups)

	return Result{
		resultType: ListedGroups,
		message: fmt.Sprintf(
			"The following groups are available in this channel:\n%s",
			bulletize(groups),
		),
	}, nil
}

func (a App) showGroup(request request) (Result, error) {
	var (
		ctx  = request.Context
		name = request.Operand
	)

	group, err := a.store.Get(ctx, name)
	if err != nil {
		return Result{}, Error{
			cause:    err,
			helpText: "Whoops, I had trouble getting that group. Please try again later!",
		}
	}

	if len(group) == 0 {
		return Result{}, Error{
			cause:    errors.New("group does not exist"),
			helpText: "Whoops, I can't find that group in this channel. (Use the /save flag to create it!)",
		}
	}

	sort.Strings(group)

	return Result{
		resultType: ShowedGroup,
		message: fmt.Sprintf(
			"The %q group has the following options:\n%s",
			name,
			bulletize(group),
		),
	}, nil
}

func (a App) saveGroup(request request) (Result, error) {
	var (
		ctx     = request.Context
		name    = request.Operand
		options = request.Args
	)

	if isForbiddenGroupName(name) {
		return Result{}, Error{
			cause: errors.Errorf("saving with forbidden group name %q", name),
			helpText: fmt.Sprintf(
				`Whoops, %q has a special meaning and can't be used as a group name. (Type "%s help" to learn more!)`,
				name,
				a.name,
			),
		}
	}

	if len(options) < 2 {
		return Result{}, Error{
			cause:    errors.New("too few options to save"),
			helpText: "Whoops, I need at least two options to save a group!",
		}
	}

	if err := a.store.Put(ctx, name, options); err != nil {
		return Result{}, Error{
			cause:    err,
			helpText: "Whoops, I had trouble saving that group. Please try again later!",
		}
	}

	sort.Strings(options)

	return Result{
		resultType: SavedGroup,
		message: fmt.Sprintf(
			"Done! The %q group was saved in this channel with the following options:\n%s",
			name,
			bulletize(options),
		),
	}, nil
}

func isForbiddenGroupName(name string) bool {
	// Keep "/" reserved as a prefix for flags. Also block "help," as it has
	// special handling.
	return name == "help" || strings.HasPrefix(name, "/")
}

func (a App) deleteGroup(request request) (Result, error) {
	var (
		ctx  = request.Context
		name = request.Operand
	)

	existed, err := a.store.Delete(ctx, name)
	if err != nil {
		return Result{}, Error{
			cause:    err,
			helpText: "Whoops, I had trouble deleting that group. Please try again later!",
		}
	}

	if !existed {
		return Result{}, Error{
			cause:    errors.New("group does not exist"),
			helpText: "Whoops, I can't find that group in this channel!",
		}
	}

	return Result{
		resultType: DeletedGroup,
		message:    fmt.Sprintf("Done! The %q group was deleted.", name),
	}, nil
}
