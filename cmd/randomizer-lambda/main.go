package main // import "go.alexhamlin.co/randomizer/cmd/randomizer-lambda"

import (
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/handlerfunc"

	"go.alexhamlin.co/randomizer/pkg/slack"
	"go.alexhamlin.co/randomizer/pkg/store"
)

func main() {
	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "SLACK_TOKEN must be provided")
		os.Exit(2)
	}

	storeFactory, err := store.DynamoDBFactoryFromEnv(os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create store: %+v\n", err)
		os.Exit(2)
	}

	app := slack.App{
		Token:        []byte(token),
		StoreFactory: storeFactory,
		DebugWriter:  os.Stderr,
	}

	proxy := handlerfunc.New(app.ServeHTTP)

	lambda.Start(
		func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			return proxy.Proxy(req)
		},
	)
}
