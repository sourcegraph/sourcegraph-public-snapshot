package apiclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

// Client is the client used to communicate with a remote job queue API.
type Client struct {
	options    Options
	client     *BaseClient
	operations *operations
}

type Options struct {
	// ExecutorName is a unique identifier for the requesting executor.
	ExecutorName string

	// PathPrefix is the path prefix added to all requests.
	PathPrefix string

	// EndpointOptions configures the target request URL.
	EndpointOptions EndpointOptions

	// BaseClientOptions configures the underlying HTTP client behavior.
	BaseClientOptions BaseClientOptions
}

type EndpointOptions struct {
	// URL is the target request URL.
	URL string

	// Username is the basic-auth username to include with all requests.
	Username string

	// Password is the basic-auth password to include with all requests.
	Password string
}

func New(options Options, observationContext *observation.Context) *Client {
	return &Client{
		options:    options,
		client:     NewBaseClient(options.BaseClientOptions),
		operations: makeOperations(observationContext),
	}
}

func (c *Client) Dequeue(ctx context.Context, queueName string, job *Job) (_ bool, err error) {
	ctx, endObservation := c.operations.dequeue.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("queueName", queueName),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.makeRequest("POST", fmt.Sprintf("%s/dequeue", queueName), DequeueRequest{
		ExecutorName: c.options.ExecutorName,
	})
	if err != nil {
		return false, err
	}

	return c.client.DoAndDecode(ctx, req, &job)
}

func (c *Client) AddExecutionLogEntry(ctx context.Context, queueName string, jobID int, entry workerutil.ExecutionLogEntry) (err error) {
	ctx, endObservation := c.operations.addExecutionLogEntry.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("queueName", queueName),
		log.Int("jobID", jobID),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.makeRequest("POST", fmt.Sprintf("%s/addExecutionLogEntry", queueName), AddExecutionLogEntryRequest{
		ExecutorName:      c.options.ExecutorName,
		JobID:             jobID,
		ExecutionLogEntry: entry,
	})
	if err != nil {
		return err
	}

	return c.client.DoAndDrop(ctx, req)
}

func (c *Client) MarkComplete(ctx context.Context, queueName string, jobID int) (err error) {
	ctx, endObservation := c.operations.markComplete.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("queueName", queueName),
		log.Int("jobID", jobID),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.makeRequest("POST", fmt.Sprintf("%s/markComplete", queueName), MarkCompleteRequest{
		ExecutorName: c.options.ExecutorName,
		JobID:        jobID,
	})
	if err != nil {
		return err
	}

	return c.client.DoAndDrop(ctx, req)
}

func (c *Client) MarkErrored(ctx context.Context, queueName string, jobID int, errorMessage string) (err error) {
	ctx, endObservation := c.operations.markErrored.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("queueName", queueName),
		log.Int("jobID", jobID),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.makeRequest("POST", fmt.Sprintf("%s/markErrored", queueName), MarkErroredRequest{
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
	ctx, endObservation := c.operations.markFailed.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("queueName", queueName),
		log.Int("jobID", jobID),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.makeRequest("POST", fmt.Sprintf("%s/markFailed", queueName), MarkErroredRequest{
		ExecutorName: c.options.ExecutorName,
		JobID:        jobID,
		ErrorMessage: errorMessage,
	})
	if err != nil {
		return err
	}

	return c.client.DoAndDrop(ctx, req)
}

func (c *Client) Heartbeat(ctx context.Context, jobIDs []int) (err error) {
	strJobIDs := make([]string, len(jobIDs))
	for _, jobID := range jobIDs {
		strJobIDs = append(strJobIDs, strconv.FormatInt(int64(jobID), 10))
	}

	ctx, endObservation := c.operations.heartbeat.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("jobIDs", strings.Join(strJobIDs, ",")),
	}})
	defer endObservation(1, observation.Args{})

	req, err := c.makeRequest("POST", "heartbeat", HeartbeatRequest{
		ExecutorName: c.options.ExecutorName,
		JobIDs:       jobIDs,
	})
	if err != nil {
		return err
	}

	return c.client.DoAndDrop(ctx, req)
}

func (c *Client) makeRequest(method, path string, payload interface{}) (*http.Request, error) {
	u, err := makeURL(
		c.options.EndpointOptions.URL,
		c.options.EndpointOptions.Username,
		c.options.EndpointOptions.Password,
		c.options.PathPrefix,
		path,
	)
	if err != nil {
		return nil, err
	}

	return MakeJSONRequest(method, u, payload)
}

func makeURL(base, username, password string, path ...string) (*url.URL, error) {
	u, err := makeRelativeURL(base, path...)
	if err != nil {
		return nil, err
	}

	u.User = url.UserPassword(username, password)
	return u, nil
}

func makeRelativeURL(base string, path ...string) (*url.URL, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return nil, err
	}

	return baseURL.ResolveReference(&url.URL{Path: filepath.Join(path...)}), nil
}
