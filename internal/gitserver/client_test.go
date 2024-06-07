package gitserver_test

import (
	"context"
	"io"
	"math/rand"
	"reflect"
	"sort"
	"sync/atomic"
	"testing"
	"testing/quick"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
)

func TestClientArchiveOptions_ProtoRoundTrip(t *testing.T) {
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

func TestClient_CreateCommitFromPatchRequest_ProtoRoundTrip(t *testing.T) {
	var diff string

	t.Run("request", func(t *testing.T) {
		fn := func(
			repo string,
			baseCommit string,
			targetRef string,
			pushRef *string,

			commitInfo struct {
				Messages    []string
				AuthorName  string
				AuthorEmail string
				Date        fuzzTime
			},

			pushConfig *protocol.PushConfig,
		) bool {
			original := protocol.CreateCommitFromPatchRequest{
				Repo:       api.RepoName(repo),
				BaseCommit: api.CommitID(baseCommit),
				TargetRef:  targetRef,
				CommitInfo: protocol.PatchCommitInfo{
					Messages:    commitInfo.Messages,
					AuthorName:  commitInfo.AuthorName,
					AuthorEmail: commitInfo.AuthorEmail,
					Date:        time.Time(commitInfo.Date),
				},
				Push:    pushConfig,
				PushRef: pushRef,
			}
			var converted protocol.CreateCommitFromPatchRequest
			converted.FromProto(original.ToMetadataProto())

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
				cli := gitserver.NewStrictMockGitserverServiceClient()
				cli.IsRepoCloneableFunc.SetDefaultHook(mockIsRepoCloneable)
				return cli
			}
		})

		client := gitserver.NewTestClient(t).WithClientSource(source)

		runTests(t, client, tc)
		if !called {
			t.Fatal("IsRepoCloneable: grpc client not called")
		}
	}
}

func TestClient_SystemsInfo(t *testing.T) {
	const (
		gitserverAddr1 = "172.16.8.1:8080"
		gitserverAddr2 = "172.16.8.2:8080"
	)

	expectedResponses := []protocol.SystemInfo{
		{
			Address:    gitserverAddr1,
			FreeSpace:  102400,
			TotalSpace: 409600,
		},
		{
			Address:    gitserverAddr2,
			FreeSpace:  51200,
			TotalSpace: 204800,
		},
	}
	sort.Slice(expectedResponses, func(i, j int) bool {
		return expectedResponses[i].Address < expectedResponses[j].Address
	})

	runTest := func(t *testing.T, client gitserver.Client) {
		ctx := context.Background()
		info, err := client.SystemsInfo(ctx)
		require.NoError(t, err, "unexpected error")

		sort.Slice(info, func(i, j int) bool {
			return info[i].Address < info[j].Address
		})

		require.Len(t, info, len(expectedResponses), "expected %d disk info(s)", len(expectedResponses))
		for i := range expectedResponses {
			require.Equal(t, info[i].Address, expectedResponses[i].Address)
			require.Equal(t, info[i].FreeSpace, expectedResponses[i].FreeSpace)
			require.Equal(t, info[i].TotalSpace, expectedResponses[i].TotalSpace)
		}
	}

	var called atomic.Bool
	source := gitserver.NewTestClientSource(t, []string{gitserverAddr1, gitserverAddr2}, func(o *gitserver.TestClientSourceOptions) {
		responseByAddress := make(map[string]*proto.DiskInfoResponse, len(expectedResponses))
		for _, response := range expectedResponses {
			responseByAddress[response.Address] = &proto.DiskInfoResponse{
				FreeSpace:   response.FreeSpace,
				TotalSpace:  response.TotalSpace,
				PercentUsed: response.PercentUsed,
			}
		}

		o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
			mockDiskInfo := func(ctx context.Context, in *proto.DiskInfoRequest, opts ...grpc.CallOption) (*proto.DiskInfoResponse, error) {
				address := cc.Target()
				response, ok := responseByAddress[address]
				if !ok {
					t.Fatalf("received unexpected address %q", address)
				}

				called.Store(true)
				return response, nil
			}
			cli := gitserver.NewStrictMockGitserverServiceClient()
			cli.DiskInfoFunc.SetDefaultHook(mockDiskInfo)
			return cli
		}
	})

	client := gitserver.NewTestClient(t).WithClientSource(source)

	runTest(t, client)
	if !called.Load() {
		t.Fatal("DiskInfo: grpc client not called")
	}
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

	called := false
	source := gitserver.NewTestClientSource(t, []string{gitserverAddr}, func(o *gitserver.TestClientSourceOptions) {
		o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
			mockDiskInfo := func(ctx context.Context, in *proto.DiskInfoRequest, opts ...grpc.CallOption) (*proto.DiskInfoResponse, error) {
				called = true
				return mockResponse, nil
			}
			cli := gitserver.NewStrictMockGitserverServiceClient()
			cli.DiskInfoFunc.SetDefaultHook(mockDiskInfo)
			return cli
		}
	})

	client := gitserver.NewTestClient(t).WithClientSource(source)

	runTest(t, client, gitserverAddr)
	if !called {
		t.Fatal("DiskInfo: grpc client not called")
	}
}

type fuzzTime time.Time

func (fuzzTime) Generate(rand *rand.Rand, _ int) reflect.Value {
	// The maximum representable year in RFC 3339 is 9999, so we'll use that as our upper bound.
	maxDate := time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)

	ts := time.Unix(rand.Int63n(maxDate.Unix()), rand.Int63n(int64(time.Second)))
	return reflect.ValueOf(fuzzTime(ts))
}

var _ quick.Generator = fuzzTime{}

func TestNewChangedFilesIteratorFromSlice(t *testing.T) {
	t.Run("IterateThroughFiles", func(t *testing.T) {
		files := []gitdomain.PathStatus{
			{Path: "file1.txt", Status: gitdomain.StatusAdded},
			{Path: "file2.txt", Status: gitdomain.StatusModified},
			{Path: "file3.txt", Status: gitdomain.StatusDeleted},
		}

		iter := gitserver.NewChangedFilesIteratorFromSlice(files)
		defer iter.Close()

		for i := 0; i < len(files); i++ {
			file, err := iter.Next()
			require.NoError(t, err)
			require.Equal(t, files[i], file)
		}

		_, err := iter.Next()
		require.Equal(t, io.EOF, err)
	})

	t.Run("EmptySlice", func(t *testing.T) {
		iter := gitserver.NewChangedFilesIteratorFromSlice(nil)
		defer iter.Close()

		_, err := iter.Next()
		require.Equal(t, io.EOF, err)
	})

	t.Run("Close", func(t *testing.T) {
		files := []gitdomain.PathStatus{
			{Path: "file1.txt", Status: gitdomain.StatusAdded},
		}

		iter := gitserver.NewChangedFilesIteratorFromSlice(files)

		// Close should be idempotent
		iter.Close()
		iter.Close()

		_, err := iter.Next()
		require.Equal(t, io.EOF, err)
	})
}
