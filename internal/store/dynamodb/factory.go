package dynamodb

import (
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/pkg/errors"

	"go.alexhamlin.co/randomizer/internal/randomizer"
)

// FactoryFromEnv returns a store.Factory whose stores are backed by Amazon
// DynamoDB.
//
// AWS configuration is read as described at
// https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html.
func FactoryFromEnv() (func(string) randomizer.Store, error) {
	cfg, err := awsConfigFromEnv()
	if err != nil {
		return nil, err
	}

	db := dynamodb.New(cfg)
	table := tableFromEnv()

	return func(partition string) randomizer.Store {
		store, err := New(db, table, partition)
		if err != nil {
			panic(err)
		}
		return store
	}, nil
}

func awsConfigFromEnv() (aws.Config, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return aws.Config{}, errors.Wrap(err, "loading AWS config")
	}

	cfg.HTTPClient = &http.Client{Timeout: 2500 * time.Millisecond}
	cfg.Retryer = aws.DefaultRetryer{NumMaxRetries: 2}

	if endpoint := os.Getenv("DYNAMODB_ENDPOINT"); endpoint != "" {
		cfg.EndpointResolver = aws.ResolveWithEndpointURL(endpoint)
	}

	return cfg, nil
}

func tableFromEnv() string {
	if table := os.Getenv("DYNAMODB_TABLE"); table != "" {
		return table
	}

	return "RandomizerGroups"
}
