package job

import (
	"context"
	"fmt"
	"net/http"

	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	internalexecutor "github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Client is the client used to communicate with a remote job queue API.
type Client struct {
	options         Options
	client          *apiclient.BaseClient
	logger          log.Logger
	metricsGatherer prometheus.Gatherer
	operations      *operations
}

type Options struct {
	// ExecutorName is a unique identifier for the requesting executor.
	ExecutorName string

	// BaseClientOptions are the underlying HTTP client options.
	BaseClientOptions apiclient.BaseClientOptions
}

func New(observationCtx *observation.Context, options Options, metricsGatherer prometheus.Gatherer) (*Client, error) {
	client, err := apiclient.NewBaseClient(options.BaseClientOptions)
	if err != nil {
		return nil, err
	}
	return &Client{
		options:         options,
		client:          client,
		logger:          log.Scoped("executor-api-queue-job-client", "The API client adapter for executors to use dbworkers over HTTP"),
		metricsGatherer: metricsGatherer,
		operations:      newOperations(observationCtx),
	}, nil
}

func (c *Client) AddExecutionLogEntry(ctx context.Context, queueName string, jobID int, entry internalexecutor.ExecutionLogEntry) (entryID int, err error) {
	ctx, _, endObservation := c.operations.addExecutionLogEntry.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("queueName", queueName),
		otlog.Int("jobID", jobID),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.client.NewJSONRequest(http.MethodPost, fmt.Sprintf("%s/addExecutionLogEntry", queueName), executor.AddExecutionLogEntryRequest{
		ExecutorName:      c.options.ExecutorName,
		JobID:             jobID,
		ExecutionLogEntry: entry,
	})
	if err != nil {
		return entryID, err
	}

	_, err = c.client.DoAndDecode(ctx, req, &entryID)
	return entryID, err
}

func (c *Client) UpdateExecutionLogEntry(ctx context.Context, queueName string, jobID, entryID int, entry internalexecutor.ExecutionLogEntry) (err error) {
	ctx, _, endObservation := c.operations.updateExecutionLogEntry.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("queueName", queueName),
		otlog.Int("jobID", jobID),
		otlog.Int("entryID", entryID),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.client.NewJSONRequest(http.MethodPost, fmt.Sprintf("%s/updateExecutionLogEntry", queueName), executor.UpdateExecutionLogEntryRequest{
		ExecutorName:      c.options.ExecutorName,
		JobID:             jobID,
		EntryID:           entryID,
		ExecutionLogEntry: entry,
	})
	if err != nil {
		return err
	}

	return c.client.DoAndDrop(ctx, req)
}
