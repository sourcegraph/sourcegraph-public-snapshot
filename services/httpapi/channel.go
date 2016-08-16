package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"context"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
)

func serveChannelListen(w http.ResponseWriter, r *http.Request) {
	ctx, cl := handlerutil.Client(r)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	channel := mux.Vars(r)["Channel"]
	stream, err := cl.Channel.Listen(ctx, &sourcegraph.ChannelListenOp{Channel: channel})
	if err != nil {
		log15.Error("serveChannelListen: failed to establish Channel.Listen stream.", "err", err, "channel", channel)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	ws, err := upgrader.Upgrade(getHijacker(w), r, nil)
	if err != nil {
		log15.Error("serveChannelListen: Upgrade to WebSocket failed.", "err", err)
		http.Error(w, "upgrade to WebSocket failed", http.StatusInternalServerError)
		return
	}

	isUnexpectedWebSocketError := func(err error) bool {
		return websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway)
	}

	// We don't interpret WebSocket client messages at the application
	// level, but we must deal with client control messages.
	go func() {
		for {
			if _, _, err := ws.NextReader(); err != nil {
				cancel()   // Close connection to gRPC server.
				ws.Close() // Close connection to browser (WebSocket client).
				if isUnexpectedWebSocketError(err) {
					log15.Error("serveChannelListen: WebSocket unexpectedly closed.", "err", err)
				}
				break
			}
		}
	}()

	isCanceled := func(err error) bool {
		return err == context.Canceled || grpc.Code(err) == codes.Canceled
	}

	// Can't return an error because we've already written a header
	// (by upgrading the connection).
	func() /* no error return */ {
		for {
			action, err := stream.Recv()
			if err != nil {
				if !isCanceled(err) {
					log15.Error("serveChannelListen: failed to receive on Channel.Listen gRPC stream.", "err", err)
				}
				break
			}

			if err := ws.WriteJSON(action); err != nil {
				if isUnexpectedWebSocketError(err) {
					log15.Error("serveChannelListen: failed to write WebSocket message.", "err", err)
				}
				break
			}
		}
		err := ws.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseGoingAway, "server encountered error"),
			time.Time{},
		)
		if err != nil && err != websocket.ErrCloseSent {
			log15.Error("serveChannelListen: failed to close WebSocket.", "err", err)
		}
	}()
}

func serveChannelSend(w http.ResponseWriter, r *http.Request) error {
	var op *sourcegraph.ChannelSendOp
	if err := json.NewDecoder(r.Body).Decode(&op); err != nil {
		return err
	}
	op.Channel = mux.Vars(r)["Channel"]

	ctx, cl := handlerutil.Client(r)
	res, err := cl.Channel.Send(ctx, op)
	if err != nil {
		log15.Error("serveChannelSend: failed to send request over Channel.Listen channel", "err", err)
		return err
	}
	return writeJSON(w, res)
}

// getHijacker gets the underlying http.Hijacker from w (which is
// usually wrapped to be a gziphandler.GzipResponseWriter, etc.). If
// it can't get it, it panics.
//
// KNOWN ISSUE: The Appdash middleware always tries to write to the
// HTTP response even if it's hijacked, so this will always result in
// the log message "http: response.Write on hijacked connection" being
// printed to stderr. That is harmless but annoying.
func getHijacker(w http.ResponseWriter) http.ResponseWriter {
	switch w2 := w.(type) {
	case gziphandler.GzipResponseWriter:
		return getHijacker(w2.ResponseWriter)
	case http.Hijacker:
		return w
	default:
		// Use reflection to handle unexported ResponseWriter wrapper
		// types that embed it, such as Appdash's
		// httptrace.responseInfoRecorder.
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
