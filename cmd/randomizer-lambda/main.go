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

	"github.com/aws/aws-lambda-go/events"
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

func (h handler) Handle(ctx context.Context, event json.RawMessage) (interface{}, error) {
	var version struct {
		Version string `json:"version"`
	}
	mustUnmarshal(event, &version)

	if strings.HasPrefix(version.Version, "2.") {
		var httpEvent events.APIGatewayV2HTTPRequest
		mustUnmarshal(event, &httpEvent)
		return httpadapter.NewV2(h.App).ProxyWithContext(ctx, httpEvent)
	}

	var httpEvent events.APIGatewayProxyRequest
	mustUnmarshal(event, &httpEvent)
	return httpadapter.New(h.App).ProxyWithContext(ctx, httpEvent)
}

func mustUnmarshal(data []byte, v interface{}) {
	if err := json.Unmarshal(data, v); err != nil {
		panic(err)
	}
}
