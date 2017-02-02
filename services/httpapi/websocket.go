package httpapi

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"reflect"

	log15 "gopkg.in/inconshreveable/log15.v2"

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

// webSocketProxy returns an HTTP handler that proxies WebSocket
// connections to the given target. Taken from bradfitz at
// https://groups.google.com/forum/#!msg/golang-nuts/KBx9pDlvFOc/QC5v-uC5UOgJ.
func webSocketProxy(dialer func() (net.Conn, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d, err := dialer()
		if err != nil {
			http.Error(w, "Error contacting backend server.", http.StatusInternalServerError)
			log15.Error("Error dialing WebSocket backend.", "err", err)
			return
		}
		hj, ok := getHijacker(w).(http.Hijacker)
		if !ok {
			log15.Error("Error hijacking HTTP connection.", "err", "not a http.Hijacker")
			http.Error(w, "Unable to proxy HTTP connection.", http.StatusInternalServerError)
			return
		}

		nc, _, err := hj.Hijack()
		if err != nil {
			log15.Error("Error hijacking HTTP connection.", "err", err)
			return
		}
		defer nc.Close()
		defer d.Close()

		if err := r.Write(d); err != nil {
			log15.Error("Error copying WebSocket request to target.", "err", err)
			return
		}

		errc := make(chan error, 2)
		cp := func(dst io.Writer, src io.Reader) {
			_, err := io.Copy(dst, src)
			errc <- err
		}
		go cp(d, nc)
		go cp(nc, d)
		<-errc
	})
}
