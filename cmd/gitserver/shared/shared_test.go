package shared

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
	"google.golang.org/grpc"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		logtest.InitWithLevel(m, log.LevelNone)
	}
	os.Exit(m.Run())
}

func TestMethodSpecificStreamInterceptor(t *testing.T) {
	tests := []struct {
		name string

		matchedMethod string
		testMethod    string

		expectedInterceptorCalled bool
	}{
		{
			name: "allowed method",

			matchedMethod: "allowedMethod",
			testMethod:    "allowedMethod",

			expectedInterceptorCalled: true,
		},
		{
			name: "not allowed method",

			matchedMethod: "allowedMethod",
			testMethod:    "otherMethod",

			expectedInterceptorCalled: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			interceptorCalled := false
			interceptor := methodSpecificStreamInterceptor(test.matchedMethod, func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
				interceptorCalled = true
				return handler(srv, ss)
			})

			handlerCalled := false
			noopHandler := func(srv any, ss grpc.ServerStream) error {
				handlerCalled = true
				return nil
			}

			err := interceptor(nil, nil, &grpc.StreamServerInfo{FullMethod: test.testMethod}, noopHandler)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if !handlerCalled {
				t.Error("expected handler to be called")
			}

			if diff := cmp.Diff(test.expectedInterceptorCalled, interceptorCalled); diff != "" {
				t.Fatalf("unexpected interceptor called value (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMethodSpecificUnaryInterceptor(t *testing.T) {
	tests := []struct {
		name string

		matchedMethod string
		testMethod    string

		expectedInterceptorCalled bool
	}{
		{
			name: "allowed method",

			matchedMethod: "allowedMethod",
			testMethod:    "allowedMethod",

			expectedInterceptorCalled: true,
		},
		{
			name: "not allowed method",

			matchedMethod: "allowedMethod",
			testMethod:    "otherMethod",

			expectedInterceptorCalled: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			interceptorCalled := false
			interceptor := methodSpecificUnaryInterceptor(test.matchedMethod, func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
				interceptorCalled = true
				return handler(ctx, req)
			})

			handlerCalled := false
			noopHandler := func(ctx context.Context, req any) (any, error) {
				handlerCalled = true
				return nil, nil
			}

			_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: test.testMethod}, noopHandler)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if !handlerCalled {
				t.Error("expected handler to be called")
			}

			if diff := cmp.Diff(test.expectedInterceptorCalled, interceptorCalled); diff != "" {
				t.Fatalf("unexpected interceptor called value (-want +got):\n%s", diff)
			}

		})
	}
}
