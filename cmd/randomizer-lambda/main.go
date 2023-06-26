// The randomizer-lambda command is an AWS Lambda handler that serves the Slack
// slash command API for the randomizer.
//
// The handler expects HTTP request events using the [Amazon API Gateway
// payload format version 2.0]. This makes it suitable for invocation through a
// [Lambda function URL], or through an AWS Lambda proxy integration in an
// Amazon API Gateway HTTP API.
//
// See the randomizer repository README for more information on configuring and
// deploying the randomizer on AWS Lambda.
//
// [Amazon API Gateway payload format version 2.0]: https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-develop-integrations-lambda.html
// [Lambda function URL]: https://docs.aws.amazon.com/lambda/latest/dg/lambda-urls.html
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"go.alexhamlin.co/randomizer/internal/slack"
	"go.alexhamlin.co/randomizer/internal/store/dynamodb"
)

func main() {
	tokenProvider, err := slack.TokenProviderFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to configure Slack token: %+v\n", err)
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
	lambda.Start(httpadapter.NewV2(app).ProxyWithContext)
}
