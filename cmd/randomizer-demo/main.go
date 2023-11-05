// The randomizer-demo command invokes the randomizer using arguments from the
// command line.
//
// The demo CLI provides a way to try the randomizer without fully deploying it
// as a Slack slash command. It supports the same slash-prefixed flag syntax as
// the slash command, and writes output in Slack's "mrkdwn" format. It supports
// the same environment variables as the randomizer-server command to configure
// storage for groups.
//
// Unlike the slash command, which splits a single argument string by
// whitespace, the demo CLI treats each CLI argument as a direct argument to
// the randomizer. Note that this allows the demo CLI to exhibit behaviors not
// normally possible with the slash command, such as randomizing or storing
// options containing whitespace.
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
		fmt.Fprintf(os.Stderr, "Failed to create store: %v\n", err)
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
