package main

import "net/http" // "net/http"

type wrapper struct {
	c *http.Client      // "net/http Client"
	x http.RoundTripper // "net/http RoundTripper"
}

func foobar(c *http.Client) error { // "net/http Client"
	w := &wrapper{
		c: &http.Client{},         // "net/http Client"
		x: http.RoundTripper(nil), // "net/http RoundTripper"
	}
	_ = w
	return nil
}

func main() {
	c := &http.Client{} // "net/http Client"
	foobar(c)
}
