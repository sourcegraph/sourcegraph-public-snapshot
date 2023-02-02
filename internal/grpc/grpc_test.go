package grpc

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestMultiplexHandlers(t *testing.T) {
	grpcServer := grpc.NewServer()
	called := false
	httpHandler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	})
	multiplexedHandler := MultiplexHandlers(grpcServer, httpHandler)

	{ // Basic HTTP request is routed to HTTP handler
		req, err := http.NewRequest("GET", "", bytes.NewReader(nil))
		require.NoError(t, err)
		called = false
		multiplexedHandler.ServeHTTP(httptest.NewRecorder(), req)
		require.True(t, called)
	}

	{ // Request with HTTP2 and application/grpc header is not routed to HTTP handler
		req, err := http.NewRequest("GET", "", bytes.NewReader(nil))
		require.NoError(t, err)
		req.Header.Add("content-type", "application/grpc")
		req.ProtoMajor = 2

		called = false
		multiplexedHandler.ServeHTTP(httptest.NewRecorder(), req)
		require.False(t, called)
	}
}
