package gitserver

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/connection"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
)

type RepositoryServiceClient interface {
	DeleteRepository(context.Context, api.RepoName) error
	FetchRepository(context.Context, api.RepoName) (lastFetched, lastChanged time.Time, err error)
}

func NewRepositoryServiceClient() RepositoryServiceClient {
	return &repositoryServiceClient{}
}

type repositoryServiceClient struct{}

func (c *repositoryServiceClient) DeleteRepository(ctx context.Context, repo api.RepoName) error {
	cc, err := c.clientForRepo(ctx, repo)
	if err != nil {
		return err
	}
	_, err = cc.DeleteRepository(ctx, &proto.DeleteRepositoryRequest{
		RepoName: string(repo),
	})
	return err
}

func (c *repositoryServiceClient) FetchRepository(ctx context.Context, repo api.RepoName) (lastFetched, lastChanged time.Time, err error) {
	cc, err := c.clientForRepo(ctx, repo)
	if err != nil {
		return lastFetched, lastChanged, err
	}
	resp, err := cc.FetchRepository(ctx, &proto.FetchRepositoryRequest{
		RepoName: string(repo),
	})
	if err != nil {
		return lastFetched, lastChanged, err
	}
	return resp.GetLastFetched().AsTime(), resp.GetLastChanged().AsTime(), nil
}

func (c *repositoryServiceClient) clientForRepo(ctx context.Context, repo api.RepoName) (proto.GitserverRepositoryServiceClient, error) {
	conn, err := connection.GlobalConns.ConnForRepo(ctx, repo)
	if err != nil {
		return nil, err
	}
	return proto.NewGitserverRepositoryServiceClient(conn), nil
}
