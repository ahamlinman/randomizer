package randomizer

import (
	"errors"
	"flag"
	"io/ioutil"
	"math/rand"
	"time"
)

// ErrTooFewOptions is returned when, after parsing arguments, there are fewer
// than two options to choose from for randomization.
var ErrTooFewOptions = errors.New("nothing to randomize")

// Main is the entrypoint to the randomizer tool.
func Main(args []string) (string, error) {
	fs := buildFlagSet()
	err := fs.Parse(args)
	if err != nil {
		return "", err
	}

	options := fs.Args()
	if len(options) < 2 {
		return "", ErrTooFewOptions
	}

	source := rand.NewSource(time.Now().UnixNano())
	rander := rand.New(source)
	selector := Selector(rander.Intn)
	return selector.PickString(options), nil
}

type flagSet struct {
	*flag.FlagSet
}

func buildFlagSet() *flagSet {
	fs := &flagSet{
		FlagSet: flag.NewFlagSet("randomizer", flag.ContinueOnError),
	}
	fs.SetOutput(ioutil.Discard)

	return fs
}
