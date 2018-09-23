package main // import "go.alexhamlin.co/randomizer/cmd/slack-lambda-handler"

import (
	"errors"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/handlerfunc"

	"go.alexhamlin.co/randomizer/pkg/slack"
	"go.alexhamlin.co/randomizer/pkg/store"
)

func main() {
	lambda.Start(handleEvent)
}

func handleEvent(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		return events.APIGatewayProxyResponse{}, errors.New("missing SLACK_TOKEN")
	}

	name := os.Getenv("SLACK_COMMAND_NAME")
	if name == "" {
		name = "/randomize"
	}

	storeFactory, err := store.DynamoDBFactoryFromEnv(os.Stderr)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	app := slack.App{
		Name:         name,
		Token:        []byte(token),
		StoreFactory: storeFactory,
		DebugWriter:  os.Stderr,
	}

	return handlerfunc.New(app.ServeHTTP).Proxy(request)
}
