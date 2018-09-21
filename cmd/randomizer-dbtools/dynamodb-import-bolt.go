package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/pkg/errors"
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

	writeRequests := make([]dynamodb.WriteRequest, 0)
	err = boltDB.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(partition []byte, b *bolt.Bucket) error {
			return b.ForEach(func(group, itemsGob []byte) error {
				var items []string
				decoder := gob.NewDecoder(bytes.NewReader(itemsGob))
				if err := decoder.Decode(&items); err != nil {
					return errors.Wrapf(err, "decoding items for %q in %q", string(group), string(partition))
				}

				var (
					partitionS = string(partition)
					groupS     = string(group)
				)

				writeRequests = append(writeRequests, dynamodb.WriteRequest{
					PutRequest: &dynamodb.PutRequest{
						Item: map[string]dynamodb.AttributeValue{
							"Partition": {S: &partitionS},
							"Group":     {S: &groupS},
							"Items":     {SS: items},
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

	input := dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]dynamodb.WriteRequest{
			dynamoDBTable: writeRequests,
		},
	}
	_, err = dynamoDB.BatchWriteItemRequest(&input).Send()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not write to DynamoDB: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("batch write sent")
}