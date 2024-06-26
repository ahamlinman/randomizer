package dynamodb

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/ahamlinman/randomizer/internal/awsconfig"
	"github.com/ahamlinman/randomizer/internal/randomizer"
)

// FactoryFromEnv returns a store.Factory whose stores are backed by Amazon
// DynamoDB.
//
// AWS configuration is read as described at
// https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html.
func FactoryFromEnv(ctx context.Context) (func(string) randomizer.Store, error) {
	cfg, err := awsConfigFromEnv(ctx)
	if err != nil {
		return nil, err
	}

	db := dynamodb.NewFromConfig(cfg)
	table := tableFromEnv()

	return func(partition string) randomizer.Store {
		store, err := New(db, table, partition)
		if err != nil {
			panic(err)
		}
		return store
	}, nil
}

func awsConfigFromEnv(ctx context.Context) (aws.Config, error) {
	var extraOptions []awsconfig.Option
	if endpoint := os.Getenv("DYNAMODB_ENDPOINT"); endpoint != "" {
		extraOptions = append(extraOptions,
			config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(_, _ string, _ ...any) (aws.Endpoint, error) {
					return aws.Endpoint{URL: endpoint}, nil
				},
			)),
		)
	}

	return awsconfig.New(ctx, extraOptions...)
}

func tableFromEnv() string {
	if table := os.Getenv("DYNAMODB_TABLE"); table != "" {
		return table
	}
	return "RandomizerGroups"
}
