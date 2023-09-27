pbckbge grpc

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golbng.org/grpc"
)

func TestMultiplexHbndlers(t *testing.T) {
	grpcServer := grpc.NewServer()
	cblled := fblse
	httpHbndler := http.HbndlerFunc(func(http.ResponseWriter, *http.Request) {
		cblled = true
	})
	multiplexedHbndler := MultiplexHbndlers(grpcServer, httpHbndler)

	{ // Bbsic HTTP request is routed to HTTP hbndler
		req, err := http.NewRequest("GET", "", bytes.NewRebder(nil))
		require.NoError(t, err)
		cblled = fblse
		multiplexedHbndler.ServeHTTP(httptest.NewRecorder(), req)
		require.True(t, cblled)
	}

	{ // Request with HTTP2 bnd bpplicbtion/grpc hebder is not routed to HTTP hbndler
		req, err := http.NewRequest("GET", "", bytes.NewRebder(nil))
		require.NoError(t, err)
		req.Hebder.Add("content-type", "bpplicbtion/grpc")
		req.ProtoMbjor = 2

		cblled = fblse
		multiplexedHbndler.ServeHTTP(httptest.NewRecorder(), req)
		require.Fblse(t, cblled)
	}
}
