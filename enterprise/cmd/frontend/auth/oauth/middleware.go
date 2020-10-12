package oauth

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/schema"
	"golang.org/x/oauth2"
)

func NewHandler(serviceType, authPrefix string, isAPIHandler bool, next http.Handler) http.Handler {
	oauthFlowHandler := http.StripPrefix(authPrefix, newOAuthFlowHandler(serviceType))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delegate to the auth flow handler
		if !isAPIHandler && strings.HasPrefix(r.URL.Path, authPrefix+"/") {
			r = withOAuthExternalHTTPClient(r)
			oauthFlowHandler.ServeHTTP(w, r)
			return
		}

		// If the actor is authenticated and not performing an OAuth flow, then proceed to
		// next.
		if actor.FromContext(r.Context()).IsAuthenticated() {
			next.ServeHTTP(w, r)
			return
		}

		// If there is only one auth provider configured, the single auth provider is a OAuth
		// instance, and it's an app request, redirect to signin immediately. The user wouldn't be
		// able to do anything else anyway; there's no point in showing them a signin screen with
		// just a single signin option.
		if pc := getExactlyOneOAuthProvider(); pc != nil && !isAPIHandler && pc.AuthPrefix == authPrefix && isHuman(r) {
			v := make(url.Values)
			v.Set("redirect", auth.SafeRedirectURL(r.URL.String()))
			v.Set("pc", pc.ConfigID().ID)
			http.Redirect(w, r, authPrefix+"/login?"+v.Encode(), http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func newOAuthFlowHandler(serviceType string) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/login", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		id := req.URL.Query().Get("pc")
		p := getProvider(serviceType, id)
		if p == nil {
			log15.Error("no OAuth provider found with ID and service type", "id", id, "serviceType", serviceType)
			http.Error(w, "Misconfigured GitHub auth provider.", http.StatusInternalServerError)
			return
		}
		p.Login.ServeHTTP(w, req)
	}))
	mux.Handle("/callback", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		state, err := DecodeState(req.URL.Query().Get("state"))
		if err != nil {
			http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not decode OAuth state from URL parameter.", http.StatusBadRequest)
			return
		}

		p := getProvider(serviceType, state.ProviderID)
		if p == nil {
			log15.Error("OAuth failed: in callback, no auth provider found with ID and service type", "id", state.ProviderID, "serviceType", serviceType)
			http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not find provider that matches the OAuth state parameter.", http.StatusBadRequest)
			return
		}
		p.Callback.ServeHTTP(w, req)
	}))
	return mux
}

// withOAuthExternalHTTPClient updates client such that the
// golang.org/x/oauth2 package will use our http client which is configured
// with proxy and TLS settings/etc.
func withOAuthExternalHTTPClient(r *http.Request) *http.Request {
	client := httpcli.ExternalHTTPClient()
	if traceLogEnabled {
		if httpClient, ok := client.(*http.Client); ok {
			loggingClient := *httpClient
			loggingClient.Transport = &loggingRoundTripper{underlying: loggingClient.Transport}
			ctx := context.WithValue(r.Context(), oauth2.HTTPClient, &loggingClient)
			return r.WithContext(ctx)
		} else {
			log15.Warn("Could not initialize logging HTTP client, because external HTTP client was not an instance of *http.Client")
		}
	}
	ctx := context.WithValue(r.Context(), oauth2.HTTPClient, client)
	return r.WithContext(ctx)
}

var traceLogEnabled, _ = strconv.ParseBool(env.Get("INSECURE_OAUTH2_LOG_TRACES", "false", "Log all OAuth2-related HTTP requests and responses. Only use during testing because the log messages will contain sensitive data."))

type loggingRoundTripper struct {
	underlying http.RoundTripper
}

func (l *loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var body string
	if req.Body != nil {
		if bodyBytes, err := ioutil.ReadAll(req.Body); err == nil {
			req.Body.Close()
			req.Body = ioutil.NopCloser(strings.NewReader(string(bodyBytes)))
			body = string(bodyBytes)
		} else {
			log15.Error("Unexpected error in OAuth2 debug log", "operation", "reading request body", "error", err)
		}
	}
	if len(body) > 1000 {
		body = body[:1000]
	}
	log.Printf(">>>>> HTTP Request: %s %s\n      Header: %v\n      Body: %s", req.Method, req.URL.String(), req.Header, body)

	resp, err := l.underlying.RoundTrip(req)
	if err != nil {
		log.Printf("<<<<< Error getting HTTP response: %s", err)
	} else {
		var respBody string
		if resp.Body != nil {
			if respBodyBytes, err := ioutil.ReadAll(resp.Body); err == nil {
				resp.Body.Close()
				resp.Body = ioutil.NopCloser(strings.NewReader(string(respBodyBytes)))
				respBody = string(respBodyBytes)
			} else {
				log15.Error("Unexpected error in OAuth2 debug log", "operation", "reading response body", "error", err)
			}
		}
		if len(respBody) > 1000 {
			respBody = respBody[:1000]
		}
		log.Printf("<<<<< HTTP Response: %s %s\n      Header: %v\n      Body: %s", req.Method, req.URL.String(), resp.Header, string(respBody))
	}
	return resp, err
}

func getExactlyOneOAuthProvider() *Provider {
	ps := providers.Providers()
	if len(ps) != 1 {
		return nil
	}
	p, ok := ps[0].(*Provider)
	if !ok {
		return nil
	}
	if !isOAuth(p.Config()) {
		return nil
	}
	return p
}

var isOAuths []func(p schema.AuthProviders) bool

func AddIsOAuth(f func(p schema.AuthProviders) bool) {
	isOAuths = append(isOAuths, f)
}

func isOAuth(p schema.AuthProviders) bool {
	for _, f := range isOAuths {
		if f(p) {
			return true
		}
	}
	return false
}

// isHuman returns true if the request probably came from a human, rather than a bot. Used to
// prevent unfurling the wrong URL preview.
func isHuman(req *http.Request) bool {
	return strings.Contains(strings.ToLower(req.UserAgent()), "mozilla")
}
