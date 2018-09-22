package main // import "go.alexhamlin.co/randomizer/cmd/slack-randomize-server"

import (
	"fmt"
	"net/http"
	"os"

	"go.alexhamlin.co/randomizer/pkg/slack"
	"go.alexhamlin.co/randomizer/pkg/store"
)

func main() {
	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "SLACK_TOKEN must be provided")
		os.Exit(2)
	}

	storeFactory, err := store.FactoryFromEnv(os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create store: %+v\n", err)
		os.Exit(2)
	}

	http.Handle("/", slack.App{
		Name:         "/randomize",
		Token:        []byte(token),
		StoreFactory: storeFactory,
	})

	fmt.Println("Starting randomizer service on :7636")
	err = http.ListenAndServe("0.0.0.0:7636", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to start server: %v\n", err)
		os.Exit(1)
	}
}
