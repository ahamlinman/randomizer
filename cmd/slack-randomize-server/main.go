package main // import "go.alexhamlin.co/randomizer/cmd/slack-randomize-server"

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"

	"go.alexhamlin.co/randomizer/pkg/randomizer"
	"go.alexhamlin.co/randomizer/pkg/slack"
	boltstore "go.alexhamlin.co/randomizer/pkg/store/bbolt"
	dynamostore "go.alexhamlin.co/randomizer/pkg/store/dynamodb"
)

type storeFunc func(string) randomizer.Store

func main() {
	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "SLACK_TOKEN must be provided")
		os.Exit(2)
	}

	getStoreFunc := getBoltStoreFunc
	if _, ok := os.LookupEnv("DYNAMODB_TABLE"); ok {
		getStoreFunc = getDynamoDBStoreFunc
	}

	getStore, err := getStoreFunc()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}

	http.HandleFunc("/", rootHandler(token, getStore))

	fmt.Println("Starting randomizer service")
	err = http.ListenAndServe("0.0.0.0:7636", nil)
	if err != nil {
		logErr(errors.Wrap(err, "starting server"))
		os.Exit(1)
	}
}

func rootHandler(token string, getStore storeFunc) http.HandlerFunc {
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

		app := randomizer.NewApp("/randomize", getStore(r.PostForm.Get("channel_id")))
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

func getBoltStoreFunc() (storeFunc, error) {
	fmt.Println("Using BoltDB for storage")

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "randomizer.db"
	}
	fmt.Printf("...with database file %q\n", dbPath)

	db, err := bolt.Open(dbPath, os.ModePerm&0644, nil)
	if err != nil {
		return nil, err
	}

	return func(channel string) randomizer.Store {
		return boltstore.New(
			db,
			boltstore.WithBucketName(channel),
		)
	}, nil
}

func getDynamoDBStoreFunc() (storeFunc, error) {
	fmt.Println("Using DynamoDB for storage")

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return nil, err
	}

	table, ok := os.LookupEnv("DYNAMODB_TABLE")
	if !ok {
		panic(errors.New("DynamoDB stores currently require a table"))
	}
	fmt.Printf("...with table %q\n", table)

	db := dynamodb.New(cfg)

	return func(channel string) randomizer.Store {
		return dynamostore.New(
			db,
			dynamostore.WithTable(table),
			dynamostore.WithPartition(channel),
		)
	}, nil
}
