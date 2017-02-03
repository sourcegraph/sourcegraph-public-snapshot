package httpapi

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/websocket"
)

var websocketUpgrader = websocket.Upgrader{}

// getHijacker gets the underlying http.Hijacker from w (which is
// usually wrapped to be a gziphandler.GzipResponseWriter, etc.). If
// it can't get it, it panics.
func getHijacker(w http.ResponseWriter) http.ResponseWriter {
	switch w2 := w.(type) {
	case gziphandler.GzipResponseWriter:
		return getHijacker(w2.ResponseWriter)
	case http.Hijacker:
		return w
	default:
		// Use the ResponseWriter field if it exists (otherwise panic).
		v := reflect.ValueOf(w)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if f := v.FieldByName("ResponseWriter"); f.IsValid() {
			return getHijacker(f.Interface().(http.ResponseWriter))
		}
		panic(fmt.Errorf("getHijacker: can't get http.Hijacker from ResponseWriter of type %T", w))
	}
}
