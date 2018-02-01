/*

Package link parses Link headers used for pagination, as defined in RFC 5988

Installation

Just go get the package:

    go get -u github.com/peterhellberg/link

Usage

A small usage example

    package main

    import (
    	"fmt"
    	"net/http"

    	"github.com/peterhellberg/link"
    )

    func main() {
    	for _, l := range link.Parse(`<https://example.com/?page=2>; rel="next"; foo="bar"`) {
    		fmt.Printf("URI: %q, Rel: %q, Extra: %+v\n", l.URI, l.Rel, l.Extra)
    		// URI: "https://example.com/?page=2", Rel: "next", Extra: map[foo:bar]
    	}

    	if resp, err := http.Get("https://api.github.com/search/code?q=Println+user:golang"); err == nil {
    		for _, l := range link.ParseResponse(resp) {
    			fmt.Printf("URI: %q, Rel: %q, Extra: %+v\n", l.URI, l.Rel, l.Extra)
    			// URI: "https://api.github.com/search/code?q=Println+user%3Agolang&page=2", Rel: "next", Extra: map[]
    			// URI: "https://api.github.com/search/code?q=Println+user%3Agolang&page=34", Rel: "last", Extra: map[]
    		}
    	}
    }

*/
package link
