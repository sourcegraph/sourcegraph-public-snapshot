package defaults

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const bufferSize = 1024 * 1024

func TestCloseGRPCConnectionCallback(t *testing.T) {
	listener := bufconn.Listen(bufferSize)
	defer listener.Close()

	// Start a fake GRPC server
	fakeServer := grpc.NewServer()
	defer fakeServer.Stop()

	go func() {
		if err := fakeServer.Serve(listener); err != nil {
			t.Errorf("gRPC server exited with error: %v", err)
			return
		}
	}()

	opts := []grpc.DialOption{
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.DialContext(context.Background(), "doesn't matter", opts...)
	if err != nil {
		t.Fatalf("failed to dial gRPC server: %v", err)
	}

	defer conn.Close() // ensure connection is closed when test ends

	ce := connAndError{conn: conn, dialErr: err}

	// Wait for connection to be ready, or give up after timeout

	connectionInitialized := make(chan struct{})

	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				state := ce.conn.GetState()
				if state != connectivity.Idle && state != connectivity.Connecting {
					close(connectionInitialized)
					return
				}
			}
		}
	}(ctx)

	select {
	case <-ctx.Done():
		t.Fatalf("failed to connect to gRPC server within %s, state: %q", timeout.String(), ce.conn.GetState().String())
	case <-connectionInitialized:
	}

	// Double check that connection is ready
	if state := ce.conn.GetState(); state != connectivity.Ready {
		t.Fatalf("expected gRPC connection to be in state %q, got state: %s", connectivity.Ready, state.String())
	}

	// Run test: run close connection callback
	closeGRPCConnection("", ce)

	// Try closing connection again, should return codes.Canceled error (i.e. connection already closed)
	err = ce.conn.Close()
	if status.Code(err) != codes.Canceled {
		t.Fatalf("expected %q code after closing connection twice, got err: %v", codes.Canceled.String(), err)
	}
}
