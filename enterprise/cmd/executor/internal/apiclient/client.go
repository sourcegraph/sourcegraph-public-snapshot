package apiclient

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Client is the client used to communicate with a remote job queue API.
type Client struct {
	options         Options
	client          *BaseClient
	logger          log.Logger
	metricsGatherer prometheus.Gatherer
	operations      *operations
}

type Options struct {
	// ExecutorName is a unique identifier for the requesting executor.
	ExecutorName string

	// PathPrefix is the path prefix added to all requests.
	PathPrefix string

	// EndpointOptions configures the target request URL.
	EndpointOptions EndpointOptions

	// BaseClientOptions are the underlying HTTP client options.
	BaseClientOptions BaseClientOptions

	// TelemetryOptions captures additional parameters sent in heartbeat requests.
	TelemetryOptions TelemetryOptions
}

type EndpointOptions struct {
	// URL is the target request URL.
	URL string

	// Token is the authorization token to include with all requests (via Authorization header).
	Token string
}

func New(options Options, metricsGatherer prometheus.Gatherer, observationContext *observation.Context) *Client {
	return &Client{
		options:         options,
		client:          NewBaseClient(options.BaseClientOptions),
		logger:          log.Scoped("executor-api-client", "The API client adapter for executors to use dbworkers over HTTP"),
		metricsGatherer: metricsGatherer,
		operations:      newOperations(observationContext),
	}
}

// TODO: Should this live here?
// TODO: Should we also add a method to download that src cli version?
// That would only work on a non-airgapped executor though.
func (c *Client) LatestSrcCLIVersion(ctx context.Context) (_ string, err error) {
	u, err := makeRelativeURL(
		c.options.EndpointOptions.URL,
		".api/src-cli/version",
	)
	if err != nil {
		return "", err
	}

	req, err := MakeJSONRequest(http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}

	type versionPayload struct {
		Version string `json:"version"`
	}
	var v versionPayload
	if _, err := c.client.DoAndDecode(ctx, req, &v); err != nil {
		return "", err
	}

	return v.Version, nil
}

func (c *Client) Dequeue(ctx context.Context, queueName string, job *executor.Job) (_ bool, err error) {
	ctx, _, endObservation := c.operations.dequeue.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("queueName", queueName),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.makeRequest("POST", fmt.Sprintf("%s/dequeue", queueName), executor.DequeueRequest{
		ExecutorName: c.options.ExecutorName,
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

	req, err := c.makeRequest("POST", fmt.Sprintf("%s/addExecutionLogEntry", queueName), executor.AddExecutionLogEntryRequest{
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

	req, err := c.makeRequest("POST", fmt.Sprintf("%s/updateExecutionLogEntry", queueName), executor.UpdateExecutionLogEntryRequest{
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

	req, err := c.makeRequest("POST", fmt.Sprintf("%s/markComplete", queueName), executor.MarkCompleteRequest{
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

	req, err := c.makeRequest("POST", fmt.Sprintf("%s/markErrored", queueName), executor.MarkErroredRequest{
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

	req, err := c.makeRequest("POST", fmt.Sprintf("%s/markFailed", queueName), executor.MarkErroredRequest{
		ExecutorName: c.options.ExecutorName,
		JobID:        jobID,
		ErrorMessage: errorMessage,
	})
	if err != nil {
		return err
	}

	return c.client.DoAndDrop(ctx, req)
}

func (c *Client) CanceledJobs(ctx context.Context, queueName string, knownIDs []int) (canceledIDs []int, err error) {
	req, err := c.makeRequest("POST", fmt.Sprintf("%s/canceledJobs", queueName), executor.CanceledJobsRequest{
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
	req, err := c.makeRequest("POST", fmt.Sprintf("%s/heartbeat", queueName), executor.HeartbeatRequest{
		ExecutorName: c.options.ExecutorName,
	})
	if err != nil {
		return err
	}

	return c.client.DoAndDrop(ctx, req)
}

func (c *Client) Heartbeat(ctx context.Context, queueName string, jobIDs []int) (knownIDs []int, err error) {
	ctx, _, endObservation := c.operations.heartbeat.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("queueName", queueName),
		otlog.String("jobIDs", intsToString(jobIDs)),
	}})
	defer endObservation(1, observation.Args{})

	metrics, err := gatherMetrics(c.logger, c.metricsGatherer)
	if err != nil {
		c.logger.Error("Failed to collect prometheus metrics for heartbeat", log.Error(err))
		// Continue, no metrics should not prevent heartbeats.
	}

	req, err := c.makeRequest("POST", fmt.Sprintf("%s/heartbeat", queueName), executor.HeartbeatRequest{
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
		return nil, err
	}

	if _, err := c.client.DoAndDecode(ctx, req, &knownIDs); err != nil {
		return nil, err
	}

	return knownIDs, nil
}

const SchemeExecutorToken = "token-executor"

func (c *Client) makeRequest(method, path string, payload any) (*http.Request, error) {
	u, err := makeRelativeURL(
		c.options.EndpointOptions.URL,
		c.options.PathPrefix,
		path,
	)
	if err != nil {
		return nil, err
	}

	r, err := MakeJSONRequest(method, u, payload)
	if err != nil {
		return nil, err
	}

	r.Header.Add("Authorization", fmt.Sprintf("%s %s", SchemeExecutorToken, c.options.EndpointOptions.Token))
	return r, nil
}

func makeRelativeURL(base string, path ...string) (*url.URL, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return nil, err
	}

	return baseURL.ResolveReference(&url.URL{Path: filepath.Join(path...)}), nil
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
