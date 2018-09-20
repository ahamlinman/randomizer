package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

func main() {
	dbFile := flag.String("dbfile", "", "name of the BoltDB file to migrate from")
	dynamoTable := flag.String("tablename", "", "name of the DynamoDB table to migrate to")
	endpoint := flag.String("endpoint", "", "endpoint URL to use for DynamoDB")
	flag.Parse()

	if *dbFile == "" || *dynamoTable == "" {
		flag.Usage()
		os.Exit(2)
	}

	boltDB, err := bolt.Open(*dbFile, os.ModePerm&0644, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open Bolt database: %v\n", err)
		os.Exit(2)
	}

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not load AWS config: %v\n", err)
		os.Exit(2)
	}
	if *endpoint != "" {
		cfg.EndpointResolver = aws.ResolveWithEndpointURL(*endpoint)
	}
	dynamoDB := dynamodb.New(cfg)

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
			*dynamoTable: writeRequests,
		},
	}
	_, err = dynamoDB.BatchWriteItemRequest(&input).Send()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not write to DynamoDB: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("batch write sent")
}
