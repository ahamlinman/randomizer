/*

The randomizer-server command starts a web server that accepts Slack Slash
Command API requests and runs the randomizer in response.

*/
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"go.alexhamlin.co/randomizer/internal/slack"
	"go.alexhamlin.co/randomizer/internal/store"
)

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":7636", "address to bind the server to")
	flag.Parse()

	tokenProvider, err := slack.TokenProviderFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to configure Slack token: %+v\n", err)
		os.Exit(2)
	}

	storeFactory, err := store.FactoryFromEnv(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create store: %+v\n", err)
		os.Exit(2)
	}

	http.Handle("/", slack.App{
		TokenProvider: tokenProvider,
		StoreFactory:  storeFactory,
		DebugWriter:   os.Stderr,
	})

	fmt.Println("Starting randomizer service on", addr)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to start server: %v\n", err)
		os.Exit(1)
	}
}
