package repoupdater

import (
	"context"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
)

type Client struct {
	operations *operations
}

func New(observationContext *observation.Context) *Client {
	return &Client{
		operations: newOperations(observationContext),
	}
}

func (c *Client) RepoLookup(ctx context.Context, name api.RepoName) (repo *protocol.RepoInfo, err error) {
	ctx, _, endObservation := c.operations.repoLookup.With(ctx, &err, observation.Args{LogFields: []log.Field{}})
	defer func() {
		var logFields []log.Field
		if repo != nil {
			logFields = []log.Field{log.Int("repoID", int(repo.ID))}
		}
		endObservation(1, observation.Args{LogFields: logFields})
	}()

	result, err := repoupdater.DefaultClient.RepoLookup(ctx, protocol.RepoLookupArgs{Repo: name})
	if err != nil {
		return nil, err
	}

	return result.Repo, nil
}

func (c *Client) EnqueueRepoUpdate(ctx context.Context, name api.RepoName) (resp *protocol.RepoUpdateResponse, err error) {
	ctx, _, endObservation := c.operations.enqueueRepoUpdate.With(ctx, &err, observation.Args{LogFields: []log.Field{}})
	defer func() {
		var logFields []log.Field
		if resp != nil {
			logFields = []log.Field{log.Int("repoID", int(resp.ID))}
		}
		endObservation(1, observation.Args{LogFields: logFields})
	}()

	resp, err = repoupdater.DefaultClient.EnqueueRepoUpdate(ctx, name)
	return
}
