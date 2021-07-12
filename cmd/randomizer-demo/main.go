/*

The randomizer-demo command is a quick way to test the randomizer's
functionality before deploying it into a Slack workspace.

CLI arguments are passed directly to the randomizer app.

*/
package main

import (
	"context"
	"fmt"
	"os"

	"go.alexhamlin.co/randomizer/internal/randomizer"
	"go.alexhamlin.co/randomizer/internal/store"
)

func main() {
	storeFactory, err := store.FactoryFromEnv(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create store: %+v\n", err)
		os.Exit(2)
	}

	app := randomizer.NewApp(os.Args[0], storeFactory("Groups"))
	result, err := app.Main(context.Background(), os.Args[1:])
	if err != nil {
		err := err.(randomizer.Error)
		fmt.Fprintln(os.Stderr, err.HelpText())
		fmt.Fprintf(os.Stderr, "(%v)\n", err)
		os.Exit(1)
	}

	fmt.Println(result.Message())
}
