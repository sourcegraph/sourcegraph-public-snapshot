// The testproxy command runs a simple HTTP proxy that wraps a Sourcegraph server running with site
// config auth.provider=="http-header" to test the authentication HTTP proxy support.

// +build ignore

package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

var (
	addr       = flag.String("addr", ":4080", "HTTP listen address")
	urlStr     = flag.String("url", "http://localhost:3080", "proxy origin URL (Sourcegraph HTTP/HTTPS URL)") // CI:LOCALHOST_OK
	username   = flag.String("username", os.Getenv("USER"), "username to report to Sourcegraph")
	httpHeader = flag.String("header", "X-Forwarded-User", "name of HTTP header to add to request")
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	url, err := url.Parse(*urlStr)
	if err != nil {
		log.Fatalf("Error: Invalid -url: %s.", err)
	}
	if *username == "" {
		log.Fatal("Error: No -username specified.")
	}
	if *httpHeader == "" {
		log.Fatal("Error: No -header specified.")
	}
	log.Printf(`Listening on %s, forwarding requests to %s with added header "%s: %s"`, *addr, url, *httpHeader, *username)
	p := httputil.NewSingleHostReverseProxy(url)
	log.Fatalf("Server error: %s.", http.ListenAndServe(*addr, &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.Header.Set(*httpHeader, *username)
			r.Host = url.Host
			p.Director(r)
		},
	}))
}
