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

var (
	token        []byte
	name         string
	storeFactory store.Factory
)

func init() {
	token = []byte(os.Getenv("SLACK_TOKEN"))
	if len(token) == 0 {
		panic(errors.New("missing SLACK_TOKEN"))
	}

	name = os.Getenv("SLACK_COMMAND_NAME")
	if name == "" {
		name = "/randomize"
	}

	var err error
	storeFactory, err = store.DynamoDBFactoryFromEnv(os.Stderr)
	if err != nil {
		panic(err)
	}
}

func main() {
	lambda.Start(handleEvent)
}

func handleEvent(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	app := slack.App{
		Name:         name,
		Token:        token,
		StoreFactory: storeFactory,
		DebugWriter:  os.Stderr,
	}

	return handlerfunc.New(app.ServeHTTP).Proxy(request)
}
