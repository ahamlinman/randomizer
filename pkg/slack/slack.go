package slack // import "go.alexhamlin.co/randomizer/pkg/slack"

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"

	"go.alexhamlin.co/randomizer/pkg/randomizer"
	"go.alexhamlin.co/randomizer/pkg/store"
)

// Request represents a request for a Slack slash command. Its values can be
// obtained from query string parameters or the request body, depending on the
// configuration of the command in Slack.
type Request struct {
	Token     string
	SSLCheck  string
	ChannelID string
	Text      string
}

// ResponseType represents the manner in which a response to a Slack slash
// command request will be displayed to a user.
type ResponseType int

const (
	// TypeEphemeral causes a response to be displayed to the user only. The
	// slash command invocation will be hidden from others.
	TypeEphemeral ResponseType = iota + 1

	// TypeInChannel causes a response to be displayed in the channel to other
	// Slack users, along with the slash command invocation that triggered it.
	TypeInChannel
)

// MarshalText encodes a ResponseType into the textual representation
// understood by the Slack API.
func (t ResponseType) MarshalText() ([]byte, error) {
	switch t {
	case TypeEphemeral:
		return []byte("ephemeral"), nil

	case TypeInChannel:
		return []byte("in_channel"), nil
	}

	panic(fmt.Errorf("unknown response type code %v", t))
}

// Response represents a response to a Slack slash command.
type Response struct {
	Type ResponseType `json:"response_type"`
	Text string       `json:"text"`
}

// Send writes the JSON form of a Response to the provided writer. If the
// Response is empty, nothing will be written.
func (r Response) Send(w io.Writer) {
	if r == (Response{}) {
		return
	}

	json.NewEncoder(w).Encode(&r)
}

// ErrIncorrectToken indicates that the authentication token provided in the
// request did not match the expected value.
var ErrIncorrectToken = errors.New("invalid authentication token")

// App represents a randomizer app that interacts with the Slack slash command
// API. It provides functionality for validating authentication tokens and
// returning responses in Slack's expected formats.
type App struct {
	// Name is the name of the command as displayed in help output.
	Name string
	// Token is the expected value of the Slack authentication token.
	Token []byte
	// StoreFactory provides a Store for the Slack channel in which the request
	// was made.
	StoreFactory store.Factory
}

// Run processes a slash command request from Slack and returns an appropriate
// response. If the request token is invalid, ErrIncorrectToken will be
// returned.
func (a App) Run(r Request) (Response, error) {
	// This function "[requires] careful thought to use correctly." So hopefully
	// I managed to do that.
	if subtle.ConstantTimeCompare(a.Token, []byte(r.Token)) != 1 {
		return Response{}, ErrIncorrectToken
	}

	if r.SSLCheck == "1" {
		return Response{}, nil
	}

	app := randomizer.NewApp(a.Name, a.StoreFactory(r.ChannelID))
	result, err := app.Main(strings.Split(r.Text, " "))
	if err != nil {
		return Response{
			Type: TypeEphemeral,
			Text: err.(randomizer.Error).HelpText(),
		}, nil
	}

	switch result.Type() {
	case randomizer.ListedGroups, randomizer.ShowedGroup:
		return Response{
			Type: TypeEphemeral,
			Text: result.Message(),
		}, nil

	default:
		return Response{
			Type: TypeInChannel,
			Text: result.Message(),
		}, nil
	}
}
