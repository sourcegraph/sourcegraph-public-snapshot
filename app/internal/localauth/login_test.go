package localauth

import (
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-querystring/query"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	appauth "src.sourcegraph.com/sourcegraph/app/auth"
	"src.sourcegraph.com/sourcegraph/app/internal/apptest"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
)

// TestLogIn_disabled_404 tests that the login endpoint returns 404s
// when auth is disabled.
func TestLogIn_disabled_404(t *testing.T) {
	authutil.ActiveFlags.Source = "none"
	defer func() {
		authutil.ActiveFlags = authutil.Flags{}
	}()

	c, _ := apptest.New()

	for _, method := range []string{"GET", "POST"} {
		req, _ := http.NewRequest(method, router.Rel.URLTo(router.LogIn).String(), nil)
		resp, err := c.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if want := http.StatusNotFound; resp.StatusCode != want {
			t.Errorf("%s: got HTTP %d, want %d", method, resp.StatusCode, want)
		}
	}
}

func TestLogIn_form(t *testing.T) {
	authutil.ActiveFlags.Source = "local"
	defer func() {
		authutil.ActiveFlags = authutil.Flags{}
	}()

	c, _ := apptest.New()

	if _, err := c.GetOK(router.Rel.URLTo(router.LogIn).String()); err != nil {
		t.Fatal(err)
	}
}

func TestLogIn_submit_validPassword(t *testing.T) {
	authutil.ActiveFlags.Source = "local"
	defer func() {
		authutil.ActiveFlags = authutil.Flags{}
	}()

	c, mock := apptest.New()

	frm := sourcegraph.LoginCredentials{Login: "u", Password: "valid"}
	data, err := query.Values(frm)
	if err != nil {
		t.Fatal(err)
	}

	var calledAuthGetAccessToken, calledAuthIdentify, calledUsersGet bool
	mock.Auth.GetAccessToken_ = func(ctx context.Context, op *sourcegraph.AccessTokenRequest) (*sourcegraph.AccessTokenResponse, error) {
		if !reflect.DeepEqual(*op.ResourceOwnerPassword, frm) {
			t.Errorf("got form == %+v, want %+v", op, frm)
		}
		calledAuthGetAccessToken = true
		return &sourcegraph.AccessTokenResponse{AccessToken: "k"}, nil
	}
	mock.Auth.Identify_ = func(ctx context.Context, _ *pbtypes.Void) (*sourcegraph.AuthInfo, error) {
		calledAuthIdentify = true
		return &sourcegraph.AuthInfo{UID: 123}, nil
	}
	mock.Users.Get_ = func(ctx context.Context, userSpec *sourcegraph.UserSpec) (*sourcegraph.User, error) {
		calledUsersGet = true
		return &sourcegraph.User{Login: "u"}, nil
	}

	resp, err := c.PostFormNoFollowRedirects(router.Rel.URLTo(router.LogIn).String(), data)
	if err != nil {
		t.Fatal(err)
	}

	// Check redirected to user page.
	if want := http.StatusSeeOther; resp.StatusCode != want {
		t.Errorf("got HTTP %d, want %d", resp.StatusCode, want)
	}
	if want, got := router.Rel.URLToUser("u").String(), resp.Header.Get("location"); got != want {
		t.Errorf("got Location %q, want %q", got, want)
	}

	// Check that user session cookie is set.
	cookie, err := appauth.ReadSessionCookieFromResponse(resp)
	if err != nil {
		t.Fatal(err)
	}
	if want := (&appauth.Session{AccessToken: "k"}); !reflect.DeepEqual(cookie, want) {
		t.Errorf("got cookie %+v, want %+v", cookie, want)
	}

	if !calledAuthGetAccessToken {
		t.Error("!calledAuthGetAccessToken")
	}
	if !calledAuthIdentify {
		t.Error("!calledAuthIdentify")
	}
	if !calledUsersGet {
		t.Error("!calledUsersGet")
	}
}

func TestLogIn_submit_userNotFound(t *testing.T) {
	authutil.ActiveFlags.Source = "local"
	defer func() {
		authutil.ActiveFlags = authutil.Flags{}
	}()

	c, mock := apptest.New()

	frm := sourcegraph.LoginCredentials{Login: "u", Password: "p"}
	data, err := query.Values(frm)
	if err != nil {
		t.Fatal(err)
	}

	var calledAuthGetAccessToken bool
	mock.Auth.GetAccessToken_ = func(ctx context.Context, op *sourcegraph.AccessTokenRequest) (*sourcegraph.AccessTokenResponse, error) {
		calledAuthGetAccessToken = true
		return nil, grpc.Errorf(codes.NotFound, "user not found")
	}

	resp, err := c.PostFormNoFollowRedirects(router.Rel.URLTo(router.LogIn).String(), data)
	if err != nil {
		t.Fatal(err)
	}

	// Check that login form is re-rendered.
	if want := http.StatusOK; resp.StatusCode != want {
		t.Errorf("got HTTP %d, want %d", resp.StatusCode, want)
	}

	// Check that user session cookie is NOT set.
	if _, err := appauth.ReadSessionCookieFromResponse(resp); err != appauth.ErrNoSession {
		t.Fatalf("got err %v, want ErrNoSession", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(string(body), formErrorNoUserExists) {
		t.Error("form error not found")
	}

	if !calledAuthGetAccessToken {
		t.Error("!calledAuthGetAccessToken")
	}
}

func TestLogIn_submit_badPassword(t *testing.T) {
	authutil.ActiveFlags.Source = "local"
	defer func() {
		authutil.ActiveFlags = authutil.Flags{}
	}()

	c, mock := apptest.New()

	frm := sourcegraph.LoginCredentials{Login: "u", Password: "bad"}
	data, err := query.Values(frm)
	if err != nil {
		t.Fatal(err)
	}

	var calledAuthGetAccessToken bool
	mock.Auth.GetAccessToken_ = func(ctx context.Context, op *sourcegraph.AccessTokenRequest) (*sourcegraph.AccessTokenResponse, error) {
		calledAuthGetAccessToken = true
		return nil, grpc.Errorf(codes.PermissionDenied, "bad password")
	}

	resp, err := c.PostFormNoFollowRedirects(router.Rel.URLTo(router.LogIn).String(), data)
	if err != nil {
		t.Fatal(err)
	}

	// Check that login form is re-rendered.
	if want := http.StatusOK; resp.StatusCode != want {
		t.Errorf("got HTTP %d, want %d", resp.StatusCode, want)
	}

	// Check that user session cookie is NOT set.
	if _, err := appauth.ReadSessionCookieFromResponse(resp); err != appauth.ErrNoSession {
		t.Fatalf("got err %v, want ErrNoSession", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(string(body), formErrorWrongPassword) {
		t.Error("form error not found")
	}

	if !calledAuthGetAccessToken {
		t.Error("!calledAuthGetAccessToken")
	}
}
