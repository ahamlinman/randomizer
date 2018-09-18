/*

The dynamodb-provision command requests the creation of an Amazon DynamoDB
table using the randomizer's schema.

AWS configuration is read from the environment and files as described at
https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html.

To view the available options for provisioning, run with the "-help" flag.

*/
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	dynamostore "go.alexhamlin.co/randomizer/pkg/store/dynamodb"
)

func main() {
	readCap := flag.Int64("readcap", 10, "number of read capacity units to provision")
	writeCap := flag.Int64("writecap", 10, "number of write capacity units to provision")
	table := flag.String("table", dynamostore.DefaultTable, "name of the table to provision")
	endpoint := flag.String("endpoint", "", "endpoint URL to use for DynamoDB")
	flag.Parse()

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not load AWS config: %v\n", err)
		os.Exit(2)
	}

	if *endpoint != "" {
		cfg.EndpointResolver = aws.ResolveWithEndpointURL(*endpoint)
	}

	db := dynamodb.New(cfg)
	store := dynamostore.New(db, dynamostore.WithTable(*table))

	err = store.Provision(*readCap, *writeCap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "provisioning failed: %+v\n", err)
		os.Exit(1)
	}

	fmt.Println("provisioning request successful")
}
