package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/ahamlinman/randomizer/pkg/randomizer"
	"github.com/ahamlinman/randomizer/pkg/slack"
	boltstore "github.com/ahamlinman/randomizer/pkg/store/bbolt"
	bolt "github.com/coreos/bbolt"
	"github.com/pkg/errors"
)

func main() {
	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "SLACK_TOKEN must be provided")
		os.Exit(2)
	}

	db, err := bolt.Open("randomizer.db", os.ModePerm&0644, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}

	app := randomizer.NewApp("/randomize", boltstore.New(db))

	handler := func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		reqToken := r.PostForm.Get("token")
		// TODO: Vulnerable to timing attacks, if it actually matters
		// (I might fix it just for fun)
		if reqToken != token {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.Header().Add("Content-Type", "application/json")

		result, err := app.Main(strings.Split(r.PostForm.Get("text"), " "))
		if err != nil {
			slack.Response{
				Type: slack.TypeEphemeral,
				Text: err.(randomizer.Error).HelpText(),
			}.Send(w)
			return
		}

		switch result.Type() {
		case randomizer.ListedGroups, randomizer.ShowedGroup:
			slack.Response{
				Type: slack.TypeEphemeral,
				Text: result.Message(),
			}.Send(w)

		default:
			slack.Response{
				Type: slack.TypeInChannel,
				Text: result.Message(),
			}.Send(w)
		}
	}

	http.HandleFunc("/", handler)

	fmt.Println("Starting randomizer service")
	err = http.ListenAndServe("0.0.0.0:7636", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", errors.Wrap(err, "starting server"))
		os.Exit(1)
	}
}
