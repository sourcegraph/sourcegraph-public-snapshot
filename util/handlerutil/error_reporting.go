package handlerutil

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"reflect"

	"strconv"
	"strings"

	"sourcegraph.com/sourcegraph/appdash/httptrace"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/util/envutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
	"sourcegraph.com/sourcegraph/sourcegraph/util/traceutil/appdashctx"

	"github.com/getsentry/raven-go"
	"github.com/gorilla/mux"
)

var ravenClient *raven.Client

func init() {
	if dsn := conf.PrivateRavenDSN; dsn != "" {
		var err error
		ravenClient, err = raven.NewClient(dsn, nil)
		if err != nil {
			log.Fatalf("Error initializing Raven (Sentry) error reporter (with SG_APP_RAVEN_DSN=%q): %s.", dsn, err)
		}
		ravenClient.DropHandler = func(pkt *raven.Packet) {
			log.Println("WARNING: dropped error report because buffer is full:", pkt)
		}
	}
}

// reportError reports an error to Sentry.
func reportError(r *http.Request, status int, err error, panicked bool) {
	if ravenClient == nil {
		return
	}
	if status > 0 && status < 500 {
		// Not a reportable error.
		return
	}

	// Catch panics here to be extra sure we don't disrupt the request handling.
	defer func() {
		if rv := recover(); rv != nil {
			log.Println("WARNING: panic in HTTP handler error reporter: (recovered)", rv)
		}
	}()

	var errIface raven.Interface
	if panicked {
		errIface = raven.NewException(err, raven.NewStacktrace(4, 2, []string{"sourcegraph.com/sourcegraph/"}))
	} else {
		errIface = raven.NewException(err, nil)
	}

	h := raven.NewHttp(r)
	h.Cookies = "" // Don't send session cookies (which have auth secrets).
	delete(h.Headers, "Cookie")
	delete(h.Headers, "Authorization")

	pkt := raven.NewPacket(err.Error(), errIface, h)

	addTag := func(key, val string) {
		pkt.Tags = append(pkt.Tags, raven.Tag{Key: key, Value: val})
	}

	// Add appdash span ID.
	spanID, _ := httptrace.GetSpanID(r.Header)
	if spanID != nil {
		appdashURL := appdashctx.AppdashURLSafe(httpctx.FromRequest(r))

		if spanID.Trace != 0 {
			addTag("Appdash trace", spanID.Trace.String())
			if appdashURL != nil {
				pkt.Extra["Appdash trace"] = appdashURL.ResolveReference(&url.URL{
					Path: fmt.Sprintf("/traces/%v", spanID.Trace),
				}).String()
			}
		}
		if spanID.Span != 0 {
			addTag("Appdash span", spanID.Span.String())
			if appdashURL != nil {
				pkt.Extra["Appdash span"] = appdashURL.ResolveReference(&url.URL{
					Path: fmt.Sprintf("/traces/%v/%v", spanID.Trace, spanID.Span),
				}).String()
			}
		}
		if spanID.Parent != 0 {
			addTag("Appdash parent", spanID.Parent.String())
		}
	}

	// Add request context tags.
	ctx := httpctx.FromRequest(r)
	actor := auth.ActorFromContext(ctx)
	if actor.IsAuthenticated() {
		addTag("Authed", "yes")
		addTag("Authed UID", strconv.Itoa(actor.UID))
		addTag("Authed user", actor.Login)
	} else {
		addTag("Authed", "no")
	}
	if len(actor.Scope) > 0 {
		addTag("Actor scope", "yes")
		pkt.Extra["Actor scope"] = actor.Scope
	}
	if routeName := httpctx.RouteName(r); routeName != "" {
		addTag("Route", routeName)
	}
	if routeVars := mux.Vars(r); len(routeVars) > 0 {
		pkt.Extra["Route vars"] = routeVars
		for k, v := range routeVars {
			if v == "" {
				continue
			}
			addTag("Route "+k, v)

			// Allow filtering by repo owner.
			if k == "Repo" {
				parts := strings.Split(v, "/")
				if len(parts) == 3 {
					addTag("Route RepoSpec Owner", parts[1])
					addTag("Route RepoSpec Name", parts[2])
				}
			}
		}
	}

	// Add deployment tags.
	if envutil.GitCommitID != "" {
		addTag("Deployed commit", envutil.GitCommitID)
	}

	// Add error information.
	if panicked {
		addTag("Error type", "panic")
	} else {
		addTag("Error type", reflect.TypeOf(err).String())
		pkt.Extra["Error value"] = fmt.Sprintf("%#v", err)
	}

	ravenClient.Capture(pkt, nil)
}
