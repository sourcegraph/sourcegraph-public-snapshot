package files

import (
	"context"
	"fmt"
	"io"
	"net/http"

	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Client interacts with the files store.
type Client struct {
	client     *apiclient.BaseClient
	logger     log.Logger
	operations *operations
}

// New creates a new Client based on the provided Options.
func New(options apiclient.BaseClientOptions, observationContext *observation.Context) (*Client, error) {
	client, err := apiclient.NewBaseClient(options)
	if err != nil {
		return nil, err
	}
	return &Client{
		client:     client,
		logger:     log.Scoped("executor-api-files-client", "The API client adapter for executors to interact with the Files over HTTP"),
		operations: newOperations(observationContext),
	}, nil
}

func (c *Client) Exists(ctx context.Context, bucket string, key string) (exists bool, err error) {
	ctx, _, endObservation := c.operations.exists.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("bucket", bucket),
		otlog.String("key", key),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.client.NewRequest(http.MethodHead, fmt.Sprintf("%s/%s", bucket, key), nil)
	if err != nil {
		return false, err
	}

	if err = c.client.DoAndDrop(ctx, req); err != nil {
		var unexpectedStatusCodeError *apiclient.UnexpectedStatusCodeErr
		if errors.As(err, &unexpectedStatusCodeError) {
			if unexpectedStatusCodeError.StatusCode == http.StatusNotFound {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func (c *Client) Get(ctx context.Context, bucket string, key string) (content io.ReadCloser, err error) {
	ctx, _, endObservation := c.operations.get.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("bucket", bucket),
		otlog.String("key", key),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.client.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", bucket, key), nil)
	if err != nil {
		return nil, err
	}

	_, body, err := c.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	return body, nil
}
