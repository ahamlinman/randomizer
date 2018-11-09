package randomizer

import (
	"bytes"
	"regexp"
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
	name string

	n           countFlag
	listGroups  bool
	showGroup   string
	saveGroup   string
	deleteGroup string

	args []string
}

func buildFlagSet(name string) *flagSet {
	return &flagSet{
		name: name,
		n:    countFlag{count: 1},
	}
}

var flagRegexp = regexp.MustCompile(`^/(help|n|list|show|save|delete)$`)

func (fs *flagSet) Parse(args []string) (err error) {
	// TODO: This function is an almost 60-line crappy hack to allow
	// experimentation with the new options style. Clean it up if it seems like
	// it's going to work out.

	fs.args = nil

	defer func() {
		if r := recover(); r != nil {
			err = Error{
				cause:    errors.Errorf("recovered from panic: %v", r),
				helpText: fs.buildUsage(),
			}
		}
	}()

	consume := func(n int) {
		args = args[n:]
	}

	for len(args) > 0 {
		match := flagRegexp.FindStringSubmatch(args[0])
		switch {
		case match == nil:
			fs.args = append(fs.args, args[0])
			consume(1)

		case match[1] == "help":
			return Error{
				cause:    errors.New("help requested"),
				helpText: fs.buildUsage(),
			}

		case match[1] == "n":
			if err := fs.n.Set(args[1]); err != nil {
				return Error{
					cause:    err,
					helpText: fs.buildUsage(),
				}
			}
			consume(2)

		case match[1] == "save":
			// TODO: Make this more robust
			fs.saveGroup = args[1]
			consume(2)

		case match[1] == "list":
			fs.listGroups = true
			return nil

		case match[1] == "show":
			// TODO: Make this more robust
			fs.showGroup = args[1]
			return nil

		case match[1] == "delete":
			// TODO: Make this more robust
			fs.deleteGroup = args[1]
			return nil

		default:
			panic("wtf")
		}
	}

	if len(fs.args) == 1 && fs.args[0] == "help" {
		return Error{
			cause:    errors.New("help requested"),
			helpText: fs.buildUsage(),
		}
	}

	return nil
}

func (fs *flagSet) Args() []string {
	return fs.args
}

var usageTmpl = template.Must(template.New("").Parse(
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

func (fs *flagSet) buildUsage() string {
	var buf bytes.Buffer
	usageTmpl.Execute(&buf, struct{ Name string }{fs.name})
	return buf.String()
}
