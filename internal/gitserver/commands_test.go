package gitserver

import (
	"context"
	"io"
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"

	mockrequire "github.com/derision-test/go-mockgen/v2/testutil/require"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/google/go-cmp/cmp"
	godiff "github.com/sourcegraph/go-diff/diff"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	v1 "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// Generate a random archive format.
func (f ArchiveFormat) Generate(rand *rand.Rand, _ int) reflect.Value {
	choices := []ArchiveFormat{ArchiveFormatZip, ArchiveFormatTar}
	index := rand.Intn(len(choices))

	return reflect.ValueOf(choices[index])
}

func TestDiffWithSubRepoFiltering(t *testing.T) {
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{
		UID: 1,
	})

	checker := getTestSubRepoPermsChecker("file1.1", "file2")
	testCases := []struct {
		label             string
		expectedDiffFiles []string
		expectedFileStat  *godiff.Stat
		diffFile          string
	}{
		{
			label:             "adding files",
			expectedDiffFiles: []string{"file1", "file3", "file3.3"},
			expectedFileStat:  &godiff.Stat{Added: 3},
			diffFile:          "testdata/TestDiffWithSubRepoFiltering/adding_files",
		},
		{
			label: "changing filename",
			// Additional git commands used:
			// "mv file1.1 file_can_access",
			// "git add file_can_access",
			// makeGitCommit("rename", 7),
			expectedDiffFiles: []string{"file_can_access"},
			expectedFileStat:  &godiff.Stat{Added: 1},
			diffFile:          "testdata/TestDiffWithSubRepoFiltering/changing_filename",
		},
		{
			label: "file modified",
			// Additional git commands used:
			// "echo new_file_content > file2",
			// "echo more_new_file_content > file1",
			// "git add file2",
			// "git add file1",
			// makeGitCommit("edit_files", 7),
			expectedDiffFiles: []string{"file1"}, // file2 is updated but user doesn't have access
			expectedFileStat:  &godiff.Stat{Changed: 1},
			diffFile:          "testdata/TestDiffWithSubRepoFiltering/file_modified",
		},
		{
			label: "diff for commit w/ no access returns empty result",
			// Additional git commands used:
			// "echo new_file_content > file2",
			// "git add file2",
			// makeGitCommit("no_access", 7),
			expectedDiffFiles: []string{},
			expectedFileStat:  &godiff.Stat{},
			diffFile:          "testdata/TestDiffWithSubRepoFiltering/diff_for_commit_with_no_access_returns_empty_result",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			if tc.diffFile == "" {
				t.Fatal("diffFile must be specified")
			}
			diff, err := os.ReadFile(tc.diffFile)
			require.NoError(t, err)
			c := NewTestClient(t).WithClientSource(NewTestClientSource(t, []string{"test"}, func(o *TestClientSourceOptions) {
				o.ClientFunc = func(conn *grpc.ClientConn) proto.GitserverServiceClient {
					c := NewMockGitserverServiceClient()
					d := NewMockGitserverService_RawDiffClient()
					d.RecvFunc.PushReturn(&proto.RawDiffResponse{Chunk: diff}, nil)
					d.RecvFunc.PushReturn(nil, io.EOF)
					c.RawDiffFunc.SetDefaultReturn(d, nil)
					return c
				}
			})).WithChecker(checker)

			iter, err := c.Diff(ctx, "repo", DiffOptions{Base: "base", Head: "head"})
			if err != nil {
				t.Fatalf("error fetching diff: %s", err)
			}
			defer iter.Close()

			stat := &godiff.Stat{}
			fileNames := make([]string, 0, 3)
			for {
				file, err := iter.Next()
				if err == io.EOF {
					break
				} else if err != nil {
					t.Error(err)
				}

				fileNames = append(fileNames, file.NewName)

				fileStat := file.Stat()
				stat.Added += fileStat.Added
				stat.Changed += fileStat.Changed
				stat.Deleted += fileStat.Deleted
			}
			if diff := cmp.Diff(fileNames, tc.expectedDiffFiles); diff != "" {
				t.Fatal(diff)
			}
			if diff := cmp.Diff(stat, tc.expectedFileStat); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

// get a test sub-repo permissions checker which allows access to all files (so should be a no-op)
func getTestSubRepoPermsChecker(noAccessPaths ...string) authz.SubRepoPermissionChecker {
	checker := authz.NewMockSubRepoPermissionChecker()
	checker.EnabledFunc.SetDefaultHook(func() bool {
		return true
	})
	checker.PermissionsFunc.SetDefaultHook(func(ctx context.Context, i int32, content authz.RepoContent) (authz.Perms, error) {
		for _, noAccessPath := range noAccessPaths {
			if content.Path == noAccessPath {
				return authz.None, nil
			}
		}
		return authz.Read, nil
	})
	usePermissionsForFilePermissionsFunc(checker)
	return checker
}

func usePermissionsForFilePermissionsFunc(m *authz.MockSubRepoPermissionChecker) {
	m.FilePermissionsFuncFunc.SetDefaultHook(func(ctx context.Context, userID int32, repo api.RepoName) (authz.FilePermissionFunc, error) {
		return func(path string) (authz.Perms, error) {
			return m.Permissions(ctx, userID, authz.RepoContent{Repo: repo, Path: path})
		}, nil
	})
}

func TestClient_StreamBlameFile(t *testing.T) {
	t.Run("firstChunk memoization", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				bc := NewMockGitserverService_BlameClient()
				bc.RecvFunc.PushReturn(&proto.BlameResponse{Hunk: &proto.BlameHunk{Commit: "deadbeef"}}, nil)
				bc.RecvFunc.PushReturn(&proto.BlameResponse{Hunk: &proto.BlameHunk{Commit: "deadbeef2"}}, nil)
				bc.RecvFunc.PushReturn(nil, io.EOF)
				c.BlameFunc.SetDefaultReturn(bc, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		hr, err := c.StreamBlameFile(context.Background(), "repo", "file", &BlameOptions{})
		require.NoError(t, err)

		// This chunk comes from the memoized first message.
		h, err := hr.Read()
		require.NoError(t, err)
		require.Equal(t, h.CommitID, api.CommitID("deadbeef"))

		// This chunk is returned from Recv inside the hunk reader.
		h, err = hr.Read()
		require.NoError(t, err)
		require.Equal(t, h.CommitID, api.CommitID("deadbeef2"))

		// Done.
		_, err = hr.Read()
		require.Error(t, err)
		require.Equal(t, io.EOF, err)

		require.NoError(t, hr.Close())
	})
	t.Run("checks for subrepo permissions on the path", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				bc := NewMockGitserverService_BlameClient()
				bc.RecvFunc.SetDefaultHook(func() (*proto.BlameResponse, error) {
					t.Fatal("should not be called")
					return nil, nil
				})
				c.BlameFunc.SetDefaultReturn(bc, nil)
				return c
			}
		})

		srp := authz.NewMockSubRepoPermissionChecker()
		srp.EnabledFunc.SetDefaultReturn(true)
		srp.EnabledForRepoFunc.SetDefaultReturn(true, nil)
		srp.PermissionsFunc.SetDefaultHook(func(ctx context.Context, userID int32, content authz.RepoContent) (authz.Perms, error) {
			require.Equal(t, "file", content.Path)
			return authz.None, nil
		})
		c := NewTestClient(t).WithClientSource(source).WithChecker(srp)

		ctx := actor.WithActor(context.Background(), actor.FromUser(1))
		_, err := c.StreamBlameFile(ctx, "repo", "file", &BlameOptions{})
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
	})
	t.Run("revision not found errors are returned early", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				bc := NewMockGitserverService_BlameClient()
				s, err := status.New(codes.NotFound, "commit not found").WithDetails(&proto.RevisionNotFoundPayload{Repo: "repo", Spec: "deadbeef"})
				require.NoError(t, err)
				bc.RecvFunc.PushReturn(nil, s.Err())
				c.BlameFunc.SetDefaultReturn(bc, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		_, err := c.StreamBlameFile(context.Background(), "repo", "file", &BlameOptions{})
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
	})
	t.Run("file not found errors are returned early", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				bc := NewMockGitserverService_BlameClient()
				s, err := status.New(codes.NotFound, "file not found").WithDetails(&proto.FileNotFoundPayload{Repo: "repo", Commit: "deadbeef", Path: []byte("file")})
				require.NoError(t, err)
				bc.RecvFunc.PushReturn(nil, s.Err())
				c.BlameFunc.SetDefaultReturn(bc, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		_, err := c.StreamBlameFile(context.Background(), "repo", "file", &BlameOptions{})
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
	})
	t.Run("empty blame doesn't fail", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				bc := NewMockGitserverService_BlameClient()
				bc.RecvFunc.PushReturn(nil, io.EOF)
				c.BlameFunc.SetDefaultReturn(bc, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		r, err := c.StreamBlameFile(context.Background(), "repo", "file", &BlameOptions{})
		require.NoError(t, err)
		h, err := r.Read()
		require.Equal(t, io.EOF, err)
		require.Nil(t, h)
		require.NoError(t, r.Close())
	})
}

func TestClient_GetDefaultBranch(t *testing.T) {
	t.Run("correctly returns server response", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				c.DefaultBranchFunc.SetDefaultReturn(&proto.DefaultBranchResponse{RefName: "refs/heads/master", Commit: "deadbeef"}, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		refName, sha, err := c.GetDefaultBranch(context.Background(), "repo", false)
		require.NoError(t, err)
		require.Equal(t, "refs/heads/master", refName)
		require.Equal(t, api.CommitID("deadbeef"), sha)
	})
	t.Run("returns empty for common errors", func(t *testing.T) {
		calls := 0
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				if calls == 0 {
					s, err := status.New(codes.NotFound, "bad revision").WithDetails(&proto.RevisionNotFoundPayload{Repo: "repo", Spec: "deadbeef"})
					require.NoError(t, err)
					c.DefaultBranchFunc.PushReturn(nil, s.Err())
					calls++
					return c
				}
				s, err := status.New(codes.NotFound, "repo cloning").WithDetails(&proto.RepoNotFoundPayload{Repo: "repo", CloneInProgress: true})
				require.NoError(t, err)
				c.DefaultBranchFunc.PushReturn(nil, s.Err())
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		// First request fails with clone error
		refName, sha, err := c.GetDefaultBranch(context.Background(), "repo", false)
		require.NoError(t, err)
		require.Equal(t, "", refName)
		require.Equal(t, api.CommitID(""), sha)
		// First request fails with bad rev error
		refName, sha, err = c.GetDefaultBranch(context.Background(), "repo", false)
		require.NoError(t, err)
		require.Equal(t, "", refName)
		require.Equal(t, api.CommitID(""), sha)
	})
}

func TestClient_MergeBase(t *testing.T) {
	t.Run("correctly returns server response", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				c.MergeBaseFunc.SetDefaultReturn(&proto.MergeBaseResponse{MergeBaseCommitSha: "deadbeef"}, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		sha, err := c.MergeBase(context.Background(), "repo", "master", "b2")
		require.NoError(t, err)
		require.Equal(t, api.CommitID("deadbeef"), sha)
	})
	t.Run("returns empty for empty merge base", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				c.MergeBaseFunc.SetDefaultReturn(&proto.MergeBaseResponse{MergeBaseCommitSha: ""}, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		sha, err := c.MergeBase(context.Background(), "repo", "master", "b2")
		require.NoError(t, err)
		require.Equal(t, api.CommitID(""), sha)
	})
	t.Run("revision not found", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				s, err := status.New(codes.NotFound, "bad revision").WithDetails(&proto.RevisionNotFoundPayload{Repo: "repo", Spec: "deadbeef"})
				require.NoError(t, err)
				c.MergeBaseFunc.SetDefaultReturn(nil, s.Err())
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		_, err := c.MergeBase(context.Background(), "repo", "master", "b2")
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
	})
}

func TestClient_NewFileReader(t *testing.T) {
	t.Run("firstChunk memoization", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				rfc := NewMockGitserverService_ReadFileClient()
				rfc.RecvFunc.PushReturn(&proto.ReadFileResponse{Data: []byte("part1\n")}, nil)
				rfc.RecvFunc.PushReturn(&proto.ReadFileResponse{Data: []byte("part2\n")}, nil)
				rfc.RecvFunc.PushReturn(nil, io.EOF)
				c.ReadFileFunc.SetDefaultReturn(rfc, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		r, err := c.NewFileReader(context.Background(), "repo", "deadbeef", "file")
		require.NoError(t, err)

		content, err := io.ReadAll(r)
		require.NoError(t, err)
		require.NoError(t, r.Close())
		require.Equal(t, "part1\npart2\n", string(content))
	})
	t.Run("firstChunk error memoization", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				rfc := NewMockGitserverService_ReadFileClient()
				rfc.RecvFunc.PushReturn(nil, io.EOF)
				c.ReadFileFunc.SetDefaultReturn(rfc, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		r, err := c.NewFileReader(context.Background(), "repo", "deadbeef", "file")
		require.NoError(t, err)

		content, err := io.ReadAll(r)
		require.NoError(t, err)
		require.NoError(t, r.Close())
		require.Equal(t, "", string(content))
	})
	t.Run("checks for subrepo permissions on the path", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				rc := NewMockGitserverService_ReadFileClient()
				rc.RecvFunc.SetDefaultHook(func() (*proto.ReadFileResponse, error) {
					t.Fatal("should not be called")
					return nil, nil
				})
				c.ReadFileFunc.SetDefaultReturn(rc, nil)
				return c
			}
		})

		srp := authz.NewMockSubRepoPermissionChecker()
		srp.EnabledFunc.SetDefaultReturn(true)
		srp.EnabledForRepoFunc.SetDefaultReturn(true, nil)
		srp.PermissionsFunc.SetDefaultHook(func(ctx context.Context, userID int32, content authz.RepoContent) (authz.Perms, error) {
			require.Equal(t, "file", content.Path)
			return authz.None, nil
		})
		c := NewTestClient(t).WithClientSource(source).WithChecker(srp)

		ctx := actor.WithActor(context.Background(), actor.FromUser(1))
		_, err := c.NewFileReader(ctx, "repo", "HEAD", "file")
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
	})
	t.Run("file not found errors are returned early", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				rfc := NewMockGitserverService_ReadFileClient()
				s, err := status.New(codes.NotFound, "bad actor").WithDetails(&proto.FileNotFoundPayload{})
				require.NoError(t, err)
				rfc.RecvFunc.PushReturn(nil, s.Err())
				c.ReadFileFunc.SetDefaultReturn(rfc, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		_, err := c.NewFileReader(context.Background(), "repo", "deadbeef", "file")
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
	})
	t.Run("revision not found errors are returned early", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				rfc := NewMockGitserverService_ReadFileClient()
				s, err := status.New(codes.NotFound, "revision not found").WithDetails(&proto.RevisionNotFoundPayload{})
				require.NoError(t, err)
				rfc.RecvFunc.PushReturn(nil, s.Err())
				c.ReadFileFunc.SetDefaultReturn(rfc, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		_, err := c.NewFileReader(context.Background(), "repo", "deadbeef", "file")
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
	})
	t.Run("empty file", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				rfc := NewMockGitserverService_ReadFileClient()
				rfc.RecvFunc.PushReturn(nil, io.EOF)
				c.ReadFileFunc.SetDefaultReturn(rfc, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		r, err := c.NewFileReader(context.Background(), "repo", "deadbeef", "file")
		require.NoError(t, err)
		content, err := io.ReadAll(r)
		require.NoError(t, err)
		require.Empty(t, content)
		require.NoError(t, r.Close())
	})
}

func TestClient_GetCommit(t *testing.T) {
	t.Run("correctly returns server response", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				c.GetCommitFunc.SetDefaultReturn(&proto.GetCommitResponse{Commit: &proto.GitCommit{Oid: "deadbeef"}}, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		commit, err := c.GetCommit(context.Background(), "repo", "deadbeef")
		require.NoError(t, err)
		require.Equal(t, api.CommitID("deadbeef"), commit.ID)
	})
	t.Run("checks for subrepo permissions", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				c.GetCommitFunc.SetDefaultHook(func(ctx context.Context, req *proto.GetCommitRequest, co ...grpc.CallOption) (*proto.GetCommitResponse, error) {
					require.Equal(t, true, req.GetIncludeModifiedFiles())
					// Only modified "file".
					return &proto.GetCommitResponse{Commit: &proto.GitCommit{Oid: "deadbeef"}, ModifiedFiles: [][]byte{[]byte("file")}}, nil
				})
				return c
			}
		})

		srp := authz.NewMockSubRepoPermissionChecker()
		srp.EnabledFunc.SetDefaultReturn(true)
		srp.EnabledForRepoFunc.SetDefaultReturn(true, nil)
		srp.PermissionsFunc.SetDefaultHook(func(ctx context.Context, userID int32, content authz.RepoContent) (authz.Perms, error) {
			require.Equal(t, "file", content.Path)
			return authz.None, nil
		})
		c := NewTestClient(t).WithClientSource(source).WithChecker(srp)

		ctx := actor.WithActor(context.Background(), actor.FromUser(1))
		_, err := c.GetCommit(ctx, "repo", "deadbeef")
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
	})
	t.Run("checks for subrepo permissions some files visible", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				c.GetCommitFunc.SetDefaultHook(func(ctx context.Context, req *proto.GetCommitRequest, co ...grpc.CallOption) (*proto.GetCommitResponse, error) {
					require.Equal(t, true, req.GetIncludeModifiedFiles())
					return &proto.GetCommitResponse{Commit: &proto.GitCommit{Oid: "deadbeef"}, ModifiedFiles: [][]byte{[]byte("file"), []byte("file2")}}, nil
				})
				return c
			}
		})

		srp := authz.NewMockSubRepoPermissionChecker()
		srp.EnabledFunc.SetDefaultReturn(true)
		srp.EnabledForRepoFunc.SetDefaultReturn(true, nil)
		srp.PermissionsFunc.SetDefaultHook(func(ctx context.Context, userID int32, content authz.RepoContent) (authz.Perms, error) {
			if content.Path == "file2" {
				return authz.Read, nil
			}
			require.Equal(t, "file", content.Path)
			return authz.None, nil
		})
		c := NewTestClient(t).WithClientSource(source).WithChecker(srp)

		ctx := actor.WithActor(context.Background(), actor.FromUser(1))
		_, err := c.GetCommit(ctx, "repo", "deadbeef")
		require.NoError(t, err)
	})
	t.Run("returns correct error for not found", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				s, err := status.New(codes.NotFound, "bad revision").WithDetails(&proto.RevisionNotFoundPayload{Repo: "repo", Spec: "deadbeef"})
				require.NoError(t, err)
				c.GetCommitFunc.PushReturn(nil, s.Err())
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		_, err := c.GetCommit(context.Background(), "repo", "deadbeef")
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
	})
}

func TestClient_ArchiveReader(t *testing.T) {
	t.Run("firstChunk memoization", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				rfc := NewMockGitserverService_ArchiveClient()
				rfc.RecvFunc.PushReturn(&proto.ArchiveResponse{Data: []byte("part1\n")}, nil)
				rfc.RecvFunc.PushReturn(&proto.ArchiveResponse{Data: []byte("part2\n")}, nil)
				rfc.RecvFunc.PushReturn(nil, io.EOF)
				c.ArchiveFunc.SetDefaultReturn(rfc, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		r, err := c.ArchiveReader(context.Background(), "repo", ArchiveOptions{Treeish: "deadbeef", Format: ArchiveFormatTar, Paths: []string{"file"}})
		require.NoError(t, err)

		content, err := io.ReadAll(r)
		require.NoError(t, err)
		require.NoError(t, r.Close())
		require.Equal(t, "part1\npart2\n", string(content))
	})
	t.Run("firstChunk error memoization", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				rfc := NewMockGitserverService_ArchiveClient()
				rfc.RecvFunc.PushReturn(nil, io.EOF)
				c.ArchiveFunc.SetDefaultReturn(rfc, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		r, err := c.ArchiveReader(context.Background(), "repo", ArchiveOptions{Treeish: "deadbeef", Format: ArchiveFormatTar, Paths: []string{"file"}})
		require.NoError(t, err)

		content, err := io.ReadAll(r)
		require.NoError(t, err)
		require.NoError(t, r.Close())
		require.Equal(t, "", string(content))
	})
	t.Run("revision not found errors are returned early", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				rfc := NewMockGitserverService_ArchiveClient()
				s, err := status.New(codes.NotFound, "revision not found").WithDetails(&proto.RevisionNotFoundPayload{})
				require.NoError(t, err)
				rfc.RecvFunc.PushReturn(nil, s.Err())
				c.ArchiveFunc.SetDefaultReturn(rfc, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		_, err := c.ArchiveReader(context.Background(), "repo", ArchiveOptions{Treeish: "deadbeef", Format: ArchiveFormatTar, Paths: []string{"file"}})
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
	})
	t.Run("checks for subrepo permissions enabled on the repo", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				ac := NewMockGitserverService_ArchiveClient()
				ac.RecvFunc.SetDefaultHook(func() (*proto.ArchiveResponse, error) {
					t.Fatal("should not be called")
					return nil, nil
				})
				c.ArchiveFunc.SetDefaultReturn(ac, nil)
				return c
			}
		})

		srp := authz.NewMockSubRepoPermissionChecker()
		srp.EnabledFunc.SetDefaultReturn(true)
		srp.EnabledForRepoFunc.SetDefaultReturn(true, nil)
		c := NewTestClient(t).WithClientSource(source).WithChecker(srp)

		ctx := actor.WithActor(context.Background(), actor.FromUser(1))
		_, err := c.ArchiveReader(ctx, "repo", ArchiveOptions{})
		require.Error(t, err)
	})
	t.Run("empty archive", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				rfc := NewMockGitserverService_ArchiveClient()
				rfc.RecvFunc.PushReturn(nil, io.EOF)
				c.ArchiveFunc.SetDefaultReturn(rfc, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		r, err := c.ArchiveReader(context.Background(), "repo", ArchiveOptions{Treeish: "deadbeef", Format: ArchiveFormatTar, Paths: []string{"file"}})
		require.NoError(t, err)
		content, err := io.ReadAll(r)
		require.NoError(t, err)
		require.Empty(t, content)
		require.NoError(t, r.Close())
	})
}

func TestClient_Diff(t *testing.T) {
	var testDiff = []byte(`diff --git INSTALL.md INSTALL.md
index e5af166..d44c3fc 100644
--- INSTALL.md
+++ INSTALL.md
@@ -3,10 +3,10 @@
 Line 1
 Line 2
 Line 3
-Line 4
+This is cool: Line 4
 Line 5
 Line 6
-Line 7
-Line 8
+Another Line 7
+Foobar Line 8
 Line 9
 Line 10
`)
	t.Run("firstChunk memoization", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				rfc := NewMockGitserverService_RawDiffClient()
				rfc.RecvFunc.PushReturn(&proto.RawDiffResponse{Chunk: testDiff[:len(testDiff)/2]}, nil)
				rfc.RecvFunc.PushReturn(&proto.RawDiffResponse{Chunk: testDiff[len(testDiff)/2:]}, nil)
				rfc.RecvFunc.PushReturn(nil, io.EOF)
				c.RawDiffFunc.SetDefaultReturn(rfc, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		r, err := c.Diff(context.Background(), "repo", DiffOptions{})
		require.NoError(t, err)

		fd, err := r.Next()
		require.NoError(t, err)
		require.NoError(t, r.Close())
		// Verify the parsing works.
		require.Equal(t, "INSTALL.md", fd.OrigName)
	})
	t.Run("firstChunk error memoization", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				rfc := NewMockGitserverService_RawDiffClient()
				rfc.RecvFunc.PushReturn(nil, io.EOF)
				c.RawDiffFunc.SetDefaultReturn(rfc, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		r, err := c.Diff(context.Background(), "repo", DiffOptions{})
		require.NoError(t, err)

		_, err = r.Next()
		require.True(t, errors.Is(err, io.EOF))
		require.NoError(t, r.Close())
	})
	t.Run("custom context amount", func(t *testing.T) {
		var zeroContextDiff = []byte(`diff --git f f
index 0ef51c52043997fdd257a0b77d761e9ca58bcc1f..58692a00a73d1f78df00014edf4ef39ef4ba0019 100644
--- f
+++ f
@@ -1 +1 @@
-line1
+line1.1
@@ -5 +5 @@ lin4
-line5
+line5.5
`)
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				rfc := NewMockGitserverService_RawDiffClient()
				rfc.RecvFunc.PushReturn(&v1.RawDiffResponse{
					Chunk: zeroContextDiff,
				}, nil)
				rfc.RecvFunc.PushReturn(nil, io.EOF)
				c.RawDiffFunc.SetDefaultReturn(rfc, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		r, err := c.Diff(context.Background(), "repo", DiffOptions{
			InterHunkContext: pointers.Ptr(0),
			ContextLines:     pointers.Ptr(0),
		})
		require.NoError(t, err)

		hunk, err := r.Next()
		require.NoError(t, err)
		require.Equal(t, &godiff.FileDiff{
			OrigName: "f",
			NewName:  "f",
			Extended: []string{"diff --git f f", "index 0ef51c52043997fdd257a0b77d761e9ca58bcc1f..58692a00a73d1f78df00014edf4ef39ef4ba0019 100644"},
			Hunks: []*godiff.Hunk{
				{
					OrigStartLine:   1,
					OrigLines:       1,
					OrigNoNewlineAt: 0,
					NewStartLine:    1,
					NewLines:        1,
					Section:         "",
					StartPosition:   1,
					Body:            []byte("-line1\n+line1.1\n"),
				},
				{
					OrigStartLine:   5,
					OrigLines:       1,
					OrigNoNewlineAt: 0,
					NewStartLine:    5,
					NewLines:        1,
					Section:         "lin4",
					StartPosition:   4,
					Body:            []byte("-line5\n+line5.5\n"),
				},
			},
		}, hunk)

		_, err = r.Next()
		require.True(t, errors.Is(err, io.EOF))
		require.NoError(t, r.Close())
	})
	t.Run("revision not found errors are returned early", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				rfc := NewMockGitserverService_RawDiffClient()
				s, err := status.New(codes.NotFound, "revision not found").WithDetails(&proto.RevisionNotFoundPayload{})
				require.NoError(t, err)
				rfc.RecvFunc.PushReturn(nil, s.Err())
				c.RawDiffFunc.SetDefaultReturn(rfc, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		_, err := c.Diff(context.Background(), "repo", DiffOptions{})
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
	})
}

func TestClient_ResolveRevision(t *testing.T) {
	t.Run("correctly returns server response", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				c.ResolveRevisionFunc.SetDefaultReturn(&proto.ResolveRevisionResponse{CommitSha: "deadbeef"}, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		sha, err := c.ResolveRevision(context.Background(), "repo", "HEAD", ResolveRevisionOptions{})
		require.NoError(t, err)
		require.Equal(t, api.CommitID("deadbeef"), sha)
	})
	t.Run("returns common errors correctly", func(t *testing.T) {
		calls := 0
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				if calls == 0 {
					s, err := status.New(codes.NotFound, "bad revision").WithDetails(&proto.RevisionNotFoundPayload{Repo: "repo", Spec: "deadbeef"})
					require.NoError(t, err)
					c.ResolveRevisionFunc.PushReturn(nil, s.Err())
					calls++
					return c
				}
				s, err := status.New(codes.NotFound, "repo cloning").WithDetails(&proto.RepoNotFoundPayload{Repo: "repo", CloneInProgress: true})
				require.NoError(t, err)
				c.ResolveRevisionFunc.PushReturn(nil, s.Err())
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		// First request fails with revision error
		_, err := c.ResolveRevision(context.Background(), "repo", "HEAD", ResolveRevisionOptions{})
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
		// First request fails with clone error
		_, err = c.ResolveRevision(context.Background(), "repo", "HEAD", ResolveRevisionOptions{})
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RepoNotExistError](err))
	})
}

func TestClient_RevAtTime(t *testing.T) {
	t.Run("correctly returns server response", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				c.RevAtTimeFunc.SetDefaultReturn(&proto.RevAtTimeResponse{CommitSha: "deadbeef"}, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		sha, found, err := c.RevAtTime(context.Background(), "repo", "HEAD", time.Now())
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, api.CommitID("deadbeef"), sha)
	})

	t.Run("correctly returns not found on empty sha", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				c.RevAtTimeFunc.SetDefaultReturn(&proto.RevAtTimeResponse{CommitSha: ""}, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		_, found, err := c.RevAtTime(context.Background(), "repo", "HEAD", time.Now())
		require.NoError(t, err)
		require.False(t, found)
	})

	t.Run("returns common errors correctly", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				s, err := status.New(codes.NotFound, "revision not found").WithDetails(&proto.RevisionNotFoundPayload{
					Repo: "repo",
					Spec: "HEAD",
				})
				require.NoError(t, err)
				c.RevAtTimeFunc.PushReturn(nil, s.Err())
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		_, _, err := c.RevAtTime(context.Background(), "repo", "HEAD", time.Now())
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
	})
}

func TestClient_ListRefs(t *testing.T) {
	t.Run("correctly returns server response", func(t *testing.T) {
		now := time.Now().UTC()
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				ss := NewMockGitserverService_ListRefsClient()
				ss.RecvFunc.SetDefaultReturn(nil, io.EOF)
				ss.RecvFunc.PushReturn(&proto.ListRefsResponse{Refs: []*proto.GitRef{
					{
						RefName:      []byte("refs/heads/master"),
						TargetCommit: "deadbeef",
						CreatedAt:    timestamppb.New(now),
					},
				}}, nil)
				c.ListRefsFunc.SetDefaultReturn(ss, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		refs, err := c.ListRefs(context.Background(), "repo", ListRefsOpts{})
		require.NoError(t, err)
		require.Equal(t, []gitdomain.Ref{
			{
				Name:        "refs/heads/master",
				CommitID:    "deadbeef",
				CreatedDate: now,
			},
		}, refs)
	})
	t.Run("returns well known error types", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				s, err := status.New(codes.NotFound, "repo cloning").WithDetails(&proto.RepoNotFoundPayload{Repo: "repo", CloneInProgress: true})
				require.NoError(t, err)
				c.ListRefsFunc.PushReturn(nil, s.Err())
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		// Should fail with clone error
		_, err := c.ListRefs(context.Background(), "repo", ListRefsOpts{})
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RepoNotExistError](err))
	})
}

func TestClient_ContributorCounts(t *testing.T) {
	t.Run("correctly returns server response", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				c.ContributorCountsFunc.SetDefaultReturn(&proto.ContributorCountsResponse{
					Counts: []*proto.ContributorCount{
						{
							Author: &proto.GitSignature{
								Name:  []byte("Foo"),
								Email: []byte("foo@sourcegraph.com"),
							},
							Count: 1,
						},
					},
				}, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		res, err := c.ContributorCount(context.Background(), "repo", ContributorOptions{Range: "asd", After: time.Now(), Path: "path"})
		require.NoError(t, err)
		require.Equal(t, []*gitdomain.ContributorCount{{Name: "Foo", Email: "foo@sourcegraph.com", Count: 1}}, res)
	})

	t.Run("returns common errors correctly", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				s, err := status.New(codes.NotFound, "revision not found").WithDetails(&proto.RevisionNotFoundPayload{
					Repo: "repo",
					Spec: "HEAD",
				})
				require.NoError(t, err)
				c.ContributorCountsFunc.PushReturn(nil, s.Err())
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		_, err := c.ContributorCount(context.Background(), "repo", ContributorOptions{})
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
	})
}

func TestClient_FirstEverCommit(t *testing.T) {
	t.Run("correctly returns server response", func(t *testing.T) {

		expectedCommit := &gitdomain.Commit{
			ID:        "deadbeef",
			Author:    gitdomain.Signature{Name: "Foo", Email: "foo@bar.com"},
			Committer: &gitdomain.Signature{Name: "Bar", Email: "bar@bar.com"},
			Message:   "Initial commit",
			Parents:   []api.CommitID{},
		}
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {

				c := NewMockGitserverServiceClient()
				c.FirstEverCommitFunc.SetDefaultReturn(&proto.FirstEverCommitResponse{
					Commit: expectedCommit.ToProto(),
				}, nil)

				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		actualCommit, err := c.FirstEverCommit(context.Background(), "repo")
		require.NoError(t, err)

		if diff := cmp.Diff(expectedCommit, actualCommit, cmpopts.EquateEmpty()); diff != "" {
			t.Fatalf("unexpected commit (-want +got):\n%s", diff)
		}
	})
	t.Run("returns well known error types", func(t *testing.T) {
		t.Run("repository not found", func(t *testing.T) {
			source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
				o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
					c := NewMockGitserverServiceClient()
					s, err := status.New(codes.NotFound, "repository not found").WithDetails(&proto.RepoNotFoundPayload{Repo: "repo", CloneInProgress: true})
					require.NoError(t, err)
					c.FirstEverCommitFunc.PushReturn(nil, s.Err())
					return c
				}
			})

			c := NewTestClient(t).WithClientSource(source)

			// Should fail with clone error
			_, err := c.FirstEverCommit(context.Background(), "repo")
			require.Error(t, err)
			require.True(t, errors.HasType[*gitdomain.RepoNotExistError](err))
		})

		t.Run("empty repository", func(t *testing.T) {
			source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
				o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
					c := NewMockGitserverServiceClient()
					s, err := status.New(codes.FailedPrecondition, "empty repo").WithDetails(&proto.RevisionNotFoundPayload{Repo: "repo", Spec: "HEAD"})
					require.NoError(t, err)
					c.FirstEverCommitFunc.SetDefaultReturn(nil, s.Err())
					return c
				}
			})

			c := NewTestClient(t).WithClientSource(source)

			// Should fail with RepositoryEmptyError
			_, err := c.FirstEverCommit(context.Background(), "repo")
			require.Error(t, err)
			require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
		})
	})
}

func TestClient_GetBehindAhead(t *testing.T) {
	t.Run("correctly returns server response", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				c.BehindAheadFunc.SetDefaultReturn(&proto.BehindAheadResponse{
					Behind: 5,
					Ahead:  3,
				}, nil)

				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		actualBehindAhead, err := c.BehindAhead(context.Background(), "repo", "left", "right")
		require.NoError(t, err)

		expected := &gitdomain.BehindAhead{
			Behind: 5,
			Ahead:  3,
		}

		if diff := cmp.Diff(expected, actualBehindAhead, cmpopts.EquateEmpty()); diff != "" {
			t.Fatalf("unexpected behind/ahead (-want +got):\n%s", diff)
		}
	})

	t.Run("returns well known error types", func(t *testing.T) {
		t.Run("repository not found", func(t *testing.T) {
			source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
				o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
					c := NewMockGitserverServiceClient()
					s, err := status.New(codes.NotFound, "repository not found").WithDetails(&proto.RepoNotFoundPayload{Repo: "repo", CloneInProgress: true})
					require.NoError(t, err)
					c.BehindAheadFunc.PushReturn(nil, s.Err())
					return c
				}
			})

			c := NewTestClient(t).WithClientSource(source)

			// Should fail with clone error
			_, err := c.BehindAhead(context.Background(), "repo", "left", "right")
			require.Error(t, err)
			require.True(t, errors.HasType[*gitdomain.RepoNotExistError](err))
		})

		t.Run("revision not found", func(t *testing.T) {
			source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
				o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
					c := NewMockGitserverServiceClient()
					s, err := status.New(codes.NotFound, "revision not found").WithDetails(&proto.RevisionNotFoundPayload{Repo: "repo", Spec: "right"})
					require.NoError(t, err)
					c.BehindAheadFunc.SetDefaultReturn(nil, s.Err())
					return c
				}
			})

			c := NewTestClient(t).WithClientSource(source)

			// Should fail with RevisionNotFoundError
			_, err := c.BehindAhead(context.Background(), "repo", "left", "right")
			require.Error(t, err)
			require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
		})
	})
}

func TestClient_ChangedFiles(t *testing.T) {
	t.Run("correctly returns server response", func(t *testing.T) {
		expectedChanges := []gitdomain.PathStatus{
			{Path: "file1.txt", Status: gitdomain.StatusAdded},
			{Path: "file2.txt", Status: gitdomain.StatusModified},
			{Path: "file3.txt", Status: gitdomain.StatusDeleted},
		}
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				ss := NewMockGitserverService_ChangedFilesClient()
				ss.RecvFunc.SetDefaultReturn(nil, io.EOF)
				ss.RecvFunc.PushReturn(&proto.ChangedFilesResponse{
					Files: []*proto.ChangedFile{
						{Path: []byte("file1.txt"), Status: proto.ChangedFile_STATUS_ADDED},
						{Path: []byte("file2.txt"), Status: proto.ChangedFile_STATUS_MODIFIED},
						{Path: []byte("file3.txt"), Status: proto.ChangedFile_STATUS_DELETED},
					},
				}, nil)
				c.ChangedFilesFunc.SetDefaultReturn(ss, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		changedFilesIter, err := c.ChangedFiles(context.Background(), "repo", "base", "head")
		require.NoError(t, err)

		defer changedFilesIter.Close()

		var actualChanges []gitdomain.PathStatus

		for {
			change, err := changedFilesIter.Next()
			if err == io.EOF {
				break
			}

			require.NoError(t, err)
			actualChanges = append(actualChanges, change)
		}

		require.Equal(t, expectedChanges, actualChanges)
	})

	t.Run("returns well known error types", func(t *testing.T) {
		t.Run("repository not found", func(t *testing.T) {
			source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
				o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
					c := NewMockGitserverServiceClient()
					s, err := status.New(codes.NotFound, "repository not found").WithDetails(&proto.RepoNotFoundPayload{Repo: "repo", CloneInProgress: true})
					require.NoError(t, err)
					c.ChangedFilesFunc.PushReturn(nil, s.Err())
					return c
				}
			})

			c := NewTestClient(t).WithClientSource(source)

			// Should fail with clone error
			iter, initialErr := c.ChangedFiles(context.Background(), "repo", "base", "head")

			var iterErr error
			if iter != nil {
				defer iter.Close()
				_, iterErr = iter.Next()
			}

			// Check to see if either the initial error or the error from the iterator is a RepoNotExistError
			require.True(t,
				errors.HasType[*gitdomain.RepoNotExistError](initialErr) ||
					errors.HasType[*gitdomain.RepoNotExistError](iterErr))
		})

		t.Run("revision not found", func(t *testing.T) {
			source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
				o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
					c := NewMockGitserverServiceClient()
					ss := NewMockGitserverService_ChangedFilesClient()
					s, err := status.New(codes.NotFound, "revision not found").WithDetails(&proto.RevisionNotFoundPayload{Repo: "repo", Spec: "head"})
					require.NoError(t, err)
					ss.RecvFunc.PushReturn(nil, s.Err())
					c.ChangedFilesFunc.SetDefaultReturn(ss, nil)
					return c
				}
			})

			c := NewTestClient(t).WithClientSource(source)

			// Should fail with RevisionNotFoundError
			iter, initialErr := c.ChangedFiles(context.Background(), "repo", "base", "head")

			var iterErr error
			if iter != nil {
				defer iter.Close()
				_, iterErr = iter.Next()
			}

			// Check to see if either the initial error or the error from the iterator is a RevisionNotFoundError
			require.True(t,
				errors.HasType[*gitdomain.RevisionNotFoundError](initialErr) ||
					errors.HasType[*gitdomain.RevisionNotFoundError](iterErr))

		})
	})
}

func TestChangedFilesIterator(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		fetchCallCount := 0
		fetchFunc := func() ([]gitdomain.PathStatus, error) {
			if fetchCallCount > 0 {
				return nil, io.EOF
			}
			fetchCallCount++

			return []gitdomain.PathStatus{
				{Path: "file1.txt", Status: gitdomain.StatusAdded},
				{Path: "file2.txt", Status: gitdomain.StatusModified},
			}, nil
		}
		closeFunc := func() {}

		iter := newChangedFilesIterator(fetchFunc, closeFunc)
		defer iter.Close()

		// Test fetching the first item
		item, err := iter.Next()
		require.NoError(t, err)
		require.Equal(t, gitdomain.PathStatus{Path: "file1.txt", Status: gitdomain.StatusAdded}, item)

		// Test fetching the second item
		item, err = iter.Next()
		require.NoError(t, err)
		require.Equal(t, gitdomain.PathStatus{Path: "file2.txt", Status: gitdomain.StatusModified}, item)

		// Test fetching when there are no more items
		_, err = iter.Next()
		require.Equal(t, io.EOF, err)
	})

	t.Run("fetch called multiple times", func(t *testing.T) {
		var emptySliceReturned bool

		var fetchCount int
		fetchFunc := func() ([]gitdomain.PathStatus, error) {
			fetchCount++
			switch fetchCount {
			case 1:
				return []gitdomain.PathStatus{
					{Path: "file1.txt", Status: gitdomain.StatusAdded},
					{Path: "file2.txt", Status: gitdomain.StatusModified},
				}, nil
			case 2:
				return []gitdomain.PathStatus{
					{Path: "file3.txt", Status: gitdomain.StatusAdded},
				}, nil
			case 3:
				// Ensure that fetch function is called if we returned no data
				// but we haven't signaled the end of stream with io.EOF
				emptySliceReturned = true
				return nil, nil
			default:
				return nil, io.EOF
			}
		}
		closeFunc := func() {}

		iter := newChangedFilesIterator(fetchFunc, closeFunc)
		defer iter.Close()

		// Test fetching the first item
		item, err := iter.Next()
		require.NoError(t, err)
		require.Equal(t, gitdomain.PathStatus{Path: "file1.txt", Status: gitdomain.StatusAdded}, item)
		require.Equal(t, 1, fetchCount)

		// Test fetching the second item
		item, err = iter.Next()
		require.NoError(t, err)
		require.Equal(t, gitdomain.PathStatus{Path: "file2.txt", Status: gitdomain.StatusModified}, item)
		require.Equal(t, 1, fetchCount)

		// Test fetching the third item (should trigger a new fetch)
		item, err = iter.Next()
		require.NoError(t, err)
		require.Equal(t, gitdomain.PathStatus{Path: "file3.txt", Status: gitdomain.StatusAdded}, item)
		require.Equal(t, 2, fetchCount)

		// Test fetching when there are no more items (should trigger two new fetches (since the third fetch returns an empty slice)
		_, err = iter.Next()
		require.Equal(t, io.EOF, err)
		require.True(t, emptySliceReturned)
		require.Equal(t, 4, fetchCount)
	})

	t.Run("next returns an error", func(t *testing.T) {
		fetchFunc := func() ([]gitdomain.PathStatus, error) {
			return nil, io.ErrUnexpectedEOF
		}
		closeFunc := func() {}

		iter := newChangedFilesIterator(fetchFunc, closeFunc)
		defer iter.Close()

		// Test fetching when an error occurs
		_, err := iter.Next()
		require.Equal(t, io.ErrUnexpectedEOF, err)
	})

	t.Run("close", func(t *testing.T) {
		fetchFunc := func() ([]gitdomain.PathStatus, error) {
			return nil, nil
		}

		closeCount := 0
		closeFunc := func() {
			closeCount++
		}

		iter := newChangedFilesIterator(fetchFunc, closeFunc)

		// Test closing the iterator
		iter.Close()
		require.Equal(t, closeCount, 1)

		// Test closing the iterator multiple times
		iter.Close()
		require.Equal(t, closeCount, 1)
	})
}

func TestClient_GetObject(t *testing.T) {
	t.Run("correctly returns server response", func(t *testing.T) {
		expectedID := gitdomain.OID{0xde, 0xad, 0xbe, 0xef, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				c.GetObjectFunc.SetDefaultReturn(&proto.GetObjectResponse{
					Object: &proto.GitObject{
						Id:   expectedID[:],
						Type: proto.GitObject_OBJECT_TYPE_BLOB,
					},
				}, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		object, err := c.GetObject(context.Background(), "repo", "deadbeef")
		require.NoError(t, err)
		require.Equal(t, &gitdomain.GitObject{
			ID:   expectedID,
			Type: gitdomain.ObjectTypeBlob,
		}, object)
	})

	t.Run("returns well known error types", func(t *testing.T) {
		t.Run("repository not found", func(t *testing.T) {
			source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
				o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
					c := NewMockGitserverServiceClient()
					s, err := status.New(codes.NotFound, "repository not found").WithDetails(&proto.RepoNotFoundPayload{Repo: "repo", CloneInProgress: true})
					require.NoError(t, err)
					c.GetObjectFunc.PushReturn(nil, s.Err())
					return c
				}
			})

			c := NewTestClient(t).WithClientSource(source)

			_, err := c.GetObject(context.Background(), "repo", "deadbeef")
			require.Error(t, err)
			require.True(t, errors.HasType[*gitdomain.RepoNotExistError](err))
		})

		t.Run("object not found", func(t *testing.T) {
			source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
				o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
					c := NewMockGitserverServiceClient()
					s, err := status.New(codes.NotFound, "object not found").WithDetails(&proto.RevisionNotFoundPayload{Repo: "repo", Spec: "deadbeef"})
					require.NoError(t, err)
					c.GetObjectFunc.SetDefaultReturn(nil, s.Err())
					return c
				}
			})

			c := NewTestClient(t).WithClientSource(source)

			_, err := c.GetObject(context.Background(), "repo", "deadbeef")
			require.Error(t, err)
			require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
		})
	})
}

func TestClient_Stat(t *testing.T) {
	t.Run("correctly returns server response", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				c.StatFunc.SetDefaultReturn(&proto.StatResponse{
					FileInfo: &proto.FileInfo{
						Name: []byte("file"),
						Size: 10,
						Mode: 0644,
					},
				}, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		res, err := c.Stat(context.Background(), "repo", "HEAD", "file")
		require.NoError(t, err)
		require.Equal(t, "file", res.Name())
	})

	t.Run("returns common errors correctly", func(t *testing.T) {
		t.Run("RevisionNotFound", func(t *testing.T) {
			source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
				o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
					c := NewMockGitserverServiceClient()
					s, err := status.New(codes.NotFound, "revision not found").WithDetails(&proto.RevisionNotFoundPayload{
						Repo: "repo",
						Spec: "HEAD",
					})
					require.NoError(t, err)
					c.StatFunc.PushReturn(nil, s.Err())
					return c
				}
			})

			c := NewTestClient(t).WithClientSource(source)

			_, err := c.Stat(context.Background(), "repo", "HEAD", "file")
			require.Error(t, err)
			require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
		})

		t.Run("FileNotFound", func(t *testing.T) {
			source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
				o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
					c := NewMockGitserverServiceClient()
					s, err := status.New(codes.NotFound, "file not found").WithDetails(&proto.FileNotFoundPayload{
						Repo: "repo",
						Path: []byte("file"),
					})
					require.NoError(t, err)
					c.StatFunc.PushReturn(nil, s.Err())
					return c
				}
			})

			c := NewTestClient(t).WithClientSource(source)

			_, err := c.Stat(context.Background(), "repo", "HEAD", "file")
			require.Error(t, err)
			require.True(t, os.IsNotExist(err))
		})
	})
	t.Run("subrepo permissions", func(t *testing.T) {
		ctx := actor.WithActor(context.Background(), actor.FromUser(1))

		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				c.StatFunc.SetDefaultReturn(&proto.StatResponse{
					FileInfo: &proto.FileInfo{
						Name: []byte("file"),
						Size: 10,
						Mode: 0644,
					},
				}, nil)
				return c
			}
		})

		checker := getTestSubRepoPermsChecker("file")
		c := NewTestClient(t).WithClientSource(source).WithChecker(checker)

		_, err := c.Stat(ctx, "repo", "HEAD", "file")
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
	})
}

func TestClient_ReadDir(t *testing.T) {
	t.Run("correctly returns server response", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				s := NewMockGitserverService_ReadDirClient()
				s.RecvFunc.PushReturn(&proto.ReadDirResponse{
					FileInfo: []*proto.FileInfo{
						{
							Name: []byte("file"),
							Size: 10,
							Mode: 0644,
						},
					},
				}, nil)
				s.RecvFunc.PushReturn(&proto.ReadDirResponse{
					FileInfo: []*proto.FileInfo{
						{
							Name: []byte("dir/file"),
							Size: 12,
							Mode: 0644,
						},
					},
				}, nil)
				s.RecvFunc.PushReturn(nil, io.EOF)
				c.ReadDirFunc.SetDefaultReturn(s, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		it, err := c.ReadDir(context.Background(), "repo", "HEAD", "", true)
		require.NoError(t, err)

		fd, err := it.Next()
		require.NoError(t, err)
		require.Equal(t, "file", fd.Name())

		fd, err = it.Next()
		require.NoError(t, err)
		require.Equal(t, "dir/file", fd.Name())

		_, err = it.Next()
		require.Equal(t, io.EOF, err)

		it.Close()
	})

	t.Run("correctly memoizes multiple results in one chunk", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				s := NewMockGitserverService_ReadDirClient()
				s.RecvFunc.PushReturn(&proto.ReadDirResponse{
					FileInfo: []*proto.FileInfo{
						{
							Name: []byte("file"),
							Size: 10,
							Mode: 0644,
						},
						{
							Name: []byte("dir/file"),
							Size: 12,
							Mode: 0644,
						},
					},
				}, nil)
				s.RecvFunc.PushReturn(nil, io.EOF)
				c.ReadDirFunc.SetDefaultReturn(s, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		it, err := c.ReadDir(context.Background(), "repo", "HEAD", "", true)
		require.NoError(t, err)

		fd, err := it.Next()
		require.NoError(t, err)
		require.Equal(t, "file", fd.Name())

		fd, err = it.Next()
		require.NoError(t, err)
		require.Equal(t, "dir/file", fd.Name())

		_, err = it.Next()
		require.Equal(t, io.EOF, err)

		it.Close()
	})

	t.Run("returns common errors correctly", func(t *testing.T) {
		t.Run("RevisionNotFound", func(t *testing.T) {
			source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
				o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
					c := NewMockGitserverServiceClient()
					ss := NewMockGitserverService_ReadDirClient()
					s, err := status.New(codes.NotFound, "revision not found").WithDetails(&proto.RevisionNotFoundPayload{
						Repo: "repo",
						Spec: "HEAD",
					})
					ss.RecvFunc.SetDefaultReturn(nil, s.Err())
					require.NoError(t, err)
					c.ReadDirFunc.PushReturn(ss, nil)
					return c
				}
			})

			c := NewTestClient(t).WithClientSource(source)

			_, err := c.ReadDir(context.Background(), "repo", "HEAD", "file", true)
			require.Error(t, err)
			require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
		})

		t.Run("FileNotFound", func(t *testing.T) {
			source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
				o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
					c := NewMockGitserverServiceClient()
					ss := NewMockGitserverService_ReadDirClient()
					s, err := status.New(codes.NotFound, "file not found").WithDetails(&proto.FileNotFoundPayload{
						Repo: "repo",
						Path: []byte("file"),
					})
					ss.RecvFunc.SetDefaultReturn(nil, s.Err())
					require.NoError(t, err)
					c.ReadDirFunc.PushReturn(ss, nil)
					return c
				}
			})

			c := NewTestClient(t).WithClientSource(source)

			_, err := c.ReadDir(context.Background(), "repo", "HEAD", "file", true)
			require.Error(t, err)
			require.True(t, os.IsNotExist(err))
		})
	})
	t.Run("subrepo permissions", func(t *testing.T) {
		ctx := actor.WithActor(context.Background(), actor.FromUser(1))

		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				s := NewMockGitserverService_ReadDirClient()
				s.RecvFunc.PushReturn(&proto.ReadDirResponse{
					FileInfo: []*proto.FileInfo{
						{
							Name: []byte("file"),
							Size: 10,
							Mode: 0644,
						},
					},
				}, nil)
				s.RecvFunc.PushReturn(&proto.ReadDirResponse{
					FileInfo: []*proto.FileInfo{
						{
							Name: []byte("dir/file"),
							Size: 12,
							Mode: 0644,
						},
					},
				}, nil)
				s.RecvFunc.PushReturn(nil, io.EOF)
				c.ReadDirFunc.SetDefaultReturn(s, nil)
				return c
			}
		})

		checker := getTestSubRepoPermsChecker("file")
		c := NewTestClient(t).WithClientSource(source).WithChecker(checker)

		it, err := c.ReadDir(ctx, "repo", "HEAD", "file", true)
		require.NoError(t, err)
		fd, err := it.Next()
		require.NoError(t, err)
		require.Equal(t, "dir/file", fd.Name())
		_, err = it.Next()
		require.Equal(t, io.EOF, err)

		it.Close()
	})
}

func TestClient_Commits(t *testing.T) {
	t.Run("correctly returns server response", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				s := NewMockGitserverService_CommitLogClient()
				s.RecvFunc.PushReturn(&proto.CommitLogResponse{
					Commits: []*proto.GetCommitResponse{
						{
							Commit: &v1.GitCommit{
								Oid:       "2b2289762392764ed127587b0d5fd88a2f16b7c1",
								Author:    &v1.GitSignature{Name: []byte("Bar Author"), Email: []byte("bar@sourcegraph.com"), Date: timestamppb.New(mustParseTime(time.RFC3339, "2006-01-02T15:04:06Z"))},
								Committer: &v1.GitSignature{Name: []byte("c"), Email: []byte("c@c.com"), Date: timestamppb.New(mustParseTime(time.RFC3339, "2006-01-02T15:04:07Z"))},
								Message:   []byte("bar"),
								Parents:   []string{"5fab3adc1e398e749749271d14ab843759b192cf"},
							},
							ModifiedFiles: [][]byte{[]byte("file")},
						},
					},
				}, nil)
				s.RecvFunc.PushReturn(&proto.CommitLogResponse{
					Commits: []*proto.GetCommitResponse{
						{
							Commit: &v1.GitCommit{
								Oid:       "5fab3adc1e398e749749271d14ab843759b192cf",
								Author:    &v1.GitSignature{Name: []byte("Foo Author"), Email: []byte("foo@sourcegraph.com"), Date: timestamppb.New(mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z"))},
								Committer: &v1.GitSignature{Name: []byte("c"), Email: []byte("c@c.com"), Date: timestamppb.New(mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z"))},
								Message:   []byte("foo"),
								Parents:   nil,
							},
							ModifiedFiles: [][]byte{[]byte("otherfile")},
						},
					},
				}, nil)
				s.RecvFunc.PushReturn(nil, io.EOF)
				c.CommitLogFunc.SetDefaultReturn(s, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		cs, err := c.Commits(context.Background(), "repo", CommitsOptions{AllRefs: true})
		require.NoError(t, err)

		require.Equal(t, []*gitdomain.Commit{
			{
				ID:        "2b2289762392764ed127587b0d5fd88a2f16b7c1",
				Author:    gitdomain.Signature{Name: "Bar Author", Email: "bar@sourcegraph.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
				Committer: &gitdomain.Signature{Name: "c", Email: "c@c.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
				Message:   "bar",
				Parents:   []api.CommitID{"5fab3adc1e398e749749271d14ab843759b192cf"},
			},
			{
				ID:        "5fab3adc1e398e749749271d14ab843759b192cf",
				Author:    gitdomain.Signature{Name: "Foo Author", Email: "foo@sourcegraph.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Committer: &gitdomain.Signature{Name: "c", Email: "c@c.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Message:   "foo",
				Parents:   []api.CommitID{},
			},
		}, cs)
	})

	t.Run("returns common errors correctly", func(t *testing.T) {
		t.Run("RevisionNotFound", func(t *testing.T) {
			source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
				o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
					c := NewMockGitserverServiceClient()
					ss := NewMockGitserverService_CommitLogClient()
					s, err := status.New(codes.NotFound, "revision not found").WithDetails(&proto.RevisionNotFoundPayload{
						Repo: "repo",
						Spec: "HEAD",
					})
					ss.RecvFunc.SetDefaultReturn(nil, s.Err())
					require.NoError(t, err)
					c.CommitLogFunc.PushReturn(ss, nil)
					return c
				}
			})

			c := NewTestClient(t).WithClientSource(source)

			_, err := c.Commits(context.Background(), "repo", CommitsOptions{AllRefs: true})
			require.Error(t, err)
			require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
		})
	})
	t.Run("subrepo permissions", func(t *testing.T) {
		ctx := actor.WithActor(context.Background(), actor.FromUser(1))

		allGitCommits := []*gitdomain.Commit{
			{
				ID:        "5f2834f77bc9df150aad4067819960034faedb61",
				Author:    gitdomain.Signature{Name: "Foobar Author", Email: "foobar@sourcegraph.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
				Committer: &gitdomain.Signature{Name: "c", Email: "c@c.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
				Message:   "foobar",
				Parents:   []api.CommitID{"2b2289762392764ed127587b0d5fd88a2f16b7c1"},
			},
			{
				ID:        "2b2289762392764ed127587b0d5fd88a2f16b7c1",
				Author:    gitdomain.Signature{Name: "Bar Author", Email: "bar@sourcegraph.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
				Committer: &gitdomain.Signature{Name: "c", Email: "c@c.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
				Message:   "bar",
				Parents:   []api.CommitID{"5fab3adc1e398e749749271d14ab843759b192cf"},
			},
			{
				ID:        "5fab3adc1e398e749749271d14ab843759b192cf",
				Author:    gitdomain.Signature{Name: "Foo Author", Email: "foo@sourcegraph.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Committer: &gitdomain.Signature{Name: "c", Email: "c@c.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Message:   "foo",
				Parents:   []api.CommitID{},
			},
		}

		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				s := NewMockGitserverService_CommitLogClient()
				s.RecvFunc.PushReturn(&proto.CommitLogResponse{
					Commits: []*proto.GetCommitResponse{
						{
							Commit: &v1.GitCommit{
								Oid:       "5f2834f77bc9df150aad4067819960034faedb61",
								Author:    &v1.GitSignature{Name: []byte("Foobar Author"), Email: []byte("foobar@sourcegraph.com"), Date: timestamppb.New(mustParseTime(time.RFC3339, "2006-01-02T15:04:06Z"))},
								Committer: &v1.GitSignature{Name: []byte("c"), Email: []byte("c@c.com"), Date: timestamppb.New(mustParseTime(time.RFC3339, "2006-01-02T15:04:07Z"))},
								Message:   []byte("foobar"),
								Parents:   []string{"2b2289762392764ed127587b0d5fd88a2f16b7c1"},
							},
							ModifiedFiles: [][]byte{[]byte("file3")},
						},
					},
				}, nil)
				s.RecvFunc.PushReturn(&proto.CommitLogResponse{
					Commits: []*proto.GetCommitResponse{
						{
							Commit: &v1.GitCommit{
								Oid:       "2b2289762392764ed127587b0d5fd88a2f16b7c1",
								Author:    &v1.GitSignature{Name: []byte("Bar Author"), Email: []byte("bar@sourcegraph.com"), Date: timestamppb.New(mustParseTime(time.RFC3339, "2006-01-02T15:04:06Z"))},
								Committer: &v1.GitSignature{Name: []byte("c"), Email: []byte("c@c.com"), Date: timestamppb.New(mustParseTime(time.RFC3339, "2006-01-02T15:04:07Z"))},
								Message:   []byte("bar"),
								Parents:   []string{"5fab3adc1e398e749749271d14ab843759b192cf"},
							},
							ModifiedFiles: [][]byte{[]byte("file2.2"), []byte("file2")},
						},
					},
				}, nil)
				s.RecvFunc.PushReturn(&proto.CommitLogResponse{
					Commits: []*proto.GetCommitResponse{
						{
							Commit: &v1.GitCommit{
								Oid:       "5fab3adc1e398e749749271d14ab843759b192cf",
								Author:    &v1.GitSignature{Name: []byte("Foo Author"), Email: []byte("foo@sourcegraph.com"), Date: timestamppb.New(mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z"))},
								Committer: &v1.GitSignature{Name: []byte("c"), Email: []byte("c@c.com"), Date: timestamppb.New(mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z"))},
								Message:   []byte("foo"),
								Parents:   nil,
							},
							ModifiedFiles: [][]byte{[]byte("file1")},
						},
					},
				}, nil)
				s.RecvFunc.PushReturn(nil, io.EOF)
				c.CommitLogFunc.SetDefaultHook(func(ctx context.Context, clr *v1.CommitLogRequest, co ...grpc.CallOption) (v1.GitserverService_CommitLogClient, error) {
					// With subrepo permissions on, we always need to request modified files.
					require.True(t, clr.IncludeModifiedFiles)
					return s, nil
				})
				return c
			}
		})

		t.Run("if no read perms on at least one file in the commit should filter out commit", func(t *testing.T) {
			checker := getTestSubRepoPermsChecker("file2", "file3")
			c := NewTestClient(t).WithClientSource(source).WithChecker(checker)

			cs, err := c.Commits(ctx, "repo", CommitsOptions{AllRefs: true})
			require.NoError(t, err)
			require.Equal(t, allGitCommits[1:3], cs)
		})

		t.Run("fetch more pages when commits are filtered out", func(t *testing.T) {
			c := NewMockGitserverServiceClient()
			commits := []*proto.GetCommitResponse{
				{
					Commit: &v1.GitCommit{
						Oid: "8",
					},
					ModifiedFiles: [][]byte{[]byte("file2")},
				},
				{
					Commit: &v1.GitCommit{
						Oid: "7",
					},
					ModifiedFiles: [][]byte{[]byte("file1")},
				},
				{
					Commit: &v1.GitCommit{
						Oid: "6",
					},
					ModifiedFiles: [][]byte{[]byte("file2")},
				},
				{
					Commit: &v1.GitCommit{
						Oid: "5",
					},
					ModifiedFiles: [][]byte{[]byte("file1")},
				},
				{
					Commit: &v1.GitCommit{
						Oid: "4",
					},
					ModifiedFiles: [][]byte{[]byte("file1")},
				},
				{
					Commit: &v1.GitCommit{
						Oid: "3",
					},
					ModifiedFiles: [][]byte{[]byte("file1")},
				},
				{
					Commit: &v1.GitCommit{
						Oid: "2",
					},
					ModifiedFiles: [][]byte{[]byte("file2")},
				},
				{
					Commit: &v1.GitCommit{
						Oid: "1",
					},
					ModifiedFiles: [][]byte{[]byte("file1")},
				},
			}
			source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
				o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
					c.CommitLogFunc.SetDefaultHook(func(ctx context.Context, clr *v1.CommitLogRequest, co ...grpc.CallOption) (v1.GitserverService_CommitLogClient, error) {
						// With subrepo permissions on, we always need to request modified files.
						require.True(t, clr.IncludeModifiedFiles)
						s := NewMockGitserverService_CommitLogClient()
						for range clr.MaxCommits {
							s.RecvFunc.PushReturn(&v1.CommitLogResponse{
								Commits: commits[:1],
							}, nil)
							commits = commits[1:]
						}
						s.RecvFunc.PushReturn(nil, io.EOF)
						return s, nil
					})
					return c
				}
			})

			checker := getTestSubRepoPermsChecker("file2")
			cli := NewTestClient(t).WithClientSource(source).WithChecker(checker)

			cs, err := cli.Commits(ctx, "repo", CommitsOptions{
				AllRefs: true,
				N:       3,
			})
			require.NoError(t, err)
			require.Equal(t, []*gitdomain.Commit{
				{
					ID: "7",
					Author: gitdomain.Signature{
						Date: time.Unix(0, 0).UTC(),
					},
					Committer: &gitdomain.Signature{
						Date: time.Unix(0, 0).UTC(),
					},
					Parents: []api.CommitID{},
				},
				{
					ID: "5",
					Author: gitdomain.Signature{
						Date: time.Unix(0, 0).UTC(),
					},
					Committer: &gitdomain.Signature{
						Date: time.Unix(0, 0).UTC(),
					},
					Parents: []api.CommitID{},
				},
				{
					ID: "4",
					Author: gitdomain.Signature{
						Date: time.Unix(0, 0).UTC(),
					},
					Committer: &gitdomain.Signature{
						Date: time.Unix(0, 0).UTC(),
					},
					Parents: []api.CommitID{},
				},
			}, cs)

			// First page [8,7,6] contains only one visible commit [7], second page
			// contains [5,4,3], and we have enough for N=3 with [7,5,4].
			mockrequire.CalledN(t, c.CommitLogFunc, 2)
		})
	})
}

func TestClient_MergeBaseOctopus(t *testing.T) {
	t.Run("correctly returns server response", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				c.MergeBaseOctopusFunc.SetDefaultReturn(&proto.MergeBaseOctopusResponse{MergeBaseCommitSha: "deadbeef"}, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		sha, err := c.MergeBaseOctopus(context.Background(), "repo", "master", "b2")
		require.NoError(t, err)
		require.Equal(t, api.CommitID("deadbeef"), sha)
	})
	t.Run("returns empty for empty merge base", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				c.MergeBaseOctopusFunc.SetDefaultReturn(&proto.MergeBaseOctopusResponse{MergeBaseCommitSha: ""}, nil)
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		sha, err := c.MergeBaseOctopus(context.Background(), "repo", "master", "b2")
		require.NoError(t, err)
		require.Equal(t, api.CommitID(""), sha)
	})
	t.Run("revision not found", func(t *testing.T) {
		source := NewTestClientSource(t, []string{"gitserver"}, func(o *TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				c := NewMockGitserverServiceClient()
				s, err := status.New(codes.NotFound, "bad revision").WithDetails(&proto.RevisionNotFoundPayload{Repo: "repo"})
				require.NoError(t, err)
				c.MergeBaseOctopusFunc.SetDefaultReturn(nil, s.Err())
				return c
			}
		})

		c := NewTestClient(t).WithClientSource(source)

		_, err := c.MergeBaseOctopus(context.Background(), "repo", "master", "b2")
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
	})
}

func mustParseTime(layout, value string) time.Time {
	tm, err := time.Parse(layout, value)
	if err != nil {
		panic(err.Error())
	}
	return tm
}
