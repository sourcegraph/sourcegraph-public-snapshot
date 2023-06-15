package rfc

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

func TestAllocateRandomPort(t *testing.T) {
	socket, err := allocateRandomPort()
	if err != nil {
		t.Fatal(err)
	}
	defer socket.Close()

	// Check that a port was allocated
	port := socket.Addr().(*net.TCPAddr).Port
	if port == 0 {
		t.Errorf("Expected non-zero port, got %d", port)
	}

	// Check that we can listen on the port
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatal(err)
	}
	listener.Close()
}

func TestAuthResponseHandler(t *testing.T) {
	sendCode := make(chan string, 1)
	gracefulShutdown := false

	handler := authResponseHandler(sendCode, &gracefulShutdown)
	req, _ := http.NewRequest("GET", "/?code=abc123", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}
	if w.Body.String() != "'sg' authentication complete. You may close this window." {
		t.Errorf("Expected response body to match, got %s", w.Body.String())
	}
	if gracefulShutdown != true {
		t.Error("Expected gracefulShutdown to be true")
	}
	code := <-sendCode
	if code != "abc123" {
		t.Errorf("Expected auth code abc123, got %s", code)
	}
}

func TestStartAuthHandlerServer(t *testing.T) {
	sendCode := make(chan string, 1)
	socket, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer socket.Close()

	startAuthHandlerServer(socket, "/auth", sendCode)

	const fakeAuthCode = "AAABBB"

	// Make request to auth endpoint
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/auth?code=%s",
		socket.Addr().(*net.TCPAddr).Port, fakeAuthCode))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	// Check auth code is sent on channel
	authCode := <-sendCode
	if authCode != fakeAuthCode {
		t.Error("Expected non-empty auth code")
	}
}

func TestHandleAuthResponse(t *testing.T) {
	redirectUrl, sendCode, err := handleAuthResponse()
	if err != nil {
		t.Fatal(err)
	}

	// Check redirect URL is properly formed
	host, port, _ := net.SplitHostPort(redirectUrl.Host)
	if redirectUrl.Scheme != "http" || host != "localhost" || port == "0" {
		t.Errorf("Expected redirect URL to be http://localhost, got %s", redirectUrl.String())
	}

	const fakeAuthCode = "XXXYYYZZZ"

	query := redirectUrl.Query()
	query.Add("code", fakeAuthCode)
	redirectUrl.RawQuery = query.Encode()

	// Make request to auth endpoint
	resp, err := http.Get(redirectUrl.String())
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	// Check auth code is sent on channel
	authCode := <-sendCode
	if authCode != fakeAuthCode {
		t.Error("Expected non-empty auth code")
	}
}

type mockConfig struct {
	code  string
	url   *url.URL
	token *oauth2.Token
}

func (th *mockConfig) SetRedirectURL(url *url.URL) {}

func (th *mockConfig) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return th.url.String()
}

func (th *mockConfig) Exchange(ctx context.Context, code string,
	opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	if code != th.code {
		return nil, fmt.Errorf("Code mismatch. Wanted '%s' but got '%s", th.code, code)
	}
	return th.token, nil
}

func TestGetTokenFromWeb(t *testing.T) {
	sendCode := make(chan string, 1)

	responseFactory := func() (*url.URL, chan string, error) {
		return &url.URL{Scheme: "http", Host: "localhost"}, sendCode, nil
	}
	ctx := context.Background()
	config := &mockConfig{
		code: "QQQAAAZZZ",
		token: &oauth2.Token{
			AccessToken:  "foo-foo-giggly-pufs",
			RefreshToken: "mary-had-a-little-lamb",
		},
		url: &url.URL{
			Host: "somewhere-far-away:11111",
		},
	}
	buf := bytes.Buffer{}
	out := std.NewOutput(bufio.NewWriter(&buf), false)

	go func() {
		sendCode <- config.code
	}()

	token, err := getTokenFromWeb(responseFactory, ctx, config, out)
	if err != nil {
		t.Fatal(err)
	}

	if token != config.token {
		t.Error("Expected non-nil token")
	}
}
