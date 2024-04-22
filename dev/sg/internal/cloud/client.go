package cloud

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	cloudapiv1 "github.com/sourcegraph/cloud-api/go/cloudapi/v1"
	"github.com/sourcegraph/cloud-api/go/cloudapi/v1/cloudapiv1connect"
	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const HeaderUserToken = "X-GCP-User-Token"
const APIEndpoint = "https://cloud-ops-dev.sgdev.org/api"

type Client struct {
	client cloudapiv1connect.InstanceServiceClient
	token  string
}

func NewClient(ctx context.Context, endpoint string) (*Client, error) {
	client := cloudapiv1connect.NewInstanceServiceClient(
		http.DefaultClient,
		endpoint,
	)

	token, err := run.Cmd(ctx, "gcloud auth print-identity-token").Run().String()
	if err != nil {
		return nil, errors.Newf("failed to get gcloud auth token: %v", err)
	}

	return &Client{client: client, token: token}, nil
}

func newRequestWithToken[T any](token string, message *T) *connect.Request[T] {
	req := connect.NewRequest(message)
	req.Header().Add(HeaderUserToken, token)
	return req
}

func (c *Client) ListInstances(ctx context.Context) ([]*cloudapiv1.Instance, error) {
	req := newRequestWithToken(c.token, &cloudapiv1.ListInstancesRequest{})
	res, err := c.client.ListInstances(
		ctx,
		req,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list instances")
	}

	return res.Msg.Instances, nil
}
