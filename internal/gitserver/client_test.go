package gitserver_test

import (
	"archive/zip"
	"bytes"
	"container/list"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"testing/quick"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server/perforce"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func newMockDB() database.DB {
	db := database.NewMockDB()
	db.GitserverReposFunc.SetDefaultReturn(database.NewMockGitserverRepoStore())
	db.FeatureFlagsFunc.SetDefaultReturn(database.NewMockFeatureFlagStore())

	r := database.NewMockRepoStore()
	r.GetByNameFunc.SetDefaultHook(func(ctx context.Context, repoName api.RepoName) (*types.Repo, error) {
		return &types.Repo{
			Name: repoName,
			ExternalRepo: api.ExternalRepoSpec{
				ServiceType: extsvc.TypeGitHub,
			},
		}, nil
	})
	db.ReposFunc.SetDefaultReturn(r)

	return db
}

func TestClient_Archive_ProtoRoundTrip(t *testing.T) {
	var diff string

	fn := func(original gitserver.ArchiveOptions) bool {

		var converted gitserver.ArchiveOptions
		converted.FromProto(original.ToProto("test"))

		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(fn, nil); err != nil {
		t.Errorf("ArchiveOptions proto roundtrip failed (-want +got):\n%s", diff)
	}
}

func TestClient_IsRepoCloneale_ProtoRoundTrip(t *testing.T) {
	var diff string

	fn := func(original protocol.IsRepoCloneableResponse) bool {
		var converted protocol.IsRepoCloneableResponse
		converted.FromProto(original.ToProto())

		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(fn, nil); err != nil {
		t.Errorf("IsRepoCloneableResponse proto roundtrip failed (-want +got):\n%s", diff)
	}
}

func TestClient_RepoStats_ProtoRoundTrip(t *testing.T) {
	var diff string

	fn := func(updatedAt fuzzTime, gitDirBytes int64) bool {
		original := protocol.ReposStats{
			UpdatedAt:   time.Time(updatedAt),
			GitDirBytes: gitDirBytes,
		}

		var converted protocol.ReposStats
		converted.FromProto(original.ToProto())

		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(fn, nil); err != nil {
		t.Errorf("ReposStats proto roundtrip failed (-want +got):\n%s", diff)
	}
}

func TestClient_RepoUpdateRequest_ProtoRoundTrip(t *testing.T) {
	var diff string
	t.Run("request", func(t *testing.T) {
		fn := func(repo api.RepoName, since int64) bool {
			original := protocol.RepoUpdateRequest{
				Repo:  repo,
				Since: time.Duration(since),
			}

			var converted protocol.RepoUpdateRequest
			converted.FromProto(original.ToProto())

			if diff = cmp.Diff(original, converted); diff != "" {
				return false
			}

			return true
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Errorf("RepoUpdateRequest proto roundtrip failed (-want +got):\n%s", diff)
		}
	})

	t.Run("response", func(t *testing.T) {
		fn := func(lastFetched fuzzTime, lastChanged fuzzTime, err string) bool {
			lastFetchedPtr := time.Time(lastFetched)
			lastChangedPtr := time.Time(lastChanged)

			original := protocol.RepoUpdateResponse{
				LastFetched: &lastFetchedPtr,
				LastChanged: &lastChangedPtr,
				Error:       err,
			}
			var converted protocol.RepoUpdateResponse
			converted.FromProto(original.ToProto())

			if diff = cmp.Diff(original, converted); diff != "" {
				return false
			}

			return true
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Errorf("RepoUpdateResponse proto roundtrip failed (-want +got):\n%s", diff)
		}
	})
}

func TestClient_CreateCommitFromPatchRequest_ProtoRoundTrip(t *testing.T) {
	var diff string

	t.Run("request", func(t *testing.T) {
		fn := func(
			repo string,
			baseCommit string,
			patch []byte,
			targetRef string,
			uniqueRef bool,
			pushRef *string,

			commitInfo struct {
				Messages    []string
				AuthorName  string
				AuthorEmail string
				Date        fuzzTime
			},

			pushConfig *protocol.PushConfig,
			gitApplyArgs []string,
		) bool {
			original := protocol.CreateCommitFromPatchRequest{
				Repo:       api.RepoName(repo),
				BaseCommit: api.CommitID(baseCommit),
				Patch:      patch,
				TargetRef:  targetRef,
				UniqueRef:  uniqueRef,
				CommitInfo: protocol.PatchCommitInfo{
					Messages:    commitInfo.Messages,
					AuthorName:  commitInfo.AuthorName,
					AuthorEmail: commitInfo.AuthorEmail,
					Date:        time.Time(commitInfo.Date),
				},
				Push:         pushConfig,
				PushRef:      pushRef,
				GitApplyArgs: gitApplyArgs,
			}
			var converted protocol.CreateCommitFromPatchRequest
			converted.FromProto(original.ToProto())

			if diff = cmp.Diff(original, converted); diff != "" {
				return false
			}

			return true
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Errorf("CreateCommitFromPatchRequest proto roundtrip failed (-want +got):\n%s", diff)
		}
	})

	t.Run("response", func(t *testing.T) {
		fn := func(original protocol.CreateCommitFromPatchResponse) bool {
			var converted protocol.CreateCommitFromPatchResponse
			converted.FromProto(original.ToProto())

			if diff = cmp.Diff(original, converted); diff != "" {
				return false
			}

			return true
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Errorf("CreateCommitFromPatchResponse proto roundtrip failed (-want +got):\n%s", diff)
		}
	})
}

func TestClient_BatchLog_ProtoRoundTrip(t *testing.T) {
	var diff string

	t.Run("request", func(t *testing.T) {
		fn := func(original protocol.BatchLogRequest) bool {
			var converted protocol.BatchLogRequest
			converted.FromProto(original.ToProto())

			if diff = cmp.Diff(original, converted); diff != "" {
				return false
			}

			return true
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Errorf("BatchChangesLogResponse proto roundtrip failed (-want +got):\n%s", diff)
		}
	})

	t.Run("response", func(t *testing.T) {
		fn := func(original protocol.BatchLogResponse) bool {
			var converted protocol.BatchLogResponse
			converted.FromProto(original.ToProto())

			if diff = cmp.Diff(original, converted); diff != "" {
				return false
			}

			return true
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Errorf("BatchChangesLogResponse proto roundtrip failed (-want +got):\n%s", diff)
		}
	})

}

func TestClient_RepoCloneProgress_ProtoRoundTrip(t *testing.T) {
	var diff string

	fn := func(original protocol.RepoCloneProgress) bool {
		var converted protocol.RepoCloneProgress
		converted.FromProto(original.ToProto())

		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(fn, nil); err != nil {
		t.Errorf("RepoCloneProgress proto roundtrip failed (-want +got):\n%s", diff)
	}
}

func TestClient_P4ExecRequest_ProtoRoundTrip(t *testing.T) {
	var diff string

	fn := func(original protocol.P4ExecRequest) bool {
		var converted protocol.P4ExecRequest
		converted.FromProto(original.ToProto())

		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(fn, nil); err != nil {
		t.Errorf("P4ExecRequest proto roundtrip failed (-want +got):\n%s", diff)
	}
}

func TestClient_RepoClone_ProtoRoundTrip(t *testing.T) {
	var diff string

	fn := func(original protocol.RepoCloneResponse) bool {
		var converted protocol.RepoCloneResponse
		converted.FromProto(original.ToProto())

		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(fn, nil); err != nil {
		t.Errorf("RepoCloneResponse proto roundtrip failed (-want +got):\n%s", diff)
	}
}

func TestClient_ListGitolite_ProtoRoundTrip(t *testing.T) {
	var diff string

	fn := func(original gitolite.Repo) bool {
		var converted gitolite.Repo
		converted.FromProto(original.ToProto())

		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(fn, nil); err != nil {
		t.Errorf("ListGitoliteRepo proto roundtrip failed (-want +got):\n%s", diff)
	}
}

func TestClient_Remove(t *testing.T) {
	test := func(t *testing.T, called *bool) {
		db := database.NewMockDB()
		repo := api.RepoName("github.com/sourcegraph/sourcegraph")
		addrs := []string{"172.16.8.1:8080", "172.16.8.2:8080"}

		expected := "http://172.16.8.1:8080"

		source := gitserver.NewTestClientSource(t, db, addrs, func(o *gitserver.TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				mockRepoDelete := func(ctx context.Context, in *proto.RepoDeleteRequest, opts ...grpc.CallOption) (*proto.RepoDeleteResponse, error) {
					*called = true
					return nil, nil
				}
				return &mockClient{
					mockRepoDelete: mockRepoDelete,
				}
			}
		})
		cli := gitserver.NewTestClient(
			httpcli.DoerFunc(func(r *http.Request) (*http.Response, error) {
				switch r.URL.String() {
				// Ensure that the request was received by the "expected" gitserver instance - where
				// expected is the gitserver instance according to the Rendezvous hashing scheme.
				// For anything else apart from this we return an error.
				case expected + "/delete":
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewBufferString("{}")),
					}, nil
				default:
					return nil, errors.Newf("unexpected URL: %q", r.URL.String())
				}
			}),

			source,
		)

		err := cli.Remove(context.Background(), repo)
		if err != nil {
			t.Fatalf("expected URL %q, but got err %q", expected, err)
		}

		err = cli.RemoveFrom(context.Background(), repo, "172.16.8.1:8080")
		if err != nil {
			t.Fatalf("expected URL %q, but got err %q", expected, err)
		}
	}

	t.Run("GRPC", func(t *testing.T) {
		called := false
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				ExperimentalFeatures: &schema.ExperimentalFeatures{
					EnableGRPC: true,
				},
			},
		})
		t.Cleanup(func() {
			conf.Mock(nil)
		})
		test(t, &called)
		if !called {
			t.Fatal("grpc client not called")
		}
	})
	t.Run("HTTP", func(t *testing.T) {
		called := false
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				ExperimentalFeatures: &schema.ExperimentalFeatures{
					EnableGRPC: false,
				},
			},
		})
		t.Cleanup(func() {
			conf.Mock(nil)
		})
		test(t, &called)
		if called {
			t.Fatal("grpc client called")
		}
	})

}

func TestClient_ArchiveReader(t *testing.T) {
	root := gitserver.CreateRepoDir(t)

	type test struct {
		name string

		remote      string
		revision    string
		want        map[string]string
		clientErr   error
		readerError error
		skipReader  bool
	}

	tests := []test{
		{
			name: "simple",

			remote:   createSimpleGitRepo(t, root),
			revision: "HEAD",
			want: map[string]string{
				"dir1/":      "",
				"dir1/file1": "infile1",
				"file 2":     "infile2",
			},
			skipReader: false,
		},
		{
			name: "repo-with-dotgit-dir",

			remote:   createRepoWithDotGitDir(t, root),
			revision: "HEAD",
			want: map[string]string{
				"file1":            "hello\n",
				".git/mydir/file2": "milton\n",
				".git/mydir/":      "",
				".git/":            "",
			},
			skipReader: false,
		},
		{
			name: "not-found",

			revision:   "HEAD",
			clientErr:  errors.New("repository does not exist: not-found"),
			skipReader: false,
		},
		{
			name: "revision-not-found",

			remote:      createRepoWithDotGitDir(t, root),
			revision:    "revision-not-found",
			clientErr:   nil,
			readerError: &gitdomain.RevisionNotFoundError{Repo: "revision-not-found", Spec: "revision-not-found"},
			skipReader:  true,
		},
	}

	runArchiveReaderTestfunc := func(t *testing.T, mkClient func(t *testing.T, addrs []string) gitserver.Client, name api.RepoName, test test) {
		t.Run(string(name), func(t *testing.T) {
			// Setup: Prepare the test Gitserver server + register the gRPC server
			s := &server.Server{
				Logger:   logtest.Scoped(t),
				ReposDir: filepath.Join(root, "repos"),
				DB:       newMockDB(),
				GetRemoteURLFunc: func(_ context.Context, name api.RepoName) (string, error) {
					if test.remote != "" {
						return test.remote, nil
					}
					return "", errors.Errorf("no remote for %s", test.name)
				},
				GetVCSSyncer: func(ctx context.Context, name api.RepoName) (server.VCSSyncer, error) {
					return server.NewGitRepoSyncer(wrexec.NewNoOpRecordingCommandFactory()), nil
				},
				RecordingCommandFactory: wrexec.NewNoOpRecordingCommandFactory(),
			}

			grpcServer := defaults.NewServer(logtest.Scoped(t))

			proto.RegisterGitserverServiceServer(grpcServer, &server.GRPCServer{Server: s})
			handler := internalgrpc.MultiplexHandlers(grpcServer, s.Handler())
			srv := httptest.NewServer(handler)
			defer srv.Close()

			u, _ := url.Parse(srv.URL)

			addrs := []string{u.Host}
			cli := mkClient(t, addrs)
			ctx := context.Background()

			if test.remote != "" {
				if _, err := cli.RequestRepoUpdate(ctx, name, 0); err != nil {
					t.Fatal(err)
				}
			}

			rc, err := cli.ArchiveReader(ctx, nil, name, gitserver.ArchiveOptions{Treeish: test.revision, Format: gitserver.ArchiveFormatZip})
			if have, want := fmt.Sprint(err), fmt.Sprint(test.clientErr); have != want {
				t.Errorf("archive: have err %v, want %v", have, want)
			}
			if rc == nil {
				return
			}
			t.Cleanup(func() {
				if err := rc.Close(); err != nil {
					t.Fatal(err)
				}
			})

			data, readErr := io.ReadAll(rc)
			if readErr != nil {
				if readErr.Error() != test.readerError.Error() {
					t.Errorf("archive: have reader err %v, want %v", readErr.Error(), test.readerError.Error())
				}

				if test.skipReader {
					return
				}

				t.Fatal(readErr)
			}

			zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
			if err != nil {
				t.Fatal(err)
			}

			got := map[string]string{}
			for _, f := range zr.File {
				r, err := f.Open()
				if err != nil {
					t.Errorf("failed to open %q because %s", f.Name, err)
					continue
				}
				contents, err := io.ReadAll(r)
				_ = r.Close()
				if err != nil {
					t.Errorf("Read(%q): %s", f.Name, err)
					continue
				}
				got[f.Name] = string(contents)
			}

			if !cmp.Equal(test.want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(test.want, got))
			}
		})
	}

	t.Run("grpc", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				ExperimentalFeatures: &schema.ExperimentalFeatures{
					EnableGRPC: true,
				},
			},
		})
		t.Cleanup(func() {
			conf.Mock(nil)
		})
		for _, test := range tests {
			db := database.NewMockDB()
			repoName := api.RepoName(test.name)
			called := false

			mkClient := func(t *testing.T, addrs []string) gitserver.Client {
				t.Helper()

				source := gitserver.NewTestClientSource(t, db, addrs, func(o *gitserver.TestClientSourceOptions) {
					o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
						base := proto.NewGitserverServiceClient(cc)

						mockArchive := func(ctx context.Context, in *proto.ArchiveRequest, opts ...grpc.CallOption) (proto.GitserverService_ArchiveClient, error) {
							called = true
							return base.Archive(ctx, in, opts...)
						}
						mockRepoUpdate := func(ctx context.Context, in *proto.RepoUpdateRequest, opts ...grpc.CallOption) (*proto.RepoUpdateResponse, error) {
							base := proto.NewGitserverServiceClient(cc)
							return base.RepoUpdate(ctx, in, opts...)
						}
						return &mockClient{
							mockArchive:    mockArchive,
							mockRepoUpdate: mockRepoUpdate,
						}
					}
				})

				return gitserver.NewTestClient(&http.Client{}, source)
			}

			runArchiveReaderTestfunc(t, mkClient, repoName, test)
			if !called {
				t.Error("archiveReader: GitserverServiceClient should have been called")
			}

		}
	})

	t.Run("http", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				ExperimentalFeatures: &schema.ExperimentalFeatures{
					EnableGRPC: false,
				},
			},
		})
		t.Cleanup(func() {
			conf.Mock(nil)
		})

		for _, test := range tests {
			db := database.NewMockDB()
			repoName := api.RepoName(test.name)
			called := false

			mkClient := func(t *testing.T, addrs []string) gitserver.Client {
				t.Helper()

				source := gitserver.NewTestClientSource(t, db, addrs, func(o *gitserver.TestClientSourceOptions) {
					o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
						mockArchive := func(ctx context.Context, in *proto.ArchiveRequest, opts ...grpc.CallOption) (proto.GitserverService_ArchiveClient, error) {
							called = true
							base := proto.NewGitserverServiceClient(cc)
							return base.Archive(ctx, in, opts...)
						}
						return &mockClient{mockArchive: mockArchive}
					}
				})

				return gitserver.NewTestClient(&http.Client{}, source)
			}

			runArchiveReaderTestfunc(t, mkClient, repoName, test)
			if called {
				t.Error("archiveReader: GitserverServiceClient should have been called")
			}

		}

	})
}

func createRepoWithDotGitDir(t *testing.T, root string) string {
	t.Helper()
	b64 := func(s string) string {
		t.Helper()
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	dir := filepath.Join(root, "remotes", "repo-with-dot-git-dir")

	// This repo was synthesized by hand to contain a file whose path is `.git/mydir/file2` (the Git
	// CLI will not let you create a file with a `.git` path component).
	//
	// The synthesized bad commit is:
	//
	// commit aa600fc517ea6546f31ae8198beb1932f13b0e4c (HEAD -> master)
	// Author: Quinn Slack <qslack@qslack.com>
	// 	Date:   Tue Jun 5 16:17:20 2018 -0700
	//
	// wip
	//
	// diff --git a/.git/mydir/file2 b/.git/mydir/file2
	// new file mode 100644
	// index 0000000..82b919c
	// --- /dev/null
	// +++ b/.git/mydir/file2
	// @@ -0,0 +1 @@
	// +milton
	files := map[string]string{
		"config": `
[core]
repositoryformatversion=0
filemode=true
`,
		"HEAD":              `ref: refs/heads/master`,
		"refs/heads/master": `aa600fc517ea6546f31ae8198beb1932f13b0e4c`,
		"objects/e7/9c5e8f964493290a409888d5413a737e8e5dd5": b64("eAFLyslPUrBgyMzLLMlMzOECACgtBOw="),
		"objects/ce/013625030ba8dba906f756967f9e9ca394464a": b64("eAFLyslPUjBjyEjNycnnAgAdxQQU"),
		"objects/82/b919c9c565d162c564286d9d6a2497931be47e": b64("eAFLyslPUjBnyM3MKcnP4wIAIw8ElA=="),
		"objects/e5/231c1d547df839dce09809e43608fe6c537682": b64("eAErKUpNVTAzYTAxAAIFvfTMEgbb8lmsKdJ+zz7ukeMOulcqZqOllmloYGBmYqKQlpmTashwjtFMlZl7xe2VbN/DptXPm7N4ipsXACOoGDo="),
		"objects/da/5ecc846359eaf23e8abe907b3125fdd7abdbc0": b64("eAErKUpNVTA2ZjA0MDAzMVFIy8xJNWJo2il58mjqxaSjKRq5c7NUpk+WflIHABZRD2I="),
		"objects/d0/01d287018593691c36042e1c8089fde7415296": b64("eAErKUpNVTA2ZjA0MDAzMVFIy8xJNWQ4x2imysy94vZKtu9h0+rnzVk8xc0LAP2TDiQ="),
		"objects/b4/009ecbf1eba01c5279f25840e2afc0d15f5005": b64("eAGdjdsJAjEQRf1OFdOAMpPN5gEitiBWEJIRBzcJu2b7N2IHfh24nMtJrRTpQA4PfWOGjEhZe4fk5zDZQGmyaDRT8ujDI7MzNOtgVdz7s21w26VWuC8xveC8vr+8/nBKrVxgyF4bJBfgiA5RjXUEO/9xVVKlS1zUB/JxNbA="),
		"objects/3d/779a05641b4ee6f1bc1e0b52de75163c2a2669": b64("eAErKUpNVTA2YjAxAAKF3MqUzCKGW3FnWpIjX32y69o3odpQ9e/11bcPAAAipRGQ"),
		"objects/aa/600fc517ea6546f31ae8198beb1932f13b0e4c": b64("eAGdjlkKAjEQBf3OKfoCSmfpLCDiFcQTZDodHHQWxwxe3xFv4FfBKx4UT8PQNzDa7doiAkLGataFXCg12lRYMEVM4qzHWMUz2eCjUXNeZGzQOdwkd1VLl1EzmZCqoehQTK6MRVMlRFJ5bbdpgcvajyNcH5nvcHy+vjz/cOBpOIEmE41D7xD2GBDVtm6BTf64qnc/qw9c4UKS"),
		"objects/e6/9de29bb2d1d6434b8b29ae775ad8c2e48c5391": b64("eAFLyslPUjBgAAAJsAHw"),
	}
	for name, data := range files {
		name = filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(name, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

func createSimpleGitRepo(t *testing.T, root string) string {
	t.Helper()
	dir := filepath.Join(root, "remotes", "simple")

	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}

	for _, cmd := range []string{
		"git init",
		"mkdir dir1",
		"echo -n infile1 > dir1/file1",
		"touch --date=2006-01-02T15:04:05Z dir1 dir1/file1 || touch -t 200601021704.05 dir1 dir1/file1",
		"git add dir1/file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_AUTHOR_DATE=2006-01-02T15:04:05Z GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"echo -n infile2 > 'file 2'",
		"touch --date=2014-05-06T19:20:21Z 'file 2' || touch -t 201405062120.21 'file 2'",
		"git add 'file 2'",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_AUTHOR_DATE=2006-01-02T15:04:05Z GIT_COMMITTER_DATE=2014-05-06T19:20:21Z git commit -m commit2 --author='a <a@a.com>' --date 2014-05-06T19:20:21Z",
		"git branch test-ref HEAD~1",
		"git branch test-nested-ref test-ref",
	} {
		c := exec.Command("bash", "-c", `GIT_CONFIG_GLOBAL="" GIT_CONFIG_SYSTEM="" `+cmd)
		c.Dir = dir
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fatalf("Command %q failed. Output was:\n\n%s", cmd, out)
		}
	}

	return dir
}

type mockP4ExecClient struct {
	isEndOfStream bool
	Err           error
	grpc.ClientStream
}

func (m *mockP4ExecClient) Recv() (*proto.P4ExecResponse, error) {
	if m.isEndOfStream {
		return nil, io.EOF
	}

	if m.Err != nil {
		s, _ := status.FromError(m.Err)
		return nil, s.Err()

	}

	response := &proto.P4ExecResponse{
		Data: []byte("example output"),
	}

	// Set the end-of-stream condition
	m.isEndOfStream = true

	return response, nil
}

func TestClient_P4ExecGRPC(t *testing.T) {
	_ = gitserver.CreateRepoDir(t)
	type test struct {
		name     string
		host     string
		user     string
		password string
		args     []string
		err      error
		wantBody string
		wantErr  string
	}
	tests := []test{
		{
			name:     "check request body",
			host:     "ssl:111.222.333.444:1666",
			user:     "admin",
			password: "pa$$word",
			args:     []string{"protects"},
			wantBody: "example output",
			wantErr:  "<nil>",
		},
		{
			name:    "error response",
			err:     errors.New("example error"),
			wantErr: "rpc error: code = Unknown desc = example error",
		},
		{
			name:    "context cancellation",
			err:     status.New(codes.Canceled, context.Canceled.Error()).Err(),
			wantErr: context.Canceled.Error(),
		},
		{
			name:    "context expiration",
			err:     status.New(codes.DeadlineExceeded, context.DeadlineExceeded.Error()).Err(),
			wantErr: context.DeadlineExceeded.Error(),
		},
	}

	ctx := context.Background()
	runTest := func(t *testing.T, test test, cli gitserver.Client, called bool) {
		t.Run(test.name, func(t *testing.T) {
			t.Log(test.name)

			rc, _, _ := cli.P4Exec(ctx, test.host, test.user, test.password, test.args...)

			var body []byte
			var err error
			if rc != nil {
				defer func() { _ = rc.Close() }()

				body, err = io.ReadAll(rc)
				if err != nil {
					if diff := cmp.Diff(test.wantErr, fmt.Sprintf("%v", err)); diff != "" {
						t.Fatalf("Mismatch (-want +got):\n%s", diff)
					}
				}
			}

			if diff := cmp.Diff(test.wantBody, string(body)); diff != "" {
				t.Fatalf("Mismatch (-want +got):\n%s", diff)
			}
		})

	}

	t.Run("GRPC", func(t *testing.T) {
		for _, test := range tests {
			db := database.NewMockDB()
			conf.Mock(&conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					ExperimentalFeatures: &schema.ExperimentalFeatures{
						EnableGRPC: true,
					},
				},
			})
			t.Cleanup(func() {
				conf.Mock(nil)
			})

			const gitserverAddr = "172.16.8.1:8080"
			addrs := []string{gitserverAddr}
			called := false

			source := gitserver.NewTestClientSource(t, db, addrs, func(o *gitserver.TestClientSourceOptions) {
				o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
					mockP4Exec := func(ctx context.Context, in *proto.P4ExecRequest, opts ...grpc.CallOption) (proto.GitserverService_P4ExecClient, error) {
						called = true
						return &mockP4ExecClient{
							Err: test.err,
						}, nil
					}

					return &mockClient{mockP4Exec: mockP4Exec}
				}
			})

			cli := gitserver.NewTestClient(&http.Client{}, source)
			runTest(t, test, cli, called)

			if !called {
				t.Fatal("GRPC should be called")
			}
		}

	})

}

func TestClient_P4Exec(t *testing.T) {
	_ = gitserver.CreateRepoDir(t)
	type test struct {
		name     string
		host     string
		user     string
		password string
		args     []string
		handler  http.HandlerFunc
		wantBody string
		wantErr  string
	}
	tests := []test{
		{
			name:     "check request body",
			host:     "ssl:111.222.333.444:1666",
			user:     "admin",
			password: "pa$$word",
			args:     []string{"protects"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.ProtoMajor == 2 {
					// Ignore attempted gRPC connections
					w.WriteHeader(http.StatusNotImplemented)
					return
				}

				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}

				wantBody := `{"p4port":"ssl:111.222.333.444:1666","p4user":"admin","p4passwd":"pa$$word","args":["protects"]}`
				if diff := cmp.Diff(wantBody, string(body)); diff != "" {
					t.Fatalf("Mismatch (-want +got):\n%s", diff)
				}

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("example output"))
			},
			wantBody: "example output",
			wantErr:  "<nil>",
		},
		{
			name: "error response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.ProtoMajor == 2 {
					// Ignore attempted gRPC connections
					w.WriteHeader(http.StatusNotImplemented)
					return
				}

				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("example error"))
			},
			wantErr: "unexpected status code: 400 - example error",
		},
	}

	ctx := context.Background()
	runTest := func(t *testing.T, test test, cli gitserver.Client, called bool) {
		t.Run(test.name, func(t *testing.T) {
			t.Log(test.name)

			rc, _, err := cli.P4Exec(ctx, test.host, test.user, test.password, test.args...)
			if diff := cmp.Diff(test.wantErr, fmt.Sprintf("%v", err)); diff != "" {
				t.Fatalf("Mismatch (-want +got):\n%s", diff)
			}

			var body []byte
			if rc != nil {
				defer func() { _ = rc.Close() }()

				body, err = io.ReadAll(rc)
				if err != nil {
					t.Fatal(err)
				}
			}

			if diff := cmp.Diff(test.wantBody, string(body)); diff != "" {
				t.Fatalf("Mismatch (-want +got):\n%s", diff)
			}
		})

	}
	t.Run("HTTP", func(t *testing.T) {
		db := database.NewMockDB()
		for _, test := range tests {
			conf.Mock(&conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					ExperimentalFeatures: &schema.ExperimentalFeatures{
						EnableGRPC: false,
					},
				},
			})
			t.Cleanup(func() {
				conf.Mock(nil)
			})

			testServer := httptest.NewServer(test.handler)
			defer testServer.Close()

			u, _ := url.Parse(testServer.URL)
			addrs := []string{u.Host}
			source := gitserver.NewTestClientSource(t, db, addrs)
			called := false

			cli := gitserver.NewTestClient(&http.Client{}, source)
			runTest(t, test, cli, called)

			if called {
				t.Fatal("handler shoulde be called")
			}
		}

	})
}

func TestClient_ResolveRevisions(t *testing.T) {
	root := t.TempDir()
	remote := createSimpleGitRepo(t, root)
	// These hashes should be stable since we set the timestamps
	// when creating the commits.
	hash1 := "b6602ca96bdc0ab647278577a3c6edcb8fe18fb0"
	hash2 := "c5151eceb40d5e625716589b745248e1a6c6228d"

	tests := []struct {
		input []protocol.RevisionSpecifier
		want  []string
		err   error
	}{{
		input: []protocol.RevisionSpecifier{{}},
		want:  []string{hash2},
	}, {
		input: []protocol.RevisionSpecifier{{RevSpec: "HEAD"}},
		want:  []string{hash2},
	}, {
		input: []protocol.RevisionSpecifier{{RevSpec: "HEAD~1"}},
		want:  []string{hash1},
	}, {
		input: []protocol.RevisionSpecifier{{RevSpec: "test-ref"}},
		want:  []string{hash1},
	}, {
		input: []protocol.RevisionSpecifier{{RevSpec: "test-nested-ref"}},
		want:  []string{hash1},
	}, {
		input: []protocol.RevisionSpecifier{{RefGlob: "refs/heads/test-*"}},
		want:  []string{hash1, hash1}, // two hashes because to refs point to that hash
	}, {
		input: []protocol.RevisionSpecifier{{RevSpec: "test-fake-ref"}},
		err:   &gitdomain.RevisionNotFoundError{Repo: api.RepoName(remote), Spec: "test-fake-ref"},
	}}

	logger := logtest.Scoped(t)
	db := newMockDB()
	ctx := context.Background()

	s := server.Server{
		Logger:   logger,
		ReposDir: filepath.Join(root, "repos"),
		GetRemoteURLFunc: func(_ context.Context, name api.RepoName) (string, error) {
			return remote, nil
		},
		GetVCSSyncer: func(ctx context.Context, name api.RepoName) (server.VCSSyncer, error) {
			return server.NewGitRepoSyncer(wrexec.NewNoOpRecordingCommandFactory()), nil
		},
		DB:                      db,
		Perforce:                perforce.NewService(ctx, observation.TestContextTB(t), logger, db, list.New()),
		RecordingCommandFactory: wrexec.NewNoOpRecordingCommandFactory(),
	}

	grpcServer := defaults.NewServer(logtest.Scoped(t))
	proto.RegisterGitserverServiceServer(grpcServer, &server.GRPCServer{Server: &s})

	handler := internalgrpc.MultiplexHandlers(grpcServer, s.Handler())
	srv := httptest.NewServer(handler)

	defer srv.Close()

	u, _ := url.Parse(srv.URL)
	addrs := []string{u.Host}
	source := gitserver.NewTestClientSource(t, db, addrs)

	cli := gitserver.NewTestClient(&http.Client{}, source)

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			_, err := cli.RequestRepoUpdate(ctx, api.RepoName(remote), 0)
			require.NoError(t, err)

			got, err := cli.ResolveRevisions(ctx, api.RepoName(remote), test.input)
			if test.err != nil {
				require.Equal(t, test.err, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}

}

func TestClient_BatchLogGRPC(t *testing.T) {
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				EnableGRPC: true,
			},
		},
	})
	t.Cleanup(func() {
		conf.Mock(nil)
	})

	addrs := []string{"172.16.8.1:8080"}

	called := false

	db := database.NewMockDB()
	source := gitserver.NewTestClientSource(t, db, addrs, func(o *gitserver.TestClientSourceOptions) {
		o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
			mockBatchLog := func(ctx context.Context, in *proto.BatchLogRequest, opts ...grpc.CallOption) (*proto.BatchLogResponse, error) {
				called = true

				var req protocol.BatchLogRequest
				req.FromProto(in)

				var results []protocol.BatchLogResult
				for _, repoCommit := range req.RepoCommits {
					results = append(results, protocol.BatchLogResult{
						RepoCommit:    repoCommit,
						CommandOutput: fmt.Sprintf("out<%s: %s@%s>", addrs[0], repoCommit.Repo, repoCommit.CommitID),
						CommandError:  "",
					})

				}

				var resp protocol.BatchLogResponse
				resp.Results = results
				return resp.ToProto(), nil
			}

			return &mockClient{mockBatchLog: mockBatchLog}
		}
	})

	cli := gitserver.NewTestClient(&http.Client{}, source)

	opts := gitserver.BatchLogOptions{
		RepoCommits: []api.RepoCommit{
			{Repo: api.RepoName("github.com/test/foo"), CommitID: api.CommitID("deadbeef01")},
			{Repo: api.RepoName("github.com/test/bar"), CommitID: api.CommitID("deadbeef02")},
			{Repo: api.RepoName("github.com/test/baz"), CommitID: api.CommitID("deadbeef03")},
			{Repo: api.RepoName("github.com/test/bonk"), CommitID: api.CommitID("deadbeef04")},
			{Repo: api.RepoName("github.com/test/quux"), CommitID: api.CommitID("deadbeef05")},
			{Repo: api.RepoName("github.com/test/honk"), CommitID: api.CommitID("deadbeef06")},
			{Repo: api.RepoName("github.com/test/xyzzy"), CommitID: api.CommitID("deadbeef07")},
			{Repo: api.RepoName("github.com/test/lorem"), CommitID: api.CommitID("deadbeef08")},
			{Repo: api.RepoName("github.com/test/ipsum"), CommitID: api.CommitID("deadbeef09")},
			{Repo: api.RepoName("github.com/test/fnord"), CommitID: api.CommitID("deadbeef10")},
		},
		Format: "--format=test",
	}

	results := map[api.RepoCommit]gitserver.RawBatchLogResult{}
	var mu sync.Mutex

	if err := cli.BatchLog(context.Background(), opts, func(repoCommit api.RepoCommit, gitLogResult gitserver.RawBatchLogResult) error {
		mu.Lock()
		defer mu.Unlock()

		results[repoCommit] = gitLogResult
		return nil
	}); err != nil {
		t.Fatalf("unexpected error performing batch log: %s", err)
	}

	expectedResults := map[api.RepoCommit]gitserver.RawBatchLogResult{
		// Shard 1
		{Repo: "github.com/test/baz", CommitID: "deadbeef03"}:  {Stdout: "out<172.16.8.1:8080: github.com/test/baz@deadbeef03>"},
		{Repo: "github.com/test/quux", CommitID: "deadbeef05"}: {Stdout: "out<172.16.8.1:8080: github.com/test/quux@deadbeef05>"},
		{Repo: "github.com/test/honk", CommitID: "deadbeef06"}: {Stdout: "out<172.16.8.1:8080: github.com/test/honk@deadbeef06>"},

		// Shard 2
		{Repo: "github.com/test/bar", CommitID: "deadbeef02"}:   {Stdout: "out<172.16.8.1:8080: github.com/test/bar@deadbeef02>"},
		{Repo: "github.com/test/xyzzy", CommitID: "deadbeef07"}: {Stdout: "out<172.16.8.1:8080: github.com/test/xyzzy@deadbeef07>"},

		// Shard 3
		{Repo: "github.com/test/foo", CommitID: "deadbeef01"}:   {Stdout: "out<172.16.8.1:8080: github.com/test/foo@deadbeef01>"},
		{Repo: "github.com/test/bonk", CommitID: "deadbeef04"}:  {Stdout: "out<172.16.8.1:8080: github.com/test/bonk@deadbeef04>"},
		{Repo: "github.com/test/lorem", CommitID: "deadbeef08"}: {Stdout: "out<172.16.8.1:8080: github.com/test/lorem@deadbeef08>"},
		{Repo: "github.com/test/ipsum", CommitID: "deadbeef09"}: {Stdout: "out<172.16.8.1:8080: github.com/test/ipsum@deadbeef09>"},
		{Repo: "github.com/test/fnord", CommitID: "deadbeef10"}: {Stdout: "out<172.16.8.1:8080: github.com/test/fnord@deadbeef10>"},
	}
	if diff := cmp.Diff(expectedResults, results); diff != "" {
		t.Errorf("unexpected results (-want +got):\n%s", diff)
	}

	if !called {
		t.Error("expected mockBatchLog to be called")
	}
}

func TestClient_BatchLog(t *testing.T) {
	db := database.NewMockDB()
	addrs := []string{"172.16.8.1:8080", "172.16.8.2:8080", "172.16.8.3:8080"}
	source := gitserver.NewTestClientSource(t, db, addrs)

	cli := gitserver.NewTestClient(
		httpcli.DoerFunc(func(r *http.Request) (*http.Response, error) {
			var req protocol.BatchLogRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				return nil, err
			}

			var results []protocol.BatchLogResult
			for _, repoCommit := range req.RepoCommits {
				results = append(results, protocol.BatchLogResult{
					RepoCommit:    repoCommit,
					CommandOutput: fmt.Sprintf("out<%s: %s@%s>", r.URL.String(), repoCommit.Repo, repoCommit.CommitID),
					CommandError:  "",
				})
			}

			encoded, _ := json.Marshal(protocol.BatchLogResponse{Results: results})
			body := io.NopCloser(strings.NewReader(strings.TrimSpace(string(encoded))))
			return &http.Response{StatusCode: 200, Body: body}, nil
		}),
		source,
	)

	opts := gitserver.BatchLogOptions{
		RepoCommits: []api.RepoCommit{
			{Repo: api.RepoName("github.com/test/foo"), CommitID: api.CommitID("deadbeef01")},
			{Repo: api.RepoName("github.com/test/bar"), CommitID: api.CommitID("deadbeef02")},
			{Repo: api.RepoName("github.com/test/baz"), CommitID: api.CommitID("deadbeef03")},
			{Repo: api.RepoName("github.com/test/bonk"), CommitID: api.CommitID("deadbeef04")},
			{Repo: api.RepoName("github.com/test/quux"), CommitID: api.CommitID("deadbeef05")},
			{Repo: api.RepoName("github.com/test/honk"), CommitID: api.CommitID("deadbeef06")},
			{Repo: api.RepoName("github.com/test/xyzzy"), CommitID: api.CommitID("deadbeef07")},
			{Repo: api.RepoName("github.com/test/lorem"), CommitID: api.CommitID("deadbeef08")},
			{Repo: api.RepoName("github.com/test/ipsum"), CommitID: api.CommitID("deadbeef09")},
			{Repo: api.RepoName("github.com/test/fnord"), CommitID: api.CommitID("deadbeef10")},
		},
		Format: "--format=test",
	}

	results := map[api.RepoCommit]gitserver.RawBatchLogResult{}
	var mu sync.Mutex

	if err := cli.BatchLog(context.Background(), opts, func(repoCommit api.RepoCommit, gitLogResult gitserver.RawBatchLogResult) error {
		mu.Lock()
		defer mu.Unlock()

		results[repoCommit] = gitLogResult
		return nil
	}); err != nil {
		t.Fatalf("unexpected error performing batch log: %s", err)
	}

	expectedResults := map[api.RepoCommit]gitserver.RawBatchLogResult{
		// Shard 1
		{Repo: "github.com/test/baz", CommitID: "deadbeef03"}:  {Stdout: "out<http://172.16.8.1:8080/batch-log: github.com/test/baz@deadbeef03>"},
		{Repo: "github.com/test/quux", CommitID: "deadbeef05"}: {Stdout: "out<http://172.16.8.1:8080/batch-log: github.com/test/quux@deadbeef05>"},
		{Repo: "github.com/test/honk", CommitID: "deadbeef06"}: {Stdout: "out<http://172.16.8.1:8080/batch-log: github.com/test/honk@deadbeef06>"},

		// Shard 2
		{Repo: "github.com/test/bar", CommitID: "deadbeef02"}:   {Stdout: "out<http://172.16.8.2:8080/batch-log: github.com/test/bar@deadbeef02>"},
		{Repo: "github.com/test/xyzzy", CommitID: "deadbeef07"}: {Stdout: "out<http://172.16.8.2:8080/batch-log: github.com/test/xyzzy@deadbeef07>"},

		// Shard 3
		{Repo: "github.com/test/foo", CommitID: "deadbeef01"}:   {Stdout: "out<http://172.16.8.3:8080/batch-log: github.com/test/foo@deadbeef01>"},
		{Repo: "github.com/test/bonk", CommitID: "deadbeef04"}:  {Stdout: "out<http://172.16.8.3:8080/batch-log: github.com/test/bonk@deadbeef04>"},
		{Repo: "github.com/test/lorem", CommitID: "deadbeef08"}: {Stdout: "out<http://172.16.8.3:8080/batch-log: github.com/test/lorem@deadbeef08>"},
		{Repo: "github.com/test/ipsum", CommitID: "deadbeef09"}: {Stdout: "out<http://172.16.8.3:8080/batch-log: github.com/test/ipsum@deadbeef09>"},
		{Repo: "github.com/test/fnord", CommitID: "deadbeef10"}: {Stdout: "out<http://172.16.8.3:8080/batch-log: github.com/test/fnord@deadbeef10>"},
	}
	if diff := cmp.Diff(expectedResults, results); diff != "" {
		t.Errorf("unexpected results (-want +got):\n%s", diff)
	}
}

func TestLocalGitCommand(t *testing.T) {
	// creating a repo with 1 committed file
	root := gitserver.CreateRepoDir(t)

	for _, cmd := range []string{
		"git init",
		"echo -n infile1 > file1",
		"touch --date=2006-01-02T15:04:05Z file1 || touch -t 200601021704.05 file1",
		"git add file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_AUTHOR_DATE=2006-01-02T15:04:05Z GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	} {
		c := exec.Command("bash", "-c", `GIT_CONFIG_GLOBAL="" GIT_CONFIG_SYSTEM="" `+cmd)
		c.Dir = root
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fatalf("Command %q failed. Output was:\n\n%s", cmd, out)
		}
	}

	ctx := context.Background()
	command := gitserver.NewLocalGitCommand(api.RepoName(filepath.Base(root)), "log")
	command.ReposDir = filepath.Dir(root)

	stdout, stderr, err := command.DividedOutput(ctx)
	if err != nil {
		t.Fatalf("Local git command run failed. Command: %q Error:\n\n%s", command, err)
	}
	if len(stderr) > 0 {
		t.Fatalf("Local git command run failed. Command: %q Error:\n\n%s", command, stderr)
	}

	stringOutput := string(stdout)
	if !strings.Contains(stringOutput, "commit1") {
		t.Fatalf("No commit message in git log output. Output: %s", stringOutput)
	}
	if command.ExitStatus() != 0 {
		t.Fatalf("Local git command finished with non-zero status. Status: %d", command.ExitStatus())
	}
}

func TestClient_ReposStats(t *testing.T) {
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				EnableGRPC: false,
			},
		},
	})
	t.Cleanup(func() {
		conf.Mock(nil)
	})

	const gitserverAddr = "172.16.8.1:8080"
	now := time.Now().UTC()
	addrs := []string{gitserverAddr}

	expected := fmt.Sprintf("http://%s", gitserverAddr)
	wantStats := protocol.ReposStats{
		UpdatedAt:   now,
		GitDirBytes: 1337,
	}

	db := database.NewMockDB()
	source := gitserver.NewTestClientSource(t, db, addrs)
	cli := gitserver.NewTestClient(
		httpcli.DoerFunc(func(r *http.Request) (*http.Response, error) {
			switch r.URL.String() {
			case expected + "/repos-stats":
				encoded, _ := json.Marshal(wantStats)
				body := io.NopCloser(strings.NewReader(strings.TrimSpace(string(encoded))))
				return &http.Response{
					StatusCode: 200,
					Body:       body,
				}, nil
			default:
				return nil, errors.Newf("unexpected URL: %q", r.URL.String())
			}
		}),
		source,
	)

	gotStatsMap, err := cli.ReposStats(context.Background())
	if err != nil {
		t.Fatalf("expected URL %q, but got err %q", expected, err)
	}

	assert.Equal(t, wantStats, *gotStatsMap[gitserverAddr])
}
func TestClient_ReposStatsGRPC(t *testing.T) {
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				EnableGRPC: true,
			},
		},
	})
	t.Cleanup(func() {
		conf.Mock(nil)
	})

	const gitserverAddr = "172.16.8.1:8080"
	now := time.Now().UTC()
	wantStats := protocol.ReposStats{
		UpdatedAt:   now,
		GitDirBytes: 1337,
	}
	called := false
	db := database.NewMockDB()
	source := gitserver.NewTestClientSource(t, db, []string{gitserverAddr}, func(o *gitserver.TestClientSourceOptions) {
		o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
			mockRepoStats := func(ctx context.Context, in *proto.ReposStatsRequest, opts ...grpc.CallOption) (*proto.ReposStatsResponse, error) {
				called = true
				return wantStats.ToProto(), nil
			}
			return &mockClient{mockRepoStats: mockRepoStats}
		}
	})

	cli := gitserver.NewTestClient(http.DefaultClient, source)

	gotStatsMap, err := cli.ReposStats(context.Background())
	if err != nil {
		t.Fatalf("expected URL %q, but got err %q", wantStats, err)
	}

	if !called {
		t.Fatal("ReposStats: grpc client not called")
	}

	assert.Equal(t, wantStats, *gotStatsMap[gitserverAddr])
}

func TestClient_IsRepoCloneableGRPC(t *testing.T) {

	type test struct {
		name          string
		repo          api.RepoName
		mockResponse  *protocol.IsRepoCloneableResponse
		wantErr       bool
		wantErrString string
	}

	const gitserverAddr = "172.16.8.1:8080"
	testCases := []test{
		{
			name: "cloneable",
			repo: "github.com/sourcegraph/sourcegraph",
			mockResponse: &protocol.IsRepoCloneableResponse{
				Cloneable: true,
			},
		},
		{
			name: "not found",
			repo: "github.com/nonexistent/repo",
			mockResponse: &protocol.IsRepoCloneableResponse{
				Cloneable: false,
				Reason:    "repository not found",
			},
			wantErr:       true,
			wantErrString: "unable to clone repo (name=\"github.com/nonexistent/repo\" notfound=true) because repository not found",
		},
		{
			name: "other error",
			repo: "github.com/sourcegraph/sourcegraph",
			mockResponse: &protocol.IsRepoCloneableResponse{
				Cloneable: false,
				Reason:    "some other error",
			},
			wantErr:       true,
			wantErrString: "unable to clone repo (name=\"github.com/sourcegraph/sourcegraph\" notfound=false) because some other error",
		},
	}
	runTests := func(t *testing.T, client gitserver.Client, tc test, called bool) {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			err := client.IsRepoCloneable(ctx, tc.repo)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				if err.Error() != tc.wantErrString {
					t.Errorf("got error %q, want %q", err.Error(), tc.wantErrString)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})

	}

	t.Run("GRPC", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				ExperimentalFeatures: &schema.ExperimentalFeatures{
					EnableGRPC: true,
				},
			},
		})
		t.Cleanup(func() {
			conf.Mock(nil)
		})

		for _, tc := range testCases {

			called := false
			db := database.NewMockDB()
			source := gitserver.NewTestClientSource(t, db, []string{gitserverAddr}, func(o *gitserver.TestClientSourceOptions) {
				o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
					mockIsRepoCloneable := func(ctx context.Context, in *proto.IsRepoCloneableRequest, opts ...grpc.CallOption) (*proto.IsRepoCloneableResponse, error) {
						called = true
						if api.RepoName(in.Repo) != tc.repo {
							t.Errorf("got %q, want %q", in.Repo, tc.repo)
						}
						return tc.mockResponse.ToProto(), nil
					}
					return &mockClient{mockIsRepoCloneable: mockIsRepoCloneable}
				}
			})

			client := gitserver.NewTestClient(http.DefaultClient, source)

			runTests(t, client, tc, called)
			if !called {
				t.Fatal("IsRepoCloneable: grpc client not called")
			}
		}
	})

	t.Run("HTTP", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				ExperimentalFeatures: &schema.ExperimentalFeatures{
					EnableGRPC: false,
				},
			},
		})
		t.Cleanup(func() {
			conf.Mock(nil)
		})
		expected := fmt.Sprintf("http://%s", gitserverAddr)

		for _, tc := range testCases {

			called := false
			db := database.NewMockDB()
			source := gitserver.NewTestClientSource(t, db, []string{gitserverAddr}, func(o *gitserver.TestClientSourceOptions) {
				o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
					mockIsRepoCloneable := func(ctx context.Context, in *proto.IsRepoCloneableRequest, opts ...grpc.CallOption) (*proto.IsRepoCloneableResponse, error) {
						called = true
						if api.RepoName(in.Repo) != tc.repo {
							t.Errorf("got %q, want %q", in.Repo, tc.repo)
						}
						return tc.mockResponse.ToProto(), nil
					}
					return &mockClient{mockIsRepoCloneable: mockIsRepoCloneable}
				}
			})

			client := gitserver.NewTestClient(
				httpcli.DoerFunc(func(r *http.Request) (*http.Response, error) {
					switch r.URL.String() {
					case expected + "/is-repo-cloneable":
						encoded, _ := json.Marshal(tc.mockResponse)
						body := io.NopCloser(strings.NewReader(strings.TrimSpace(string(encoded))))
						return &http.Response{
							StatusCode: 200,
							Body:       body,
						}, nil
					default:
						return nil, errors.Newf("unexpected URL: %q", r.URL.String())
					}
				}),
				source,
			)

			runTests(t, client, tc, called)
			if called {
				t.Fatal("IsRepoCloneable: http client should be called")
			}
		}
	})

}

type mockClient struct {
	mockBatchLog                    func(ctx context.Context, in *proto.BatchLogRequest, opts ...grpc.CallOption) (*proto.BatchLogResponse, error)
	mockCreateCommitFromPatchBinary func(ctx context.Context, in *proto.CreateCommitFromPatchBinaryRequest, opts ...grpc.CallOption) (*proto.CreateCommitFromPatchBinaryResponse, error)
	mockExec                        func(ctx context.Context, in *proto.ExecRequest, opts ...grpc.CallOption) (proto.GitserverService_ExecClient, error)
	mockGetObject                   func(ctx context.Context, in *proto.GetObjectRequest, opts ...grpc.CallOption) (*proto.GetObjectResponse, error)
	mockIsRepoCloneable             func(ctx context.Context, in *proto.IsRepoCloneableRequest, opts ...grpc.CallOption) (*proto.IsRepoCloneableResponse, error)
	mockListGitolite                func(ctx context.Context, in *proto.ListGitoliteRequest, opts ...grpc.CallOption) (*proto.ListGitoliteResponse, error)
	mockRepoClone                   func(ctx context.Context, in *proto.RepoCloneRequest, opts ...grpc.CallOption) (*proto.RepoCloneResponse, error)
	mockRepoCloneProgress           func(ctx context.Context, in *proto.RepoCloneProgressRequest, opts ...grpc.CallOption) (*proto.RepoCloneProgressResponse, error)
	mockRepoDelete                  func(ctx context.Context, in *proto.RepoDeleteRequest, opts ...grpc.CallOption) (*proto.RepoDeleteResponse, error)
	mockRepoStats                   func(ctx context.Context, in *proto.ReposStatsRequest, opts ...grpc.CallOption) (*proto.ReposStatsResponse, error)
	mockRepoUpdate                  func(ctx context.Context, in *proto.RepoUpdateRequest, opts ...grpc.CallOption) (*proto.RepoUpdateResponse, error)
	mockArchive                     func(ctx context.Context, in *proto.ArchiveRequest, opts ...grpc.CallOption) (proto.GitserverService_ArchiveClient, error)
	mockSearch                      func(ctx context.Context, in *proto.SearchRequest, opts ...grpc.CallOption) (proto.GitserverService_SearchClient, error)
	mockP4Exec                      func(ctx context.Context, in *proto.P4ExecRequest, opts ...grpc.CallOption) (proto.GitserverService_P4ExecClient, error)
}

// BatchLog implements v1.GitserverServiceClient.
func (mc *mockClient) BatchLog(ctx context.Context, in *proto.BatchLogRequest, opts ...grpc.CallOption) (*proto.BatchLogResponse, error) {
	return mc.mockBatchLog(ctx, in, opts...)
}

// GetObject implements v1.GitserverServiceClient.
func (mc *mockClient) GetObject(ctx context.Context, in *proto.GetObjectRequest, opts ...grpc.CallOption) (*proto.GetObjectResponse, error) {
	return mc.mockGetObject(ctx, in, opts...)
}

// ListGitolite implements v1.GitserverServiceClient.
func (mc *mockClient) ListGitolite(ctx context.Context, in *proto.ListGitoliteRequest, opts ...grpc.CallOption) (*proto.ListGitoliteResponse, error) {
	return mc.mockListGitolite(ctx, in, opts...)
}

// P4Exec implements v1.GitserverServiceClient.
func (mc *mockClient) P4Exec(ctx context.Context, in *proto.P4ExecRequest, opts ...grpc.CallOption) (proto.GitserverService_P4ExecClient, error) {
	return mc.mockP4Exec(ctx, in, opts...)
}

// CreateCommitFromPatchBinary implements v1.GitserverServiceClient.
func (mc *mockClient) CreateCommitFromPatchBinary(ctx context.Context, in *proto.CreateCommitFromPatchBinaryRequest, opts ...grpc.CallOption) (*proto.CreateCommitFromPatchBinaryResponse, error) {
	return mc.mockCreateCommitFromPatchBinary(ctx, in, opts...)
}

// RepoUpdate implements v1.GitserverServiceClient
func (mc *mockClient) RepoUpdate(ctx context.Context, in *proto.RepoUpdateRequest, opts ...grpc.CallOption) (*proto.RepoUpdateResponse, error) {
	return mc.mockRepoUpdate(ctx, in, opts...)
}

// RepoDelete implements v1.GitserverServiceClient
func (mc *mockClient) RepoDelete(ctx context.Context, in *proto.RepoDeleteRequest, opts ...grpc.CallOption) (*proto.RepoDeleteResponse, error) {
	return mc.mockRepoDelete(ctx, in, opts...)
}

// RepoCloneProgress implements v1.GitserverServiceClient
func (mc *mockClient) RepoCloneProgress(ctx context.Context, in *proto.RepoCloneProgressRequest, opts ...grpc.CallOption) (*proto.RepoCloneProgressResponse, error) {
	return mc.mockRepoCloneProgress(ctx, in, opts...)
}

// Exec implements v1.GitserverServiceClient
func (mc *mockClient) Exec(ctx context.Context, in *proto.ExecRequest, opts ...grpc.CallOption) (proto.GitserverService_ExecClient, error) {
	return mc.mockExec(ctx, in, opts...)
}

// RepoClone implements v1.GitserverServiceClient
func (mc *mockClient) RepoClone(ctx context.Context, in *proto.RepoCloneRequest, opts ...grpc.CallOption) (*proto.RepoCloneResponse, error) {
	return mc.mockRepoClone(ctx, in, opts...)
}

func (ms *mockClient) IsRepoCloneable(ctx context.Context, in *proto.IsRepoCloneableRequest, opts ...grpc.CallOption) (*proto.IsRepoCloneableResponse, error) {
	return ms.mockIsRepoCloneable(ctx, in, opts...)
}

// ReposStats implements v1.GitserverServiceClient
func (ms *mockClient) ReposStats(ctx context.Context, in *proto.ReposStatsRequest, opts ...grpc.CallOption) (*proto.ReposStatsResponse, error) {
	return ms.mockRepoStats(ctx, in, opts...)
}

// Search implements v1.GitserverServiceClient
func (ms *mockClient) Search(ctx context.Context, in *proto.SearchRequest, opts ...grpc.CallOption) (proto.GitserverService_SearchClient, error) {
	return ms.mockSearch(ctx, in, opts...)
}

func (mc *mockClient) Archive(ctx context.Context, in *proto.ArchiveRequest, opts ...grpc.CallOption) (proto.GitserverService_ArchiveClient, error) {
	return mc.mockArchive(ctx, in, opts...)
}

var _ proto.GitserverServiceClient = &mockClient{}

var _ proto.GitserverService_P4ExecClient = &mockP4ExecClient{}

type fuzzTime time.Time

func (fuzzTime) Generate(rand *rand.Rand, _ int) reflect.Value {
	// The maximum representable year in RFC 3339 is 9999, so we'll use that as our upper bound.
	maxDate := time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)

	ts := time.Unix(rand.Int63n(maxDate.Unix()), rand.Int63n(int64(time.Second)))
	return reflect.ValueOf(fuzzTime(ts))
}

var _ quick.Generator = fuzzTime{}
