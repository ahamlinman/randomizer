package randomizer

import (
	"bytes"
	"flag"
	"io/ioutil"
	"text/template"
)

type flagSet struct {
	*flag.FlagSet
	name string

	count int
	all   bool

	listGroups  bool
	showGroup   string
	saveGroup   string
	deleteGroup string
}

func buildFlagSet(name string) *flagSet {
	fs := &flagSet{
		FlagSet: flag.NewFlagSet("", flag.ContinueOnError),
		name:    name,
	}
	fs.SetOutput(ioutil.Discard)

	fs.IntVar(&fs.count, "n", 1, "number of items to pick")
	fs.BoolVar(&fs.all, "all", false, "pick all items in a random order")

	fs.BoolVar(&fs.listGroups, "list", false, "list all known groups")
	fs.StringVar(&fs.showGroup, "show", "", "show the options in the specified group")
	fs.StringVar(&fs.saveGroup, "save", "", "save options into the specified group")
	fs.StringVar(&fs.deleteGroup, "delete", "", "delete the specified group")

	return fs
}

var usageTmpl = template.Must(template.New("").Parse(
	`{{.Name}} helps you pick options randomly out of a list.

*Example:* {{.Name}} one two three
> I choose *three*!

You can choose more than one option at a time. The selected options will be given back in a random order.

*Example:* {{.Name}} -n 2 one two three
> I choose *two* and *one*!

*Example:* {{.Name}} -all one two three
> I choose *two* and *three* and *one*!

You can also create *groups* for the current channel or DM.

*Save a group:* {{.Name}} -save first3 one two three
*Randomize from a group:* {{.Name}} +first3
*Combine groups with other options:* {{.Name}} -n 3 +first3 +next3 seven eight
*Remove some options from consideration:* {{.Name}} +first3 +next3 -two -five
*List groups:* {{.Name}} -list
*Show options in a group:* {{.Name}} -show first3
*Delete a group:* {{.Name}} -delete first3

Note that the selection is weighted. An option is more likely to be picked if it is given multiple times. This also applies when multiple groups are given, and an option is in more than one of them.`))

func (fs *flagSet) buildUsage() string {
	var buf bytes.Buffer
	usageTmpl.Execute(&buf, struct{ Name string }{fs.name})
	return buf.String()
}
