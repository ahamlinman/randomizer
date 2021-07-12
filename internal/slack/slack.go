package slack

import (
	"context"
	"crypto/subtle"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"

	"go.alexhamlin.co/randomizer/internal/randomizer"
	"go.alexhamlin.co/randomizer/internal/store"
)

// App provides HTTP handling logic that allows the randomizer to be integrated
// as a slash command in a Slack workspace.
type App struct {
	// Token is the expected value of the authentication token provided by Slack.
	// This can be obtained from the slash command configuration.
	Token []byte
	// StoreFactory provides a Store for the Slack channel in which the request
	// was made.
	StoreFactory store.Factory
	// DebugWriter, if non-nil, will be used to print information about errors
	// that occur while handling each request.
	DebugWriter io.Writer
}

// ServeHTTP handles the GET or POST request made by Slack when the randomizer
// slash command is invoked. (The HTTP method used by Slack can be selected
// when configuring the slash command.)
func (a App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params, err := a.getRequestParams(r)
	if err != nil {
		a.log("Failed to get request params: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !a.hasValidToken(params) {
		a.log("Invalid token in request\n")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if params.Get("ssl_check") == "1" {
		a.log("Handled SSL check\n")
		return
	}

	result, err := a.runRandomizer(r.Context(), params)
	if err != nil {
		a.log("Error from randomizer: %v\n", err)
		a.writeError(w, err)
		return
	}

	a.writeResult(w, result)
}

func (a App) getRequestParams(r *http.Request) (url.Values, error) {
	switch r.Method {
	case http.MethodPost:
		// This implicitly assumes an application/x-www-form-urlencoded body, per
		// https://api.slack.com/slash-commands#app_command_handling.
		err := r.ParseForm()
		return r.PostForm, errors.Wrap(err, "reading POST form data")

	case http.MethodGet:
		return r.URL.Query(), nil
	}

	return nil, errors.Errorf("unsupported method %v", r.Method)
}

func (a App) hasValidToken(params url.Values) bool {
	token := params.Get("token")

	// This function "[requires] careful thought to use correctly." Hopefully
	// I've managed to do that.
	return subtle.ConstantTimeCompare(a.Token, []byte(token)) == 1
}

func (a App) runRandomizer(ctx context.Context, params url.Values) (randomizer.Result, error) {
	var (
		name      = params.Get("command")
		channelID = params.Get("channel_id")
		args      = strings.Fields(params.Get("text"))
	)

	app := randomizer.NewApp(name, a.StoreFactory(channelID))
	return app.Main(ctx, args)
}

func (a App) writeResult(w http.ResponseWriter, result randomizer.Result) {
	a.writeResponse(w, response{
		Text: result.Message(),
		Type: slackTypeForResultType[result.Type()],
	})
}

var slackTypeForResultType = map[randomizer.ResultType]responseType{
	randomizer.Selection:    typeInChannel,
	randomizer.SavedGroup:   typeInChannel,
	randomizer.DeletedGroup: typeInChannel,
	// Default to typeEphemeral for everything else, to keep channels clean.
}

func (a App) writeError(w http.ResponseWriter, err error) {
	a.writeResponse(w, response{
		Text: err.(randomizer.Error).HelpText(),
		Type: typeEphemeral,
	})
}

func (a App) writeResponse(w http.ResponseWriter, response response) {
	w.Header().Add("Content-Type", "application/json")

	_, err := response.WriteTo(w)
	if err != nil {
		a.log("Error writing response: %v\n", err)
	}
}

func (a App) log(format string, v ...interface{}) {
	if a.DebugWriter == nil {
		return
	}

	fmt.Fprintf(a.DebugWriter, format, v...)
}
