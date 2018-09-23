package store // import "go.alexhamlin.co/randomizer/pkg/store"

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"

	"go.alexhamlin.co/randomizer/pkg/randomizer"
	boltstore "go.alexhamlin.co/randomizer/pkg/store/bbolt"
	dynamostore "go.alexhamlin.co/randomizer/pkg/store/dynamodb"
)

// Factory represents a type for functions that produce a store for the
// randomizer to use for a given "partition" (e.g. Slack channel).
//
// A Factory may panic if it requires a non-empty partition and no partition is
// given.
type Factory func(partition string) randomizer.Store

// FactoryFromEnv constructs and returns a Factory based on available
// environment variables.
//
// By default, a store backed by a local Bolt database (using the CoreOS
// "bbolt" fork) is returned, using the database file from the DB_PATH
// environment variable. If this variable is unset, the file "randomizer.db" in
// the current directory will be used. The database file is automatically
// created if it does not yet exist.
//
// If the DYNAMODB, DYNAMODB_TABLE, or DYNAMODB_ENDPOINT environment variables
// are set, a store backed by an Amazon DynamoDB table is returned.
// Requirements for the table are described by the dynamodb package in this
// module. AWS configuration is read as described at
// https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html.
func FactoryFromEnv(debug io.Writer) (Factory, error) {
	if debug == nil {
		debug = ioutil.Discard
	}

	if envHasAny("DYNAMODB", "DYNAMODB_TABLE", "DYNAMODB_ENDPOINT") {
		return DynamoDBFactoryFromEnv(debug)
	}

	return BoltFactoryFromEnv(debug)
}

func envHasAny(names ...string) bool {
	for _, name := range names {
		if _, ok := os.LookupEnv(name); ok {
			return true
		}
	}
	return false
}

// DynamoDBFactoryFromEnv constructs and returns a Factory for DynamoDB-backed
// stores based on available environment variables. See FactoryFromEnv for more
// information.
func DynamoDBFactoryFromEnv(debug io.Writer) (Factory, error) {
	return dynamoDBFactory(
		debug,
		os.Getenv("DYNAMODB_TABLE"),
		os.Getenv("DYNAMODB_ENDPOINT"),
	)
}

func dynamoDBFactory(debug io.Writer, table, endpoint string) (Factory, error) {
	fmt.Fprintln(debug, "Using DynamoDB for storage")

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return nil, errors.Wrap(err, "loading AWS config")
	}

	cfg.HTTPClient = &http.Client{Timeout: 2500 * time.Millisecond}
	cfg.Retryer = aws.DefaultRetryer{NumMaxRetries: 2}

	if endpoint != "" {
		fmt.Fprintf(debug, "-> Endpoint: %s\n", endpoint)
		cfg.EndpointResolver = aws.ResolveWithEndpointURL(endpoint)
	} else {
		fmt.Fprintf(debug, "-> Endpoint: AWS %s\n", cfg.Region)
	}

	db := dynamodb.New(cfg)

	if table == "" {
		table = "RandomizerGroups"
	}
	fmt.Fprintf(debug, "-> Table: %s\n", table)

	return func(partition string) randomizer.Store {
		store, err := dynamostore.New(db, table, partition)
		if err != nil {
			panic(err)
		}
		return store
	}, nil
}

// BoltFactoryFromEnv constructs and returns a Factory for Bolt-backed stores
// based on available environment variables. See FactoryFromEnv for more
// information.
func BoltFactoryFromEnv(debug io.Writer) (Factory, error) {
	return boltFactory(debug, os.Getenv("DB_PATH"))
}

func boltFactory(debug io.Writer, path string) (Factory, error) {
	fmt.Fprintln(debug, "Using Bolt for storage")

	if path == "" {
		path = "randomizer.db"
	}
	fmt.Fprintf(debug, "-> Database: %s\n", path)

	db, err := bolt.Open(path, os.ModePerm&0644, nil)
	if err != nil {
		return nil, err
	}

	return func(partition string) randomizer.Store {
		store, err := boltstore.New(db, partition)
		if err != nil {
			panic(err)
		}
		return store
	}, nil
}
