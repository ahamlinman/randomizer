package main

import (
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

	if err == nil {
		fmt.Println(result)
		return
	}

	if err, ok := err.(randomizer.Error); ok {
		fmt.Fprintln(os.Stderr, err.HelpText())
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Whoops, we had a problem… %v\n", err)
}
