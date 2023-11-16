package requestclient_test

import (
	"context"
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc/peer"

	"github.com/sourcegraph/sourcegraph/internal/requestclient"
)

func TestPropagator(t *testing.T) {
	tests := []struct {
		name string

		requestClient *requestclient.Client
		requestPeer   *peer.Peer

		wantClient *requestclient.Client
	}{
		{
			name: "no client or peer",

			wantClient: &requestclient.Client{},
		},

		{
			name: "client with no peer",
			requestClient: &requestclient.Client{
				IP:           "192.168.1.1",
				ForwardedFor: "192.168.1.2",
			},

			wantClient: &requestclient.Client{
				IP:           "192.168.1.1",
				ForwardedFor: "192.168.1.2",
			},
		},

		{
			name: "peer only (nil client)",
			requestPeer: &peer.Peer{
				Addr: &net.IPAddr{IP: net.ParseIP("192.168.1.1")},
			},

			wantClient: &requestclient.Client{
				IP: "192.168.1.1",
			},
		},
		{
			name: "peer only (non-nil empty client)",

			requestClient: &requestclient.Client{},
			requestPeer: &peer.Peer{
				Addr: &net.IPAddr{IP: net.ParseIP("192.168.1.1")},
			},

			wantClient: &requestclient.Client{
				IP: "192.168.1.1",
			},
		},

		{
			name: "client should override peer",

			requestClient: &requestclient.Client{
				IP:           "192.168.1.1",
				ForwardedFor: "192.168.1.2",
			},
			requestPeer: &peer.Peer{
				Addr: &net.IPAddr{IP: net.ParseIP("192.168.1.3")},
			},

			wantClient: &requestclient.Client{
				IP:           "192.168.1.1",
				ForwardedFor: "192.168.1.2",
			},
		},

		{
			name: "client for ForwardedFor, peer for IP",

			requestClient: &requestclient.Client{
				ForwardedFor: "192.168.1.2",
			},
			requestPeer: &peer.Peer{
				Addr: &net.IPAddr{IP: net.ParseIP("192.168.1.3")},
			},

			wantClient: &requestclient.Client{
				IP:           "192.168.1.3",
				ForwardedFor: "192.168.1.2",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requestCtx := context.Background()
			if test.requestClient != nil {
				requestCtx = requestclient.WithClient(requestCtx, test.requestClient)
			}

			if test.requestPeer != nil {
				requestCtx = peer.NewContext(requestCtx, test.requestPeer)
			}

			propagator := &requestclient.Propagator{}
			md := propagator.FromContext(requestCtx)

			resultCtx := propagator.InjectContext(requestCtx, md)
			if diff := cmp.Diff(
				test.wantClient,
				requestclient.FromContext(resultCtx),
				// Ignore unexported fields in Client - this test only tests
				// the exported API surface (hence why we run this test in
				// package requestclient_test)
				cmpopts.IgnoreUnexported(requestclient.Client{}),
			); diff != "" {
				t.Errorf("Client mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
