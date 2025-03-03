package dynamodb

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/ahamlinman/randomizer/internal/awsconfig"
	"github.com/ahamlinman/randomizer/internal/randomizer"
	"github.com/ahamlinman/randomizer/internal/store/registry"
)

func init() {
	registry.Provide("dynamodb", FactoryFromEnv,
		"DYNAMODB", "DYNAMODB_TABLE", "DYNAMODB_ENDPOINT")
}

// FactoryFromEnv returns a store.Factory whose stores are backed by Amazon
// DynamoDB.
//
// AWS configuration is read as described at
// https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html.
func FactoryFromEnv(ctx context.Context) (func(string) randomizer.Store, error) {
	cfg, err := awsconfig.New(ctx)
	if err != nil {
		return nil, err
	}

	table := tableFromEnv()
	db := dynamodb.NewFromConfig(cfg, func(opts *dynamodb.Options) {
		if endpoint := os.Getenv("DYNAMODB_ENDPOINT"); endpoint != "" {
			opts.BaseEndpoint = aws.String(endpoint)
		}
	})

	return func(partition string) randomizer.Store {
		store, err := New(db, table, partition)
		if err != nil {
			panic(err)
		}
		return store
	}, nil
}

func tableFromEnv() string {
	if table := os.Getenv("DYNAMODB_TABLE"); table != "" {
		return table
	}
	return "RandomizerGroups"
}
