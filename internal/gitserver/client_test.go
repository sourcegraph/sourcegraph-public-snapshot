package gitserver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"testing/quick"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

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
			converted.FromProto(original.ToMetadataProto(), original.Patch)

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
		repo := api.RepoName("github.com/sourcegraph/sourcegraph")
		addrs := []string{"172.16.8.1:8080", "172.16.8.2:8080"}

		expected := "http://172.16.8.1:8080"

		source := gitserver.NewTestClientSource(t, addrs, func(o *gitserver.TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				mockRepoDelete := func(ctx context.Context, in *proto.RepoDeleteRequest, opts ...grpc.CallOption) (*proto.RepoDeleteResponse, error) {
					*called = true
					return nil, nil
				}
				return &gitserver.MockGRPCClient{
					MockRepoDelete: mockRepoDelete,
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
	}

	t.Run("GRPC", func(t *testing.T) {
		called := false
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				ExperimentalFeatures: &schema.ExperimentalFeatures{
					EnableGRPC: boolPointer(true),
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
					EnableGRPC: boolPointer(false),
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
		name string

		host     string
		user     string
		password string
		args     []string

		mockErr error

		wantBody                    string
		wantReaderConstructionError string
		wantReaderError             string
	}
	tests := []test{
		{
			name: "check request body",

			host:     "ssl:111.222.333.444:1666",
			user:     "admin",
			password: "pa$$word",
			args:     []string{"protects"},

			wantBody:                    "example output",
			wantReaderConstructionError: "<nil>",
			wantReaderError:             "<nil>",
		},
		{
			name: "error response",

			mockErr:                     errors.New("example error"),
			wantReaderConstructionError: "<nil>",
			wantReaderError:             "rpc error: code = Unknown desc = example error",
		},
		{
			name: "context cancellation",

			mockErr:                     status.New(codes.Canceled, context.Canceled.Error()).Err(),
			wantReaderConstructionError: "<nil>",
			wantReaderError:             context.Canceled.Error(),
		},
		{
			name: "context expiration",

			mockErr:                     status.New(codes.DeadlineExceeded, context.DeadlineExceeded.Error()).Err(),
			wantReaderConstructionError: "<nil>",
			wantReaderError:             context.DeadlineExceeded.Error(),
		},
		{
			name: "invalid credentials - reported on reader instantiation",

			mockErr:                     status.New(codes.InvalidArgument, "that is totally wrong").Err(),
			wantReaderConstructionError: status.New(codes.InvalidArgument, "that is totally wrong").Err().Error(),
			wantReaderError:             status.New(codes.InvalidArgument, "that is totally wrong").Err().Error(),
		},
		{
			name: "permission denied - reported on reader instantiation",

			mockErr:                     status.New(codes.PermissionDenied, "you can't do this").Err(),
			wantReaderConstructionError: status.New(codes.PermissionDenied, "you can't do this").Err().Error(),
			wantReaderError:             status.New(codes.PermissionDenied, "you can't do this").Err().Error(),
		},
	}

	ctx := context.Background()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			conf.Mock(&conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					ExperimentalFeatures: &schema.ExperimentalFeatures{
						EnableGRPC: boolPointer(true),
					},
				},
			})
			t.Cleanup(func() {
				conf.Mock(nil)
			})

			const gitserverAddr = "172.16.8.1:8080"
			addrs := []string{gitserverAddr}
			called := false

			source := gitserver.NewTestClientSource(t, addrs, func(o *gitserver.TestClientSourceOptions) {
				o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
					mockP4Exec := func(ctx context.Context, in *proto.P4ExecRequest, opts ...grpc.CallOption) (proto.GitserverService_P4ExecClient, error) {
						called = true
						return &mockP4ExecClient{
							Err: test.mockErr,
						}, nil
					}

					return &gitserver.MockGRPCClient{MockP4Exec: mockP4Exec}
				}
			})

			cli := gitserver.NewTestClient(&http.Client{}, source)
			rc, _, err := cli.P4Exec(ctx, test.host, test.user, test.password, test.args...)
			if diff := cmp.Diff(test.wantReaderConstructionError, fmt.Sprintf("%v", err)); diff != "" {
				t.Errorf("error when creating reader mismatch (-want +got):\n%s", diff)
			}

			var body []byte
			if rc != nil {
				t.Cleanup(func() {
					_ = rc.Close()
				})

				body, err = io.ReadAll(rc)
				if err != nil {
					if diff := cmp.Diff(test.wantReaderError, fmt.Sprintf("%v", err)); diff != "" {
						t.Errorf("Mismatch (-want +got):\n%s", diff)
					}
				}
			}

			if diff := cmp.Diff(test.wantBody, string(body)); diff != "" {
				t.Fatalf("Mismatch (-want +got):\n%s", diff)
			}

			if !called {
				t.Fatal("GRPC should be called")
			}
		})
	}
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
		for _, test := range tests {
			conf.Mock(&conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					ExperimentalFeatures: &schema.ExperimentalFeatures{
						EnableGRPC: boolPointer(false),
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
			source := gitserver.NewTestClientSource(t, addrs)
			called := false

			cli := gitserver.NewTestClient(&http.Client{}, source)
			runTest(t, test, cli, called)

			if called {
				t.Fatal("handler shoulde be called")
			}
		}

	})
}

func TestClient_BatchLogGRPC(t *testing.T) {
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				EnableGRPC: boolPointer(true),
			},
		},
	})
	t.Cleanup(func() {
		conf.Mock(nil)
	})

	addrs := []string{"172.16.8.1:8080"}

	called := false

	source := gitserver.NewTestClientSource(t, addrs, func(o *gitserver.TestClientSourceOptions) {
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

			return &gitserver.MockGRPCClient{MockBatchLog: mockBatchLog}
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
	addrs := []string{"172.16.8.1:8080", "172.16.8.2:8080", "172.16.8.3:8080"}
	source := gitserver.NewTestClientSource(t, addrs, func(o *gitserver.TestClientSourceOptions) {
		o.ClientFunc = func(conn *grpc.ClientConn) proto.GitserverServiceClient {
			mockBatchLog := func(ctx context.Context, in *proto.BatchLogRequest, opts ...grpc.CallOption) (*proto.BatchLogResponse, error) {
				var out []*proto.BatchLogResult

				for _, repoCommit := range in.GetRepoCommits() {
					out = append(out, &proto.BatchLogResult{
						RepoCommit:    repoCommit,
						CommandOutput: fmt.Sprintf("out<%s: %s@%s>", fmt.Sprintf("http://%s/batch-log", conn.Target()), repoCommit.GetRepo(), repoCommit.GetCommit()),
						CommandError:  nil,
					})
				}

				return &proto.BatchLogResponse{
					Results: out,
				}, nil
			}

			return &gitserver.MockGRPCClient{
				MockBatchLog: mockBatchLog,
			}
		}
	})

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
	runTests := func(t *testing.T, client gitserver.Client, tc test) {
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
					EnableGRPC: boolPointer(true),
				},
			},
		})
		t.Cleanup(func() {
			conf.Mock(nil)
		})

		for _, tc := range testCases {

			called := false
			source := gitserver.NewTestClientSource(t, []string{gitserverAddr}, func(o *gitserver.TestClientSourceOptions) {
				o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
					mockIsRepoCloneable := func(ctx context.Context, in *proto.IsRepoCloneableRequest, opts ...grpc.CallOption) (*proto.IsRepoCloneableResponse, error) {
						called = true
						if api.RepoName(in.Repo) != tc.repo {
							t.Errorf("got %q, want %q", in.Repo, tc.repo)
						}
						return tc.mockResponse.ToProto(), nil
					}
					return &gitserver.MockGRPCClient{MockIsRepoCloneable: mockIsRepoCloneable}
				}
			})

			client := gitserver.NewTestClient(http.DefaultClient, source)

			runTests(t, client, tc)
			if !called {
				t.Fatal("IsRepoCloneable: grpc client not called")
			}
		}
	})

	t.Run("HTTP", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				ExperimentalFeatures: &schema.ExperimentalFeatures{
					EnableGRPC: boolPointer(false),
				},
			},
		})
		t.Cleanup(func() {
			conf.Mock(nil)
		})
		expected := fmt.Sprintf("http://%s", gitserverAddr)

		for _, tc := range testCases {
			called := false
			source := gitserver.NewTestClientSource(t, []string{gitserverAddr}, func(o *gitserver.TestClientSourceOptions) {
				o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
					mockIsRepoCloneable := func(ctx context.Context, in *proto.IsRepoCloneableRequest, opts ...grpc.CallOption) (*proto.IsRepoCloneableResponse, error) {
						called = true
						if api.RepoName(in.Repo) != tc.repo {
							t.Errorf("got %q, want %q", in.Repo, tc.repo)
						}
						return tc.mockResponse.ToProto(), nil
					}
					return &gitserver.MockGRPCClient{MockIsRepoCloneable: mockIsRepoCloneable}
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

			runTests(t, client, tc)
			if called {
				t.Fatal("IsRepoCloneable: http client should be called")
			}
		}
	})
}

func TestClient_SystemsInfo(t *testing.T) {
	const gitserverAddr = "172.16.8.1:8080"
	var mockResponse = &proto.DiskInfoResponse{
		FreeSpace:  102400,
		TotalSpace: 409600,
	}

	runTest := func(t *testing.T, client gitserver.Client) {
		ctx := context.Background()
		info, err := client.SystemsInfo(ctx)
		require.NoError(t, err, "unexpected error")
		require.Len(t, info, 1, "expected 1 disk info")
		require.Equal(t, gitserverAddr, info[0].Address)
		require.Equal(t, mockResponse.FreeSpace, info[0].FreeSpace)
		require.Equal(t, mockResponse.TotalSpace, info[0].TotalSpace)
	}

	t.Run("GRPC", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				ExperimentalFeatures: &schema.ExperimentalFeatures{
					EnableGRPC: boolPointer(true),
				},
			},
		})
		t.Cleanup(func() {
			conf.Mock(nil)
		})

		called := false
		source := gitserver.NewTestClientSource(t, []string{gitserverAddr}, func(o *gitserver.TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				mockDiskInfo := func(ctx context.Context, in *proto.DiskInfoRequest, opts ...grpc.CallOption) (*proto.DiskInfoResponse, error) {
					called = true
					return mockResponse, nil
				}
				return &gitserver.MockGRPCClient{MockDiskInfo: mockDiskInfo}
			}
		})

		client := gitserver.NewTestClient(http.DefaultClient, source)

		runTest(t, client)
		if !called {
			t.Fatal("DiskInfo: grpc client not called")
		}
	})

	t.Run("HTTP", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				ExperimentalFeatures: &schema.ExperimentalFeatures{
					EnableGRPC: boolPointer(false),
				},
			},
		})
		t.Cleanup(func() {
			conf.Mock(nil)
		})
		expected := fmt.Sprintf("http://%s", gitserverAddr)

		called := false
		source := gitserver.NewTestClientSource(t, []string{gitserverAddr}, func(o *gitserver.TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				mockDiskInfo := func(ctx context.Context, in *proto.DiskInfoRequest, opts ...grpc.CallOption) (*proto.DiskInfoResponse, error) {
					called = true
					return mockResponse, nil
				}
				return &gitserver.MockGRPCClient{MockDiskInfo: mockDiskInfo}
			}
		})

		client := gitserver.NewTestClient(
			httpcli.DoerFunc(func(r *http.Request) (*http.Response, error) {
				switch r.URL.String() {
				case expected + "/disk-info":
					encoded, _ := json.Marshal(mockResponse)
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

		runTest(t, client)
		if called {
			t.Fatal("DiskInfo: http client should be called")
		}
	})
}

func TestClient_SystemInfo(t *testing.T) {
	const gitserverAddr = "172.16.8.1:8080"
	var mockResponse = &proto.DiskInfoResponse{
		FreeSpace:  102400,
		TotalSpace: 409600,
	}

	runTest := func(t *testing.T, client gitserver.Client, addr string) {
		ctx := context.Background()
		info, err := client.SystemInfo(ctx, addr)
		require.NoError(t, err, "unexpected error")
		require.Equal(t, gitserverAddr, info.Address)
		require.Equal(t, mockResponse.FreeSpace, info.FreeSpace)
		require.Equal(t, mockResponse.TotalSpace, info.TotalSpace)
	}

	t.Run("GRPC", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				ExperimentalFeatures: &schema.ExperimentalFeatures{
					EnableGRPC: boolPointer(true),
				},
			},
		})
		t.Cleanup(func() {
			conf.Mock(nil)
		})

		called := false
		source := gitserver.NewTestClientSource(t, []string{gitserverAddr}, func(o *gitserver.TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				mockDiskInfo := func(ctx context.Context, in *proto.DiskInfoRequest, opts ...grpc.CallOption) (*proto.DiskInfoResponse, error) {
					called = true
					return mockResponse, nil
				}
				return &gitserver.MockGRPCClient{MockDiskInfo: mockDiskInfo}
			}
		})

		client := gitserver.NewTestClient(http.DefaultClient, source)

		runTest(t, client, gitserverAddr)
		if !called {
			t.Fatal("DiskInfo: grpc client not called")
		}
	})

	t.Run("HTTP", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				ExperimentalFeatures: &schema.ExperimentalFeatures{
					EnableGRPC: boolPointer(false),
				},
			},
		})
		t.Cleanup(func() {
			conf.Mock(nil)
		})
		expected := fmt.Sprintf("http://%s", gitserverAddr)

		called := false
		source := gitserver.NewTestClientSource(t, []string{gitserverAddr}, func(o *gitserver.TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				mockDiskInfo := func(ctx context.Context, in *proto.DiskInfoRequest, opts ...grpc.CallOption) (*proto.DiskInfoResponse, error) {
					called = true
					return mockResponse, nil
				}
				return &gitserver.MockGRPCClient{MockDiskInfo: mockDiskInfo}
			}
		})

		client := gitserver.NewTestClient(
			httpcli.DoerFunc(func(r *http.Request) (*http.Response, error) {
				switch r.URL.String() {
				case expected + "/disk-info":
					encoded, _ := json.Marshal(mockResponse)
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

		runTest(t, client, gitserverAddr)
		if called {
			t.Fatal("DiskInfo: http client should be called")
		}
	})
}

var _ proto.GitserverService_P4ExecClient = &mockP4ExecClient{}

type fuzzTime time.Time

func (fuzzTime) Generate(rand *rand.Rand, _ int) reflect.Value {
	// The maximum representable year in RFC 3339 is 9999, so we'll use that as our upper bound.
	maxDate := time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)

	ts := time.Unix(rand.Int63n(maxDate.Unix()), rand.Int63n(int64(time.Second)))
	return reflect.ValueOf(fuzzTime(ts))
}

var _ quick.Generator = fuzzTime{}

func boolPointer(b bool) *bool {
	return &b
}
