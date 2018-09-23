package slack // import "go.alexhamlin.co/randomizer/pkg/slack"

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"go.alexhamlin.co/randomizer/pkg/randomizer"
	"go.alexhamlin.co/randomizer/pkg/store"
)

// App provides HTTP handling logic that allows the randomizer to be integrated
// as a slash command in a Slack workspace.
type App struct {
	// Name is the name of the command as displayed in help output. Ideally, it
	// should match the name of the slash command configured in Slack.
	Name string
	// Token is the expected value of the authentication token provided by Slack.
	// This can be obtained from the slash command configuration.
	Token []byte
	// StoreFactory provides a Store for the Slack channel in which the request
	// was made.
	StoreFactory store.Factory
	// LogFunc, if non-nil, will be called to print information about errors that
	// occur while handling each request.
	LogFunc func(format string, v ...interface{})
}

// ServeHTTP handles the GET or POST request made by Slack when the randomizer
// slash command is invoked. (The HTTP method used by Slack can be selected
// when configuring the slash command.)
func (a App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var params url.Values
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			a.log("Bad POST form data: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		params = r.PostForm
	} else {
		params = r.URL.Query()
	}

	// This function "[requires] careful thought to use correctly." So hopefully
	// I managed to do that.
	if subtle.ConstantTimeCompare(a.Token, []byte(params.Get("token"))) != 1 {
		a.log("Invalid token in request\n")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if params.Get("ssl_check") == "1" {
		a.log("Handled SSL check\n")
		return
	}

	app := randomizer.NewApp(a.Name, a.StoreFactory(params.Get("channel_id")))
	result, err := app.Main(strings.Split(params.Get("text"), " "))
	if err != nil {
		a.log("Error from randomizer: %v\n", err.(randomizer.Error).Cause())
		response{typeEphemeral, err.(randomizer.Error).HelpText()}.Send(w)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	switch result.Type() {
	case randomizer.ListedGroups, randomizer.ShowedGroup:
		response{typeEphemeral, result.Message()}.Send(w)

	default:
		response{typeInChannel, result.Message()}.Send(w)
	}
}

func (a App) log(format string, v ...interface{}) {
	if a.LogFunc == nil {
		return
	}

	a.LogFunc(format, v...)
}

type responseType int

const (
	typeEphemeral responseType = iota + 1
	typeInChannel
)

func (t responseType) MarshalText() ([]byte, error) {
	switch t {
	case typeEphemeral:
		return []byte("ephemeral"), nil

	case typeInChannel:
		return []byte("in_channel"), nil
	}

	panic(fmt.Errorf("unknown response type code %v", t))
}

type response struct {
	Type responseType `json:"response_type"`
	Text string       `json:"text"`
}

func (r response) Send(w io.Writer) {
	if r == (response{}) {
		return
	}

	json.NewEncoder(w).Encode(&r)
}
