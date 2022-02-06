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
	} else if ssmName, ok := os.LookupEnv("SLACK_TOKEN_SSM_NAME"); ok {
		tokenProvider = slack.AWSParameter(ssmName, getSSMTTL())
	} else {
		fmt.Fprintln(os.Stderr, "Must define SLACK_TOKEN or SLACK_TOKEN_SSM_NAME")
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

const defaultSSMTTL = 2 * time.Minute

func getSSMTTL() time.Duration {
	ttlEnv, ok := os.LookupEnv("SLACK_TOKEN_SSM_TTL")
	if !ok {
		return defaultSSMTTL
	}

	ttl, err := time.ParseDuration(ttlEnv)
	if err != nil {
		fmt.Fprintln(os.Stderr, "SLACK_TOKEN_SSM_TTL must be a valid Go duration")
		os.Exit(2)
	}

	return ttl
}
