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

var dynamoCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a DynamoDB table with the randomizer schema",
	Long: `Create a DynamoDB table with the randomizer schema.

Note that a successful exit only means that DynamoDB has received the request
to provision the table. The table itself may still take some time to create
before it can be used.`,
	Run: runDynamoDBCreate,
}

var (
	createReadCap  int64
	createWriteCap int64
)

func init() {
	dynamoCreateCmd.Flags().Int64VarP(
		&createReadCap,
		"readcap", "r", 10,
		"number of read capacity units to provision",
	)

	dynamoCreateCmd.Flags().Int64VarP(
		&createWriteCap,
		"writecap", "w", 10,
		"number of write capacity units to provision",
	)

	dynamoDBCmd.AddCommand(dynamoCreateCmd)
}

func runDynamoDBCreate(cmd *cobra.Command, args []string) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not load AWS config: %v\n", err)
		os.Exit(2)
	}

	if dynamoDBEndpoint != "" {
		cfg.EndpointResolver = aws.ResolveWithEndpointURL(dynamoDBEndpoint)
	}

	db := dynamodb.New(cfg)
	store := dynamostore.New(db, dynamostore.WithTable(dynamoDBTable))

	err = store.CreateTable(createReadCap, createWriteCap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "creation failed: %+v\n", err)
		os.Exit(1)
	}

	fmt.Println("creation request successful")
}
