package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/spf13/cobra"
)

var dynamoCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a DynamoDB table with the randomizer schema",
	Long: `Create a DynamoDB table with the randomizer schema.

Note that a successful exit only means that DynamoDB has received the request
to create the table. It may take some time for the table to become usable.

This tool does not support DynamoDB's on-demand capacity mode.`,
	Run: runDynamoDBCreate,
}

var (
	createReadCap  int64
	createWriteCap int64
)

func init() {
	dynamoCreateCmd.Flags().Int64VarP(
		&createReadCap,
		"readcap", "r", 1,
		"number of read capacity units to provision",
	)

	dynamoCreateCmd.Flags().Int64VarP(
		&createWriteCap,
		"writecap", "w", 1,
		"number of write capacity units to provision",
	)

	dynamoDBCmd.AddCommand(dynamoCreateCmd)
}

func runDynamoDBCreate(cmd *cobra.Command, args []string) {
	db := getDynamoDB()

	var (
		partitionKey = "Partition"
		groupKey     = "Group"
	)

	input := dynamodb.CreateTableInput{
		TableName: &dynamoDBTable,

		KeySchema: []dynamodb.KeySchemaElement{
			{
				AttributeName: &partitionKey,
				KeyType:       dynamodb.KeyTypeHash,
			},
			{
				AttributeName: &groupKey,
				KeyType:       dynamodb.KeyTypeRange,
			},
		},

		AttributeDefinitions: []dynamodb.AttributeDefinition{
			{
				AttributeName: &partitionKey,
				AttributeType: dynamodb.ScalarAttributeTypeS, // String
			},
			{
				AttributeName: &groupKey,
				AttributeType: dynamodb.ScalarAttributeTypeS, // String
			},
		},

		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  &createReadCap,
			WriteCapacityUnits: &createWriteCap,
		},
	}

	if _, err := db.CreateTableRequest(&input).Send(); err != nil {
		fmt.Fprintf(os.Stderr, "creation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("creation request successful")
}
