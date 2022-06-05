/*

The randomizer-lambda command provides an AWS Lambda handler for the randomizer
as a Slack Slash Command, using Amazon API Gateway's proxy mode.

*/
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

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

	handler := handler{
		slack.App{
			TokenProvider: tokenProvider,
			StoreFactory:  storeFactory,
			DebugWriter:   os.Stderr,
		},
	}
	lambda.Start(handler.Handle)
}

type handler struct {
	slack.App
}

func (h handler) Handle(ctx context.Context, event json.RawMessage) (any, error) {
	var version struct {
		Version string `json:"version"`
	}
	json.Unmarshal(event, &version) // Assume that Lambda's JSON input is valid.

	if strings.Split(version.Version, ".")[0] == "2" {
		return proxyHTTP(ctx, event, httpadapter.NewV2(h.App).ProxyWithContext)
	}
	return proxyHTTP(ctx, event, httpadapter.New(h.App).ProxyWithContext)
}

func proxyHTTP[In, Out any](
	ctx context.Context,
	event json.RawMessage,
	httpHandler func(context.Context, In) (Out, error),
) (any, error) {
	var input In
	if err := json.Unmarshal(event, &input); err != nil {
		return nil, err
	}
	return httpHandler(ctx, input)
}
