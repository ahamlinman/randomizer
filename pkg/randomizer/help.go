package randomizer

import (
	"bytes"
	"regexp"
	"text/template"
)

var groupsHelpRegexp = regexp.MustCompile("groups?")

func (a App) getHelpMessage(category string) string {
	template := helpMessageTemplate
	if groupsHelpRegexp.MatchString(category) {
		template = groupHelpMessageTemplate
	}

	var buf bytes.Buffer
	template.Execute(&buf, struct{ Name string }{a.name})
	return buf.String()
}

var helpMessageTemplate = template.Must(template.New("").Parse(
	`:game_die: *{{.Name}}*

{{.Name}} helps you pick options randomly out of a list.

*Example:* {{.Name}} one two three
> I choose *three*!

You can choose more than one option at a time. The selected options will be given back in a random order.

*Example:* {{.Name}} /n 2 one two three
> I choose *two* and *one*!

*Example:* {{.Name}} /n all one two three
> I choose *two*, *three*, and *one*!

If you use a set of options a lot (say, the names of your team members), try saving them as a *group* in the current channel or DM! Type "{{.Name}} /help groups" to learn more.

Note that the selection is weighted. An option is more likely to be picked if it is given multiple times. This also applies when multiple groups are given, and an option is in more than one of them.`))

var groupHelpMessageTemplate = template.Must(template.New("").Parse(
	`:busts_in_silhouette: *Groups*

{{.Name}} lets you save *groups* in the current channel or DM.

Never sure where to eat? Save some local restaurants into your personal DM! Need to request a code review? Save your team members' names into your team's channel! With groups, your most common choices are always at hand.

*List your current channel's groups:* {{.Name}} /list
*Save a group:* {{.Name}} /save snacks chips pretzels jerky trailmix
*Show the options in a group:* {{.Name}} /show snacks
*Delete a group:* {{.Name}} /delete snacks

Use *+* to include a group in your choices.

*Example:* {{.Name}} +snacks
> I choose *pretzels*!

Don't care for pretzels today? Use *-* to remove an option from this round (without removing it from the group permanently).

*Example:* {{.Name}} /n all +snacks -pretzels cereal
> I choose *jerky*, *trailmix*, *cereal*, and *chips*!

Like the example shows, you can still use flags like /n and add options from outside the group. (See "{{.Name}} /help" for the basics.)`))
