package main // import "go.alexhamlin.co/randomizer/cmd/slack-randomize-server"

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	bolt "github.com/coreos/bbolt"
	"github.com/pkg/errors"

	"go.alexhamlin.co/randomizer/pkg/randomizer"
	"go.alexhamlin.co/randomizer/pkg/slack"
	boltstore "go.alexhamlin.co/randomizer/pkg/store/bbolt"
)

func main() {
	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "SLACK_TOKEN must be provided")
		os.Exit(2)
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		fmt.Println("Using default DB path 'randomizer.db'")
		dbPath = "randomizer.db"
	}

	db, err := bolt.Open(dbPath, os.ModePerm&0644, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}

	http.HandleFunc("/", rootHandler(token, db))

	fmt.Println("Starting randomizer service")
	err = http.ListenAndServe("0.0.0.0:7636", nil)
	if err != nil {
		logErr(errors.Wrap(err, "starting server"))
		os.Exit(1)
	}
}

func logErr(err error) {
	fmt.Fprintf(os.Stderr, "%+v\n", errors.Cause(err))
}

func rootHandler(token string, db *bolt.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			logErr(err)
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

		app := randomizer.NewApp(
			"/randomize",
			boltstore.New(db, boltstore.WithBucketName(r.PostForm.Get("channel_id"))),
		)
		result, err := app.Main(strings.Split(r.PostForm.Get("text"), " "))
		if err != nil {
			logErr(err)
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

		fmt.Printf("Handled command from %v at %v\n", r.PostForm.Get("user_name"), time.Now())
	}
}
