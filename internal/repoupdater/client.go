package repoupdater

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sync"

	"github.com/sourcegraph/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	proto "github.com/sourcegraph/sourcegraph/internal/repoupdater/v1"
)

var (
	// DefaultClient is the default Client. Unless overwritten, it is
	// connected to the server specified by the REPO_UPDATER_URL
	// environment variable.
	DefaultClient = NewClient(repoUpdaterURLDefault())
)

func repoUpdaterURLDefault() string {
	if u := os.Getenv("REPO_UPDATER_URL"); u != "" {
		return u
	}

	return "http://repo-updater:3182"
}

// Client is a repoupdater client.
type Client struct {
	// grpcClient is a function that lazily creates a grpc client.
	// Any implementation should not recreate the client more than once.
	grpcClient func() (proto.RepoUpdaterServiceClient, error)
}

// NewClient will initiate a new repoupdater Client with the given serverURL.
func NewClient(serverURL string) *Client {
	return &Client{
		grpcClient: sync.OnceValues(func() (proto.RepoUpdaterServiceClient, error) {
			u, err := url.Parse(serverURL)
			if err != nil {
				return nil, err
			}

			l := log.Scoped("repoUpdateGRPCClient")
			conn, err := defaults.Dial(u.Host, l)
			if err != nil {
				return nil, err
			}

			client := &automaticRetryClient{base: proto.NewRepoUpdaterServiceClient(conn)}
			return client, nil
		}),
	}
}

// RepoUpdateSchedulerInfo returns information about the state of the repo in the update scheduler.
func (c *Client) RepoUpdateSchedulerInfo(
	ctx context.Context,
	args protocol.RepoUpdateSchedulerInfoArgs,
) (result *protocol.RepoUpdateSchedulerInfoResult, err error) {
	client, err := c.grpcClient()
	if err != nil {
		return nil, err
	}
	req := &proto.RepoUpdateSchedulerInfoRequest{Id: int32(args.ID)}
	resp, err := client.RepoUpdateSchedulerInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	return protocol.RepoUpdateSchedulerInfoResultFromProto(resp), nil
}

func (c *Client) RecloneRepository(ctx context.Context, repoName api.RepoName) error {
	client, err := c.grpcClient()
	if err != nil {
		return err
	}

	_, err = client.RecloneRepository(ctx, &proto.RecloneRepositoryRequest{
		RepoName: string(repoName),
	})
	return err
}

// MockEnqueueRepoUpdate mocks (*Client).EnqueueRepoUpdate for tests.
var MockEnqueueRepoUpdate func(ctx context.Context, repo api.RepoName) (*protocol.RepoUpdateResponse, error)

// EnqueueRepoUpdate requests that the named repository be updated in the near
// future. It does not wait for the update.
func (c *Client) EnqueueRepoUpdate(ctx context.Context, repo api.RepoName) (*protocol.RepoUpdateResponse, error) {
	if MockEnqueueRepoUpdate != nil {
		return MockEnqueueRepoUpdate(ctx, repo)
	}

	client, err := c.grpcClient()
	if err != nil {
		return nil, err
	}

	req := proto.EnqueueRepoUpdateRequest{Repo: string(repo)}
	resp, err := client.EnqueueRepoUpdate(ctx, &req)
	if err != nil {
		if s, ok := status.FromError(err); ok && s.Code() == codes.NotFound {
			return nil, &repoNotFoundError{repo: string(repo), responseBody: s.Message()}
		}

		return nil, err
	}

	return protocol.RepoUpdateResponseFromProto(resp), nil
}

type repoNotFoundError struct {
	repo         string
	responseBody string
}

func (repoNotFoundError) NotFound() bool { return true }
func (e *repoNotFoundError) Error() string {
	return fmt.Sprintf("repo %v not found with response: %v", e.repo, e.responseBody)
}

// MockEnqueueChangesetSync mocks (*Client).EnqueueChangesetSync for tests.
var MockEnqueueChangesetSync func(ctx context.Context, ids []int64) error

func (c *Client) EnqueueChangesetSync(ctx context.Context, ids []int64) error {
	if MockEnqueueChangesetSync != nil {
		return MockEnqueueChangesetSync(ctx, ids)
	}

	client, err := c.grpcClient()
	if err != nil {
		return err
	}

	// empty response can be ignored
	_, err = client.EnqueueChangesetSync(ctx, &proto.EnqueueChangesetSyncRequest{Ids: ids})
	return err
}
