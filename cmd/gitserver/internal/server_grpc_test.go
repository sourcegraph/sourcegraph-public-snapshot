package internal

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	mockassert "github.com/derision-test/go-mockgen/v2/testutil/assert"
	mockrequire "github.com/derision-test/go-mockgen/v2/testutil/require"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/fileutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	v1 "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestGRPCServer_Blame(t *testing.T) {
	mockSS := gitserver.NewMockGitserverService_BlameServer()
	// Add an actor to the context.
	a := actor.FromUser(1)
	mockSS.ContextFunc.SetDefaultReturn(actor.WithActor(context.Background(), a))
	t.Run("argument validation", func(t *testing.T) {
		gs := &grpcServer{}
		err := gs.Blame(&v1.BlameRequest{RepoName: "", Path: []byte("thepath")}, mockSS)
		require.ErrorContains(t, err, "repo must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
		err = gs.Blame(&v1.BlameRequest{RepoName: "therepo", Commit: ""}, mockSS)
		require.ErrorContains(t, err, "commit must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
		err = gs.Blame(&v1.BlameRequest{RepoName: "therepo", Commit: "deadbeef", Path: []byte("")}, mockSS)
		require.ErrorContains(t, err, "path must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
	})
	t.Run("checks for uncloned repo", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		fs.RepoClonedFunc.SetDefaultReturn(false, nil)
		locker := NewMockRepositoryLocker()
		locker.StatusFunc.SetDefaultReturn("cloning", true)
		gs := &grpcServer{svc: NewMockService(), fs: fs, locker: locker}
		err := gs.Blame(&v1.BlameRequest{RepoName: "therepo", Commit: "deadbeef", Path: []byte("thepath")}, mockSS)
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RepoNotFoundPayload{})
		require.Contains(t, err.Error(), "repo not found")
		mockassert.Called(t, fs.RepoClonedFunc)
		mockassert.Called(t, locker.StatusFunc)
	})
	t.Run("e2e", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		b := git.NewMockGitBackend()
		hr := git.NewMockBlameHunkReader()
		hr.ReadFunc.PushReturn(&gitdomain.Hunk{CommitID: "deadbeef"}, nil)
		hr.ReadFunc.PushReturn(nil, io.EOF)
		b.BlameFunc.PushReturn(hr, nil)
		gs := &grpcServer{
			svc: NewMockService(),
			fs:  fs,
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				return b
			},
		}

		cli := spawnServer(t, gs)
		r, err := cli.Blame(context.Background(), &v1.BlameRequest{
			RepoName: "therepo",
			Commit:   "deadbeef",
			Path:     []byte("thepath"),
		})
		require.NoError(t, err)
		for {
			msg, err := r.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				require.NoError(t, err)
			}
			if diff := cmp.Diff(&proto.BlameResponse{
				Hunk: &proto.BlameHunk{
					Commit: "deadbeef",
					Author: &v1.BlameAuthor{
						Date: timestamppb.New(time.Time{}),
					},
				},
			}, msg, cmpopts.IgnoreUnexported(proto.BlameResponse{}, proto.BlameHunk{}, proto.BlameAuthor{}, timestamppb.Timestamp{})); diff != "" {
				t.Fatalf("unexpected response (-want +got):\n%s", diff)
			}
		}

		{
			b.BlameFunc.PushReturn(nil, &os.PathError{Op: "open", Path: "thepath", Err: os.ErrNotExist})
			r, err = cli.Blame(context.Background(), &v1.BlameRequest{
				RepoName: "therepo",
				Commit:   "deadbeef",
				Path:     []byte("thepath"),
			})
			require.NoError(t, err)

			_, err := r.Recv()
			assertGRPCStatusCode(t, err, codes.NotFound)
			assertHasGRPCErrorDetailOfType(t, err, &proto.FileNotFoundPayload{})
		}

		{
			b.BlameFunc.PushReturn(nil, &gitdomain.RevisionNotFoundError{Repo: "therepo", Spec: "deadbeef"})
			r, err = cli.Blame(context.Background(), &v1.BlameRequest{
				RepoName: "therepo",
				Commit:   "deadbeef",
				Path:     []byte("thepath"),
			})
			require.NoError(t, err)

			_, err := r.Recv()
			assertGRPCStatusCode(t, err, codes.NotFound)
			assertHasGRPCErrorDetailOfType(t, err, &proto.RevisionNotFoundPayload{})
		}
	})
}

func TestGRPCServer_DefaultBranch(t *testing.T) {
	ctx := context.Background()
	t.Run("argument validation", func(t *testing.T) {
		gs := &grpcServer{}
		_, err := gs.DefaultBranch(ctx, &v1.DefaultBranchRequest{RepoName: ""})
		require.ErrorContains(t, err, "repo must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
	})
	t.Run("checks for uncloned repo", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		fs.RepoClonedFunc.SetDefaultReturn(false, nil)
		locker := NewMockRepositoryLocker()
		locker.StatusFunc.SetDefaultReturn("cloning", true)
		gs := &grpcServer{svc: NewMockService(), fs: fs, locker: locker}
		_, err := gs.DefaultBranch(ctx, &v1.DefaultBranchRequest{RepoName: "therepo"})
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RepoNotFoundPayload{})
		require.Contains(t, err.Error(), "repo not found")
		mockassert.Called(t, fs.RepoClonedFunc)
		mockassert.Called(t, locker.StatusFunc)
	})
	t.Run("e2e", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		b := git.NewMockGitBackend()
		b.SymbolicRefHeadFunc.SetDefaultReturn("refs/heads/main", nil)
		b.RevParseHeadFunc.SetDefaultReturn("deadbeef", nil)
		gs := &grpcServer{
			svc: NewMockService(),
			fs:  fs,
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				return b
			},
		}

		cli := spawnServer(t, gs)
		res, err := cli.DefaultBranch(ctx, &v1.DefaultBranchRequest{
			RepoName: "therepo",
		})
		require.NoError(t, err)
		if diff := cmp.Diff(&proto.DefaultBranchResponse{
			RefName: "refs/heads/main",
			Commit:  "deadbeef",
		}, res, cmpopts.IgnoreUnexported(proto.DefaultBranchResponse{})); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}

		// Check that RevNotFoundErrors are mapped correctly:
		b.RevParseHeadFunc.SetDefaultReturn("", &gitdomain.RevisionNotFoundError{Repo: "therepo", Spec: "HEAD"})
		_, err = cli.DefaultBranch(ctx, &v1.DefaultBranchRequest{
			RepoName: "therepo",
		})
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RevisionNotFoundPayload{})
	})
}

func TestGRPCServer_MergeBase(t *testing.T) {
	ctx := context.Background()
	t.Run("argument validation", func(t *testing.T) {
		gs := &grpcServer{}
		_, err := gs.MergeBase(ctx, &v1.MergeBaseRequest{RepoName: ""})
		require.ErrorContains(t, err, "repo must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
		_, err = gs.MergeBase(ctx, &v1.MergeBaseRequest{RepoName: "therepo", Base: []byte{}})
		require.ErrorContains(t, err, "base must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
		_, err = gs.MergeBase(ctx, &v1.MergeBaseRequest{RepoName: "therepo", Base: []byte("master")})
		require.ErrorContains(t, err, "head must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
	})
	t.Run("checks for uncloned repo", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		fs.RepoClonedFunc.SetDefaultReturn(false, nil)
		locker := NewMockRepositoryLocker()
		locker.StatusFunc.SetDefaultReturn("cloning", true)
		gs := &grpcServer{svc: NewMockService(), fs: fs, locker: locker}
		_, err := gs.MergeBase(ctx, &v1.MergeBaseRequest{RepoName: "therepo", Base: []byte("master"), Head: []byte("b2")})
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RepoNotFoundPayload{})
		require.Contains(t, err.Error(), "repo not found")
		mockassert.Called(t, fs.RepoClonedFunc)
		mockassert.Called(t, locker.StatusFunc)
	})
	t.Run("revision not found", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		gs := &grpcServer{
			svc: NewMockService(),
			fs:  fs,
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				b := git.NewMockGitBackend()
				b.MergeBaseFunc.SetDefaultReturn("", &gitdomain.RevisionNotFoundError{Repo: "therepo", Spec: "b2"})
				return b
			},
		}
		_, err := gs.MergeBase(ctx, &v1.MergeBaseRequest{RepoName: "therepo", Base: []byte("master"), Head: []byte("b2")})
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RevisionNotFoundPayload{})
		require.Contains(t, err.Error(), "revision not found")
	})
	t.Run("e2e", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		b := git.NewMockGitBackend()
		b.MergeBaseFunc.SetDefaultReturn("deadbeef", nil)
		gs := &grpcServer{
			svc: NewMockService(),
			fs:  fs,
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				return b
			},
		}

		cli := spawnServer(t, gs)
		res, err := cli.MergeBase(ctx, &v1.MergeBaseRequest{
			RepoName: "therepo",
			Base:     []byte("master"),
			Head:     []byte("b2"),
		})
		require.NoError(t, err)
		if diff := cmp.Diff(&proto.MergeBaseResponse{
			MergeBaseCommitSha: "deadbeef",
		}, res, cmpopts.IgnoreUnexported(proto.MergeBaseResponse{})); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}
	})
}

func TestGRPCServer_ReadFile(t *testing.T) {
	mockSS := gitserver.NewMockGitserverService_ReadFileServer()
	// Add an actor to the context.
	a := actor.FromUser(1)
	mockSS.ContextFunc.SetDefaultReturn(actor.WithActor(context.Background(), a))
	t.Run("argument validation", func(t *testing.T) {
		gs := &grpcServer{}
		err := gs.ReadFile(&v1.ReadFileRequest{RepoName: ""}, mockSS)
		require.ErrorContains(t, err, "repo must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
		err = gs.ReadFile(&v1.ReadFileRequest{RepoName: "therepo", Path: []byte("")}, mockSS)
		require.ErrorContains(t, err, "path must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
		err = gs.ReadFile(&v1.ReadFileRequest{RepoName: "therepo", Path: []byte("thepath"), Commit: ""}, mockSS)
		require.ErrorContains(t, err, "commit must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
	})
	t.Run("checks for uncloned repo", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		fs.RepoClonedFunc.SetDefaultReturn(false, nil)
		locker := NewMockRepositoryLocker()
		locker.StatusFunc.SetDefaultReturn("cloning", true)
		gs := &grpcServer{svc: NewMockService(), fs: fs, locker: locker}
		err := gs.ReadFile(&v1.ReadFileRequest{RepoName: "therepo", Commit: "deadbeef", Path: []byte("thepath")}, mockSS)
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RepoNotFoundPayload{})
		require.Contains(t, err.Error(), "repo not found")
		mockassert.Called(t, fs.RepoClonedFunc)
		mockassert.Called(t, locker.StatusFunc)
	})
	t.Run("e2e", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		b := git.NewMockGitBackend()
		b.ReadFileFunc.SetDefaultReturn(io.NopCloser(bytes.NewReader([]byte("filecontent"))), nil)
		gs := &grpcServer{
			svc: NewMockService(),
			fs:  fs,
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				return b
			},
		}

		cli := spawnServer(t, gs)
		r, err := cli.ReadFile(context.Background(), &v1.ReadFileRequest{
			RepoName: "therepo",
			Commit:   "deadbeef",
			Path:     []byte("thepath"),
		})
		require.NoError(t, err)
		for {
			msg, err := r.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				require.NoError(t, err)
			}
			if diff := cmp.Diff(&proto.ReadFileResponse{
				Data: []byte("filecontent"),
			}, msg, cmpopts.IgnoreUnexported(proto.ReadFileResponse{})); diff != "" {
				t.Fatalf("unexpected response (-want +got):\n%s", diff)
			}
		}

		b.ReadFileFunc.SetDefaultReturn(nil, os.ErrNotExist)
		cc, err := cli.ReadFile(context.Background(), &v1.ReadFileRequest{
			RepoName: "therepo",
			Commit:   "deadbeef",
			Path:     []byte("thepath"),
		})
		require.NoError(t, err)
		_, err = cc.Recv()
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.FileNotFoundPayload{})

		b.ReadFileFunc.SetDefaultReturn(nil, &gitdomain.RevisionNotFoundError{})
		cc, err = cli.ReadFile(context.Background(), &v1.ReadFileRequest{
			RepoName: "therepo",
			Commit:   "deadbeef",
			Path:     []byte("thepath"),
		})
		require.NoError(t, err)
		_, err = cc.Recv()
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RevisionNotFoundPayload{})
	})
}

func TestGRPCServer_Archive(t *testing.T) {
	mockSS := gitserver.NewMockGitserverService_ArchiveServer()
	// Add an actor to the context.
	a := actor.FromUser(1)
	mockSS.ContextFunc.SetDefaultReturn(actor.WithActor(context.Background(), a))
	t.Run("argument validation", func(t *testing.T) {
		gs := &grpcServer{}
		err := gs.Archive(&v1.ArchiveRequest{Repo: ""}, mockSS)
		require.ErrorContains(t, err, "repo must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)

		err = gs.Archive(&v1.ArchiveRequest{Repo: "therepo", Format: proto.ArchiveFormat_ARCHIVE_FORMAT_TAR}, mockSS)
		require.ErrorContains(t, err, "treeish must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)

		err = gs.Archive(&v1.ArchiveRequest{Repo: "therepo", Treeish: "HEAD"}, mockSS)
		require.ErrorContains(t, err, "unknown archive format")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
	})
	t.Run("checks for uncloned repo", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		fs.RepoClonedFunc.SetDefaultReturn(false, nil)
		locker := NewMockRepositoryLocker()
		locker.StatusFunc.SetDefaultReturn("cloning", true)
		gs := &grpcServer{svc: NewMockService(), fs: fs, locker: locker}
		err := gs.Archive(&v1.ArchiveRequest{Repo: "therepo", Treeish: "HEAD", Format: proto.ArchiveFormat_ARCHIVE_FORMAT_ZIP}, mockSS)
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RepoNotFoundPayload{})
		require.Contains(t, err.Error(), "repo not found")
		mockassert.Called(t, fs.RepoClonedFunc)
		mockassert.Called(t, locker.StatusFunc)
	})
	t.Run("e2e", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		b := git.NewMockGitBackend()
		b.ArchiveReaderFunc.SetDefaultReturn(io.NopCloser(bytes.NewReader([]byte("filecontent"))), nil)
		gs := &grpcServer{
			svc: NewMockService(),
			fs:  fs,
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				return b
			},
		}

		cli := spawnServer(t, gs)
		r, err := cli.Archive(context.Background(), &v1.ArchiveRequest{
			Repo:    "therepo",
			Treeish: "HEAD",
			Format:  proto.ArchiveFormat_ARCHIVE_FORMAT_ZIP,
		})
		require.NoError(t, err)
		for {
			msg, err := r.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				require.NoError(t, err)
			}
			if diff := cmp.Diff(&proto.ArchiveResponse{
				Data: []byte("filecontent"),
			}, msg, cmpopts.IgnoreUnexported(proto.ArchiveResponse{})); diff != "" {
				t.Fatalf("unexpected response (-want +got):\n%s", diff)
			}
		}

		// TODO: Do we return this?
		b.ArchiveReaderFunc.SetDefaultReturn(nil, &gitdomain.RevisionNotFoundError{})
		cc, err := cli.Archive(context.Background(), &v1.ArchiveRequest{
			Repo:    "therepo",
			Treeish: "HEAD",
			Format:  proto.ArchiveFormat_ARCHIVE_FORMAT_ZIP,
		})
		require.NoError(t, err)
		_, err = cc.Recv()
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RevisionNotFoundPayload{})
	})
}

func TestGRPCServer_GetCommit(t *testing.T) {
	// Add an actor to the context.
	a := actor.FromUser(1)
	ctx := actor.WithActor(context.Background(), a)
	t.Run("argument validation", func(t *testing.T) {
		gs := &grpcServer{}
		_, err := gs.GetCommit(ctx, &v1.GetCommitRequest{RepoName: ""})
		require.ErrorContains(t, err, "repo must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
		_, err = gs.GetCommit(ctx, &v1.GetCommitRequest{RepoName: "therepo", Commit: ""})
		require.ErrorContains(t, err, "commit must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
	})
	t.Run("checks for uncloned repo", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		fs.RepoClonedFunc.SetDefaultReturn(false, nil)
		locker := NewMockRepositoryLocker()
		locker.StatusFunc.SetDefaultReturn("cloning", true)
		gs := &grpcServer{svc: NewMockService(), fs: fs, locker: locker}
		_, err := gs.GetCommit(ctx, &v1.GetCommitRequest{RepoName: "therepo", Commit: "deadbeef"})
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RepoNotFoundPayload{})
		require.Contains(t, err.Error(), "repo not found")
		mockassert.Called(t, fs.RepoClonedFunc)
		mockassert.Called(t, locker.StatusFunc)
	})
	t.Run("e2e", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		b := git.NewMockGitBackend()
		now := time.Now()
		b.GetCommitFunc.PushReturn(&git.GitCommitWithFiles{Commit: &gitdomain.Commit{
			Committer: &gitdomain.Signature{
				Name:  "committer",
				Email: "committer@sourcegraph.com",
				Date:  now,
			},
			Author: gitdomain.Signature{
				Name:  "author",
				Email: "author@sourcegraph.com",
				Date:  now,
			},
		}}, nil)
		b.GetCommitFunc.PushReturn(&git.GitCommitWithFiles{Commit: &gitdomain.Commit{
			Committer: &gitdomain.Signature{
				Name:  "committer",
				Email: "committer@sourcegraph.com",
				Date:  now,
			},
			Author: gitdomain.Signature{
				Name:  "author",
				Email: "author@sourcegraph.com",
				Date:  now,
			},
		}, ModifiedFiles: []string{"modfile"}}, nil)
		gs := &grpcServer{
			svc: NewMockService(),
			fs:  fs,
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				return b
			},
		}

		cli := spawnServer(t, gs)
		_, err := cli.GetCommit(ctx, &v1.GetCommitRequest{
			RepoName: "therepo",
			Commit:   "deadbeef",
		})
		require.NoError(t, err)

		commit, err := cli.GetCommit(ctx, &v1.GetCommitRequest{
			RepoName:             "therepo",
			Commit:               "deadbeef",
			IncludeModifiedFiles: true,
		})
		require.NoError(t, err)
		mockrequire.CalledAtNWith(t, b.GetCommitFunc, 0, mockassert.Values(mockassert.Skip, api.CommitID("deadbeef"), false))
		mockrequire.CalledAtNWith(t, b.GetCommitFunc, 1, mockassert.Values(mockassert.Skip, api.CommitID("deadbeef"), true))
		if diff := cmp.Diff(&proto.GetCommitResponse{
			Commit: &v1.GitCommit{
				Committer: &v1.GitSignature{
					Name:  []byte("committer"),
					Email: []byte("committer@sourcegraph.com"),
					Date:  timestamppb.New(now),
				},
				Author: &v1.GitSignature{
					Name:  []byte("author"),
					Email: []byte("author@sourcegraph.com"),
					Date:  timestamppb.New(now),
				},
			},
			ModifiedFiles: [][]byte{[]byte("modfile")},
		}, commit, cmpopts.IgnoreUnexported(proto.GetCommitResponse{}, proto.GitCommit{}, proto.GitSignature{}, timestamppb.Timestamp{})); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}

		b.GetCommitFunc.PushReturn(nil, &gitdomain.RevisionNotFoundError{Repo: "therepo", Spec: "deadbeef"})
		_, err = cli.GetCommit(ctx, &v1.GetCommitRequest{
			RepoName: "therepo",
			Commit:   "deadbeef",
		})
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RevisionNotFoundPayload{})
	})
}

func TestGRPCServer_ResolveRevision(t *testing.T) {
	ctx := context.Background()
	t.Run("argument validation", func(t *testing.T) {
		gs := &grpcServer{}
		_, err := gs.ResolveRevision(ctx, &v1.ResolveRevisionRequest{RepoName: ""})
		require.ErrorContains(t, err, "repo must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
	})
	t.Run("checks for uncloned repo", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		fs.RepoClonedFunc.SetDefaultReturn(false, nil)
		locker := NewMockRepositoryLocker()
		locker.StatusFunc.SetDefaultReturn("cloning", true)
		gs := &grpcServer{svc: NewMockService(), fs: fs, locker: locker}
		_, err := gs.ResolveRevision(ctx, &v1.ResolveRevisionRequest{RepoName: "therepo"})
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RepoNotFoundPayload{})
		require.Contains(t, err.Error(), "repo not found")
		mockassert.Called(t, fs.RepoClonedFunc)
		mockassert.Called(t, locker.StatusFunc)
	})
	t.Run("e2e", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		b := git.NewMockGitBackend()
		b.ResolveRevisionFunc.SetDefaultReturn("deadbeef", nil)
		svc := NewMockService()
		gs := &grpcServer{
			svc: svc,
			fs:  fs,
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				return b
			},
		}

		cli := spawnServer(t, gs)
		res, err := cli.ResolveRevision(ctx, &v1.ResolveRevisionRequest{
			RepoName: "therepo",
		})
		require.NoError(t, err)
		if diff := cmp.Diff(&proto.ResolveRevisionResponse{
			CommitSha: "deadbeef",
		}, res, cmpopts.IgnoreUnexported(proto.ResolveRevisionResponse{})); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}

		// Check that RevNotFoundErrors are mapped correctly:
		b.ResolveRevisionFunc.SetDefaultReturn("", &gitdomain.RevisionNotFoundError{Repo: "therepo", Spec: "HEAD"})
		_, err = cli.ResolveRevision(ctx, &v1.ResolveRevisionRequest{
			RepoName: "therepo",
		})
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RevisionNotFoundPayload{})

		// Test EnsureRevision is called correctly.
		// Initially, the revision is not found.
		b.ResolveRevisionFunc.PushReturn("", &gitdomain.RevisionNotFoundError{Repo: "therepo", Spec: "HEAD"})
		// EnsureRevision was able to run a fetch, retry.
		svc.EnsureRevisionFunc.SetDefaultReturn(true)
		// After the fetch, resolve revision succeeds.
		b.ResolveRevisionFunc.PushReturn("deadbeef", nil)
		_, err = cli.ResolveRevision(ctx, &v1.ResolveRevisionRequest{
			RepoName:       "therepo",
			RevSpec:        []byte("HEAD"),
			EnsureRevision: pointers.Ptr(true),
		})
		require.NoError(t, err)
		if diff := cmp.Diff(&proto.ResolveRevisionResponse{
			CommitSha: "deadbeef",
		}, res, cmpopts.IgnoreUnexported(proto.ResolveRevisionResponse{})); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}
		mockrequire.Called(t, svc.EnsureRevisionFunc)
	})
}

func TestGRPCServer_RevAtTime(t *testing.T) {
	ctx := context.Background()
	t.Run("argument validation", func(t *testing.T) {
		gs := &grpcServer{}
		_, err := gs.RevAtTime(ctx, &v1.RevAtTimeRequest{RepoName: "", RevSpec: []byte("HEAD"), Time: timestamppb.Now()})
		require.ErrorContains(t, err, "repo must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
	})
	t.Run("checks for uncloned repo", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		fs.RepoClonedFunc.SetDefaultReturn(false, nil)
		locker := NewMockRepositoryLocker()
		locker.StatusFunc.SetDefaultReturn("cloning", true)
		gs := &grpcServer{svc: NewMockService(), fs: fs, locker: locker}
		_, err := gs.RevAtTime(ctx, &v1.RevAtTimeRequest{RepoName: "therepo", RevSpec: []byte("HEAD"), Time: timestamppb.Now()})
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RepoNotFoundPayload{})
		require.Contains(t, err.Error(), "repo not found")
		mockassert.Called(t, fs.RepoClonedFunc)
		mockassert.Called(t, locker.StatusFunc)
	})
	t.Run("e2e", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		b := git.NewMockGitBackend()
		b.RevAtTimeFunc.SetDefaultReturn("deadbeef", nil)
		svc := NewMockService()
		gs := &grpcServer{
			svc: svc,
			fs:  fs,
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				return b
			},
		}

		cli := spawnServer(t, gs)
		res, err := cli.RevAtTime(ctx, &v1.RevAtTimeRequest{
			RepoName: "therepo",
			RevSpec:  []byte("HEAD"),
			Time:     timestamppb.Now(),
		})
		require.NoError(t, err)
		if diff := cmp.Diff(&proto.RevAtTimeResponse{
			CommitSha: "deadbeef",
		}, res, cmpopts.IgnoreUnexported(proto.RevAtTimeResponse{})); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}
	})
}

func TestGRPCServer_GetObject(t *testing.T) {
	ctx := context.Background()

	t.Run("argument validation", func(t *testing.T) {
		gs := &grpcServer{}
		t.Run("repo must be specified", func(t *testing.T) {
			_, err := gs.GetObject(ctx, &v1.GetObjectRequest{})
			require.ErrorContains(t, err, "repo must be specified")
			assertGRPCStatusCode(t, err, codes.InvalidArgument)
		})

		t.Run("object name must be specified", func(t *testing.T) {
			_, err := gs.GetObject(ctx, &v1.GetObjectRequest{Repo: "therepo"})
			require.ErrorContains(t, err, "object name must be specified")
			assertGRPCStatusCode(t, err, codes.InvalidArgument)
		})
	})

	t.Run("checks for uncloned repo", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		fs.RepoClonedFunc.SetDefaultReturn(false, nil)
		locker := NewMockRepositoryLocker()
		locker.StatusFunc.SetDefaultReturn("cloning", true)
		gs := &grpcServer{svc: NewMockService(), fs: fs, locker: locker}
		_, err := gs.GetObject(ctx, &v1.GetObjectRequest{Repo: "therepo", ObjectName: "deadbeef"})
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RepoNotFoundPayload{})
		require.Contains(t, err.Error(), "repo not found")
		mockassert.Called(t, fs.RepoClonedFunc)
		mockassert.Called(t, locker.StatusFunc)
	})

	t.Run("e2e", func(t *testing.T) {
		expectedID := gitdomain.OID{0xde, 0xad, 0xbe, 0xef, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		b := git.NewMockGitBackend()

		b.GetObjectFunc.SetDefaultReturn(&gitdomain.GitObject{
			ID:   expectedID,
			Type: gitdomain.ObjectTypeBlob,
		}, nil)
		gs := &grpcServer{
			svc:    NewMockService(),
			fs:     fs,
			logger: logtest.Scoped(t),
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				return b
			},
		}

		cli := spawnServer(t, gs)

		res, err := cli.GetObject(ctx, &v1.GetObjectRequest{
			Repo:       "therepo",
			ObjectName: "deadbeef",
		})
		require.NoError(t, err)
		if diff := cmp.Diff(&proto.GetObjectResponse{
			Object: &proto.GitObject{
				Id:   expectedID[:],
				Type: proto.GitObject_OBJECT_TYPE_BLOB,
			},
		}, res, protocmp.Transform()); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}

		b.GetObjectFunc.SetDefaultReturn(nil, &gitdomain.RevisionNotFoundError{Repo: "therepo", Spec: "deadbeef"})
		_, err = cli.GetObject(ctx, &v1.GetObjectRequest{
			Repo:       "therepo",
			ObjectName: "deadbeef",
		})

		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RevisionNotFoundPayload{})
	})
}

func TestGRPCServer_ListRefs(t *testing.T) {
	ctx := context.Background()
	mockSS := gitserver.NewMockGitserverService_ListRefsServer()
	mockSS.ContextFunc.SetDefaultReturn(ctx)
	t.Run("argument validation", func(t *testing.T) {
		gs := &grpcServer{}
		err := gs.ListRefs(&v1.ListRefsRequest{RepoName: ""}, mockSS)
		require.ErrorContains(t, err, "repo must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
	})
	t.Run("checks for uncloned repo", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		fs.RepoClonedFunc.SetDefaultReturn(false, nil)
		locker := NewMockRepositoryLocker()
		locker.StatusFunc.SetDefaultReturn("cloning", true)
		gs := &grpcServer{svc: NewMockService(), fs: fs, locker: locker}
		err := gs.ListRefs(&v1.ListRefsRequest{RepoName: "therepo"}, mockSS)
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RepoNotFoundPayload{})
		require.Contains(t, err.Error(), "repo not found")
		mockassert.Called(t, fs.RepoClonedFunc)
		mockassert.Called(t, locker.StatusFunc)
	})
	t.Run("e2e", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		b := git.NewMockGitBackend()
		it := git.NewMockRefIterator()
		it.NextFunc.PushReturn(&gitdomain.Ref{Name: "refs/heads/master"}, nil)
		it.NextFunc.PushReturn(nil, io.EOF)
		b.ListRefsFunc.SetDefaultReturn(it, nil)
		gs := &grpcServer{
			svc: NewMockService(),
			fs:  fs,
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				return b
			},
		}

		cli := spawnServer(t, gs)
		cc, err := cli.ListRefs(ctx, &v1.ListRefsRequest{
			RepoName: "therepo",
		})
		require.NoError(t, err)
		refs := []*v1.GitRef{}
		for {
			resp, err := cc.Recv()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
			refs = append(refs, resp.GetRefs()...)
		}
		if diff := cmp.Diff([]*v1.GitRef{
			{
				RefName:   []byte("refs/heads/master"),
				CreatedAt: timestamppb.New(time.Time{}),
			},
		}, refs, cmpopts.IgnoreUnexported(v1.GitRef{}, timestamppb.Timestamp{})); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}
	})
}

func TestGRPCServer_RawDiff(t *testing.T) {
	mockSS := gitserver.NewMockGitserverService_RawDiffServer()
	// Add an actor to the context.
	a := actor.FromUser(1)
	mockSS.ContextFunc.SetDefaultReturn(actor.WithActor(context.Background(), a))
	t.Run("argument validation", func(t *testing.T) {
		gs := &grpcServer{}
		err := gs.RawDiff(&v1.RawDiffRequest{RepoName: ""}, mockSS)
		require.ErrorContains(t, err, "repo must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
		err = gs.RawDiff(&v1.RawDiffRequest{RepoName: "therepo"}, mockSS)
		require.ErrorContains(t, err, "base_rev_spec must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
		err = gs.RawDiff(&v1.RawDiffRequest{RepoName: "therepo", BaseRevSpec: []byte("base")}, mockSS)
		require.ErrorContains(t, err, "head_rev_spec must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
		err = gs.RawDiff(&v1.RawDiffRequest{RepoName: "therepo", BaseRevSpec: []byte("base"), HeadRevSpec: []byte("head")}, mockSS)
		require.ErrorContains(t, err, "comparison_type must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
	})
	t.Run("checks for uncloned repo", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		fs.RepoClonedFunc.SetDefaultReturn(false, nil)
		locker := NewMockRepositoryLocker()
		locker.StatusFunc.SetDefaultReturn("cloning", true)
		gs := &grpcServer{svc: NewMockService(), fs: fs, locker: locker}
		err := gs.RawDiff(&v1.RawDiffRequest{RepoName: "therepo", BaseRevSpec: []byte("base"), HeadRevSpec: []byte("head"), ComparisonType: proto.RawDiffRequest_COMPARISON_TYPE_INTERSECTION}, mockSS)
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RepoNotFoundPayload{})
		require.Contains(t, err.Error(), "repo not found")
		mockassert.Called(t, fs.RepoClonedFunc)
		mockassert.Called(t, locker.StatusFunc)
	})
	t.Run("e2e", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		b := git.NewMockGitBackend()
		b.RawDiffFunc.SetDefaultReturn(io.NopCloser(bytes.NewReader([]byte("diffcontent"))), nil)
		gs := &grpcServer{
			svc: NewMockService(),
			fs:  fs,
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				return b
			},
		}

		cli := spawnServer(t, gs)
		r, err := cli.RawDiff(context.Background(), &v1.RawDiffRequest{
			RepoName:       "therepo",
			BaseRevSpec:    []byte("base"),
			HeadRevSpec:    []byte("head"),
			ComparisonType: proto.RawDiffRequest_COMPARISON_TYPE_INTERSECTION,
		})
		require.NoError(t, err)
		for {
			msg, err := r.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				require.NoError(t, err)
			}
			if diff := cmp.Diff(&proto.RawDiffResponse{
				Chunk: []byte("diffcontent"),
			}, msg, cmpopts.IgnoreUnexported(proto.RawDiffResponse{})); diff != "" {
				t.Fatalf("unexpected response (-want +got):\n%s", diff)
			}
		}

		b.RawDiffFunc.SetDefaultReturn(nil, &gitdomain.RevisionNotFoundError{})
		r, err = cli.RawDiff(context.Background(), &v1.RawDiffRequest{
			RepoName:       "therepo",
			BaseRevSpec:    []byte("base"),
			HeadRevSpec:    []byte("head"),
			ComparisonType: proto.RawDiffRequest_COMPARISON_TYPE_INTERSECTION,
		})
		require.NoError(t, err)
		_, err = r.Recv()
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RevisionNotFoundPayload{})
	})
}

func TestGRPCServer_ContributorCounts(t *testing.T) {
	ctx := context.Background()
	t.Run("argument validation", func(t *testing.T) {
		gs := &grpcServer{}
		_, err := gs.ContributorCounts(ctx, &v1.ContributorCountsRequest{RepoName: ""})
		require.ErrorContains(t, err, "repo must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
	})
	t.Run("checks for uncloned repo", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		fs.RepoClonedFunc.SetDefaultReturn(false, nil)
		locker := NewMockRepositoryLocker()
		locker.StatusFunc.SetDefaultReturn("cloning", true)
		gs := &grpcServer{svc: NewMockService(), fs: fs, locker: locker}
		_, err := gs.ContributorCounts(ctx, &v1.ContributorCountsRequest{RepoName: "therepo"})
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RepoNotFoundPayload{})
		require.Contains(t, err.Error(), "repo not found")
		mockassert.Called(t, fs.RepoClonedFunc)
		mockassert.Called(t, locker.StatusFunc)
	})
	t.Run("e2e", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		b := git.NewMockGitBackend()
		b.ContributorCountsFunc.SetDefaultReturn([]*gitdomain.ContributorCount{{Count: 1, Name: "Foo", Email: "foo@sourcegraph.com"}}, nil)
		svc := NewMockService()
		gs := &grpcServer{
			svc: svc,
			fs:  fs,
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				return b
			},
		}

		cli := spawnServer(t, gs)
		res, err := cli.ContributorCounts(ctx, &v1.ContributorCountsRequest{
			RepoName: "therepo",
		})
		require.NoError(t, err)
		if diff := cmp.Diff(&v1.ContributorCountsResponse{
			Counts: []*v1.ContributorCount{
				{
					Author: &v1.GitSignature{
						Name:  []byte("Foo"),
						Email: []byte("foo@sourcegraph.com"),
					},
					Count: int32(1),
				},
			},
		}, res, cmpopts.IgnoreUnexported(v1.ContributorCountsResponse{}, v1.ContributorCount{}, v1.GitSignature{})); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}
	})
}

func TestGRPCServer_ChangedFiles(t *testing.T) {
	mockSS := gitserver.NewMockGitserverService_ChangedFilesServer()
	mockSS.ContextFunc.SetDefaultReturn(context.Background())
	t.Run("argument validation", func(t *testing.T) {
		t.Run("repo must be specified", func(t *testing.T) {
			gs := &grpcServer{}
			err := gs.ChangedFiles(&v1.ChangedFilesRequest{RepoName: "", Head: []byte("HEAD")}, mockSS)
			require.ErrorContains(t, err, "repo must be specified")
			assertGRPCStatusCode(t, err, codes.InvalidArgument)
		})

		t.Run("head (<tree-ish>) must be specified", func(t *testing.T) {
			gs := &grpcServer{}
			err := gs.ChangedFiles(&v1.ChangedFilesRequest{RepoName: "therepo"}, mockSS)
			require.ErrorContains(t, err, "head (<tree-ish>) must be specified")
			assertGRPCStatusCode(t, err, codes.InvalidArgument)
		})
	})
	t.Run("checks for uncloned repo", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		fs.RepoClonedFunc.SetDefaultReturn(false, nil)
		locker := NewMockRepositoryLocker()
		locker.StatusFunc.SetDefaultReturn("cloning", true)
		gs := &grpcServer{svc: NewMockService(), fs: fs, locker: locker}
		err := gs.ChangedFiles(&v1.ChangedFilesRequest{RepoName: "therepo", Base: []byte("base"), Head: []byte("head")}, mockSS)
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RepoNotFoundPayload{})
		require.Contains(t, err.Error(), "repo not found")
		mockassert.Called(t, fs.RepoClonedFunc)
		mockassert.Called(t, locker.StatusFunc)
	})
	t.Run("revision not found", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		gs := &grpcServer{
			svc: NewMockService(),
			fs:  fs,
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				b := git.NewMockGitBackend()
				b.ChangedFilesFunc.SetDefaultReturn(nil, &gitdomain.RevisionNotFoundError{Repo: "therepo", Spec: "base...head"})
				return b
			},
		}
		err := gs.ChangedFiles(&v1.ChangedFilesRequest{RepoName: "therepo", Base: []byte("base"), Head: []byte("head")}, mockSS)
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RevisionNotFoundPayload{})
		require.Contains(t, err.Error(), "revision not found")
	})
	t.Run("e2e", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		b := git.NewMockGitBackend()
		b.ChangedFilesFunc.SetDefaultReturn(&testChangedFilesIterator{
			paths: []gitdomain.PathStatus{
				{Path: "file1.txt", Status: gitdomain.StatusAdded},
				{Path: "file2.txt", Status: gitdomain.StatusModified},
				{Path: "file3.txt", Status: gitdomain.StatusDeleted},
			},
		}, nil)
		gs := &grpcServer{
			svc: NewMockService(),
			fs:  fs,
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				return b
			},
		}

		cli := spawnServer(t, gs)
		r, err := cli.ChangedFiles(context.Background(), &v1.ChangedFilesRequest{
			RepoName: "therepo",
			Base:     []byte("base"),
			Head:     []byte("head"),
		})
		require.NoError(t, err)
		var paths []*proto.ChangedFile
		for {
			msg, err := r.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				require.NoError(t, err)
			}
			paths = append(paths, msg.GetFiles()...)
		}
		if diff := cmp.Diff([]*proto.ChangedFile{
			{Path: []byte("file1.txt"), Status: proto.ChangedFile_STATUS_ADDED},
			{Path: []byte("file2.txt"), Status: proto.ChangedFile_STATUS_MODIFIED},
			{Path: []byte("file3.txt"), Status: proto.ChangedFile_STATUS_DELETED},
		}, paths, protocmp.Transform()); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}
	})
}

func TestGRPCServer_FirstCommitEver(t *testing.T) {
	ctx := context.Background()

	t.Run("argument validation", func(t *testing.T) {
		gs := &grpcServer{}
		_, err := gs.FirstEverCommit(ctx, &v1.FirstEverCommitRequest{RepoName: ""})
		require.ErrorContains(t, err, "repo must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
	})

	t.Run("checks for uncloned repo", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		fs.RepoClonedFunc.SetDefaultReturn(false, nil)
		locker := NewMockRepositoryLocker()
		locker.StatusFunc.SetDefaultReturn("cloning", true)
		gs := &grpcServer{svc: NewMockService(), fs: fs, locker: locker}
		_, err := gs.FirstEverCommit(ctx, &v1.FirstEverCommitRequest{RepoName: "therepo"})
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RepoNotFoundPayload{})
		require.Contains(t, err.Error(), "repo not found")
		mockassert.Called(t, fs.RepoClonedFunc)
		mockassert.Called(t, locker.StatusFunc)
	})

	t.Run("e2e", func(t *testing.T) {
		expectedCommit := &gitdomain.Commit{
			ID: "f0e8d8b3d070c1c89f4e634c1d9e7f7d7d6a3f9a",
			Author: gitdomain.Signature{
				Name:  "John Doe",
				Email: "john@example.com",
				Date:  time.Date(2023, 4, 1, 12, 0, 0, 0, time.UTC),
			},
			Committer: &gitdomain.Signature{
				Name:  "Jane Smith",
				Email: "jane@example.com",
				Date:  time.Date(2023, 4, 1, 12, 5, 0, 0, time.UTC),
			},
			Message: "Initial commit",
			Parents: []api.CommitID{},
		}

		fs := gitserverfs.NewMockFS()

		// First, check to see that the commit is returned correctly.

		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		b := git.NewMockGitBackend()
		b.FirstEverCommitFunc.PushReturn(expectedCommit.ID, nil)
		b.GetCommitFunc.PushReturn(&git.GitCommitWithFiles{Commit: expectedCommit}, nil)
		gs := &grpcServer{
			svc: NewMockService(),
			fs:  fs,
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				return b
			},
		}

		cli := spawnServer(t, gs)
		rawResponse, err := cli.FirstEverCommit(ctx, &v1.FirstEverCommitRequest{
			RepoName: "therepo",
		})

		actualResponse := rawResponse.GetCommit()
		if diff := cmp.Diff(expectedCommit.ToProto(), actualResponse, protocmp.Transform()); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}

		require.NoError(t, err)

		// Second, check to see that the correct error is returned if the repository is empty

		b.FirstEverCommitFunc.PushReturn("", &gitdomain.RevisionNotFoundError{Repo: "therepo", Spec: "HEAD"})
		_, err = cli.FirstEverCommit(ctx, &v1.FirstEverCommitRequest{
			RepoName: "therepo",
		})

		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RevisionNotFoundPayload{})
	})
}

func TestGRPCServer_BehindAhead(t *testing.T) {
	ctx := context.Background()

	t.Run("argument validation", func(t *testing.T) {
		gs := &grpcServer{}
		_, err := gs.BehindAhead(ctx, &proto.BehindAheadRequest{RepoName: ""})
		require.ErrorContains(t, err, "repo must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
	})

	t.Run("checks for uncloned repo", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		fs.RepoClonedFunc.SetDefaultReturn(false, nil)
		locker := NewMockRepositoryLocker()
		locker.StatusFunc.SetDefaultReturn("cloning", true)
		gs := &grpcServer{svc: NewMockService(), fs: fs, locker: locker}
		_, err := gs.BehindAhead(ctx, &proto.BehindAheadRequest{RepoName: "therepo", Left: []byte("base"), Right: []byte("head")})
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RepoNotFoundPayload{})
		require.Contains(t, err.Error(), "repo not found")
		mockassert.Called(t, fs.RepoClonedFunc)
		mockassert.Called(t, locker.StatusFunc)
	})

	t.Run("revision not found", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		gs := &grpcServer{
			svc: NewMockService(),
			fs:  fs,
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				b := git.NewMockGitBackend()
				b.BehindAheadFunc.SetDefaultReturn(&gitdomain.BehindAhead{}, &gitdomain.RevisionNotFoundError{Repo: "therepo", Spec: "base...head"})
				return b
			},
		}
		_, err := gs.BehindAhead(ctx, &proto.BehindAheadRequest{RepoName: "therepo", Left: []byte("base"), Right: []byte("head")})
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RevisionNotFoundPayload{})
		require.Contains(t, err.Error(), "revision not found")
	})

	t.Run("e2e", func(t *testing.T) {
		expectedBehindAhead := gitdomain.BehindAhead{Behind: 5, Ahead: 3}

		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		b := git.NewMockGitBackend()
		b.BehindAheadFunc.SetDefaultReturn(&expectedBehindAhead, nil)
		gs := &grpcServer{
			svc: NewMockService(),
			fs:  fs,
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				return b
			},
		}

		cli := spawnServer(t, gs)
		response, err := cli.BehindAhead(ctx, &proto.BehindAheadRequest{
			RepoName: "therepo",
			Left:     []byte("base"),
			Right:    []byte("head"),
		})
		require.NoError(t, err)

		if diff := cmp.Diff(&proto.BehindAheadResponse{
			Behind: expectedBehindAhead.Behind,
			Ahead:  expectedBehindAhead.Ahead,
		}, response, cmpopts.IgnoreUnexported(proto.BehindAheadResponse{})); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}
	})
}
func TestGRPCServer_Stat(t *testing.T) {
	ctx := context.Background()

	t.Run("argument validation", func(t *testing.T) {
		gs := &grpcServer{}
		_, err := gs.Stat(ctx, &proto.StatRequest{RepoName: ""})
		require.ErrorContains(t, err, "repo must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)

		_, err = gs.Stat(ctx, &proto.StatRequest{RepoName: "repo"})
		require.ErrorContains(t, err, "commit_sha must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
	})

	t.Run("checks for uncloned repo", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		fs.RepoClonedFunc.SetDefaultReturn(false, nil)
		locker := NewMockRepositoryLocker()
		locker.StatusFunc.SetDefaultReturn("cloning", true)
		gs := &grpcServer{svc: NewMockService(), fs: fs, locker: locker}
		_, err := gs.Stat(ctx, &proto.StatRequest{RepoName: "therepo", CommitSha: "HEAD"})
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RepoNotFoundPayload{})
		require.Contains(t, err.Error(), "repo not found")
		mockassert.Called(t, fs.RepoClonedFunc)
		mockassert.Called(t, locker.StatusFunc)
	})

	t.Run("revision not found", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		gs := &grpcServer{
			svc: NewMockService(),
			fs:  fs,
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				b := git.NewMockGitBackend()
				b.StatFunc.SetDefaultReturn(nil, &gitdomain.RevisionNotFoundError{Repo: "therepo", Spec: "base...head"})
				return b
			},
		}
		_, err := gs.Stat(ctx, &proto.StatRequest{RepoName: "therepo", CommitSha: "HEAD"})
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RevisionNotFoundPayload{})
		require.Contains(t, err.Error(), "revision not found")
	})

	t.Run("e2e", func(t *testing.T) {
		expectedStat := fileutil.FileInfo{Name_: "file"}
		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		b := git.NewMockGitBackend()
		b.StatFunc.SetDefaultReturn(&expectedStat, nil)
		gs := &grpcServer{
			svc: NewMockService(),
			fs:  fs,
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				return b
			},
		}

		cli := spawnServer(t, gs)
		response, err := cli.Stat(ctx, &proto.StatRequest{
			RepoName:  "therepo",
			CommitSha: "HEAD",
			Path:      []byte("file"),
		})
		require.NoError(t, err)

		if diff := cmp.Diff(&proto.StatResponse{
			FileInfo: gitdomain.FSFileInfoToProto(&expectedStat),
		}, response, cmpopts.IgnoreUnexported(proto.StatResponse{}, proto.FileInfo{})); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}

		b.StatFunc.SetDefaultReturn(nil, &gitdomain.RevisionNotFoundError{})
		_, err = cli.Stat(context.Background(), &v1.StatRequest{
			RepoName:  "therepo",
			CommitSha: "HEAD",
			Path:      []byte("file"),
		})
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RevisionNotFoundPayload{})

		b.StatFunc.SetDefaultReturn(nil, os.ErrNotExist)
		_, err = cli.Stat(context.Background(), &v1.StatRequest{
			RepoName:  "therepo",
			CommitSha: "HEAD",
			Path:      []byte("file"),
		})
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.FileNotFoundPayload{})
	})
}

func TestGRPCServer_ReadDir(t *testing.T) {
	ctx := context.Background()
	mockSS := gitserver.NewMockGitserverService_ReadDirServer()
	mockSS.ContextFunc.SetDefaultReturn(ctx)
	t.Run("argument validation", func(t *testing.T) {
		gs := &grpcServer{}
		err := gs.ReadDir(&v1.ReadDirRequest{RepoName: ""}, mockSS)
		require.ErrorContains(t, err, "repo must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)

		err = gs.ReadDir(&v1.ReadDirRequest{RepoName: "repo"}, mockSS)
		require.ErrorContains(t, err, "commit_sha must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
	})
	t.Run("checks for uncloned repo", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		fs.RepoClonedFunc.SetDefaultReturn(false, nil)
		locker := NewMockRepositoryLocker()
		locker.StatusFunc.SetDefaultReturn("cloning", true)
		gs := &grpcServer{svc: NewMockService(), fs: fs, locker: locker}
		err := gs.ReadDir(&v1.ReadDirRequest{RepoName: "therepo", CommitSha: "HEAD"}, mockSS)
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RepoNotFoundPayload{})
		require.Contains(t, err.Error(), "repo not found")
		mockassert.Called(t, fs.RepoClonedFunc)
		mockassert.Called(t, locker.StatusFunc)
	})
	t.Run("e2e", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		b := git.NewMockGitBackend()
		it := git.NewMockReadDirIterator()
		it.NextFunc.PushReturn(&fileutil.FileInfo{Name_: "file"}, nil)
		it.NextFunc.PushReturn(&fileutil.FileInfo{Name_: "dir/file"}, nil)
		it.NextFunc.PushReturn(nil, io.EOF)
		b.ReadDirFunc.SetDefaultReturn(it, nil)
		gs := &grpcServer{
			svc: NewMockService(),
			fs:  fs,
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				return b
			},
		}

		cli := spawnServer(t, gs)
		cc, err := cli.ReadDir(ctx, &v1.ReadDirRequest{
			RepoName:  "therepo",
			CommitSha: "HEAD",
		})
		require.NoError(t, err)
		fis := []*v1.FileInfo{}
		for {
			resp, err := cc.Recv()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
			fis = append(fis, resp.GetFileInfo()...)
		}
		if diff := cmp.Diff([]*v1.FileInfo{
			{
				Name: []byte("file"),
			},
			{
				Name: []byte("dir/file"),
			},
		}, fis, cmpopts.IgnoreUnexported(v1.FileInfo{})); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}

		it = git.NewMockReadDirIterator()
		it.NextFunc.SetDefaultReturn(nil, &gitdomain.RevisionNotFoundError{})
		it.CloseFunc.SetDefaultReturn(&gitdomain.RevisionNotFoundError{})
		b.ReadDirFunc.SetDefaultReturn(it, nil)
		cc, err = cli.ReadDir(context.Background(), &v1.ReadDirRequest{
			RepoName:  "therepo",
			CommitSha: "HEAD",
		})
		require.NoError(t, err)
		_, err = cc.Recv()
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RevisionNotFoundPayload{})

		it = git.NewMockReadDirIterator()
		it.NextFunc.SetDefaultReturn(nil, os.ErrNotExist)
		it.CloseFunc.SetDefaultReturn(os.ErrNotExist)
		b.ReadDirFunc.SetDefaultReturn(it, nil)
		cc, err = cli.ReadDir(context.Background(), &v1.ReadDirRequest{
			RepoName:  "therepo",
			CommitSha: "HEAD",
		})
		require.NoError(t, err)
		_, err = cc.Recv()
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.FileNotFoundPayload{})
	})
}

func TestGRPCServer_CommitLog(t *testing.T) {
	ctx := context.Background()
	mockSS := gitserver.NewMockGitserverService_CommitLogServer()
	mockSS.ContextFunc.SetDefaultReturn(ctx)
	t.Run("argument validation", func(t *testing.T) {
		gs := &grpcServer{}
		err := gs.CommitLog(&v1.CommitLogRequest{RepoName: ""}, mockSS)
		require.ErrorContains(t, err, "repo must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)

		err = gs.CommitLog(&v1.CommitLogRequest{RepoName: "repo"}, mockSS)
		require.ErrorContains(t, err, "must specify ranges or all_refs")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)

		err = gs.CommitLog(&v1.CommitLogRequest{RepoName: "repo", Ranges: [][]byte{[]byte("range")}, AllRefs: true}, mockSS)
		require.ErrorContains(t, err, "cannot specify both ranges and all_refs")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)

		err = gs.CommitLog(&v1.CommitLogRequest{RepoName: "repo", AllRefs: true, Order: 42}, mockSS)
		require.ErrorContains(t, err, "unknown order")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
	})
	t.Run("checks for uncloned repo", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		fs.RepoClonedFunc.SetDefaultReturn(false, nil)
		locker := NewMockRepositoryLocker()
		locker.StatusFunc.SetDefaultReturn("cloning", true)
		gs := &grpcServer{svc: NewMockService(), fs: fs, locker: locker}
		err := gs.CommitLog(&v1.CommitLogRequest{RepoName: "therepo", AllRefs: true}, mockSS)
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RepoNotFoundPayload{})
		require.Contains(t, err.Error(), "repo not found")
		mockassert.Called(t, fs.RepoClonedFunc)
		mockassert.Called(t, locker.StatusFunc)
	})
	t.Run("e2e", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		b := git.NewMockGitBackend()
		it := git.NewMockCommitLogIterator()
		it.NextFunc.PushReturn(&git.GitCommitWithFiles{Commit: &gitdomain.Commit{
			ID:        "2b2289762392764ed127587b0d5fd88a2f16b7c1",
			Author:    gitdomain.Signature{Name: "Bar Author", Email: "bar@sourcegraph.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
			Committer: &gitdomain.Signature{Name: "c", Email: "c@c.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
			Message:   "bar",
			Parents:   []api.CommitID{"5fab3adc1e398e749749271d14ab843759b192cf"},
		}, ModifiedFiles: []string{"file"}}, nil)
		it.NextFunc.PushReturn(&git.GitCommitWithFiles{Commit: &gitdomain.Commit{
			ID:        "5fab3adc1e398e749749271d14ab843759b192cf",
			Author:    gitdomain.Signature{Name: "Foo Author", Email: "foo@sourcegraph.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Committer: &gitdomain.Signature{Name: "c", Email: "c@c.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Message:   "foo",
			Parents:   nil,
		}, ModifiedFiles: []string{"otherfile"}}, nil)
		it.NextFunc.PushReturn(nil, io.EOF)
		b.CommitLogFunc.SetDefaultReturn(it, nil)
		gs := &grpcServer{
			svc: NewMockService(),
			fs:  fs,
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				return b
			},
		}

		cli := spawnServer(t, gs)
		cc, err := cli.CommitLog(ctx, &v1.CommitLogRequest{
			RepoName: "therepo",
			AllRefs:  true,
		})
		require.NoError(t, err)
		res := []*v1.GetCommitResponse{}
		for {
			resp, err := cc.Recv()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
			res = append(res, resp.GetCommits()...)
		}
		if diff := cmp.Diff([]*v1.GetCommitResponse{
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
		}, res, cmpopts.IgnoreUnexported(v1.GetCommitResponse{}, v1.GitCommit{}, v1.GitSignature{}, timestamppb.Timestamp{})); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}

		it = git.NewMockCommitLogIterator()
		it.NextFunc.SetDefaultReturn(nil, &gitdomain.RevisionNotFoundError{})
		it.CloseFunc.SetDefaultReturn(&gitdomain.RevisionNotFoundError{})
		b.CommitLogFunc.SetDefaultReturn(it, nil)
		cc, err = cli.CommitLog(context.Background(), &v1.CommitLogRequest{
			RepoName: "therepo",
			AllRefs:  true,
		})
		require.NoError(t, err)
		_, err = cc.Recv()
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RevisionNotFoundPayload{})
	})
}

func TestGRPCServer_MergeBaseOctopus(t *testing.T) {
	ctx := context.Background()
	t.Run("argument validation", func(t *testing.T) {
		gs := &grpcServer{}
		_, err := gs.MergeBaseOctopus(ctx, &v1.MergeBaseOctopusRequest{RepoName: ""})
		require.ErrorContains(t, err, "repo must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
		_, err = gs.MergeBaseOctopus(ctx, &v1.MergeBaseOctopusRequest{RepoName: "therepo", Revspecs: [][]byte{}})
		require.ErrorContains(t, err, "at least 2 revspecs must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
		_, err = gs.MergeBaseOctopus(ctx, &v1.MergeBaseOctopusRequest{RepoName: "therepo", Revspecs: [][]byte{[]byte("onlyone")}})
		require.ErrorContains(t, err, "at least 2 revspecs must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
	})
	t.Run("checks for uncloned repo", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		fs.RepoClonedFunc.SetDefaultReturn(false, nil)
		locker := NewMockRepositoryLocker()
		locker.StatusFunc.SetDefaultReturn("cloning", true)
		gs := &grpcServer{svc: NewMockService(), fs: fs, locker: locker}
		_, err := gs.MergeBaseOctopus(ctx, &v1.MergeBaseOctopusRequest{RepoName: "therepo", Revspecs: [][]byte{[]byte("one"), []byte("two"), []byte("three")}})
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RepoNotFoundPayload{})
		require.Contains(t, err.Error(), "repo not found")
		mockassert.Called(t, fs.RepoClonedFunc)
		mockassert.Called(t, locker.StatusFunc)
	})
	t.Run("revision not found", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		gs := &grpcServer{
			svc: NewMockService(),
			fs:  fs,
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				b := git.NewMockGitBackend()
				b.MergeBaseOctopusFunc.SetDefaultReturn("", &gitdomain.RevisionNotFoundError{Repo: "therepo"})
				return b
			},
		}
		_, err := gs.MergeBaseOctopus(ctx, &v1.MergeBaseOctopusRequest{RepoName: "therepo", Revspecs: [][]byte{[]byte("one"), []byte("two"), []byte("three")}})
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RevisionNotFoundPayload{})
		require.Contains(t, err.Error(), "revision not found")
	})
	t.Run("e2e", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		b := git.NewMockGitBackend()
		b.MergeBaseOctopusFunc.SetDefaultReturn("deadbeef", nil)
		gs := &grpcServer{
			svc: NewMockService(),
			fs:  fs,
			gitBackendSource: func(common.GitDir, api.RepoName) git.GitBackend {
				return b
			},
		}

		cli := spawnServer(t, gs)
		res, err := cli.MergeBaseOctopus(ctx, &v1.MergeBaseOctopusRequest{
			RepoName: "therepo",
			Revspecs: [][]byte{[]byte("one"), []byte("two"), []byte("three")},
		})
		require.NoError(t, err)
		if diff := cmp.Diff(&proto.MergeBaseOctopusResponse{
			MergeBaseCommitSha: "deadbeef",
		}, res, cmpopts.IgnoreUnexported(proto.MergeBaseOctopusResponse{})); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}
	})
}

func assertGRPCStatusCode(t *testing.T, err error, want codes.Code) {
	t.Helper()
	s, ok := status.FromError(err)
	require.True(t, ok, "expected status.FromError to succeed")
	require.Equal(t, want, s.Code())
}

func assertHasGRPCErrorDetailOfType(t *testing.T, err error, typ any) {
	t.Helper()
	s, ok := status.FromError(err)
	require.True(t, ok, "expected status.FromError to succeed")
	for _, d := range s.Details() {
		// Compare types of d and typ:
		if reflect.TypeOf(d) == reflect.TypeOf(typ) {
			return
		}
	}
	t.Fatalf("error %v does not implement error detail type %T", err, typ)
}

func spawnServer(t *testing.T, server *grpcServer) proto.GitserverServiceClient {
	t.Helper()
	grpcServer := defaults.NewServer(logtest.Scoped(t))
	proto.RegisterGitserverServiceServer(grpcServer, server)
	handler := internalgrpc.MultiplexHandlers(grpcServer, http.NotFoundHandler())
	srv := httptest.NewServer(handler)
	t.Cleanup(func() {
		srv.Close()
	})

	u, err := url.Parse(srv.URL)
	require.NoError(t, err)

	cc, err := defaults.Dial(u.Host, logtest.Scoped(t))
	require.NoError(t, err)

	return proto.NewGitserverServiceClient(cc)
}

type testChangedFilesIterator struct {
	paths []gitdomain.PathStatus
}

func (t *testChangedFilesIterator) Next() (gitdomain.PathStatus, error) {
	if len(t.paths) == 0 {
		return gitdomain.PathStatus{}, io.EOF
	}
	path := t.paths[0]
	t.paths = t.paths[1:]
	return path, nil
}

func (t *testChangedFilesIterator) Close() error {
	return nil
}

func mustParseTime(layout, value string) time.Time {
	tm, err := time.Parse(layout, value)
	if err != nil {
		panic(err.Error())
	}
	return tm
}
