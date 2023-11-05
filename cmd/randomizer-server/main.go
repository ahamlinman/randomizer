// The randomizer-server command is an HTTP server that serves the Slack slash
// command API for the randomizer.
//
// See the randomizer repository README for more information on configuring and
// deploying the server.
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
	http.Handle("/healthz",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))

	fmt.Println("Starting randomizer service on", addr)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to start server: %v\n", err)
		os.Exit(1)
	}
}
