package slack

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"go.alexhamlin.co/randomizer/internal/randomizer"
	"go.alexhamlin.co/randomizer/internal/randomizer/rndtest"
)

func TestValidRequests(t *testing.T) {
	store := make(rndtest.Store)
	app := App{
		TokenProvider: StaticToken("right"),
		StoreFactory:  func(_ string) randomizer.Store { return store },
	}

	headers := http.Header{"Content-Type": {"application/x-www-form-urlencoded"}}
	params := makeTestParams("/save test one two")
	body := strings.NewReader(params.Encode())

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header = headers
	app.ServeHTTP(resp, req)

	if resp.Result().StatusCode != http.StatusOK {
		t.Errorf("invalid status: got %v, want %v", resp.Result().StatusCode, http.StatusOK)
	}
	if len(store) < 1 {
		t.Error("/save command failed to save a new group in the store")
	}
}

func TestInvalidMethod(t *testing.T) {
	app := App{
		TokenProvider: StaticToken("right"),
		StoreFactory:  func(_ string) randomizer.Store { return rndtest.Store(nil) },
	}

	params := makeTestParams("one two three")
	params.Set("token", "right")
	reqURL := url.URL{Path: "/", RawQuery: params.Encode()}

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, reqURL.String(), nil)
	app.ServeHTTP(resp, req)

	if resp.Result().StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("invalid status: got %v, want %v", resp.Result().StatusCode, http.StatusMethodNotAllowed)
	}
	if resp.Result().Header.Get("Allow") != http.MethodPost {
		t.Errorf("invalid Accept header: got %q, want %q", resp.Result().Header.Get("Allow"), http.MethodPost)
	}
}

func TestInvalidToken(t *testing.T) {
	app := App{
		TokenProvider: StaticToken("right"),
		StoreFactory:  func(_ string) randomizer.Store { return rndtest.Store(nil) },
	}

	headers := http.Header{"Content-Type": {"application/x-www-form-urlencoded"}}
	params := makeTestParams("help")
	params.Set("token", "wrong")

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
