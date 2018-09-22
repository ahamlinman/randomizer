package main // import "go.alexhamlin.co/randomizer/cmd/slack-randomize-server"

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"

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
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}

	http.HandleFunc("/", rootHandler(slack.App{
		Name:         "/randomize",
		Token:        []byte(token),
		StoreFactory: storeFactory,
	}))

	fmt.Println("Starting randomizer service")
	err = http.ListenAndServe("0.0.0.0:7636", nil)
	if err != nil {
		logErr(errors.Wrap(err, "starting server"))
		os.Exit(1)
	}
}

func rootHandler(app slack.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			logErr(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		response, err := app.Run(slack.Request{
			Token:     r.PostForm.Get("token"),
			SSLCheck:  r.PostForm.Get("ssl_check"),
			ChannelID: r.PostForm.Get("channel_id"),
			Text:      r.PostForm.Get("text"),
		})

		if err != nil {
			logErr(err)

			if err == slack.ErrIncorrectToken {
				w.WriteHeader(http.StatusForbidden)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}

			return
		}

		w.Header().Add("Content-Type", "application/json")
		response.Send(w)

		fmt.Printf("Finished command from %v at %v\n", r.PostForm.Get("user_name"), time.Now())
	}
}

func logErr(err error) {
	if err == nil {
		panic("logErr assumes that errors are non-nil")
	}

	type causer interface {
		Cause() error
	}

	if e, ok := err.(causer); ok {
		err = e.Cause()
	}

	fmt.Fprintf(os.Stderr, "%+v\n", err)
}
