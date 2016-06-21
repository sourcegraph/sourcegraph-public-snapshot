package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httputil/httpctx"
)

// Session is the information stored in a session cookie.
type Session struct {
	// AccessToken is the user's access token. It's obtained from the
	// server using OAuth2 when the user logs in or signs up.
	AccessToken string
}

// sessionCookieName is the name of the session cookie.
const sessionCookieName = "session-oauth2-token"

// ErrNoSession indicates that there is no session cookie sent in the
// HTTP request.
var ErrNoSession = errors.New("no session cookie")

// OnlySecureCookies indicates whether or not the
// secure flag should be set for all issued cookies.
func OnlySecureCookies(ctx context.Context) bool {
	return conf.AppURL(ctx).Scheme == "https"
}

// ReadSessionCookie reads the session from the HTTP request. If there
// is no session cookie, ErrNoSession is returned.
func ReadSessionCookie(req *http.Request) (*Session, error) {
	sessionCookie, err := req.Cookie(sessionCookieName)
	if err == http.ErrNoCookie {
		return nil, ErrNoSession
	}
	if err != nil {
		return nil, err
	}
	return readSessionCookie(sessionCookie)
}

// ReadSessionCookieFromResponse reads the session from an HTTP
// response. If there is no session cookie, ErrNoSession is returned.
func ReadSessionCookieFromResponse(resp *http.Response) (*Session, error) {
	for _, c := range resp.Cookies() {
		if c.Name == sessionCookieName {
			return readSessionCookie(c)
		}
	}
	return nil, ErrNoSession
}

func readSessionCookie(c *http.Cookie) (*Session, error) {
	decoded, err := base64.StdEncoding.DecodeString(c.Value)
	if err != nil {
		return nil, err
	}

	var s Session
	if err := json.Unmarshal(decoded, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// NewSessionCookie creates a new session cookie with the given
// session information.
func NewSessionCookie(s Session, isSecure bool) (*http.Cookie, error) {
	encoded, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return &http.Cookie{
		Name:     sessionCookieName,
		Value:    base64.StdEncoding.EncodeToString(encoded),
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(time.Hour * 24 * 365 * 2),
		Secure:   isSecure,
	}, nil
}

// WriteSessionCookie writes the session cookie to the HTTP response.
func WriteSessionCookie(w http.ResponseWriter, s Session, isSecure bool) error {
	sc, err := NewSessionCookie(s, isSecure)
	if err != nil {
		return err
	}
	http.SetCookie(w, sc)
	return nil
}

// DeleteSessionCookie deletes the session cookie by sending a
// Set-Cookie header in the HTTP response to immediately expire it on
// the client.
func DeleteSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   sessionCookieName,
		Path:   "/",
		MaxAge: -1,
	})
}

// CookieMiddleware is an http.Handler middleware that authenticates
// future API requests using the OAuth2 access token from the user's
// cookie (if any). It performs no validation or authentication of the
// access token; it merely causes it to be passed along verbatim in
// outgoing API requests.
func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if sess, err := ReadSessionCookie(r); err == nil {
			ctx := httpctx.FromRequest(r)
			ctx = sourcegraph.WithCredentials(ctx, oauth2.StaticTokenSource(&oauth2.Token{TokenType: "Bearer", AccessToken: sess.AccessToken}))
			httpctx.SetForRequest(r, ctx)

			// Vary based on Authorization header if the request is
			// operating with any level of authorization, so that the
			// response can't be cached and mixed in with unauthorized
			// responses in an HTTP cache.
			w.Header().Add("vary", "Authorization")
		} else if err != ErrNoSession {
			log.Printf("%s %s: Error checking request auth info: %s (will delete session cookie).", r.Method, r.URL.RequestURI(), err)
			DeleteSessionCookie(w)
		}
		next.ServeHTTP(w, r)
	})
}
