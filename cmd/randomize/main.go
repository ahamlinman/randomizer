package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ahamlinman/randomizer/pkg/randomizer"
	bolt "github.com/coreos/bbolt"
)

func main() {
	db, err := bolt.Open("randomizer.db", os.ModePerm&0644, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}

	app := randomizer.NewApp(&boltStore{db})
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
			fmt.Fprintf(os.Stderr, "%+v", err)
			os.Exit(1)
		}
	}

	fmt.Println(result)
}
