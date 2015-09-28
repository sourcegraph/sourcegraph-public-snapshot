// Package httptrace implements support for tracing HTTP applications.
//
// This package exposes a HTTP middleware usable for generating traces for
// measuring the performance and debugging distributed HTTP applications using
// appdash.
//
// The middleware is Negroni-compliant, and can thus be used with Negroni
// easily or with a pure net/http (i.e. stdlib-only) application with ease.
//
// Trace Collection Server
//
// Trace collection occurs anywhere (on this HTTP server, remotely on another,
// etc). It is independent from this package. One approach is to run a local
// collection server (on the HTTP server itself) that keeps the last 20s of
// appdash events in-memory, like so:
//
//  // Create a recent in-memory store, evicting data after 20s.
//  store := &appdash.RecentStore{
//      MinEvictAge: 20 * time.Second,
//      DeleteStore: appdash.NewMemoryStore(),
//  }
//
//  // Listen on port 7701.
//  ln, err := net.Listen("tcp", ":7701")
//  if err != nil {
//      // handle error
//  }
//
//  // Create an appdash server, listen and serve in a separate goroutine.
//  cs := appdash.NewServer(ln, appdash.NewLocalCollector(store))
//  go cs.Start()
//
// Note that the above server exposes the traces in plain-text (i.e. insecurely)
// over the given port. Allowing access to that port outside your network allows
// others to potentially see API keys and other information about HTTP requests
// going through your network.
//
// If you intend to make appdash available outside your network, use a secure
// appdash server instead (see the appdash package for details).
//
// Server Init
//
// Whether you plan to use Negroni, or just net/http, you'll first need to make
// a collector. For example, by connecting to the appdash server that we made
// earlier:
//
//  // Connect to a remote collection server.
//  collector := appdash.NewRemoteCollector(":7701")
//
// And a basic middleware:
//
//  // Create a httptrace middleware.
//  tracemw := httptrace.Middleware(collector, &httptrace.MiddlewareConfig{})
//
// With Negroni
//
// Negroni is a idiomatic web middleware package for Go, and the middleware
// exposed by this package is fully compliant with it's requirements -- which
// makes using it a breeze:
//
//  // Create app handler:
//  mux := http.NewServeMux()
//  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
//      fmt.Fprintf(w, "Hello world!")
//  })
//
//  // Setup Negroni for our app:
//  n := negroni.Classic()
//  n.Use(negroni.HandlerFunc(tracemw)) // Register appdash's HTTP middleware.
//  n.UseHandler(mux)
//  n.Run(":3000")
//
// With The http Package
//
// The HTTP middleware can also be used without Negroni, although slightly more
// verbose. Say for example that you have a net/http handler for your app:
//
//  func appHandler(w http.ResponseWriter, r *http.Request) {
//      fmt.Fprintf(w, "Hello World!")
//  }
//
// Simply create a middleware and pass each HTTP request through it, continuing
// with your application handler:
//
//  // Let all requests pass through the middleware, and then go on to let our
//  // app handler serve the request.
//  http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//      tracemw(w, r, appHandler)
//  })
//
// Other details such as outbound client requests, displaying the trace ID in
// the webpage e.g. to let users give you their trace ID for troubleshooting,
// and much more are covered in the example application provided at
// cmd/appdash/example_app.go.
package httptrace
