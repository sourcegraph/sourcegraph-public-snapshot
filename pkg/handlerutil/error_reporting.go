package handlerutil

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/getsentry/raven-go"
	"github.com/gorilla/mux"
	opentracing "github.com/opentracing/opentracing-go"

	"sourcegraph.com/sourcegraph/sourcegraph/cli/buildvar"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
)

var ravenClient *raven.Client

func init() {
	if dsn := env.Get("SENTRY_DSN_BACKEND", "", "Sentry/Raven DSN used for tracking of backend errors"); dsn != "" {
		var err error
		ravenClient, err = raven.New(dsn)
		if err != nil {
			log.Fatalf("error initializing Sentry error reporter: %s", err)
		}
		ravenClient.DropHandler = func(pkt *raven.Packet) {
			log.Println("WARNING: dropped error report because buffer is full:", pkt)
		}
		ravenClient.SetRelease(buildvar.All.Version)
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
	addTag("trace", traceutil.SpanURL(opentracing.SpanFromContext(r.Context())))

	// Add request context tags.
	if actor := auth.ActorFromContext(r.Context()); actor.UID != "" {
		addTag("Authed", "yes")
		addTag("Authed UID", actor.UID)
		addTag("Authed user", actor.Login)
	} else {
		addTag("Authed", "no")
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

	// Add error information.
	if panicked {
		addTag("Error type", "panic")
	} else {
		addTag("Error type", reflect.TypeOf(err).String())
		pkt.Extra["Error value"] = fmt.Sprintf("%#v", err)
	}

	ravenClient.Capture(pkt, nil)
}
