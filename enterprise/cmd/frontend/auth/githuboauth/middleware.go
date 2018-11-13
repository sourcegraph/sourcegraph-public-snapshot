package githuboauth

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/auth"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

const authPrefix = auth.AuthURLPrefix + "/github"

var Middleware = &auth.Middleware{
	API: func(next http.Handler) http.Handler {
		return newOAuthHandler(true, next)
	},
	App: func(next http.Handler) http.Handler {
		return newOAuthHandler(false, next)
	},
}

func newOAuthHandler(isAPIRequest bool, next http.Handler) http.Handler {
	oauthFlowHandler := http.StripPrefix(authPrefix, newOAuthFlowHandler())
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !ffIsEnabled {
			next.ServeHTTP(w, r)
			return
		}

		// Delegate to the auth flow handler
		if !isAPIRequest && strings.HasPrefix(r.URL.Path, authPrefix+"/") {
			oauthFlowHandler.ServeHTTP(w, r)
			return
		}

		// If the actor is authenticated and not performing an OAuth flow, then proceed to
		// next.
		if actor.FromContext(r.Context()).IsAuthenticated() {
			next.ServeHTTP(w, r)
			return
		}

		// If there is only one auth provider configured, the single auth provider is a GitHub
		// instance, and it's an app request, redirect to signin immediately. The user wouldn't be
		// able to do anything else anyway; there's no point in showing them a signin screen with
		// just a single signin option.
		if ps := auth.Providers(); len(ps) == 1 && ps[0].Config().Github != nil && !isAPIRequest {
			v := make(url.Values)
			v.Set("redirect", auth.SafeRedirectURL(r.URL.String()))
			v.Set("pc", ps[0].ConfigID().ID)
			http.Redirect(w, r, authPrefix+"/login?"+v.Encode(), http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func newOAuthFlowHandler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/login", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		id := req.URL.Query().Get("pc")
		p := getProvider(id)
		if p == nil {
			log15.Error("no GitHub auth provider found with ID", "id", id)
			http.Error(w, "Misconfigured GitHub auth provider.", http.StatusInternalServerError)
		}
		p.login.ServeHTTP(w, req)
	}))
	mux.Handle("/callback", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		state, err := DecodeState(req.URL.Query().Get("state"))
		if err != nil {
			http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not decode OAuth state from URL parameter.", http.StatusBadRequest)
			return
		}

		p := getProvider(state.ProviderID)
		if p == nil {
			log15.Error("GitHub OAuth failed: in callback, no GitHub auth provider found with ID", "id", state.ProviderID)
			http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not find GitHub provider that matches the OAuth state parameter.", http.StatusBadRequest)
			return
		}
		p.callback.ServeHTTP(w, req)
	}))
	return mux
}
