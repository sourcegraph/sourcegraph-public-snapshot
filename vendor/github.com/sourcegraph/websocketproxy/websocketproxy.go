// Package websocketproxy is a reverse proxy for WebSocket connections.
package websocketproxy

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
)

var (
	// DefaultUpgrader specifies the parameters for upgrading an HTTP
	// connection to a WebSocket connection.
	DefaultUpgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	// DefaultDialer is a dialer with all fields set to the default zero values.
	DefaultDialer = websocket.DefaultDialer
)

// WebsocketProxy is an HTTP Handler that takes an incoming WebSocket
// connection and proxies it to another server.
type WebsocketProxy struct {
	// Director, if non-nil, is a function that may copy additional request
	// headers from the incoming WebSocket connection into the output headers
	// which will be forwarded to another server.
	Director func(incoming *http.Request, out http.Header)

	// Backend returns the backend URL which the proxy uses to reverse proxy
	// the incoming WebSocket connection. Request is the initial incoming and
	// unmodified request.
	Backend func(*http.Request) *url.URL

	// Upgrader specifies the parameters for upgrading a incoming HTTP
	// connection to a WebSocket connection. If nil, DefaultUpgrader is used.
	Upgrader *websocket.Upgrader

	//  Dialer contains options for connecting to the backend WebSocket server.
	//  If nil, DefaultDialer is used.
	Dialer *websocket.Dialer

	// Copy is invoked to perform the bi-direction copy of data between the
	// public client and proxy backend.
	Copy func(client *websocket.Conn, backend func() (*websocket.Conn, error)) error
}

// ProxyHandler returns a new http.Handler interface that reverse proxies the
// request to the given target.
func ProxyHandler(target *url.URL) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		NewProxy(target).ServeHTTP(rw, req)
	})
}

// NewProxy returns a new Websocket reverse proxy that rewrites the
// URL's to the scheme, host and base path provider in target.
func NewProxy(target *url.URL) *WebsocketProxy {
	backend := func(r *http.Request) *url.URL {
		// Shallow copy
		u := *target
		u.Fragment = r.URL.Fragment
		u.Path = r.URL.Path
		u.RawQuery = r.URL.RawQuery
		return &u
	}
	return &WebsocketProxy{Backend: backend}
}

// ServeHTTP implements the http.Handler that proxies WebSocket connections.
func (w *WebsocketProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if w.Backend == nil {
		log.Println("websocketproxy: backend function is not defined")
		http.Error(rw, "internal server error (code: 1)", http.StatusInternalServerError)
		return
	}

	upgrader := w.Upgrader
	if w.Upgrader == nil {
		upgrader = DefaultUpgrader
	}

	// Now upgrade the existing incoming request to a WebSocket connection.
	connPub, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		log.Printf("websocketproxy: couldn't upgrade %s\n", err)
		return
	}
	defer connPub.Close()

	// Start our proxy now, everything is ready...
	errc := make(chan error, 1)
	go func() {
		err := w.Copy(connPub, func() (*websocket.Conn, error) {
			backendURL := w.Backend(req)
			if backendURL == nil {
				return nil, fmt.Errorf("websocketproxy: backend URL is nil")
			}

			dialer := w.Dialer
			if w.Dialer == nil {
				dialer = DefaultDialer
			}

			// Pass headers from the incoming request to the dialer to forward them to
			// the final destinations.
			requestHeader := http.Header{}
			if origin := req.Header.Get("Origin"); origin != "" {
				requestHeader.Add("Origin", origin)
			}
			for _, prot := range req.Header[http.CanonicalHeaderKey("Sec-WebSocket-Protocol")] {
				requestHeader.Add("Sec-WebSocket-Protocol", prot)
			}
			for _, cookie := range req.Header[http.CanonicalHeaderKey("Cookie")] {
				requestHeader.Add("Cookie", cookie)
			}

			// Pass X-Forwarded-For headers too, code below is a part of
			// httputil.ReverseProxy. See http://en.wikipedia.org/wiki/X-Forwarded-For
			// for more information
			// TODO: use RFC7239 http://tools.ietf.org/html/rfc7239
			if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
				// If we aren't the first proxy retain prior
				// X-Forwarded-For information as a comma+space
				// separated list and fold multiple headers into one.
				if prior, ok := req.Header["X-Forwarded-For"]; ok {
					clientIP = strings.Join(prior, ", ") + ", " + clientIP
				}
				requestHeader.Set("X-Forwarded-For", clientIP)
			}

			// Set the originating protocol of the incoming HTTP request. The SSL might
			// be terminated on our site and because we doing proxy adding this would
			// be helpful for applications on the backend.
			requestHeader.Set("X-Forwarded-Proto", "http")
			if req.TLS != nil {
				requestHeader.Set("X-Forwarded-Proto", "https")
			}

			// Enable the director to copy any additional headers it desires for
			// forwarding to the remote server.
			if w.Director != nil {
				w.Director(req, requestHeader)
			}

			// Connect to the backend URL, also pass the headers we get from the requst
			// together with the Forwarded headers we prepared above.
			connBackend, _, err := dialer.Dial(backendURL.String(), requestHeader)
			if err != nil {
				return nil, fmt.Errorf("websocketproxy: couldn't dial to remote backend url %s\n", err)
			}
			return connBackend, nil
		})
		errc <- err
	}()
	<-errc
}
