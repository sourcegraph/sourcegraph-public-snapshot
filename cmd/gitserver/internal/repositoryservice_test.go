package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	mockassert "github.com/derision-test/go-mockgen/v2/testutil/assert"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
)

func TestRepositoryServiceServer_DeleteRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("argument validation", func(t *testing.T) {
		gs := &repositoryServiceServer{}
		_, err := gs.DeleteRepository(ctx, &proto.DeleteRepositoryRequest{RepoName: ""})
		require.ErrorContains(t, err, "repo_name must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
	})

	t.Run("checks for uncloned repo", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		fs.RepoClonedFunc.SetDefaultReturn(false, nil)
		gs := &repositoryServiceServer{svc: NewMockService(), fs: fs}
		_, err := gs.DeleteRepository(ctx, &proto.DeleteRepositoryRequest{RepoName: "therepo"})
		require.Error(t, err)
		assertGRPCStatusCode(t, err, codes.NotFound)
		assertHasGRPCErrorDetailOfType(t, err, &proto.RepoNotFoundPayload{})
		require.Contains(t, err.Error(), "repo not found")
		mockassert.Called(t, fs.RepoClonedFunc)
	})

	t.Run("e2e", func(t *testing.T) {
		fs := gitserverfs.NewMockFS()
		// Repo is cloned, proceed!
		fs.RepoClonedFunc.SetDefaultReturn(true, nil)
		db := dbmocks.NewMockDB()
		db.GitserverReposFunc.SetDefaultReturn(dbmocks.NewMockGitserverRepoStore())
		rs := &repositoryServiceServer{
			svc:    NewMockService(),
			fs:     fs,
			db:     db,
			logger: logtest.NoOp(t),
		}

		cli := spawnRepositoryServer(t, rs)
		_, err := cli.DeleteRepository(ctx, &proto.DeleteRepositoryRequest{
			RepoName: "therepo",
		})
		require.NoError(t, err)
		mockassert.Called(t, fs.RemoveRepoFunc)
	})
}

func TestRepositoryServiceServer_FetchRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("argument validation", func(t *testing.T) {
		gs := &repositoryServiceServer{}
		_, err := gs.FetchRepository(ctx, &proto.FetchRepositoryRequest{RepoName: ""})
		require.ErrorContains(t, err, "repo_name must be specified")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
	})

	t.Run("e2e", func(t *testing.T) {
		svc := NewMockService()
		lastFetched := time.Unix(1, 0).UTC()
		lastChanged := time.Unix(2, 0).UTC()
		svc.FetchRepositoryFunc.SetDefaultReturn(lastFetched, lastChanged, nil)
		rs := &repositoryServiceServer{
			svc: svc,
			fs:  gitserverfs.NewMockFS(),
			db:  dbmocks.NewMockDB(),
		}

		cli := spawnRepositoryServer(t, rs)
		response, err := cli.FetchRepository(ctx, &proto.FetchRepositoryRequest{
			RepoName: "therepo",
		})
		require.NoError(t, err)
		mockassert.Called(t, svc.FetchRepositoryFunc)
		require.Equal(t, lastFetched, response.LastFetched.AsTime())
		require.Equal(t, lastChanged, response.LastChanged.AsTime())
	})
}

func spawnRepositoryServer(t *testing.T, server *repositoryServiceServer) proto.GitserverRepositoryServiceClient {
	t.Helper()
	grpcServer := defaults.NewServer(logtest.Scoped(t))
	proto.RegisterGitserverRepositoryServiceServer(grpcServer, server)
	handler := internalgrpc.MultiplexHandlers(grpcServer, http.NotFoundHandler())
	srv := httptest.NewServer(handler)
	t.Cleanup(func() {
		srv.Close()
	})

	u, err := url.Parse(srv.URL)
	require.NoError(t, err)

	cc, err := defaults.Dial(u.Host, logtest.Scoped(t))
	require.NoError(t, err)

	return proto.NewGitserverRepositoryServiceClient(cc)
}
