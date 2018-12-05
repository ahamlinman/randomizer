package store // import "go.alexhamlin.co/randomizer/pkg/store"

import (
	"os"

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
func FactoryFromEnv() (func(string) randomizer.Store, error) {
	if envHasAny("DYNAMODB", "DYNAMODB_TABLE", "DYNAMODB_ENDPOINT") {
		return dynamostore.FactoryFromEnv()
	}

	return boltstore.FactoryFromEnv()
}

func envHasAny(names ...string) bool {
	for _, name := range names {
		if _, ok := os.LookupEnv(name); ok {
			return true
		}
	}
	return false
}
