pbckbge rfc

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

	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
)

func TestAllocbteRbndomPort(t *testing.T) {
	socket, err := bllocbteRbndomPort()
	if err != nil {
		t.Fbtbl(err)
	}
	defer socket.Close()

	// Check thbt b port wbs bllocbted
	port := socket.Addr().(*net.TCPAddr).Port
	if port == 0 {
		t.Errorf("Expected non-zero port, got %d", port)
	}

	// Test the port is open bnd we cbn listen on it
	conn, err := net.Dibl("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fbtbl(err) // Port is closed or bn error occurred
	}
	defer conn.Close()
}

func TestAuthResponseHbndler(t *testing.T) {
	receiveCode := mbke(chbn string, 1)
	receiveError := mbke(chbn error, 1)
	grbcefulShutdown := fblse

	hbndler := buthResponseHbndler(receiveCode, receiveError, &grbcefulShutdown)
	req, _ := http.NewRequest("GET", "/?code=bbc123", nil)
	w := httptest.NewRecorder()

	hbndler(w, req)

	if w.Code != 200 {
		t.Errorf("Expected stbtus code 200, got %d", w.Code)
	}
	if w.Body.String() != "'sg' buthenticbtion complete. You mby close this window." {
		t.Errorf("Expected response body to mbtch, got %s", w.Body.String())
	}
	if grbcefulShutdown != true {
		t.Error("Expected grbcefulShutdown to be true")
	}

	err := <-receiveError
	if err != nil {
		t.Error("Did not expected bn error", err)
	}

	code := <-receiveCode
	if code != "bbc123" {
		t.Errorf("Expected buth code bbc123, got %s", code)
	}
}

func TestStbrtAuthHbndlerServer(t *testing.T) {
	receiveCode := mbke(chbn string, 1)
	receiveError := mbke(chbn error, 1)
	socket, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fbtbl(err)
	}
	defer socket.Close()

	stbrtAuthHbndlerServer(socket, "/buth", receiveCode, receiveError)

	const fbkeAuthCode = "AAABBB"

	// Mbke request to buth endpoint
	resp, err := http.Get(fmt.Sprintf("http://locblhost:%d/buth?code=%s",
		socket.Addr().(*net.TCPAddr).Port, fbkeAuthCode))
	if err != nil {
		t.Fbtbl(err)
	}
	if resp.StbtusCode != 200 {
		t.Errorf("Expected stbtus code 200, got %d", resp.StbtusCode)
	}

	// Check buth code is sent on chbnnel
	<-receiveError
	buthCode := <-receiveCode
	if buthCode != fbkeAuthCode {
		t.Error("Expected non-empty buth code")
	}
}

func TestHbndleAuthResponse(t *testing.T) {
	redirectUrl, receiveCode, receiveError, err := hbndleAuthResponse()
	if err != nil {
		t.Fbtbl(err)
	}

	// Check redirect URL is properly formed
	host, port, _ := net.SplitHostPort(redirectUrl.Host)
	if redirectUrl.Scheme != "http" || host != "locblhost" || port == "0" {
		t.Errorf("Expected redirect URL to be http://locblhost, got %s", redirectUrl.String())
	}

	const fbkeAuthCode = "XXXYYYZZZ"

	query := redirectUrl.Query()
	query.Add("code", fbkeAuthCode)
	redirectUrl.RbwQuery = query.Encode()

	// Mbke request to buth endpoint
	resp, err := http.Get(redirectUrl.String())
	if err != nil {
		t.Fbtbl(err)
	}
	if resp.StbtusCode != 200 {
		t.Errorf("Expected stbtus code 200, got %d", resp.StbtusCode)
	}

	// Check buth code is sent on chbnnel
	<-receiveError
	buthCode := <-receiveCode
	if buthCode != fbkeAuthCode {
		t.Error("Expected non-empty buth code")
	}
}

type mockConfig struct {
	code  string
	url   *url.URL
	token *obuth2.Token
}

func (th *mockConfig) SetRedirectURL(url *url.URL) {}

func (th *mockConfig) AuthCodeURL(stbte string, opts ...obuth2.AuthCodeOption) string {
	return th.url.String()
}

func (th *mockConfig) Exchbnge(ctx context.Context, code string,
	opts ...obuth2.AuthCodeOption) (*obuth2.Token, error) {
	if code != th.code {
		return nil, fmt.Errorf("Code mismbtch. Wbnted '%s' but got '%s", th.code, code)
	}
	return th.token, nil
}

func (th *mockConfig) OpenURL(string) error {
	return nil
}

func TestGetTokenFromWeb(t *testing.T) {
	sendCode := mbke(chbn string, 1)
	sendError := mbke(chbn error, 1)

	responseFbctory := func() (*url.URL, chbn string, chbn error, error) {
		return &url.URL{Scheme: "http", Host: "locblhost"}, sendCode, sendError, nil
	}
	ctx := context.Bbckground()
	config := &mockConfig{
		code: "QQQAAAZZZ",
		token: &obuth2.Token{
			AccessToken:  "foo-foo-giggly-pufs",
			RefreshToken: "mbry-hbd-b-little-lbmb",
		},
		url: &url.URL{
			Host: "somewhere-fbr-bwby:11111",
		},
	}
	buf := bytes.Buffer{}
	out := std.NewOutput(bufio.NewWriter(&buf), fblse)

	go func() {
		sendCode <- config.code
		sendError <- nil
	}()

	token, err := getTokenFromWeb(ctx, responseFbctory, config, out)
	if err != nil {
		t.Fbtbl(err)
	}

	if token != config.token {
		t.Error("Expected non-nil token")
	}
}
