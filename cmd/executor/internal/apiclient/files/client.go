package files

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/files"
	"github.com/sourcegraph/sourcegraph/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Client interacts with the files store.
type Client struct {
	client     *apiclient.BaseClient
	logger     log.Logger
	operations *operations
}

// Compile time validation.
var _ files.Store = &Client{}

// New creates a new Client based on the provided Options.
func New(observationCtx *observation.Context, options apiclient.BaseClientOptions) (*Client, error) {
	logger := log.Scoped("executor-api-files-client")
	client, err := apiclient.NewBaseClient(logger, options)
	if err != nil {
		return nil, err
	}
	return &Client{
		client:     client,
		logger:     logger,
		operations: newOperations(observationCtx),
	}, nil
}

func (c *Client) Exists(ctx context.Context, job types.Job, bucket string, key string) (exists bool, err error) {
	ctx, _, endObservation := c.operations.exists.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("bucket", bucket),
		attribute.String("key", key),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.client.NewRequest(job.ID, job.Token, http.MethodHead, fmt.Sprintf("%s/%s", bucket, key), nil)
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

func (c *Client) Get(ctx context.Context, job types.Job, bucket string, key string) (content io.ReadCloser, err error) {
	ctx, _, endObservation := c.operations.get.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("bucket", bucket),
		attribute.String("key", key),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.client.NewRequest(job.ID, job.Token, http.MethodGet, fmt.Sprintf("%s/%s", bucket, key), nil)
	if err != nil {
		return nil, err
	}

	_, body, err := c.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	return body, nil
}
