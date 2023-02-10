package randomizer

import (
	"strings"
	"text/template"
)

func (a App) showHelp(request request) (Result, error) {
	var b strings.Builder
	helpMessageTemplate.Execute(&b, struct{ Name string }{a.name})
	return Result{
		resultType: ShowedHelp,
		message:    b.String(),
	}, nil
}

var helpMessageTemplate = template.Must(template.New("").Parse(
	`{{.Name}} randomizes the order of options in a list.

*Example:* {{.Name}} one two three
&gt; I randomized and got: *two*, *three*, *one*.

If you use a set of options a lot, try saving them as a *group* in the current channel or DM!

*Save a group:* {{.Name}} /save snacks chips pretzels trailmix
*Use a group:* {{.Name}} snacks
*List your current channel's groups:* {{.Name}} /list
*Show the options in a group:* {{.Name}} /show snacks
*Delete a group:* {{.Name}} /delete snacks`))
