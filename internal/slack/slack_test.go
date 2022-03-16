package slack

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"go.alexhamlin.co/randomizer/internal/randomizer"
	"go.alexhamlin.co/randomizer/internal/randomizer/rndtest"
)

func TestValidRequests(t *testing.T) {
	app := App{TokenProvider: StaticToken("right")}
	params := makeTestParams("/save test one two")

	for _, method := range []string{http.MethodGet, http.MethodPost} {
		t.Run(method, func(t *testing.T) {
			store := make(rndtest.Store)
			app.StoreFactory = func(_ string) randomizer.Store { return store }

			requestURL := url.URL{Path: "/"}
			if method == http.MethodGet {
				requestURL.RawQuery = params.Encode()
			}

			var body io.Reader
			headers := make(http.Header)
			if method == http.MethodPost {
				headers.Add("Content-Type", "application/x-www-form-urlencoded")
				body = strings.NewReader(params.Encode())
			}

			resp := httptest.NewRecorder()
			req := httptest.NewRequest(method, requestURL.String(), body)
			req.Header = headers
			app.ServeHTTP(resp, req)

			if resp.Result().StatusCode != http.StatusOK {
				t.Errorf("invalid status: got %v, want %v", resp.Result().StatusCode, http.StatusOK)
			}
			if len(store) < 1 {
				t.Error("/save command failed to save a new group in the store")
			}
		})
	}
}

func TestInvalidToken(t *testing.T) {
	app := App{
		TokenProvider: StaticToken("right"),
		StoreFactory:  func(_ string) randomizer.Store { return rndtest.Store(nil) },
	}

	params := makeTestParams("help")
	params.Set("token", "wrong")

	headers := make(http.Header)
	headers.Add("Content-Type", "application/x-www-form-urlencoded")

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(params.Encode()))
	req.Header = headers
	app.ServeHTTP(resp, req)

	if resp.Result().StatusCode != http.StatusForbidden {
		t.Errorf("wrong status for invalid token: got %v, want %v", resp.Result().StatusCode, http.StatusForbidden)
	}
}

func makeTestParams(text string) url.Values {
	params := make(url.Values)
	params.Add("token", "right")
	params.Add("channel_id", "C12345678")
	params.Add("command", "/randomize")
	params.Add("text", text)
	return params
}
