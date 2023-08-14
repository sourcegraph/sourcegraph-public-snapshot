package shared

import (
	"context"
	"flag"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
	"google.golang.org/grpc"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server"
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

	_, ok := s.(*server.PerforceDepotSyncer)
	if !ok {
		t.Fatalf("Want *server.PerforceDepotSyncer, got %T", s)
	}
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

func TestSetupAndClearTmp(t *testing.T) {
	root := t.TempDir()

	// All non .git paths should become .git
	mkFiles(t, root,
		"github.com/foo/baz/.git/HEAD",
		"example.org/repo/.git/HEAD",

		// Needs to be deleted
		".tmp/foo",
		".tmp/baz/bam",

		// Older tmp cleanups that failed
		".tmp-old123/foo",
	)

	tmp, err := setupAndClearTmp(logtest.Scoped(t), root)
	if err != nil {
		t.Fatal(err)
	}

	// Straight after cleaning .tmp should be empty
	assertPaths(t, filepath.Join(root, ".tmp"), ".")

	// tmp should exist
	if info, err := os.Stat(tmp); err != nil {
		t.Fatal(err)
	} else if !info.IsDir() {
		t.Fatal("tmpdir is not a dir")
	}

	// tmp should be on the same mount as root, ie root is parent.
	if filepath.Dir(tmp) != root {
		t.Fatalf("tmp is not under root: tmp=%s root=%s", tmp, root)
	}

	// Wait until async cleaning is done
	for i := 0; i < 1000; i++ {
		found := false
		files, err := os.ReadDir(root)
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range files {
			found = found || strings.HasPrefix(f.Name(), ".tmp-old")
		}
		if !found {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Only files should be the repo files
	assertPaths(
		t,
		root,
		"github.com/foo/baz/.git/HEAD",
		"example.org/repo/.git/HEAD",
		".tmp",
	)
}

func TestSetupAndClearTmp_Empty(t *testing.T) {
	root := t.TempDir()

	_, err := setupAndClearTmp(logtest.Scoped(t), root)
	if err != nil {
		t.Fatal(err)
	}

	// No files, just the empty .tmp dir should exist
	assertPaths(t, root, ".tmp")
}

// assertPaths checks that all paths under want exist. It excludes non-empty directories
func assertPaths(t *testing.T, root string, want ...string) {
	t.Helper()
	notfound := make(map[string]struct{})
	for _, p := range want {
		notfound[p] = struct{}{}
	}
	var unwanted []string
	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if empty, err := isEmptyDir(path); err != nil {
				t.Fatal(err)
			} else if !empty {
				return nil
			}
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if _, ok := notfound[rel]; ok {
			delete(notfound, rel)
		} else {
			unwanted = append(unwanted, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(notfound) > 0 {
		var paths []string
		for p := range notfound {
			paths = append(paths, p)
		}
		sort.Strings(paths)
		t.Errorf("did not find expected paths: %s", strings.Join(paths, " "))
	}
	if len(unwanted) > 0 {
		sort.Strings(unwanted)
		t.Errorf("found unexpected paths: %s", strings.Join(unwanted, " "))
	}
}

func isEmptyDir(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

func mkFiles(t *testing.T, root string, paths ...string) {
	t.Helper()
	for _, p := range paths {
		if err := os.MkdirAll(filepath.Join(root, filepath.Dir(p)), os.ModePerm); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(root, p), nil)
	}
}

func writeFile(t *testing.T, path string, content []byte) {
	t.Helper()
	err := os.WriteFile(path, content, 0o666)
	if err != nil {
		t.Fatal(err)
	}
}
