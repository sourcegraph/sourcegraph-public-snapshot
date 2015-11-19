package github

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/kr/pretty"
	"github.com/sourcegraph/go-github/github"
	"golang.org/x/oauth2"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestGetGitHubUserEmailAddresses(t *testing.T) {
	// Mock GitHub
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()
	gh := newGitHubClient(server.URL, "")

	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "Bearer beyangtoken" {
			w.Write([]byte(`{
  "login": "beyang",
  "email": "beyang@beyang.org",
  "name": "Beyang Liu"
}`))
		} else if authHeader == "Bearer othertoken" {
			w.Write([]byte(`{
  "login": "sqs",
  "email": "qslack@qslack.com",
  "name": "Quinn Slack"
}`))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	})

	mux.HandleFunc("/user/emails", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "Bearer beyangtoken" {
			w.Write([]byte(`[
  {
    "email": "beyang@beyang.org",
    "verified": true,
    "primary": true
  },
  {
    "email": "beyang@private.com",
    "verified": true,
    "primary": false
  }
]`))
		} else if authHeader == "Bearer othertoken" {
			w.Write([]byte(`[
  {
    "email": "sqs@obama.com",
    "verified": true,
    "primary": true
  }
]`))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("Unhandled requests: %+v", r.URL)
		w.WriteHeader(http.StatusInternalServerError)
	})

	// Unauthorized request should get public email and emails from
	// activity stream.
	userLogin := "beyang"
	ghuserEmail := "beyang@beyang.org"
	ghuserName := "Beyang Liu"
	ghuser := &github.User{
		Login: &userLogin,
		Email: &ghuserEmail,
		Name:  &ghuserName,
	}
	emails, err := GetGitHubUserEmailAddresses(ghuser, gh, true)
	if err != nil {
		t.Errorf("Failed to update emails from GitHub: %v", err)
	} else {
		expEmails := []*sourcegraph.EmailAddr{{Email: "beyang@beyang.org"}}
		if !reflect.DeepEqual(emails, expEmails) {
			t.Errorf("Unauthed: Expected emails did not match actual: %s", strings.Join(pretty.Diff(expEmails, emails), "\n"))
		}
	}

	// Authorized request should get private emails if/only-if the
	// authorization is for the user requested.
	ghAuth := newGitHubClient(server.URL, "beyangtoken")
	emails, err = GetGitHubUserEmailAddresses(ghuser, ghAuth, true)
	if err != nil {
		t.Errorf("Failed to update emails from GitHub: %v", err)
	} else {
		expEmails := []*sourcegraph.EmailAddr{{Email: "beyang@beyang.org", Verified: true, Primary: true}, {Email: "beyang@private.com", Verified: true}}
		if !reflect.DeepEqual(emails, expEmails) {
			t.Errorf("Authed: Expected emails did not match actual: %s", strings.Join(pretty.Diff(expEmails, emails), "\n"))
		}
	}

	ghAuthOther := newGitHubClient(server.URL, "othertoken")
	emails, err = GetGitHubUserEmailAddresses(ghuser, ghAuthOther, true)
	if err != nil {
		t.Errorf("Failed to update emails from GitHub: %v", err)
	} else {
		expEmails := []*sourcegraph.EmailAddr{{Email: "beyang@beyang.org"}}
		if !reflect.DeepEqual(emails, expEmails) {
			t.Errorf("Authed as usersDiff user: Expected emails did not match actual: %s", strings.Join(pretty.Diff(expEmails, emails), "\n"))
		}
	}
}

func newGitHubClient(serverURL string, accessToken string) (gh *github.Client) {
	if accessToken == "" {
		gh = github.NewClient(nil)
	} else {
		c := oauth2.NewClient(oauth2.NoContext, oauth2.StaticTokenSource(&oauth2.Token{TokenType: "Bearer", AccessToken: accessToken}))
		gh = github.NewClient(c)
	}
	gh.BaseURL, _ = url.Parse(serverURL)
	return gh
}
