package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/appdash/httptrace"
	"sourcegraph.com/sourcegraph/appdash/traceapp"
)

func init() {
	_, err := CLI.AddCommand("demo",
		"start a demo web app that uses appdash",
		"The demo command starts a demo web app that uses appdash.",
		&demoCmd,
	)
	if err != nil {
		log.Fatal(err)
	}
}

// DemoCmd is the command for running Appdash in demo mode.
type DemoCmd struct {
	AppdashHTTPAddr string `long:"appdash-http" description:"appdash HTTP listen address" default:":8700"`
	DemoHTTPAddr    string `long:"demo-http" description:"demo app HTTP listen address" default:":8699"`
	Debug           bool   `long:"debug" description:"debug logging"`
	Trace           bool   `long:"trace" description:"trace logging"`
}

var demoCmd DemoCmd

// Execute execudes the commands with the given arguments and returns an error,
// if any.
func (c *DemoCmd) Execute(args []string) error {
	// We create a new in-memory store. All information about traces will
	// eventually be stored here.
	store := appdash.NewMemoryStore()

	// Listen on any available TCP port locally.
	l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		log.Fatal(err)
	}
	collectorPort := l.Addr().(*net.TCPAddr).Port
	log.Printf("Appdash collector listening on tcp:%d", collectorPort)

	// Start an Appdash collection server that will listen for spans and
	// annotations and add them to the local collector (stored in-memory).
	cs := appdash.NewServer(l, appdash.NewLocalCollector(store))
	cs.Debug = c.Debug // Debug logging
	cs.Trace = c.Trace // Trace logging
	go cs.Start()

	// Print the URL at which the web UI will be running.
	appdashURLStr := "http://localhost" + c.AppdashHTTPAddr
	appdashURL, err := url.Parse(appdashURLStr)
	if err != nil {
		log.Fatalf("Error parsing http://localhost:%s: %s", c.AppdashHTTPAddr, err)
	}
	log.Printf("Appdash web UI running at %s", appdashURL)

	// Start the web UI in a separate goroutine.
	tapp := traceapp.New(nil)
	tapp.Store = store
	tapp.Queryer = store
	go func() {
		log.Fatal(http.ListenAndServe(c.AppdashHTTPAddr, tapp))
	}()

	// Print the URL at which the demo app is running.
	demoURLStr := "http://localhost" + c.DemoHTTPAddr
	demoURL, err := url.Parse(demoURLStr)
	if err != nil {
		log.Fatalf("Error parsing http://localhost:%s: %s", c.DemoHTTPAddr, err)
	}
	log.Println()
	log.Printf("Appdash demo app running at %s", demoURL)

	// The Appdash collection server that our demo app will use is running
	// locally with our HTTP server in this case, so we set this up now.
	localCollector := appdash.NewRemoteCollector(fmt.Sprintf(":%d", collectorPort))

	// Handle the root path of our app.
	http.Handle("/", &middlewareHandler{
		middleware: httptrace.Middleware(localCollector, &httptrace.MiddlewareConfig{
			RouteName:      func(r *http.Request) string { return r.URL.Path },
			SetContextSpan: requestSpans.setRequestSpan,
		}),
		next: &demoApp{collector: localCollector, baseURL: demoURL, appdashURL: appdashURL},
	})
	return http.ListenAndServe(c.DemoHTTPAddr, nil)
}

type middlewareHandler struct {
	middleware func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)
	next       http.Handler
}

func (h *middlewareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.middleware(w, r, h.next.ServeHTTP)
}

type demoApp struct {
	collector  appdash.Collector
	baseURL    *url.URL
	appdashURL *url.URL
}

func (a *demoApp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	span := requestSpans[r]

	switch r.URL.Path {
	case "/":
		io.WriteString(w, `<h1>Appdash demo</h1>
<p>Welcome! Click some links and then view the traces for each HTTP request by following the link at the bottom of the page.
<ul>
<li><a href="/api-calls">Visit a page that issues some API calls</a></li>
</ul>`)
	case "/api-calls":
		httpClient := &http.Client{
			Transport: &httptrace.Transport{Recorder: appdash.NewRecorder(span, a.collector), SetName: true},
		}
		resp, err := httpClient.Get(a.baseURL.ResolveReference(&url.URL{Path: "/endpoint-A"}).String())
		if err == nil {
			defer resp.Body.Close()
		}
		resp, err = httpClient.Get(a.baseURL.ResolveReference(&url.URL{Path: "/endpoint-B"}).String())
		if err == nil {
			defer resp.Body.Close()
		}
		resp, err = httpClient.Get(a.baseURL.ResolveReference(&url.URL{Path: "/endpoint-C"}).String())
		if err == nil {
			defer resp.Body.Close()
		}
		io.WriteString(w, `<a href="/">Home</a><br><br><p>I just made 3 API calls. Check the trace below to see them!</p>`)
	case "/endpoint-A":
		time.Sleep(250 * time.Millisecond)
		io.WriteString(w, "performed an operation!")
		return
	case "/endpoint-B":
		time.Sleep(75 * time.Millisecond)
		io.WriteString(w, "performed another operation!")
		return
	case "/endpoint-C":
		time.Sleep(300 * time.Millisecond)
		io.WriteString(w, "performed yet another operation!")
		return
	}

	spanURL := a.appdashURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/traces/%v", span.Trace)})
	io.WriteString(w, fmt.Sprintf(`<br><br><hr><a href="%s">View request trace on appdash</a> (trace ID is %s)`, spanURL, span.Trace))
}

type requestSpanMap map[*http.Request]appdash.SpanID

// requestSpans stores the appdash span ID associated with each HTTP
// request. In a real app, you would use something like
// gorilla/context instead of a map (so that entries get removed when
// handling is completed, etc.).
var requestSpans = requestSpanMap{}

func (m requestSpanMap) setRequestSpan(r *http.Request, spanID appdash.SpanID) {
	m[r] = spanID
}
