package sourcegraph

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"os/exec"
	"reflect"
	"testing"
	"time"

	"strings"
	"sync"

	"sourcegraph.com/sqs/pbtypes"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

func TestPerRPCCredentials(t *testing.T) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	s := grpc.NewServer()
	stopped := false
	cleanupBarrier := make(chan struct{})
	go func() {
		if err := s.Serve(l); err != nil && !stopped {
			t.Fatal(err)
		}
		cleanupBarrier <- struct{}{}
	}()
	defer func() {
		stopped = true
		s.Stop()
		<-cleanupBarrier
	}()

	var ms testMetaServer
	RegisterMetaServer(s, &ms)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			time.Sleep(time.Duration(i%10) * 5 * time.Millisecond)

			key := fmt.Sprintf("key%d", i)

			ctx := context.Background()
			ctx = WithGRPCEndpoint(ctx, &url.URL{Host: l.Addr().String()})
			ctx = WithCredentials(ctx, oauth2.StaticTokenSource(&oauth2.Token{TokenType: "x", AccessToken: key}))
			ctx = metadata.NewContext(ctx, metadata.MD{"want-access-token": []string{"x " + key}})
			c, _ := NewClientFromContext(ctx)
			if _, err := c.Meta.Status(ctx, &pbtypes.Void{}); err != nil {
				t.Fatal(err)
			}
		}(i)
	}
	wg.Wait()

	out, err := exec.Command("netstat", "-ntap").CombinedOutput()
	if err == nil {
		lines := bytes.Split(out, []byte("\n"))
		var conns, timeWaits int
		addr := strings.Replace(l.Addr().String(), "[::]", "::1", 1)
		for _, line := range lines {
			if bytes.Contains(line, []byte(addr)) {
				conns++
				if bytes.Contains(line, []byte("TIME_WAIT")) {
					timeWaits++
				}
			}
		}
		t.Logf("lingering connections count: %d", conns)
		t.Logf("         in TIME_WAIT state: %d", timeWaits)
		t.Log("(ideally, there should be 0 lingering connections)")
	} else {
		t.Logf("warning: error running `netstat -ntap` to check # of TIME_WAIT conns: %s", err)
	}
}

type testMetaServer struct {
	MetaServer
}

func (s *testMetaServer) Status(ctx context.Context, _ *pbtypes.Void) (*ServerStatus, error) {
	md, _ := metadata.FromContext(ctx)
	if want, got := md["want-access-token"], md["authorization"]; !reflect.DeepEqual(got, want) {
		return nil, grpc.Errorf(codes.Unknown, "got access-token %q, want %q", got, want)
	}
	return &ServerStatus{}, nil
}
