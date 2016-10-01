package main

import (
	"net/http" // "net/http"
)

func main() {
	r := &http.Request{} // "net/http Request"
	r.Header = nil       // "net/http Request Header"
	x := r
	x.Header["Content-Encoding"] = []string{"application/json"} // "net/http Request Header"
}
