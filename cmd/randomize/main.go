package main // import "go.alexhamlin.co/randomizer/cmd/randomize"

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	bolt "go.etcd.io/bbolt"

	"go.alexhamlin.co/randomizer/pkg/randomizer"
	boltstore "go.alexhamlin.co/randomizer/pkg/store/bbolt"
	dynamostore "go.alexhamlin.co/randomizer/pkg/store/dynamodb"
)

func main() {
	store, err := getStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}

	app := randomizer.NewApp("randomize", store)
	result, err := app.Main(os.Args[1:])
	if err != nil {
		err := err.(randomizer.Error)
		fmt.Fprintln(os.Stderr, err.HelpText())
		fmt.Fprintf(os.Stderr, "\n%+v\n", err.Cause())
		os.Exit(1)
	}

	fmt.Println(result.Message())
}

func getStore() (randomizer.Store, error) {
	if endpoint, ok := os.LookupEnv("DYNAMODB"); ok {
		cfg, err := external.LoadDefaultAWSConfig()
		if err != nil {
			return nil, err
		}

		if endpoint != "" {
			cfg.EndpointResolver = aws.ResolveWithEndpointURL(endpoint)
		}

		db := dynamodb.New(cfg)
		return dynamostore.New(db), nil
	}

	db, err := bolt.Open("randomizer.db", os.ModePerm&0644, nil)
	if err != nil {
		return nil, err
	}
	return boltstore.New(db), nil
}
