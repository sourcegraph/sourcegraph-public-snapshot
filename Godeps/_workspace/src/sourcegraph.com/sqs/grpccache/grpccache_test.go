package grpccache_test

import (
	"net"
	"reflect"
	"testing"
	"time"

	"strconv"

	"sourcegraph.com/sqs/grpccache"
	"sourcegraph.com/sqs/grpccache/testpb"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func TestGRPCCache(t *testing.T) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}

	var ts testServer
	gs := grpc.NewServer()
	testpb.RegisterTestServer(gs, &testpb.CachedTestServer{TestServer: &ts})
	go func() {
		if err := gs.Serve(l); err != nil {
			t.Log("warning: Serve:", err)
		}
	}()
	defer gs.Stop()

	cc, err := grpc.Dial(l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	c := &testpb.CachedTestClient{TestClient: testpb.NewTestClient(cc), Cache: &grpccache.Cache{}}
	c.Cache.Log = true

	ctx := context.Background()

	if want := 0; len(ts.calls) != want {
		t.Errorf("got %d calls (%+v), want %d", len(ts.calls), ts.calls, want)
	}

	noopCtxFunc := func(ctx context.Context) context.Context { return ctx }

	testNotCached := func(op *testpb.TestOp, ctxFunc func(context.Context) context.Context) {
		if ctxFunc == nil {
			ctxFunc = noopCtxFunc
		}

		beforeNumCalls := len(ts.calls)
		r, err := c.TestMethod(ctxFunc(ctx), op)
		if err != nil {
			t.Fatal(err)
		}
		if want := (&testpb.TestResult{X: op.A}); !reflect.DeepEqual(r, want) {
			t.Errorf("got %#v, want %#v", r, want)
		}
		if want := beforeNumCalls + 1; len(ts.calls) != want {
			t.Errorf("server did not handle call %+v (client handled it from cache), wanted it to be uncached", op)
		}
	}

	testCached := func(op *testpb.TestOp, ctxFunc func(context.Context) context.Context) {
		if ctxFunc == nil {
			ctxFunc = noopCtxFunc
		}

		beforeNumCalls := len(ts.calls)
		r, err := c.TestMethod(ctxFunc(ctx), op)
		if err != nil {
			t.Fatal(err)
		}
		if want := (&testpb.TestResult{X: op.A}); !reflect.DeepEqual(r, want) {
			t.Errorf("got %#v, want %#v", r, want)
		}
		if want := beforeNumCalls; len(ts.calls) != want {
			t.Errorf("server handled call %+v, wanted it to be client-cached", op)
		}
	}

	// Test caching (with no expiration)
	ts.maxAge = 999 * time.Hour
	testNotCached(&testpb.TestOp{A: 1}, nil)
	testCached(&testpb.TestOp{A: 1}, nil)
	testNotCached(&testpb.TestOp{A: 2}, nil)
	testNotCached(&testpb.TestOp{A: 2, B: []*testpb.T{{A: true}}}, nil)
	testCached(&testpb.TestOp{A: 2}, nil)
	testCached(&testpb.TestOp{A: 2, B: []*testpb.T{{A: true}}}, nil)
	testCached(&testpb.TestOp{A: 1}, nil)
	testNotCached(&testpb.TestOp{A: 3}, nil)

	// Test cache expiration
	ts.maxAge = time.Millisecond * 250
	testNotCached(&testpb.TestOp{A: 100}, nil)
	testCached(&testpb.TestOp{A: 100}, nil)
	testCached(&testpb.TestOp{A: 100}, nil)
	testNotCached(&testpb.TestOp{A: 111}, nil)
	time.Sleep(ts.maxAge)
	testNotCached(&testpb.TestOp{A: 100}, nil)
	testNotCached(&testpb.TestOp{A: 111}, nil)
	testCached(&testpb.TestOp{A: 100}, nil)
	testCached(&testpb.TestOp{A: 100}, nil)

	c.Cache.Clear()

	// Test cache max size
	c.Cache.MaxSize = 8
	testNotCached(&testpb.TestOp{A: 200}, nil)
	testCached(&testpb.TestOp{A: 200}, nil)
	testNotCached(&testpb.TestOp{A: 201}, nil)
	testCached(&testpb.TestOp{A: 201}, nil)
	testNotCached(&testpb.TestOp{A: 202}, nil) // exceeds max size
	testNotCached(&testpb.TestOp{A: 202}, nil)
	c.Cache.MaxSize = 0
	testNotCached(&testpb.TestOp{A: 202}, nil)
	testCached(&testpb.TestOp{A: 202}, nil)

	// Test gzip above a certain length
	c.Cache.MaxSize = 10000
	orig := grpccache.MinByteGzip
	grpccache.MinByteGzip = 1
	testNotCached(&testpb.TestOp{A: 302}, nil)
	testCached(&testpb.TestOp{A: 302}, nil)
	grpccache.MinByteGzip = orig
	c.Cache.MaxSize = 0

	// Test KeyPart
	kp := 0
	c.Cache.KeyPart = func(context.Context) string {
		kp++
		return strconv.Itoa(kp)
	}
	testNotCached(&testpb.TestOp{A: 400}, nil)
	testNotCached(&testpb.TestOp{A: 400}, nil)
	c.Cache.KeyPart = nil

	// Test NoCache
	testNotCached(&testpb.TestOp{A: 500}, grpccache.NoCache)
	testNotCached(&testpb.TestOp{A: 500}, grpccache.NoCache)
}

type testServer struct {
	calls []*testpb.TestOp

	maxAge time.Duration
}

func (s *testServer) TestMethod(ctx context.Context, op *testpb.TestOp) (*testpb.TestResult, error) {
	s.calls = append(s.calls, op)

	{
		// Only the last call to SetCacheControl should take effect. Make
		// a call here that will be overridden by the following call, to
		// ensure that that indeed is the case.
		otherMaxAge := time.Duration(0)
		if s.maxAge == 0 {
			otherMaxAge = 5 * time.Second
		}
		grpccache.SetCacheControl(ctx, grpccache.CacheControl{MaxAge: otherMaxAge})
	}

	// Set cache control.
	grpccache.SetCacheControl(ctx, grpccache.CacheControl{MaxAge: s.maxAge})

	return &testpb.TestResult{X: op.A}, nil
}
