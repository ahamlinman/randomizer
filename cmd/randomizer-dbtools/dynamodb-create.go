package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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

	_, err := db.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
		TableName: &dynamoDBTable,
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: &partitionKey,
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: &groupKey,
				KeyType:       types.KeyTypeRange,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: &partitionKey,
				AttributeType: types.ScalarAttributeTypeS, // String
			},
			{
				AttributeName: &groupKey,
				AttributeType: types.ScalarAttributeTypeS, // String
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  &createReadCap,
			WriteCapacityUnits: &createWriteCap,
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "creation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("creation request successful")
}
