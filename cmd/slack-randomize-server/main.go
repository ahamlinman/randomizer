package main // import "go.alexhamlin.co/randomizer/cmd/slack-randomize-server"

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"go.alexhamlin.co/randomizer/pkg/slack"
	"go.alexhamlin.co/randomizer/pkg/store"
)

func main() {
	var (
		addr string
		name string
	)

	flag.StringVar(&addr, "addr", ":7636", "address to bind the server to")
	flag.StringVar(&name, "name", "/randomize", "name of the command to show in help text")
	flag.Parse()

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
		Name:         name,
		Token:        []byte(token),
		StoreFactory: storeFactory,
		LogFunc:      log.New(os.Stderr, "", 0).Printf,
	})

	fmt.Println("Starting randomizer service on", addr)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to start server: %v\n", err)
		os.Exit(1)
	}
}
