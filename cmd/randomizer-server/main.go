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
	"log/slog"
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
		slog.Error("Unable to configure Slack token", "err", err)
		os.Exit(2)
	}

	storeFactory, err := store.FactoryFromEnv(context.Background())
	if err != nil {
		slog.Error("Unable to create store", "err", err)
		os.Exit(2)
	}

	mux := http.NewServeMux()
	mux.Handle("/", slack.App{
		TokenProvider: tokenProvider,
		StoreFactory:  storeFactory,
		Logger:        slog.Default(),
	})
	mux.Handle("/healthz",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))

	srv := &http.Server{Addr: addr, Handler: mux}
	go func() {
		slog.Info("Starting randomizer server", "addr", addr)
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Unable to start server", "err", err)
			os.Exit(1)
		}
	}()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, exitSignals...)
	<-exit
	signal.Stop(exit)

	slog.Info("Shutting down; interrupt again to force exit")
	err = srv.Shutdown(context.Background())
	if err != nil {
		slog.Error("Unable to shut down gracefully", "err", err)
	}
}
