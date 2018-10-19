package randomizer

import (
	"bytes"
	"flag"
	"io/ioutil"
	"strconv"
	"text/template"

	"github.com/pkg/errors"
)

// countFlag represents the possible values of the "-n" flag, which may be an
// integer count or the string "all" (representing as many options as are
// available).
type countFlag struct {
	all   bool
	count int
}

func (c *countFlag) String() string {
	if c == nil {
		return "0"
	}

	return strconv.Itoa(c.count)
}

func (c *countFlag) Set(val string) error {
	if val == "all" {
		*c = countFlag{all: true}
		return nil
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		return errors.WithStack(err)
	}
	*c = countFlag{count: count}
	return nil
}

// validateRange ensures that the count provided by the user will work for the
// given number of options.
func (c *countFlag) validateRange(n int) error {
	if c.all {
		return nil
	}

	if c.count < 1 {
		return Error{
			cause:    errors.New("count too small"),
			helpText: "Whoops, I can't pick less than one option!",
		}
	}

	if c.count > n {
		return Error{
			cause:    errors.New("count too large"),
			helpText: "Whoops, I can't pick more options than I was given!",
		}
	}

	return nil
}

type flagSet struct {
	*flag.FlagSet
	name string

	n countFlag

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

	fs.Var(&fs.n, "n", `number of items to pick (or "all" for all options)`)

	fs.BoolVar(&fs.listGroups, "list", false, "list all known groups")
	fs.StringVar(&fs.showGroup, "show", "", "show the options in the specified group")
	fs.StringVar(&fs.saveGroup, "save", "", "save options into the specified group")
	fs.StringVar(&fs.deleteGroup, "delete", "", "delete the specified group")

	return fs
}

func (fs *flagSet) Parse(args []string) error {
	err := fs.FlagSet.Parse(args)

	if err != nil || (len(args) == 1 && args[0] == "help") {
		if err == nil {
			err = errors.New("help requested in args")
		} else {
			err = errors.Wrap(err, "parsing flags")
		}

		return Error{
			cause:    err,
			helpText: fs.buildUsage(),
		}
	}

	return nil
}

var usageTmpl = template.Must(template.New("").Parse(
	`{{.Name}} helps you pick options randomly out of a list.

*Example:* {{.Name}} one two three
> I choose *three*!

You can choose more than one option at a time. The selected options will be given back in a random order.

*Example:* {{.Name}} -n 2 one two three
> I choose *two* and *one*!

*Example:* {{.Name}} -n all one two three
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
