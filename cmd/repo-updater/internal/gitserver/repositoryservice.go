package gitserver

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/connection"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type RepositoryServiceClient interface {
	DeleteRepository(context.Context, api.RepoName) error
	FetchRepository(context.Context, api.RepoName) (lastFetched, lastChanged time.Time, err error)
}

func NewRepositoryServiceClient(scope string) RepositoryServiceClient {
	return &repositoryServiceClient{
		operations: getOperations(),
		scope:      scope,
	}
}

type repositoryServiceClient struct {
	operations *operations
	scope      string
}

func (c *repositoryServiceClient) DeleteRepository(ctx context.Context, repo api.RepoName) (err error) {
	ctx, _, endObservation := c.operations.deleteRepository.With(ctx,
		&err,
		observation.Args{
			Attrs: []attribute.KeyValue{
				repo.Attr(),
			},
			MetricLabelValues: []string{c.scope},
		},
	)
	defer endObservation(1, observation.Args{})

	cc, err := c.clientForRepo(ctx, repo)
	if err != nil {
		return err
	}

	_, err = cc.DeleteRepository(ctx, &proto.DeleteRepositoryRequest{
		RepoName: string(repo),
	}, defaults.RetryPolicy...)

	return err
}

func (c *repositoryServiceClient) FetchRepository(ctx context.Context, repo api.RepoName) (lastFetched, lastChanged time.Time, err error) {
	ctx, _, endObservation := c.operations.fetchRepository.With(ctx,
		&err,
		observation.Args{
			Attrs: []attribute.KeyValue{
				repo.Attr(),
			},
			MetricLabelValues: []string{c.scope},
		},
	)
	defer endObservation(1, observation.Args{})

	cc, err := c.clientForRepo(ctx, repo)
	if err != nil {
		return lastFetched, lastChanged, err
	}

	resp, err := cc.FetchRepository(ctx, &proto.FetchRepositoryRequest{
		RepoName: string(repo),
	}, defaults.RetryPolicy...)
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
