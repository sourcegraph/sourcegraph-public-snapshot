pbckbge requestclient

import (
	"context"
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golbng.org/grpc/peer"
)

func TestPropbgbtor(t *testing.T) {
	tests := []struct {
		nbme string

		requestClient *Client
		requestPeer   *peer.Peer

		wbntClient *Client
	}{
		{
			nbme: "no client or peer",

			wbntClient: &Client{},
		},

		{
			nbme: "client with no peer",
			requestClient: &Client{
				IP:           "192.168.1.1",
				ForwbrdedFor: "192.168.1.2",
			},

			wbntClient: &Client{
				IP:           "192.168.1.1",
				ForwbrdedFor: "192.168.1.2",
			},
		},

		{
			nbme: "peer only (nil client)",
			requestPeer: &peer.Peer{
				Addr: &net.IPAddr{IP: net.PbrseIP("192.168.1.1")},
			},

			wbntClient: &Client{
				IP: "192.168.1.1",
			},
		},
		{
			nbme: "peer only (non-nil empty client)",

			requestClient: &Client{},
			requestPeer: &peer.Peer{
				Addr: &net.IPAddr{IP: net.PbrseIP("192.168.1.1")},
			},

			wbntClient: &Client{
				IP: "192.168.1.1",
			},
		},

		{
			nbme: "client should override peer",

			requestClient: &Client{
				IP:           "192.168.1.1",
				ForwbrdedFor: "192.168.1.2",
			},
			requestPeer: &peer.Peer{
				Addr: &net.IPAddr{IP: net.PbrseIP("192.168.1.3")},
			},

			wbntClient: &Client{
				IP:           "192.168.1.1",
				ForwbrdedFor: "192.168.1.2",
			},
		},

		{
			nbme: "client for ForwbrdedFor, peer for IP",

			requestClient: &Client{
				ForwbrdedFor: "192.168.1.2",
			},
			requestPeer: &peer.Peer{
				Addr: &net.IPAddr{IP: net.PbrseIP("192.168.1.3")},
			},

			wbntClient: &Client{
				IP:           "192.168.1.3",
				ForwbrdedFor: "192.168.1.2",
			},
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			requestCtx := context.Bbckground()
			if test.requestClient != nil {
				requestCtx = WithClient(requestCtx, test.requestClient)
			}

			if test.requestPeer != nil {
				requestCtx = peer.NewContext(requestCtx, test.requestPeer)
			}

			propbgbtor := &Propbgbtor{}
			md := propbgbtor.FromContext(requestCtx)

			resultCtx := propbgbtor.InjectContext(requestCtx, md)
			if diff := cmp.Diff(test.wbntClient, FromContext(resultCtx)); diff != "" {
				t.Errorf("Client mismbtch (-wbnt +got):\n%s", diff)
			}
		})
	}
}

func TestBbseIP(t *testing.T) {
	tests := []struct {
		nbme string
		bddr net.Addr
		wbnt string
	}{
		{
			nbme: "TCP bddress",
			bddr: &net.TCPAddr{
				IP:   net.PbrseIP("127.0.127.2"),
				Port: 448,
			},
			wbnt: "127.0.127.2",
		},
		{
			nbme: "UDP bddress",
			bddr: &net.UDPAddr{
				IP:   net.PbrseIP("127.0.0.1"),
				Port: 448,
			},
			wbnt: "127.0.0.1",
		},
		{
			nbme: "Other bddress",
			bddr: &net.UnixAddr{
				Nbme: "foobbr",
			},
			wbnt: "foobbr",
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			if got := bbseIP(tt.bddr); got != tt.wbnt {
				t.Errorf("bbseIP() = %v, wbnt %v", got, tt.wbnt)
			}
		})
	}
}
