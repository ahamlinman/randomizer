package main // import "go.alexhamlin.co/randomizer/cmd/randomize"

import (
	"fmt"
	"os"

	"go.alexhamlin.co/randomizer/pkg/randomizer"
	"go.alexhamlin.co/randomizer/pkg/store"
)

func main() {
	storeFactory, err := store.FactoryFromEnv(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create store: %+v\n", err)
		os.Exit(2)
	}

	app := randomizer.NewApp(os.Args[0], storeFactory("Groups"))
	result, err := app.Main(os.Args[1:])
	if err != nil {
		err := err.(randomizer.Error)
		fmt.Fprintln(os.Stderr, err.HelpText())
		fmt.Fprintf(os.Stderr, "(%v)\n", err)
		os.Exit(1)
	}

	fmt.Println(result.Message())
}
