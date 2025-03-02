package randomizer

import "strings"

func (a App) showHelp(request request) (Result, error) {
	return Result{
		resultType: ShowedHelp,
		message:    strings.ReplaceAll(helpMessageTemplate, "{{.Name}}", a.name),
	}, nil
}

// helpMessageTemplate is written with text/template syntax for familiarity.
// However, text/template uses reflection in a way that disables dead code
// elimination for the _entire_ program, so we instead use plan string
// replacement to substitute our one value.
const helpMessageTemplate = `{{.Name}} randomizes the order of options in a list.

*Example:* {{.Name}} one two three
&gt; I randomized and got: *two*, *three*, *one*.

If you use a set of options a lot, try saving them as a *group* in the current channel or DM!

*Save a group:* {{.Name}} /save snacks chips pretzels trailmix
*Use a group:* {{.Name}} snacks
*List your current channel's groups:* {{.Name}} /list
*Show the options in a group:* {{.Name}} /show snacks
*Delete a group:* {{.Name}} /delete snacks`
