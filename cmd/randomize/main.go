package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ahamlinman/randomizer/pkg/randomizer"
)

func main() {
	app := randomizer.NewApp()
	result, err := app.Main(os.Args[1:])

	switch err {
	case randomizer.ErrTooFewOptions:
		fmt.Fprintln(os.Stderr, "Hey, I need things to randomize!")
		os.Exit(1)

	case flag.ErrHelp:
		fmt.Fprintln(os.Stderr, "Usage: Just give me things to randomize!")
		return

	default:
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	fmt.Println(result)
}
