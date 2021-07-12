package dynamodb

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	xrayawsv2 "github.com/aws/aws-xray-sdk-go/instrumentation/awsv2"
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

func awsConfigFromEnv() (aws.Config, error) {
	options := []func(*config.LoadOptions) error{
		config.WithHTTPClient(&http.Client{Timeout: 2500 * time.Millisecond}),
		config.WithRetryer(
			func() aws.Retryer {
				return retry.AddWithMaxAttempts(retry.NewStandard(), 2)
			},
		),
	}

	if endpoint := os.Getenv("DYNAMODB_ENDPOINT"); endpoint != "" {
		options = append(options,
			config.WithEndpointResolver(aws.EndpointResolverFunc(
				func(_, _ string) (aws.Endpoint, error) {
					return aws.Endpoint{URL: endpoint}, nil
				},
			)),
		)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), options...)
	if err != nil {
		return aws.Config{}, errors.Wrap(err, "loading AWS config")
	}

	if useXRay := os.Getenv("DYNAMODB_XRAY_TRACING"); useXRay == "1" {
		xrayawsv2.AWSV2Instrumentor(&cfg.APIOptions)
	}

	return cfg, nil
}

func tableFromEnv() string {
	if table := os.Getenv("DYNAMODB_TABLE"); table != "" {
		return table
	}

	return "RandomizerGroups"
}
