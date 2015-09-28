// webapp: a standalone example Negroni / Gorilla based webapp.
//
// This example demonstrates basic usage of Appdash in a Negroni / Gorilla
// based web application. The entire application is ran locally (i.e. on the
// same server) -- even the Appdash web UI.
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/appdash/httptrace"
	"sourcegraph.com/sourcegraph/appdash/traceapp"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

// Used to store the SpanID in a request's context (see gorilla/context docs
// for more information).
const CtxSpanID = 0

// We want to create HTTP clients recording to this collector inside our Home
// handler below, so we use a global variable (for simplicity sake) to store
// the collector in use. We could also use gorilla/context to store it.
var collector appdash.Collector

func main() {
	// Create a recent in-memory store, evicting data after 20s.
	//
	// The store defines where information about traces (i.e. spans and
	// annotations) will be stored during the lifetime of the application. This
	// application uses a MemoryStore store wrapped by a RecentStore with an
	// eviction time of 20s (i.e. all data after 20s is deleted from memory).
	memStore := appdash.NewMemoryStore()
	store := &appdash.RecentStore{
		MinEvictAge: 20 * time.Second,
		DeleteStore: memStore,
	}

	// Start the Appdash web UI on port 8700.
	//
	// This is the actual Appdash web UI -- usable as a Go package itself, We
	// embed it directly into our application such that visiting the web server
	// on HTTP port 8700 will bring us to the web UI, displaying information
	// about this specific web-server (another alternative would be to connect
	// to a centralized Appdash collection server).
	tapp := traceapp.New(nil)
	tapp.Store = store
	tapp.Queryer = memStore
	log.Println("Appdash web UI running on HTTP :8700")
	go func() {
		log.Fatal(http.ListenAndServe(":8700", tapp))
	}()

	// We will use a local collector (as we are running the Appdash web UI
	// embedded within our app).
	//
	// A collector is responsible for collecting the information about traces
	// (i.e. spans and annotations) and placing them into a store. In this app
	// we use a local collector (we could also use a remote collector, sending
	// the information to a remote Appdash collection server).
	collector = appdash.NewLocalCollector(store)

	// Create the appdash/httptrace middleware.
	//
	// Here we initialize the appdash/httptrace middleware. It is a Negroni
	// compliant HTTP middleware that will generate HTTP events for Appdash to
	// display. We could also instruct Appdash with events manually, if we
	// wanted to.
	tracemw := httptrace.Middleware(collector, &httptrace.MiddlewareConfig{
		RouteName: func(r *http.Request) string { return r.URL.Path },
		SetContextSpan: func(r *http.Request, spanID appdash.SpanID) {
			context.Set(r, CtxSpanID, spanID)
		},
	})

	// Setup our router (for information, see the gorilla/mux docs):
	router := mux.NewRouter()
	router.HandleFunc("/", Home)
	router.HandleFunc("/endpoint", Endpoint)

	// Setup Negroni for our app (for information, see the negroni docs):
	n := negroni.Classic()
	n.Use(negroni.HandlerFunc(tracemw)) // Register appdash's HTTP middleware.
	n.UseHandler(router)
	n.Run(":8699")
}

// Home is the homepage handler for our app.
func Home(w http.ResponseWriter, r *http.Request) {
	// Grab the span from the gorilla context. We do this so that we can grab
	// the span.Trace ID and link directly to the trace on the web-page itself!
	span := context.Get(r, CtxSpanID).(appdash.SpanID)

	// We're going to make some API requests, so we create a HTTP client using
	// a appdash/httptrace transport here. The transport will inform Appdash of
	// the HTTP events occuring.
	httpClient := &http.Client{
		Transport: &httptrace.Transport{
			Recorder: appdash.NewRecorder(span, collector),
			SetName:  true,
		},
	}

	// Make three API requests using our HTTP client.
	for i := 0; i < 3; i++ {
		resp, err := httpClient.Get("http://localhost:8699/endpoint")
		if err != nil {
			log.Println("/endpoint:", err)
			continue
		}
		resp.Body.Close()
	}

	// Render the page.
	fmt.Fprintf(w, `<p>Three API requests have been made!</p>`)
	fmt.Fprintf(w, `<p><a href="http://localhost:8700/traces/%s" target="_">View the trace (ID:%s)</a></p>`, span.Trace, span.Trace)
}

// Endpoint is an example API endpoint. In a real application, the backend of
// your service would be contacting several external and internal API endpoints
// which may be the bottleneck of your application.
//
// For example purposes we just sleep for 200ms before responding to simulate a
// slow API endpoint as the bottleneck of your application.
func Endpoint(w http.ResponseWriter, r *http.Request) {
	time.Sleep(200 * time.Millisecond)
	fmt.Fprintf(w, "Slept for 200ms!")
}
