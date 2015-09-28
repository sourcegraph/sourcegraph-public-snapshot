package router_util

import (
	"bytes"
	"log"
	"net/url"

	"github.com/sourcegraph/mux"
)

var charmap = map[byte]string{
	'?': "%3F",
}

func EscapePath(in string) string {
	var buf bytes.Buffer
	for _, c := range []byte(in) {
		if v, ok := charmap[c]; ok {
			buf.WriteString(v)
		} else {
			buf.WriteByte(c)
		}
	}
	return buf.String()
}

func URLTo(router *mux.Router, routeName string, params ...string) *url.URL {
	route := router.Get(routeName)
	if route == nil {
		log.Panicf("no such route: %q (params: %v)", routeName, params)
	}
	u, err := route.URLPath(params...)
	if err != nil {
		log.Printf("Route error: failed to make URL for route %q (params: %v): %s", routeName, params, err)
		return &url.URL{}
	}
	return u
}

func MapToArray(m map[string]string) (a []string) {
	for k, v := range m {
		a = append(a, k, v)
	}
	return
}
