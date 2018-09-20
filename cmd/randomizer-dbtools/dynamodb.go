package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/spf13/cobra"

	dynamostore "go.alexhamlin.co/randomizer/pkg/store/dynamodb"
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
		"table", "t", dynamostore.DefaultTable,
		"name of the DynamoDB table to work with",
	)

	dynamoDBCmd.PersistentFlags().StringVarP(
		&dynamoDBEndpoint,
		"endpoint", "e", "",
		"endpoint URL for DynamoDB API requests",
	)

	rootCmd.AddCommand(dynamoDBCmd)
}

func getDynamoDB() *dynamodb.DynamoDB {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not load AWS config: %v\n", err)
		os.Exit(2)
	}

	if dynamoDBEndpoint != "" {
		cfg.EndpointResolver = aws.ResolveWithEndpointURL(dynamoDBEndpoint)
	}

	return dynamodb.New(cfg)
}
