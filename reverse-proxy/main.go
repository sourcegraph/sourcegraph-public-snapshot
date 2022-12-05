package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func main() {
	u, err := url.Parse("https://petri-s-organization.us.auth0.com/")
	if err != nil {
		log.Fatal(err)
	}
	urlDirector := httputil.NewSingleHostReverseProxy(u).Director
	director := func(req *http.Request) {
		orig := req.URL.String()
		urlDirector(req)
		req.URL.Host = "petri-s-organization.us.auth0.com"

		req.Host = req.URL.Host
		log.Println(orig, "->", req.URL)
	}
	err = http.ListenAndServeTLS(":444", "server.crt", "server.key", &httputil.ReverseProxy{
		Director: director,
		ModifyResponse: func(resp *http.Response) error {
			if len(resp.Header["Content-Encoding"]) > 0 && resp.Header["Content-Encoding"][0] == "gzip" {
				reader, _ := gzip.NewReader(resp.Body)
				body, _ := io.ReadAll(reader)
				newBody := strings.ReplaceAll(string(body), "https://petri-s-organization.us.auth0.com/", "https://127.0.0.1:444/")
				var b bytes.Buffer
				writer := gzip.NewWriter(&b)
				writer.Write([]byte(newBody))
				writer.Close()
				readB := b.Bytes()
				log.Println(string(readB))
				resp.Body = ioutil.NopCloser(bytes.NewReader(readB))
			}
			return nil
		},
	})
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
