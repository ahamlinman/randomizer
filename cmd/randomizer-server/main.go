// The randomizer-server command is an HTTP server that serves the Slack slash
// command API for the randomizer.
//
// See the randomizer repository README for more information on configuring and
// deploying the server.
package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/ahamlinman/randomizer/internal/slack"
	"github.com/ahamlinman/randomizer/internal/store"
)

var exitSignals = []os.Signal{os.Interrupt}

var (
	flagAddr    = flag.String("addr", ":7636", "address to bind the server to")
	flagLogJSON = flag.Bool("log-json", false, "log JSON to stderr instead of text")
)

func main() {
	flag.Parse()

	logger := slog.Default()
	if *flagLogJSON {
		logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))
	}

	tokenProvider, err := slack.TokenProviderFromEnv()
	if err != nil {
		logger.Error("Failed to configure Slack token", "err", err)
		os.Exit(2)
	}

	storeFactory, err := store.FactoryFromEnv(context.Background())
	if err != nil {
		logger.Error("Failed to create store", "err", err)
		os.Exit(2)
	}

	mux := http.NewServeMux()
	mux.Handle("/", slack.App{
		TokenProvider: tokenProvider,
		StoreFactory:  storeFactory,
		Logger:        logger,
	})
	mux.Handle("GET /healthz",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))

	srv := &http.Server{Addr: *flagAddr, Handler: mux}
	srvErr := make(chan error, 1)
	go func() {
		logger.Info("Starting randomizer server", "addr", *flagAddr)
		srvErr <- srv.ListenAndServe()
	}()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, exitSignals...)

	select {
	case err := <-srvErr:
		logger.Error("Failed to start server", "err", err)
		os.Exit(1)

	case <-exit:
	}

	signal.Stop(exit)
	logger.Info("Shutting down; interrupt again to force exit")
	err = srv.Shutdown(context.Background())
	if err != nil {
		logger.Error("Failed to shut down gracefully", "err", err)
	}
}
