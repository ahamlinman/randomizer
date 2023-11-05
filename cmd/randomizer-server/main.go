// The randomizer-server command is an HTTP server that serves the Slack slash
// command API for the randomizer.
//
// See the randomizer repository README for more information on configuring and
// deploying the server.
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"

	"go.alexhamlin.co/randomizer/internal/slack"
	"go.alexhamlin.co/randomizer/internal/store"
)

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":7636", "address to bind the server to")
	flag.Parse()

	tokenProvider, err := slack.TokenProviderFromEnv()
	if err != nil {
		log.Printf("Unable to configure Slack token: %+v\n", err)
		os.Exit(2)
	}

	storeFactory, err := store.FactoryFromEnv(context.Background())
	if err != nil {
		log.Printf("Unable to create store: %+v\n", err)
		os.Exit(2)
	}

	mux := http.NewServeMux()
	mux.Handle("/", slack.App{
		TokenProvider: tokenProvider,
		StoreFactory:  storeFactory,
		DebugWriter:   os.Stderr,
	})
	mux.Handle("/healthz",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))

	srv := &http.Server{Addr: addr, Handler: mux}
	go func() {
		log.Printf("Starting randomizer server on %s", addr)
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Unable to start server: %v\n", err)
			os.Exit(1)
		}
	}()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, exitSignals...)
	<-exit
	signal.Stop(exit)

	log.Print("Shutting down; interrupt again to force exit")
	err = srv.Shutdown(context.Background())
	if err != nil {
		log.Printf("Unable to shut down gracefully: %v", err)
	}
}
