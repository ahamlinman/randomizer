/*

The randomizer-lambda command provides an AWS Lambda handler for the randomizer
as a Slack Slash Command, using Amazon API Gateway's proxy mode.

*/
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"go.alexhamlin.co/randomizer/internal/slack"
	"go.alexhamlin.co/randomizer/internal/store/dynamodb"
)

func main() {
	var tokenProvider slack.TokenProvider
	if token, ok := os.LookupEnv("SLACK_TOKEN"); ok {
		tokenProvider = slack.StaticToken(token)
	} else if path, ok := os.LookupEnv("SLACK_TOKEN_SSM_PATH"); ok {
		tokenProvider = slack.AWSParameter(path, 2*time.Minute)
	} else {
		fmt.Fprintln(os.Stderr, "must define SLACK_TOKEN or SLACK_TOKEN_SSM_PATH")
		os.Exit(2)
	}

	storeFactory, err := dynamodb.FactoryFromEnv(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create store: %+v\n", err)
		os.Exit(2)
	}

	app := slack.App{
		TokenProvider: tokenProvider,
		StoreFactory:  storeFactory,
		DebugWriter:   os.Stderr,
	}
	lambda.Start(httpadapter.New(app).ProxyWithContext)
}
