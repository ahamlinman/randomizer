package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var dynamoImportBoltCmd = &cobra.Command{
	Use:   "import-bolt",
	Short: "Import a BoltDB database into a DynamoDB table",
	Long: `Import a BoltDB database into a DynamoDB table.

Note that the DynamoDB table must already have been created with the randomizer
schema, e.g. by using "randomizer-dbtools dynamodb create".
`,
	Run: runDynamoDBImportBolt,
}

var boltDBFile string

func init() {
	dynamoImportBoltCmd.Flags().StringVarP(
		&boltDBFile,
		"file", "f", "",
		"location of the BoltDB file to import (required)",
	)
	dynamoImportBoltCmd.MarkFlagRequired("file")

	dynamoDBCmd.AddCommand(dynamoImportBoltCmd)
}

func runDynamoDBImportBolt(cmd *cobra.Command, args []string) {
	boltDB, err := bolt.Open(boltDBFile, os.ModePerm&0644, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open Bolt database: %v\n", err)
		os.Exit(2)
	}

	dynamoDB := getDynamoDB()

	writeRequests := make([]types.WriteRequest, 0)
	err = boltDB.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(partition []byte, b *bolt.Bucket) error {
			return b.ForEach(func(group, itemsGob []byte) error {
				var (
					partitionStr = string(partition)
					groupStr     = string(group)
				)

				var items []string
				decoder := gob.NewDecoder(bytes.NewReader(itemsGob))
				err := decoder.Decode(&items)
				if err != nil {
					return fmt.Errorf("decoding items for %q in %q: %w", groupStr, partitionStr, err)
				}

				writeRequests = append(writeRequests, types.WriteRequest{
					PutRequest: &types.PutRequest{
						Item: map[string]types.AttributeValue{
							"Partition": &types.AttributeValueMemberS{Value: partitionStr},
							"Group":     &types.AttributeValueMemberS{Value: groupStr},
							"Items":     &types.AttributeValueMemberSS{Value: items},
						},
					},
				})

				return nil
			})
		})
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not read from Bolt DB: %v\n", err)
		os.Exit(1)
	}

	_, err = dynamoDB.BatchWriteItem(context.Background(), &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			dynamoDBTable: writeRequests,
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not write to DynamoDB: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("batch write sent")
}
