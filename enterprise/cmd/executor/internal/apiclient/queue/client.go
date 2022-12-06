package queue

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

	// TelemetryOptions captures additional parameters sent in heartbeat requests.
	TelemetryOptions TelemetryOptions

	// ResourceOptions inform the frontend how large of a VM the job will be executed in.
	// This can be used to replace magic variables in the job payload indicating how much
	// the task should be able to comfortably consume.
	ResourceOptions ResourceOptions
}

type ResourceOptions struct {
	// NumCPUs is the number of virtual CPUs a job can safely utilize.
	NumCPUs int

	// Memory is the maximum amount of memory a job can safely utilize.
	Memory string

	// DiskSpace is the maximum amount of disk a job can safely utilize.
	DiskSpace string
}

func New(observationCtx *observation.Context, options Options, metricsGatherer prometheus.Gatherer) (*Client, error) {
	client, err := apiclient.NewBaseClient(options.BaseClientOptions)
	if err != nil {
		return nil, err
	}
	return &Client{
		options:         options,
		client:          client,
		logger:          log.Scoped("executor-api-queue-client", "The API client adapter for executors to use dbworkers over HTTP"),
		metricsGatherer: metricsGatherer,
		operations:      newOperations(observationCtx),
	}, nil
}

func (c *Client) Dequeue(ctx context.Context, queueName string, job *executor.Job) (_ bool, err error) {
	ctx, _, endObservation := c.operations.dequeue.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("queueName", queueName),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.client.NewJSONRequest(http.MethodPost, fmt.Sprintf("%s/dequeue", queueName), executor.DequeueRequest{
		Version:      version.Version(),
		ExecutorName: c.options.ExecutorName,
		NumCPUs:      c.options.ResourceOptions.NumCPUs,
		Memory:       c.options.ResourceOptions.Memory,
		DiskSpace:    c.options.ResourceOptions.DiskSpace,
	})
	if err != nil {
		return false, err
	}

	return c.client.DoAndDecode(ctx, req, &job)
}

func (c *Client) AddExecutionLogEntry(ctx context.Context, queueName string, jobID int, entry workerutil.ExecutionLogEntry) (entryID int, err error) {
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

func (c *Client) UpdateExecutionLogEntry(ctx context.Context, queueName string, jobID, entryID int, entry workerutil.ExecutionLogEntry) (err error) {
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

func (c *Client) MarkComplete(ctx context.Context, queueName string, jobID int) (err error) {
	ctx, _, endObservation := c.operations.markComplete.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("queueName", queueName),
		otlog.Int("jobID", jobID),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.client.NewJSONRequest(http.MethodPost, fmt.Sprintf("%s/markComplete", queueName), executor.MarkCompleteRequest{
		ExecutorName: c.options.ExecutorName,
		JobID:        jobID,
	})
	if err != nil {
		return err
	}

	return c.client.DoAndDrop(ctx, req)
}

func (c *Client) MarkErrored(ctx context.Context, queueName string, jobID int, errorMessage string) (err error) {
	ctx, _, endObservation := c.operations.markErrored.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("queueName", queueName),
		otlog.Int("jobID", jobID),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.client.NewJSONRequest(http.MethodPost, fmt.Sprintf("%s/markErrored", queueName), executor.MarkErroredRequest{
		ExecutorName: c.options.ExecutorName,
		JobID:        jobID,
		ErrorMessage: errorMessage,
	})
	if err != nil {
		return err
	}

	return c.client.DoAndDrop(ctx, req)
}

func (c *Client) MarkFailed(ctx context.Context, queueName string, jobID int, errorMessage string) (err error) {
	ctx, _, endObservation := c.operations.markFailed.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("queueName", queueName),
		otlog.Int("jobID", jobID),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.client.NewJSONRequest(http.MethodPost, fmt.Sprintf("%s/markFailed", queueName), executor.MarkErroredRequest{
		ExecutorName: c.options.ExecutorName,
		JobID:        jobID,
		ErrorMessage: errorMessage,
	})
	if err != nil {
		return err
	}

	return c.client.DoAndDrop(ctx, req)
}

// TODO: Remove this in Sourcegraph 4.4.
func (c *Client) CanceledJobs(ctx context.Context, queueName string, knownIDs []int) (canceledIDs []int, err error) {
	req, err := c.client.NewJSONRequest(http.MethodPost, fmt.Sprintf("%s/canceledJobs", queueName), executor.CanceledJobsRequest{
		KnownJobIDs:  knownIDs,
		ExecutorName: c.options.ExecutorName,
	})
	if err != nil {
		return nil, err
	}

	if _, err := c.client.DoAndDecode(ctx, req, &canceledIDs); err != nil {
		return nil, err
	}

	return canceledIDs, nil
}

func (c *Client) Ping(ctx context.Context, queueName string, jobIDs []int) (err error) {
	req, err := c.client.NewJSONRequest(http.MethodPost, fmt.Sprintf("%s/heartbeat", queueName), executor.HeartbeatRequest{
		ExecutorName: c.options.ExecutorName,
	})
	if err != nil {
		return err
	}

	return c.client.DoAndDrop(ctx, req)
}

func (c *Client) Heartbeat(ctx context.Context, queueName string, jobIDs []int) (knownIDs, cancelIDs []int, err error) {
	ctx, _, endObservation := c.operations.heartbeat.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("queueName", queueName),
		otlog.String("jobIDs", intsToString(jobIDs)),
	}})
	defer endObservation(1, observation.Args{})

	metrics, err := gatherMetrics(c.logger, c.metricsGatherer)
	if err != nil {
		c.logger.Error("Failed to collect prometheus metrics for heartbeat", log.Error(err))
		// Continue, no metric errors should prevent heartbeats.
	}

	req, err := c.client.NewJSONRequest(http.MethodPost, fmt.Sprintf("%s/heartbeat", queueName), executor.HeartbeatRequest{
		// Request the new-fashioned payload.
		Version: executor.ExecutorAPIVersion2,

		ExecutorName: c.options.ExecutorName,
		JobIDs:       jobIDs,

		OS:              c.options.TelemetryOptions.OS,
		Architecture:    c.options.TelemetryOptions.Architecture,
		DockerVersion:   c.options.TelemetryOptions.DockerVersion,
		ExecutorVersion: c.options.TelemetryOptions.ExecutorVersion,
		GitVersion:      c.options.TelemetryOptions.GitVersion,
		IgniteVersion:   c.options.TelemetryOptions.IgniteVersion,
		SrcCliVersion:   c.options.TelemetryOptions.SrcCliVersion,

		PrometheusMetrics: metrics,
	})
	if err != nil {
		return nil, nil, err
	}

	// Do the request and get the reader for the response body.
	_, body, err := c.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	// Now read the response body into a buffer, so that we can decode it twice.
	// This will always be small, so no problem that we don't stream this.
	defer body.Close()
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, nil, err
	}

	// First, try to unmarshal the response into a V2 response object.
	var respV2 executor.HeartbeatResponse
	if err := json.Unmarshal(bodyBytes, &respV2); err == nil {
		// If that works, we can return the data.
		return respV2.KnownIDs, respV2.CancelIDs, nil
	}

	// If unmarshalling fails, try to parse it as a V1 payload.
	var respV1 []int
	if err := json.Unmarshal(bodyBytes, &respV1); err != nil {
		return nil, nil, err
	}

	// If that works, we also have to fetch canceled jobs separately, as we
	// are talking to a pre-4.3 Sourcegraph API and that doesn't return canceled
	// jobs as part of heartbeats.

	cancelIDs, err = c.CanceledJobs(ctx, queueName, jobIDs)
	if err != nil {
		return nil, nil, err
	}

	return respV1, cancelIDs, nil
}

func intsToString(ints []int) string {
	segments := make([]string, 0, len(ints))
	for _, id := range ints {
		segments = append(segments, strconv.Itoa(id))
	}

	return strings.Join(segments, ", ")
}

func gatherMetrics(logger log.Logger, gatherer prometheus.Gatherer) (string, error) {
	maxDuration := 3 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), maxDuration)
	defer cancel()
	go func() {
		<-ctx.Done()
		if ctx.Err() == context.DeadlineExceeded {
			logger.Warn("gathering metrics took longer than expected", log.Duration("maxDuration", maxDuration))
		}
	}()
	mfs, err := gatherer.Gather()
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	enc := expfmt.NewEncoder(&buf, expfmt.FmtText)
	for _, mf := range mfs {
		if err := enc.Encode(mf); err != nil {
			return "", errors.Wrap(err, "encoding metric family")
		}
	}
	return buf.String(), nil
}
