package main // import "go.alexhamlin.co/randomizer/cmd/slack-lambda-handler"

import (
	"errors"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

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

	if os.Getenv("DYNAMODB_TABLE") == "" {
		return events.APIGatewayProxyResponse{}, errors.New("missing DYNAMODB_TABLE")
	}

	name := os.Getenv("SLACK_COMMAND_NAME")
	if name == "" {
		name = "/randomize"
	}

	storeFactory, err := store.FactoryFromEnv(os.Stderr)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	app := slack.App{
		Name:         name,
		Token:        []byte(token),
		StoreFactory: storeFactory,
	}

	result, err := app.Run(slack.Request{
		Token:     request.QueryStringParameters["token"],
		SSLCheck:  request.QueryStringParameters["ssl_check"],
		ChannelID: request.QueryStringParameters["channel_id"],
		Text:      request.QueryStringParameters["text"],
	})

	if err == slack.ErrIncorrectToken {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusForbidden,
		}, nil
	} else if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	if result == (slack.Response{}) {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       result.String(),
	}, nil
}
