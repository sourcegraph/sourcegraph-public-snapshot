package main

import "net/http" // "net/http"

func main() {
	var x *http.Client // "net/http Client"
	var y int
	z := http.RoundTripper(nil) // "net/http RoundTripper"
	_ = x
	_ = y
	_ = &http.Client{ // "net/http Client"
		Transport: z, // "net/http Client Transport"
	}
}
