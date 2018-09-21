package store

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

	var (
		dynamo         = os.Getenv("DYNAMODB")
		dynamoTable    = os.Getenv("DYNAMODB_TABLE")
		dynamoEndpoint = os.Getenv("DYNAMODB_ENDPOINT")
	)
	if dynamo != "" || dynamoTable != "" || dynamoEndpoint != "" {
		fmt.Fprintln(debug, "Using DynamoDB for storage")
		return dynamoFactory(debug, dynamoTable, dynamoEndpoint)
	}

	fmt.Fprintln(debug, "Using Bolt for storage")
	return boltFactory(debug, os.Getenv("DB_PATH"))
}

func dynamoFactory(debug io.Writer, table, endpoint string) (Factory, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return nil, errors.Wrap(err, "loading AWS config")
	}

	cfg.HTTPClient = &http.Client{Timeout: 2500 * time.Millisecond}
	cfg.Retryer = aws.DefaultRetryer{NumMaxRetries: 2}

	if endpoint != "" {
		fmt.Fprintln(debug, "\t-> Using endpoint", endpoint)
		cfg.EndpointResolver = aws.ResolveWithEndpointURL(endpoint)
	} else {
		fmt.Fprintln(debug, "\t-> Using AWS production endpoint")
	}

	db := dynamodb.New(cfg)

	if table != "" {
		fmt.Fprintln(debug, "\t-> Using table", table)
	} else {
		fmt.Fprintln(debug, "\t-> Using default table")
	}

	return func(partition string) randomizer.Store {
		var options []dynamostore.Option

		if table == "" {
			options = []dynamostore.Option{
				dynamostore.WithPartition(partition),
			}
		} else {
			options = []dynamostore.Option{
				dynamostore.WithTable(table),
				dynamostore.WithPartition(partition),
			}
		}

		return dynamostore.New(db, options...)
	}, nil
}

func boltFactory(debug io.Writer, path string) (Factory, error) {
	if path == "" {
		path = "randomizer.db"
	}
	fmt.Fprintln(debug, "\t-> Using database", path)

	db, err := bolt.Open(path, os.ModePerm&0644, nil)
	if err != nil {
		return nil, err
	}

	return func(partition string) randomizer.Store {
		return boltstore.New(db, boltstore.WithBucketName(partition))
	}, nil
}
