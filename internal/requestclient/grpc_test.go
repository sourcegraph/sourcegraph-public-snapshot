package requestclient

import (
	"context"
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func TestInterceptors(t *testing.T) {
	const ipAddress = "127.0.2.1"

	tests := []struct {
		name     string
		peer     *peer.Peer
		wantPeer bool
	}{
		{
			name:     "no peer",
			peer:     nil,
			wantPeer: false,
		},
		{
			name:     "with peer",
			peer:     &peer.Peer{Addr: &net.IPAddr{IP: net.ParseIP(ipAddress)}},
			wantPeer: true,
		},
	}

	t.Run("unary", func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				ctx := context.Background()
				if test.peer != nil {
					ctx = peer.NewContext(ctx, test.peer)
				}

				req := "foo"
				info := &grpc.UnaryServerInfo{}

				called := false
				handler := func(ctx context.Context, req interface{}) (interface{}, error) {
					called = true

					if !test.wantPeer {
						c := FromContext(ctx)
						if c != nil {
							t.Error("client set in context")
						}

						return "foo", nil
					}

					client := FromContext(ctx)
					if client == nil {
						t.Fatal("client not set in context")
					}

					if diff := cmp.Diff(client.IP, ipAddress); diff != "" {
						t.Errorf("IP mismatch (-want +got):\n%s", diff)
					}

					return "foo", nil
				}

				resp, err := UnaryServerInterceptor(ctx, req, info, handler)
				if err != nil {
					t.Fatal(err)
				}

				if !called {
					t.Fatal("handler not called")
				}

				if resp != req {
					t.Errorf("got %v, want %v", resp, req)
				}
			})

		}
	})

	t.Run("stream", func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				ctx := context.Background()
				if test.peer != nil {
					ctx = peer.NewContext(ctx, test.peer)
				}

				req := "foo"

				called := false
				handler := func(_ any, ss grpc.ServerStream) error {
					called = true

					if !test.wantPeer {
						c := FromContext(ss.Context())
						if c != nil {
							t.Error("client set in context")
						}

						return ss.SendMsg("foo")
					}

					client := FromContext(ss.Context())
					if client == nil {
						t.Fatal("client not set in context")
					}

					if diff := cmp.Diff(client.IP, ipAddress); diff != "" {
						t.Errorf("IP mismatch (-want +got):\n%s", diff)
					}

					return ss.SendMsg("foo")
				}

				srv := struct{}{}

				ss := newMockStream(ctx)
				info := &grpc.StreamServerInfo{}

				err := StreamServerInterceptor(srv, ss, info, handler)
				if err != nil {
					t.Fatal(err)
				}

				if !called {
					t.Fatal("handler not called")
				}

				resp := ss.GetServerMessage()
				if resp != req {
					t.Errorf("got %v, want %v", resp, req)
				}
			})

		}
	})
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

// mockStream is a mock implementation of grpc.ServerStream. You can
// inspect the messages sent to and from the client.
type mockStream struct {
	ctx                context.Context
	sentFromServer     chan any
	receivedFromClient chan any
}

func newMockStream(ctx context.Context) *mockStream {
	return &mockStream{
		ctx:                ctx,
		sentFromServer:     make(chan any, 1),
		receivedFromClient: make(chan any, 1),
	}
}

func (m *mockStream) SetHeader(md metadata.MD) error {
	// No-op for testing
	return nil
}

func (m *mockStream) SendHeader(md metadata.MD) error {
	// No-op for testing
	return nil

}

func (m *mockStream) SetTrailer(md metadata.MD) {
	// No-op for testing
}

func (m *mockStream) Context() context.Context {
	return m.ctx
}

func (m *mockStream) SendMsg(message any) error {
	// Save the message to be asserted in tests
	m.sentFromServer <- message
	return nil
}

func (m *mockStream) RecvMsg(message any) error {
	// Save the message to be asserted in tests
	m.receivedFromClient <- message
	return nil
}

// GetServerMessage returns next message sent from the server to the
// client.
func (m mockStream) GetServerMessage() any {
	return <-m.sentFromServer
}

//	GetClientMessage returns next message sent from the client to the
//
// server.
func (m mockStream) GetClientMessage() any {
	return <-m.receivedFromClient
}

var _ grpc.ServerStream = &mockStream{}
