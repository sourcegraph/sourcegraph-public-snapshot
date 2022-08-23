package batches

import (
	"context"
	"io"
	"net/http"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Client is the client used to communicate with a remote job queue API.
type Client struct {
	options    Options
	client     *apiclient.BaseClient
	operations *operations
}

type Options struct {
	// BaseClientOptions are the underlying HTTP client options.
	BaseClientOptions apiclient.BaseClientOptions
}

func New(options Options, observationContext *observation.Context) *Client {
	return &Client{
		options:    options,
		client:     apiclient.NewBaseClient(options.BaseClientOptions),
		operations: newOperations(observationContext),
	}
}

func (c *Client) Get(ctx context.Context, path string) (body io.ReadCloser, err error) {
	ctx, _, endObservation := c.operations.get.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("path", path),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	_, resBody, err := c.client.Do(ctx, req)
	return resBody, err
}
