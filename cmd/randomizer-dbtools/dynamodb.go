package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/spf13/cobra"
)

var dynamoDBCmd = &cobra.Command{
	Use:   "dynamodb",
	Short: "Work with DynamoDB-based randomizer stores",
	Long: `Work with DynamoDB-based randomizer stores.

AWS configuration is read from the environment and files as described at
https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html.`,
}

var (
	dynamoDBTable    string
	dynamoDBEndpoint string
)

func init() {
	dynamoDBCmd.PersistentFlags().StringVarP(
		&dynamoDBTable,
		"table", "t", "RandomizerGroups",
		"name of the DynamoDB table to work with",
	)

	dynamoDBCmd.PersistentFlags().StringVarP(
		&dynamoDBEndpoint,
		"endpoint", "e", "",
		"endpoint URL for DynamoDB API requests",
	)

	rootCmd.AddCommand(dynamoDBCmd)
}

func getDynamoDB() *dynamodb.Client {
	var options []func(*config.LoadOptions) error
	if dynamoDBEndpoint != "" {
		options = append(options,
			config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(_, _ string, _ ...any) (aws.Endpoint, error) {
					return aws.Endpoint{URL: dynamoDBEndpoint}, nil
				},
			)),
		)
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), options...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not load AWS config: %v\n", err)
		os.Exit(2)
	}

	return dynamodb.NewFromConfig(cfg)
}
