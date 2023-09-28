package shared

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		logtest.InitWithLevel(m, log.LevelNone)
	}
	os.Exit(m.Run())
}

func TestGetVCSSyncer(t *testing.T) {
	tempReposDir, err := os.MkdirTemp("", "TestGetVCSSyncer")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempReposDir); err != nil {
			t.Fatal(err)
		}
	})
	tempCoursierCacheDir := filepath.Join(tempReposDir, "coursier")

	repo := api.RepoName("foo/bar")
	extsvcStore := dbmocks.NewMockExternalServiceStore()
	repoStore := dbmocks.NewMockRepoStore()

	repoStore.GetByNameFunc.SetDefaultHook(func(ctx context.Context, name api.RepoName) (*types.Repo, error) {
		return &types.Repo{
			ExternalRepo: api.ExternalRepoSpec{
				ServiceType: extsvc.TypePerforce,
			},
			Sources: map[string]*types.SourceInfo{
				"a": {
					ID:       "abc",
					CloneURL: "example.com",
				},
			},
		}, nil
	})

	extsvcStore.GetByIDFunc.SetDefaultHook(func(ctx context.Context, i int64) (*types.ExternalService, error) {
		return &types.ExternalService{
			ID:          1,
			Kind:        extsvc.KindPerforce,
			DisplayName: "test",
			Config:      extsvc.NewEmptyConfig(),
		}, nil
	})

	s, err := getVCSSyncer(context.Background(), &newVCSSyncerOpts{
		externalServiceStore: extsvcStore,
		repoStore:            repoStore,
		depsSvc:              new(dependencies.Service),
		repo:                 repo,
		reposDir:             tempReposDir,
		coursierCacheDir:     tempCoursierCacheDir,
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, "perforce", s.Type())
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
