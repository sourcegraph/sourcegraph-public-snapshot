package middleware

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// CanonicalURL is an HTTP middleware that intercepts HTTP requests to URLs not matching the scheme
// (http/https) or host of the `appURL`. For these intercepted requests, it returns a redirect to
// the same request URI on the canonical `appURL` scheme and host.
//
// It is intended to force redirects to HTTPS and to avoid confusion by clients that access
// Sourcegraph via a URL other than the canonical one, which may mean the user's requests are
// bypassing authentication or load-balancing.
func CanonicalURL(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conf := conf.Get()

		appURLStr := conf.AppURL
		if appURLStr == "" {
			appURLStr = globals.AppURL.String() // default
		}
		appURL, err := url.Parse(appURLStr)
		if err == nil && !appURL.IsAbs() {
			err = errors.New("non-absolute URL")
		}
		if err != nil {
			text := "Misconfigured appURL value in site configuration."
			log15.Error(text, "invalidValue", appURLStr, "err", err)
			http.Error(w, text, http.StatusInternalServerError)
			return
		}

		httpToHTTPSRedirect := parseStringOrBool(conf.HttpToHttpsRedirect, "off")
		var requireSchemeMatch bool
		switch httpToHTTPSRedirect {
		case "off":
			// noop
		case "on", "load-balanced":
			requireSchemeMatch = true
		default:
			text := "Misconfigured httpToHttpsRedirect value in site configuration."
			log15.Error(text, "invalidValue", httpToHTTPSRedirect)
			http.Error(w, text, http.StatusInternalServerError)
			return
		}

		if requireSchemeMatch && appURL.Scheme != "https" {
			// It wouldn't make sense to redirect to HTTPS if the appURL is not HTTPS.
			text := "Misconfigured appURL and httpToHttpsRedirect values in site configuration."
			log15.Error(text+" If httpToHttpsRedirect is enabled, the appURL scheme must be https.", "appURL", appURLStr, "httpToHttpsRedirect", httpToHTTPSRedirect)
			http.Error(w, text, http.StatusInternalServerError)
			return
		}

		var canonicalURLRedirect bool
		if conf.ExperimentalFeatures != nil {
			switch conf.ExperimentalFeatures.CanonicalURLRedirect {
			case "enabled", "": // default enabled
				canonicalURLRedirect = true
			case "disabled":
				// noop
			default:
				text := "Misconfigured experimentalFeatures.canonicalURLRedirect values in site configuration."
				log15.Error(text, "invalidValue", conf.ExperimentalFeatures.CanonicalURLRedirect)
				http.Error(w, text, http.StatusInternalServerError)
				return
			}
		}

		requireHostMatch := conf.ExperimentalFeatures != nil && canonicalURLRedirect
		useXForwardedProto := httpToHTTPSRedirect == "load-balanced"
		if reqURL := getRequestURL(r, useXForwardedProto); (requireHostMatch && reqURL.Host != appURL.Host) || (requireSchemeMatch && !doesSchemeMatch(r, appURL, useXForwardedProto)) {
			// Redirect.
			dest := appURL.ResolveReference(&url.URL{Path: reqURL.Path, RawQuery: reqURL.RawQuery, Fragment: reqURL.Fragment})
			http.Redirect(w, r, dest.String(), http.StatusMovedPermanently)
			return
		}

		// No redirect needed.
		next.ServeHTTP(w, r)
	})
}

// doesSchemeMatch returns true if and only if the request matches the app URL scheme.  Because the
// request URL typically has no scheme set, we use http.Request.TLS to determine if the request's
// scheme was "https". If useXForwardedProto is true, then use that while ignoring the scheme of the
// actual request.
func doesSchemeMatch(r *http.Request, appURL *url.URL, useXForwardedProto bool) bool {
	if useXForwardedProto {
		if v := r.Header.Get("X-Forwarded-Proto"); v != "" {
			return v == appURL.Scheme
		}
	}
	if appURL.Scheme == "https" && r.TLS == nil {
		return false
	}
	if appURL.Scheme == "http" && r.TLS != nil {
		return false
	}
	return true
}

func getRequestURL(r *http.Request, useXForwardedProto bool) *url.URL {
	u := *r.URL // copy
	u.Host = r.Host
	if useXForwardedProto {
		if v := r.Header.Get("X-Forwarded-Proto"); v != "" {
			u.Scheme = v
		}
	}
	return &u
}

// parseStringOrBool converts true to "on", false to "off", and returns strings as-is.
func parseStringOrBool(v interface{}, defaultValue string) string {
	if v == nil {
		return defaultValue
	}
	if s, ok := v.(string); ok {
		return s
	}
	if v.(bool) {
		return "on"
	}
	return "off"
}
