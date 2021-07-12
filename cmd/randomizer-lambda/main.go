/*

The randomizer-lambda command provides an AWS Lambda handler for the randomizer
as a Slack Slash Command, using Amazon API Gateway's proxy mode.

*/
package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"go.alexhamlin.co/randomizer/internal/slack"
	"go.alexhamlin.co/randomizer/internal/store/dynamodb"
)

func main() {
	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "SLACK_TOKEN must be provided in environment")
		os.Exit(2)
	}

	storeFactory, err := dynamodb.FactoryFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create store: %+v\n", err)
		os.Exit(2)
	}

	app := slack.App{
		Token:        []byte(token),
		StoreFactory: storeFactory,
		DebugWriter:  os.Stderr,
	}
	lambda.Start(httpadapter.New(app).ProxyWithContext)
}
