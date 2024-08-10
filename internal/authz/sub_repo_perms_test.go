package authz

import (
	"context"
	"io/fs"
	"net/netip"
	"testing"

	"github.com/gobwas/glob"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/fileutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestFilterActorPaths(t *testing.T) {
	tests := []struct {
		name            string
		paths           []string
		enabledFunc     func() bool
		permissionsFunc func(context.Context, int32, api.RepoName) (FilePermissionFunc, error)
		ipSource        IPSource

		expectedPaths []string
		expectedError error
	}{
		{
			name:        "basic filtering",
			paths:       []string{"file1", "file2", "file3"},
			enabledFunc: func() bool { return true },
			permissionsFunc: func(context.Context, int32, api.RepoName) (FilePermissionFunc, error) {
				return func(path string, ip netip.Addr) (Perms, error) {
					if path == "file1" {
						return Read, nil
					}
					return None, nil
				}, nil
			},
			ipSource: IPSourceFunc(func() (netip.Addr, error) {
				ip := netip.MustParseAddr("127.0.0.1")
				return ip, nil
			}),
			expectedPaths: []string{"file1"},
			expectedError: nil,
		},
		{
			name:        "IP address check",
			paths:       []string{"file1", "file2", "file3"},
			enabledFunc: func() bool { return true },
			permissionsFunc: func(context.Context, int32, api.RepoName) (FilePermissionFunc, error) {
				return func(path string, ip netip.Addr) (Perms, error) {
					if path == "file2" && ip == netip.MustParseAddr("192.168.1.1") {
						return Read, nil
					}
					return None, nil
				}, nil
			},
			ipSource: IPSourceFunc(func() (netip.Addr, error) {
				ip := netip.MustParseAddr("192.168.1.1")
				return ip, nil
			}),

			expectedPaths: []string{"file2"},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewMockSubRepoPermissionChecker()
			ctx := context.Background()
			a := &actor.Actor{UID: 1}
			ctx = actor.WithActor(ctx, a)
			repo := api.RepoName("foo")

			checker.EnabledFunc.SetDefaultHook(tt.enabledFunc)
			checker.FilePermissionsFuncFunc.SetDefaultHook(tt.permissionsFunc)

			filtered, err := FilterActorPaths(ctx, checker, a, tt.ipSource, repo, tt.paths)
			if tt.expectedError != nil {
				require.ErrorIs(t, err, tt.expectedError)

				return
			}

			require.NoError(t, err)
			if diff := cmp.Diff(tt.expectedPaths, filtered); diff != "" {
				t.Fatal(diff)
			}
		})
	}

	t.Run("propagates error if ip source function fails", func(t *testing.T) {
		checker := NewMockSubRepoPermissionChecker()
		checker.EnabledFunc.SetDefaultHook(func() bool { return true })
		checker.FilePermissionsFuncFunc.SetDefaultHook(func(context.Context, int32, api.RepoName) (FilePermissionFunc, error) {
			return func(path string, ip netip.Addr) (Perms, error) {
				return Read, nil
			}, nil
		})

		expectedErr := errors.New("getting the IP failed for some unknown reason")
		ipSource := IPSourceFunc(func() (netip.Addr, error) {
			return netip.Addr{}, errors.Wrap(expectedErr, "hmm...")
		})

		ctx := context.Background()
		a := &actor.Actor{UID: 1}
		ctx = actor.WithActor(ctx, a)
		repo := api.RepoName("foo")

		_, err := FilterActorPaths(ctx, checker, a, ipSource, repo, []string{"file1"})
		require.ErrorIs(t, err, expectedErr)
	})
}

func TestCanReadAllPaths(t *testing.T) {
	tests := []struct {
		name            string
		paths           []string
		enabledFunc     func() bool
		permissionsFunc func(context.Context, int32, api.RepoName) (FilePermissionFunc, error)
		ipSource        IPSource

		expectedCanReadAll bool
		expectedCanReadAny bool
	}{
		{
			name:        "can read all paths",
			paths:       []string{"file1", "file2", "file3"},
			enabledFunc: func() bool { return true },
			permissionsFunc: func(context.Context, int32, api.RepoName) (FilePermissionFunc, error) {
				return func(path string, ip netip.Addr) (Perms, error) {
					return Read, nil
				}, nil
			},
			ipSource: IPSourceFunc(func() (netip.Addr, error) {
				return netip.MustParseAddr("127.0.0.1"), nil
			}),
			expectedCanReadAll: true,
			expectedCanReadAny: true,
		},
		{
			name:        "can read all but one",
			paths:       []string{"file1", "file2", "file3"},
			enabledFunc: func() bool { return true },
			permissionsFunc: func(context.Context, int32, api.RepoName) (FilePermissionFunc, error) {
				return func(path string, ip netip.Addr) (Perms, error) {
					if path == "file2" {
						return None, nil
					}
					return Read, nil
				}, nil
			},
			ipSource: IPSourceFunc(func() (netip.Addr, error) {
				return netip.MustParseAddr("127.0.0.1"), nil
			}),
			expectedCanReadAll: false,
			expectedCanReadAny: true,
		},
		{
			name:        "can only read one",
			paths:       []string{"file1", "file2", "file3"},
			enabledFunc: func() bool { return true },
			permissionsFunc: func(context.Context, int32, api.RepoName) (FilePermissionFunc, error) {
				return func(path string, ip netip.Addr) (Perms, error) {
					if path == "file2" {
						return Read, nil
					}
					return None, nil
				}, nil
			},
			ipSource: IPSourceFunc(func() (netip.Addr, error) {
				return netip.MustParseAddr("127.0.0.1"), nil
			}),
			expectedCanReadAll: false,
			expectedCanReadAny: true,
		},
		{
			name:        "IP address check - pass",
			paths:       []string{"file1", "file2", "file3"},
			enabledFunc: func() bool { return true },
			permissionsFunc: func(context.Context, int32, api.RepoName) (FilePermissionFunc, error) {
				return func(path string, ip netip.Addr) (Perms, error) {
					if path == "file2" && ip == netip.MustParseAddr("192.168.1.1") {
						return Read, nil
					}
					return None, nil
				}, nil
			},
			ipSource: IPSourceFunc(func() (netip.Addr, error) {
				return netip.MustParseAddr("192.168.1.1"), nil
			}),
			expectedCanReadAll: false,
			expectedCanReadAny: true,
		},
		{
			name:        "IP address check - fail",
			paths:       []string{"file1", "file2", "file3"},
			enabledFunc: func() bool { return true },
			permissionsFunc: func(context.Context, int32, api.RepoName) (FilePermissionFunc, error) {
				return func(path string, ip netip.Addr) (Perms, error) {
					if path == "file2" && ip == netip.MustParseAddr("192.168.1.1") {
						return Read, nil
					}
					return None, nil
				}, nil
			},
			ipSource: IPSourceFunc(func() (netip.Addr, error) {
				return netip.MustParseAddr("127.0.0.1"), nil
			}),
			expectedCanReadAll: false,
			expectedCanReadAny: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			checker := NewMockSubRepoPermissionChecker()
			checker.EnabledFunc.SetDefaultHook(tt.enabledFunc)
			checker.FilePermissionsFuncFunc.SetDefaultHook(tt.permissionsFunc)

			ctx := context.Background()
			a := &actor.Actor{UID: 1}
			ctx = actor.WithActor(ctx, a)
			repo := api.RepoName("foo")

			canReadAll, err := CanReadAllPaths(ctx, checker, a, tt.ipSource, repo, tt.paths)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.expectedCanReadAll, canReadAll); diff != "" {
				t.Fatalf("CanReadAllPaths: expected %v, got %v", tt.expectedCanReadAll, canReadAll)
			}

			canReadAny, err := CanReadAnyPath(ctx, checker, a, tt.ipSource, repo, tt.paths)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.expectedCanReadAny, canReadAny); diff != "" {
				t.Fatalf("CanReadAnyPath: expected %v, got %v", tt.expectedCanReadAny, canReadAny)
			}
		})
	}

	t.Run("CanReadAllPaths IP source error propagation", func(t *testing.T) {
		checker := NewMockSubRepoPermissionChecker()
		checker.EnabledFunc.SetDefaultHook(func() bool { return true })
		checker.FilePermissionsFuncFunc.SetDefaultHook(func(context.Context, int32, api.RepoName) (FilePermissionFunc, error) {
			return func(path string, ip netip.Addr) (Perms, error) {
				return Read, nil
			}, nil
		})

		ipSource := IPSourceFunc(func() (netip.Addr, error) {
			return netip.Addr{}, errors.New("IP source error")
		})

		ctx := context.Background()
		a := &actor.Actor{UID: 1}
		ctx = actor.WithActor(ctx, a)
		repo := api.RepoName("foo")
		paths := []string{"file1", "file2", "file3"}

		expectedError := errors.New("getting the IP address for checking permissions: IP source error")

		canReadAll, err := CanReadAllPaths(ctx, checker, a, ipSource, repo, paths)
		if err == nil || err.Error() != expectedError.Error() {
			t.Errorf("CanReadAllPaths error: expected %v, got %v", expectedError, err)
		}
		if canReadAll {
			t.Errorf("CanReadAllPaths: expected false, got true")
		}
	})

	t.Run("CanReadAnyPath IP source error propagation", func(t *testing.T) {
		checker := NewMockSubRepoPermissionChecker()
		checker.EnabledFunc.SetDefaultHook(func() bool { return true })
		checker.FilePermissionsFuncFunc.SetDefaultHook(func(context.Context, int32, api.RepoName) (FilePermissionFunc, error) {
			return func(path string, ip netip.Addr) (Perms, error) {
				return Read, nil
			}, nil
		})

		ipSource := IPSourceFunc(func() (netip.Addr, error) {
			return netip.Addr{}, errors.New("IP source error")
		})

		ctx := context.Background()
		a := &actor.Actor{UID: 1}
		ctx = actor.WithActor(ctx, a)
		repo := api.RepoName("foo")
		paths := []string{"file1", "file2", "file3"}

		expectedError := errors.New("getting the IP address for checking permissions: IP source error")

		canReadAny, err := CanReadAnyPath(ctx, checker, a, ipSource, repo, paths)
		if err == nil || err.Error() != expectedError.Error() {
			t.Errorf("CanReadAnyPath error: expected %v, got %v", expectedError, err)
		}
		if canReadAny {
			t.Errorf("CanReadAnyPath: expected false, got true")
		}
	})
}

func TestSubRepoEnabled(t *testing.T) {
	t.Run("checker is nil", func(t *testing.T) {
		if SubRepoEnabled(nil) {
			t.Errorf("expected checker to be invalid since it is nil")
		}
	})
	t.Run("checker is not enabled", func(t *testing.T) {
		checker := NewMockSubRepoPermissionChecker()
		checker.EnabledFunc.SetDefaultHook(func() bool {
			return false
		})
		if SubRepoEnabled(checker) {
			t.Errorf("expected checker to be invalid since it is disabled")
		}
	})
	t.Run("checker is enabled", func(t *testing.T) {
		checker := NewMockSubRepoPermissionChecker()
		checker.EnabledFunc.SetDefaultHook(func() bool {
			return true
		})
		if !SubRepoEnabled(checker) {
			t.Errorf("expected checker to be valid since it is enabled")
		}
	})
}

func TestFileInfoPath(t *testing.T) {
	t.Run("adding trailing slash to directory", func(t *testing.T) {
		fi := &fileutil.FileInfo{
			Name_: "app",
			Mode_: fs.ModeDir,
		}
		assert.Equal(t, "app/", fileInfoPath(fi))
	})
	t.Run("doesn't add trailing slash if not directory", func(t *testing.T) {
		fi := &fileutil.FileInfo{
			Name_: "my-file.txt",
		}
		assert.Equal(t, "my-file.txt", fileInfoPath(fi))
	})
}

func TestGlobMatchOnlyDirectories(t *testing.T) {
	g, err := glob.Compile("**/", '/')
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, g.Match("foo/"))
	assert.True(t, g.Match("foo/thing/"))
	assert.False(t, g.Match("foo/thing"))
	assert.False(t, g.Match("/foo/thing"))
}
