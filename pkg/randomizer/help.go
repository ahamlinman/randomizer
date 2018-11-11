package randomizer

import (
	"bytes"
	"text/template"
)

func buildHelpMessage(name string) string {
	var buf bytes.Buffer
	helpMessageTemplate.Execute(&buf, struct{ Name string }{name})
	return buf.String()
}

var helpMessageTemplate = template.Must(template.New("").Parse(
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
