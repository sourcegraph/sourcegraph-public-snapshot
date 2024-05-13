package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	mockassert "github.com/derision-test/go-mockgen/v2/testutil/assert"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/internal/api"
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

func TestRepositoryServiceServer_ListRepositories(t *testing.T) {
	ctx := context.Background()

	t.Run("argument validation", func(t *testing.T) {
		gs := &repositoryServiceServer{}
		_, err := gs.ListRepositories(ctx, &proto.ListRepositoriesRequest{PageSize: 0})
		require.ErrorContains(t, err, "page_size must be > 0")
		assertGRPCStatusCode(t, err, codes.InvalidArgument)
	})

	t.Run("e2e", func(t *testing.T) {
		tts := []struct {
			repos    []string
			pageSize uint32
		}{
			// No repos
			{pageSize: 10},
			// 1 repo, 1 not full page
			{repos: []string{"repo1"}, pageSize: 10},
			// 10 repos, 1 full page
			{repos: []string{"repo1", "repo2", "repo3", "repo4", "repo5", "repo6", "repo7", "repo8", "repo9", "repo10"}, pageSize: 10},
			// 10 repos, 10 full pages
			{repos: []string{"repo1", "repo2", "repo3", "repo4", "repo5", "repo6", "repo7", "repo8", "repo9", "repo10"}, pageSize: 1},
		}
		for i, tt := range tts {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				fs := gitserverfs.NewMockFS()
				fs.ForEachRepoFunc.SetDefaultHook(func(cb func(api.RepoName, common.GitDir) bool) error {
					for _, repo := range tt.repos {
						if cb(api.RepoName(repo), common.GitDir("/data/repos/"+repo)) {
							break
						}
					}

					return nil
				})

				rs := &repositoryServiceServer{
					fs: fs,
				}

				cli := spawnRepositoryServer(t, rs)
				var haveRepos []string
				var pageToken string
				for {
					response, err := cli.ListRepositories(ctx, &proto.ListRepositoriesRequest{
						PageSize:  tt.pageSize,
						PageToken: pageToken,
					})
					require.NoError(t, err)
					for _, r := range response.GetRepositories() {
						haveRepos = append(haveRepos, r.Name)
					}
					if response.GetNextPageToken() == "" {
						break
					}
					pageToken = response.GetNextPageToken()
				}
				require.Equal(t, tt.repos, haveRepos)
			})
		}
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
