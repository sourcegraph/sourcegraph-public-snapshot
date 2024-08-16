package queue

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/cmdlogger"
	internalexecutor "github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/executor/types"
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

// Compile time validation.
var _ workerutil.Store[types.Job] = &Client{}
var _ cmdlogger.ExecutionLogEntryStore = &Client{}

func New(observationCtx *observation.Context, options Options, metricsGatherer prometheus.Gatherer) (*Client, error) {
	logger := log.Scoped("executor-api-queue-client")
	client, err := apiclient.NewBaseClient(logger, options.BaseClientOptions)
	if err != nil {
		return nil, err
	}
	return &Client{
		options:         options,
		client:          client,
		logger:          logger,
		metricsGatherer: metricsGatherer,
		operations:      newOperations(observationCtx),
	}, nil
}

func (c *Client) QueuedCount(ctx context.Context) (int, error) {
	return 0, errors.New("unimplemented")
}

func (c *Client) Dequeue(ctx context.Context, workerHostname string, extraArguments any) (job types.Job, _ bool, err error) {
	var queueAttr attribute.KeyValue
	var endpoint string
	dequeueRequest := types.DequeueRequest{
		Version:      version.Version(),
		ExecutorName: c.options.ExecutorName,
		NumCPUs:      c.options.ResourceOptions.NumCPUs,
		Memory:       c.options.ResourceOptions.Memory,
		DiskSpace:    c.options.ResourceOptions.DiskSpace,
	}

	if len(c.options.QueueNames) > 0 {
		queueAttr = attribute.StringSlice("queueNames", c.options.QueueNames)
		endpoint = "/dequeue"
		dequeueRequest.Queues = c.options.QueueNames
	} else {
		queueAttr = attribute.String("queueName", c.options.QueueName)
		endpoint = fmt.Sprintf("%s/dequeue", c.options.QueueName)
	}

	ctx, _, endObservation := c.operations.dequeue.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		queueAttr,
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.client.NewJSONRequest(http.MethodPost, endpoint, dequeueRequest)
	if err != nil {
		return job, false, err
	}

	decoded, err := c.client.DoAndDecode(ctx, req, &job)
	return job, decoded, err
}

func (c *Client) MarkComplete(ctx context.Context, job types.Job) (_ bool, err error) {
	queue := c.inferQueueName(job)
	ctx, _, endObservation := c.operations.markComplete.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("queueName", queue),
		attribute.Int("jobID", job.ID),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.client.NewJSONJobRequest(job.ID, http.MethodPost, fmt.Sprintf("%s/markComplete", queue), job.Token, types.MarkCompleteRequest{
		JobOperationRequest: types.JobOperationRequest{
			ExecutorName: c.options.ExecutorName,
			JobID:        job.ID,
		},
	})
	if err != nil {
		return false, err
	}

	if err = c.client.DoAndDrop(ctx, req); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Client) MarkErrored(ctx context.Context, job types.Job, failureMessage string) (_ bool, err error) {
	queue := c.inferQueueName(job)
	ctx, _, endObservation := c.operations.markErrored.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("queueName", queue),
		attribute.Int("jobID", job.ID),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.client.NewJSONJobRequest(job.ID, http.MethodPost, fmt.Sprintf("%s/markErrored", queue), job.Token, types.MarkErroredRequest{
		JobOperationRequest: types.JobOperationRequest{
			ExecutorName: c.options.ExecutorName,
			JobID:        job.ID,
		},
		ErrorMessage: failureMessage,
	})
	if err != nil {
		return false, err
	}

	if err = c.client.DoAndDrop(ctx, req); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Client) MarkFailed(ctx context.Context, job types.Job, failureMessage string) (_ bool, err error) {
	queue := c.inferQueueName(job)
	ctx, _, endObservation := c.operations.markFailed.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("queueName", queue),
		attribute.Int("jobID", job.ID),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.client.NewJSONJobRequest(job.ID, http.MethodPost, fmt.Sprintf("%s/markFailed", queue), job.Token, types.MarkErroredRequest{
		JobOperationRequest: types.JobOperationRequest{
			ExecutorName: c.options.ExecutorName,
			JobID:        job.ID,
		},
		ErrorMessage: failureMessage,
	})
	if err != nil {
		return false, err
	}

	if err = c.client.DoAndDrop(ctx, req); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Client) Heartbeat(ctx context.Context, jobIDs []string) (knownIDs, cancelIDs []string, err error) {
	metrics, err := gatherMetrics(c.logger, c.metricsGatherer)
	if err != nil {
		c.logger.Error("Failed to collect prometheus metrics for heartbeat", log.Error(err))
		// Continue, no metric errors should prevent heartbeats.
	}

	var queueAttr attribute.KeyValue
	var endpoint string
	var payload any
	// We are using the newer multi-queue API. It is safe to send jobIds as strings in that case.
	if len(c.options.QueueNames) > 0 {
		queueAttr = attribute.StringSlice("queueNames", c.options.QueueNames)
		queueJobIDs, parseErr := ParseJobIDs(jobIDs)
		if parseErr != nil {
			c.logger.Error("failed to parse job IDs", log.Error(parseErr))
			return nil, nil, err
		}
		endpoint = "/heartbeat"
		payload = types.HeartbeatRequest{
			ExecutorName:      c.options.ExecutorName,
			QueueNames:        c.options.QueueNames,
			JobIDsByQueue:     queueJobIDs,
			OS:                c.options.TelemetryOptions.OS,
			Architecture:      c.options.TelemetryOptions.Architecture,
			DockerVersion:     c.options.TelemetryOptions.DockerVersion,
			ExecutorVersion:   c.options.TelemetryOptions.ExecutorVersion,
			GitVersion:        c.options.TelemetryOptions.GitVersion,
			IgniteVersion:     c.options.TelemetryOptions.IgniteVersion,
			SrcCliVersion:     c.options.TelemetryOptions.SrcCliVersion,
			PrometheusMetrics: metrics,
		}
	} else {
		// If queueName is set, then we cannot be sure whether Sourcegraph is new enough (since Heartbeat can't provide
		// that context). So to be safe, we send jobIds as ints. If Sourcegraph is older, it expects ints anyway. If
		// it is newer, it knows how to convert the values to strings.
		// TODO remove in Sourcegraph 5.2.
		var jobIDsInt []int
		for _, jobID := range jobIDs {
			jobIDInt, convErr := strconv.Atoi(jobID)
			if convErr != nil {
				c.logger.Error("failed to convert job ID to int", log.String("jobID", jobID), log.Error(convErr))
				return nil, nil, err
			}
			jobIDsInt = append(jobIDsInt, jobIDInt)
		}

		queueAttr = attribute.String("queueName", c.options.QueueName)
		endpoint = fmt.Sprintf("%s/heartbeat", c.options.QueueName)
		payload = types.HeartbeatRequestV1{
			// TODO: This field is set to become unnecessary in Sourcegraph 5.2.
			Version:      types.ExecutorAPIVersion2,
			ExecutorName: c.options.ExecutorName,
			JobIDs:       jobIDsInt,

			OS:              c.options.TelemetryOptions.OS,
			Architecture:    c.options.TelemetryOptions.Architecture,
			DockerVersion:   c.options.TelemetryOptions.DockerVersion,
			ExecutorVersion: c.options.TelemetryOptions.ExecutorVersion,
			GitVersion:      c.options.TelemetryOptions.GitVersion,
			IgniteVersion:   c.options.TelemetryOptions.IgniteVersion,
			SrcCliVersion:   c.options.TelemetryOptions.SrcCliVersion,

			PrometheusMetrics: metrics,
		}
	}

	ctx, _, endObservation := c.operations.heartbeat.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		queueAttr,
		attribute.StringSlice("jobIDs", jobIDs),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.client.NewJSONRequest(http.MethodPost, endpoint, payload)

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
	var resp types.HeartbeatResponse
	if unmarshalErr := json.Unmarshal(bodyBytes, &resp); unmarshalErr != nil {
		return nil, nil, unmarshalErr
	}
	return resp.KnownIDs, resp.CancelIDs, nil
}

type JobIDsParseError struct {
	JobIDs []string
}

func (e JobIDsParseError) Error() string {
	return fmt.Sprintf("failed to parse one or more unexpected job ID formats: %s", strings.Join(e.JobIDs, ", "))
}

// ParseJobIDs attempts to split the job IDs on a separator character in order to categorize them by queue
// name, returning a list of types.QueueJobIDs.
// The expected format is <job id>-<queue name>, e.g. "42-batches".
func ParseJobIDs(jobIDs []string) ([]types.QueueJobIDs, error) {
	var queueJobIDs []types.QueueJobIDs
	queueIds := map[string][]string{}
	var invalidIds []string

	for _, stringID := range jobIDs {
		id, queueName, found := strings.Cut(stringID, "-")
		if !found {
			invalidIds = append(invalidIds, stringID)
		} else {
			queueIds[queueName] = append(queueIds[queueName], id)
		}
	}
	if len(invalidIds) > 0 {
		return nil, JobIDsParseError{JobIDs: invalidIds}
	}

	for q, ids := range queueIds {
		queueJobIDs = append(queueJobIDs, types.QueueJobIDs{QueueName: q, JobIDs: ids})
	}
	sort.Slice(queueJobIDs, func(i, j int) bool {
		return queueJobIDs[i].QueueName < queueJobIDs[j].QueueName
	})
	return queueJobIDs, nil
}

func gatherMetrics(logger log.Logger, gatherer prometheus.Gatherer) (string, error) {
	maxDuration := 3 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), maxDuration)
	defer cancel()
	context.AfterFunc(ctx, func() {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Warn("gathering metrics took longer than expected", log.Duration("maxDuration", maxDuration))
		}
	})
	mfs, err := gatherer.Gather()
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	enc := expfmt.NewEncoder(&buf, expfmt.NewFormat(expfmt.TypeTextPlain))
	for _, mf := range mfs {
		if err = enc.Encode(mf); err != nil {
			return "", errors.Wrap(err, "encoding metric family")
		}
	}
	return buf.String(), nil
}

func (c *Client) Ping(ctx context.Context) (err error) {
	var req *http.Request
	if len(c.options.QueueNames) > 0 {
		req, err = c.client.NewJSONRequest(http.MethodPost, "/heartbeat", types.HeartbeatRequest{
			ExecutorName: c.options.ExecutorName,
			QueueNames:   c.options.QueueNames,
		})
	} else {
		req, err = c.client.NewJSONRequest(http.MethodPost, fmt.Sprintf("%s/heartbeat", c.options.QueueName), types.HeartbeatRequest{
			ExecutorName: c.options.ExecutorName,
		})
	}
	if err != nil {
		return err
	}

	return c.client.DoAndDrop(ctx, req)
}

func (c *Client) AddExecutionLogEntry(ctx context.Context, job types.Job, entry internalexecutor.ExecutionLogEntry) (entryID int, err error) {
	queue := c.inferQueueName(job)

	ctx, _, endObservation := c.operations.addExecutionLogEntry.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("queueName", queue),
		attribute.Int("jobID", job.ID),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.client.NewJSONJobRequest(job.ID, http.MethodPost, fmt.Sprintf("%s/addExecutionLogEntry", queue), job.Token, types.AddExecutionLogEntryRequest{
		JobOperationRequest: types.JobOperationRequest{
			ExecutorName: c.options.ExecutorName,
			JobID:        job.ID,
		},
		ExecutionLogEntry: entry,
	})
	if err != nil {
		return entryID, err
	}

	_, err = c.client.DoAndDecode(ctx, req, &entryID)
	return entryID, err
}

func (c *Client) UpdateExecutionLogEntry(ctx context.Context, job types.Job, entryID int, entry internalexecutor.ExecutionLogEntry) (err error) {
	queue := c.inferQueueName(job)

	ctx, _, endObservation := c.operations.updateExecutionLogEntry.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("queueName", queue),
		attribute.Int("jobID", job.ID),
		attribute.Int("entryID", entryID),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.client.NewJSONJobRequest(job.ID, http.MethodPost, fmt.Sprintf("%s/updateExecutionLogEntry", queue), job.Token, types.UpdateExecutionLogEntryRequest{
		JobOperationRequest: types.JobOperationRequest{
			ExecutorName: c.options.ExecutorName,
			JobID:        job.ID,
		},
		EntryID:           entryID,
		ExecutionLogEntry: entry,
	})
	if err != nil {
		return err
	}

	return c.client.DoAndDrop(ctx, req)
}

// inferQueueName returns the queue name if it is specified on the job, which is the case
// when an executor is configured to listen to multiple queues. If the queue name is empty,
// return the specific queue that is configured.
func (c *Client) inferQueueName(job types.Job) string {
	if job.Queue != "" {
		return job.Queue
	} else {
		return c.options.QueueName
	}
}
