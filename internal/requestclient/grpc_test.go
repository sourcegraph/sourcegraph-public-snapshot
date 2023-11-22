package requestclient

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/peer"
)

func TestPropagator(t *testing.T) {
	tests := []struct {
		name string

		requestClient *Client
		requestPeer   *peer.Peer

		wantClient *Client
	}{
		{
			name: "no client or peer",

			wantClient: &Client{},
		},

		{
			name: "client with no peer",
			requestClient: &Client{
				IP:           "192.168.1.1",
				ForwardedFor: "192.168.1.2",
			},

			wantClient: &Client{
				IP:           "192.168.1.1",
				ForwardedFor: "192.168.1.2",
			},
		},

		{
			name: "peer only (nil client)",
			requestPeer: &peer.Peer{
				Addr: &net.IPAddr{IP: net.ParseIP("192.168.1.1")},
			},

			wantClient: &Client{
				IP: "192.168.1.1",
			},
		},
		{
			name: "peer only (non-nil empty client)",

			requestClient: &Client{},
			requestPeer: &peer.Peer{
				Addr: &net.IPAddr{IP: net.ParseIP("192.168.1.1")},
			},

			wantClient: &Client{
				IP: "192.168.1.1",
			},
		},

		{
			name: "client should override peer",

			requestClient: &Client{
				IP:           "192.168.1.1",
				ForwardedFor: "192.168.1.2",
			},
			requestPeer: &peer.Peer{
				Addr: &net.IPAddr{IP: net.ParseIP("192.168.1.3")},
			},

			wantClient: &Client{
				IP:           "192.168.1.1",
				ForwardedFor: "192.168.1.2",
			},
		},

		{
			name: "client for ForwardedFor, peer for IP",

			requestClient: &Client{
				ForwardedFor: "192.168.1.2",
			},
			requestPeer: &peer.Peer{
				Addr: &net.IPAddr{IP: net.ParseIP("192.168.1.3")},
			},

			wantClient: &Client{
				IP:           "192.168.1.3",
				ForwardedFor: "192.168.1.2",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requestCtx := context.Background()
			if test.requestClient != nil {
				requestCtx = WithClient(requestCtx, test.requestClient)
			}

			if test.requestPeer != nil {
				requestCtx = peer.NewContext(requestCtx, test.requestPeer)
			}

			propagator := &Propagator{}
			md := propagator.FromContext(requestCtx)

			resultCtx := propagator.InjectContext(requestCtx, md)

			// Explicitly compare exported fields because cmp.Diff doesn't work
			// when there are unexported fields
			rc := FromContext(resultCtx)
			require.NotNil(t, rc)
			assert.Equal(t, test.wantClient.IP, rc.IP)
			assert.Equal(t, test.wantClient.ForwardedFor, rc.ForwardedFor)
			assert.Equal(t, test.wantClient.UserAgent, rc.UserAgent)
		})
	}
}

func TestBaseIP(t *testing.T) {
	tests := []struct {
		name string
		addr net.Addr
		want string
	}{
		{
			name: "TCP address",
			addr: &net.TCPAddr{
				IP:   net.ParseIP("127.0.127.2"),
				Port: 448,
			},
			want: "127.0.127.2",
		},
		{
			name: "UDP address",
			addr: &net.UDPAddr{
				IP:   net.ParseIP("127.0.0.1"),
				Port: 448,
			},
			want: "127.0.0.1",
		},
		{
			name: "Other address",
			addr: &net.UnixAddr{
				Name: "foobar",
			},
			want: "foobar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := baseIP(tt.addr); got != tt.want {
				t.Errorf("baseIP() = %v, want %v", got, tt.want)
			}
		})
	}
}
