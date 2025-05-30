// Package slack supports invoking the randomizer as a Slack slash command.
package slack

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/ahamlinman/randomizer/internal/randomizer"
)

// App provides HTTP handling logic that allows the randomizer to be integrated
// as a slash command in a Slack workspace.
//
// Note that App currently only supports static verification tokens to check
// that a request legitimately originated from Slack, and does not support the
// newer signed secrets functionality.
type App struct {
	// TokenProvider provides the expected value of the slash command verification
	// token generated by Slack. This can be obtained from the slash command
	// configuration.
	TokenProvider TokenProvider
	// StoreFactory provides a Store for the Slack channel in which the request
	// was made.
	StoreFactory func(partition string) randomizer.Store
	// Logger, if non-nil, will be used to report information about errors that
	// occur while handling each request.
	Logger *slog.Logger
}

// ServeHTTP handles the POST request that Slack makes to invoke the randomizer
// slash command.
func (a App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Add("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		a.logErr(err, "Failed to read POST form")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tokenIsValid, err := a.isTokenValid(r.Context(), r.PostForm)
	if err != nil {
		a.logErr(err, "Failed to validate token")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !tokenIsValid {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if r.PostForm.Get("ssl_check") == "1" {
		return
	}

	result, err := a.runRandomizer(r.Context(), r.PostForm)
	if err != nil {
		a.logErr(err, "Failed to run randomizer")
		a.writeError(w, err)
		return
	}

	a.writeResult(w, result)
}

func (a App) isTokenValid(ctx context.Context, params url.Values) (ok bool, _ error) {
	gotToken := params.Get("token")
	wantToken, err := a.TokenProvider(ctx)
	if err != nil {
		return false, err
	}

	subtle.WithDataIndependentTiming(func() {
		ok = subtle.ConstantTimeCompare([]byte(gotToken), []byte(wantToken)) == 1
	})
	return
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

type response struct {
	Type responseType `json:"response_type"`
	Text string       `json:"text"`
}

type responseType string

const (
	typeEphemeral responseType = "ephemeral"
	typeInChannel responseType = "in_channel"
)

func (a App) writeResult(w http.ResponseWriter, result randomizer.Result) {
	rtype := typeEphemeral
	switch result.Type() {
	case randomizer.Selection, randomizer.SavedGroup, randomizer.DeletedGroup:
		rtype = typeInChannel
	}

	a.writeResponse(w, response{
		Text: result.Message(),
		Type: rtype,
	})
}

func (a App) writeError(w http.ResponseWriter, err error) {
	a.writeResponse(w, response{
		Text: err.(randomizer.Error).HelpText(),
		Type: typeEphemeral,
	})
}

func (a App) writeResponse(w http.ResponseWriter, response response) {
	w.Header().Add("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		a.logErr(err, "Failed to write response")
	}
}

func (a App) logErr(err error, msg string) {
	if a.Logger != nil {
		a.Logger.Error(msg, "err", err)
	}
}
