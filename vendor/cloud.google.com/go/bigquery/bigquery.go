// Copyright 2015 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bigquery

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/bigquery/internal"
	cloudinternal "cloud.google.com/go/internal"
	"cloud.google.com/go/internal/detect"
	"cloud.google.com/go/internal/trace"
	"cloud.google.com/go/internal/version"
	gax "github.com/googleapis/gax-go/v2"
	bq "google.golang.org/api/bigquery/v2"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

const (
	// Scope is the Oauth2 scope for the service.
	// For relevant BigQuery scopes, see:
	// https://developers.google.com/identity/protocols/googlescopes#bigqueryv2
	Scope           = "https://www.googleapis.com/auth/bigquery"
	userAgentPrefix = "gcloud-golang-bigquery"
)

var xGoogHeader = fmt.Sprintf("gl-go/%s gccl/%s", version.Go(), internal.Version)

func setClientHeader(headers http.Header) {
	headers.Set("x-goog-api-client", xGoogHeader)
}

// Client may be used to perform BigQuery operations.
type Client struct {
	// Location, if set, will be used as the default location for all subsequent
	// dataset creation and job operations. A location specified directly in one of
	// those operations will override this value.
	Location string

	projectID string
	bqs       *bq.Service
	rc        *readClient

	// governs use of preview query features.
	enableQueryPreview bool
}

// DetectProjectID is a sentinel value that instructs NewClient to detect the
// project ID. It is given in place of the projectID argument. NewClient will
// use the project ID from the given credentials or the default credentials
// (https://developers.google.com/accounts/docs/application-default-credentials)
// if no credentials were provided. When providing credentials, not all
// options will allow NewClient to extract the project ID. Specifically a JWT
// does not have the project ID encoded.
const DetectProjectID = "*detect-project-id*"

// NewClient constructs a new Client which can perform BigQuery operations.
// Operations performed via the client are billed to the specified GCP project.
//
// If the project ID is set to DetectProjectID, NewClient will attempt to detect
// the project ID from credentials.
//
// This client supports enabling query-related preview features via environmental
// variables.  By setting the environment variable QUERY_PREVIEW_ENABLED to the string
// "TRUE", the client will enable preview features, though behavior may still be
// controlled via the bigquery service as well.  Currently, the feature(s) in scope
// include: stateless queries (query execution without corresponding job metadata).
func NewClient(ctx context.Context, projectID string, opts ...option.ClientOption) (*Client, error) {
	o := []option.ClientOption{
		option.WithScopes(Scope),
		option.WithUserAgent(fmt.Sprintf("%s/%s", userAgentPrefix, internal.Version)),
	}
	o = append(o, opts...)
	bqs, err := bq.NewService(ctx, o...)
	if err != nil {
		return nil, fmt.Errorf("bigquery: constructing client: %w", err)
	}

	// Handle project autodetection.
	projectID, err = detect.ProjectID(ctx, projectID, "", opts...)
	if err != nil {
		return nil, err
	}

	var preview bool
	if v, ok := os.LookupEnv("QUERY_PREVIEW_ENABLED"); ok {
		if strings.ToUpper(v) == "TRUE" {
			preview = true
		}
	}

	c := &Client{
		projectID:          projectID,
		bqs:                bqs,
		enableQueryPreview: preview,
	}
	return c, nil
}

// EnableStorageReadClient sets up Storage API connection to be used when fetching
// large datasets from tables, jobs or queries.
// Currently out of pagination methods like PageInfo().Token and RowIterator.StartIndex
// are not supported when the Storage API is enabled.
// Calling this method twice will return an error.
func (c *Client) EnableStorageReadClient(ctx context.Context, opts ...option.ClientOption) error {
	if c.isStorageReadAvailable() {
		return fmt.Errorf("failed: storage read client already set up")
	}
	rc, err := newReadClient(ctx, c.projectID, opts...)
	if err != nil {
		return err
	}
	c.rc = rc
	return nil
}

func (c *Client) isStorageReadAvailable() bool {
	return c.rc != nil
}

// Project returns the project ID or number for this instance of the client, which may have
// either been explicitly specified or autodetected.
func (c *Client) Project() string {
	return c.projectID
}

// Close closes any resources held by the client.
// Close should be called when the client is no longer needed.
// It need not be called at program exit.
func (c *Client) Close() error {
	if c.isStorageReadAvailable() {
		err := c.rc.close()
		if err != nil {
			return err
		}
	}
	return nil
}

// Calls the Jobs.Insert RPC and returns a Job.
func (c *Client) insertJob(ctx context.Context, job *bq.Job, media io.Reader, mediaOpts ...googleapi.MediaOption) (*Job, error) {
	call := c.bqs.Jobs.Insert(c.projectID, job).Context(ctx)
	setClientHeader(call.Header())
	if media != nil {
		call.Media(media, mediaOpts...)
	}
	var res *bq.Job
	var err error
	invoke := func() error {
		sCtx := trace.StartSpan(ctx, "bigquery.jobs.insert")
		res, err = call.Do()
		trace.EndSpan(sCtx, err)
		return err
	}
	// A job with a client-generated ID can be retried; the presence of the
	// ID makes the insert operation idempotent.
	// We don't retry if there is media, because it is an io.Reader. We'd
	// have to read the contents and keep it in memory, and that could be expensive.
	// TODO(jba): Look into retrying if media != nil.
	if job.JobReference != nil && media == nil {
		// We deviate from default retries due to BigQuery wanting to retry structured internal job errors.
		err = runWithRetryExplicit(ctx, invoke, jobRetryReasons)
	} else {
		err = invoke()
	}
	if err != nil {
		return nil, err
	}
	return bqToJob(res, c)
}

// runQuery invokes the optimized query path.
// Due to differences in options it supports, it cannot be used for all existing
// jobs.insert requests that are query jobs.
func (c *Client) runQuery(ctx context.Context, queryRequest *bq.QueryRequest) (*bq.QueryResponse, error) {
	call := c.bqs.Jobs.Query(c.projectID, queryRequest).Context(ctx)
	setClientHeader(call.Header())

	var res *bq.QueryResponse
	var err error
	invoke := func() error {
		sCtx := trace.StartSpan(ctx, "bigquery.jobs.query")
		res, err = call.Do()
		trace.EndSpan(sCtx, err)
		return err
	}

	// We control request ID, so we can always runWithRetry.
	err = runWithRetryExplicit(ctx, invoke, jobRetryReasons)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Convert a number of milliseconds since the Unix epoch to a time.Time.
// Treat an input of zero specially: convert it to the zero time,
// rather than the start of the epoch.
func unixMillisToTime(m int64) time.Time {
	if m == 0 {
		return time.Time{}
	}
	return time.Unix(0, m*1e6)
}

// runWithRetry calls the function until it returns nil or a non-retryable error, or
// the context is done.
// See the similar function in ../storage/invoke.go. The main difference is the
// reason for retrying.
func runWithRetry(ctx context.Context, call func() error) error {
	return runWithRetryExplicit(ctx, call, defaultRetryReasons)
}

func runWithRetryExplicit(ctx context.Context, call func() error, allowedReasons []string) error {
	// These parameters match the suggestions in https://cloud.google.com/bigquery/sla.
	backoff := gax.Backoff{
		Initial:    1 * time.Second,
		Max:        32 * time.Second,
		Multiplier: 2,
	}
	return cloudinternal.Retry(ctx, backoff, func() (stop bool, err error) {
		err = call()
		if err == nil {
			return true, nil
		}
		return !retryableError(err, allowedReasons), err
	})
}

var (
	defaultRetryReasons = []string{"backendError", "rateLimitExceeded"}
	jobRetryReasons     = []string{"backendError", "rateLimitExceeded", "internalError"}
	retry5xxCodes       = []int{
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
	}
)

// retryableError is the unary retry predicate for this library.  In addition to structured error
// reasons, it specifies some HTTP codes (500, 502, 503, 504) and network/transport reasons.
func retryableError(err error, allowedReasons []string) bool {
	if err == nil {
		return false
	}
	if err == io.ErrUnexpectedEOF {
		return true
	}
	// Special case due to http2: https://github.com/googleapis/google-cloud-go/issues/1793
	// Due to Go's default being higher for streams-per-connection than is accepted by the
	// BQ backend, it's possible to get streams refused immediately after a connection is
	// started but before we receive SETTINGS frame from the backend.  This generally only
	// happens when we try to enqueue > 100 requests onto a newly initiated connection.
	if err.Error() == "http2: stream closed" {
		return true
	}

	switch e := err.(type) {
	case *googleapi.Error:
		// We received a structured error from backend.
		var reason string
		if len(e.Errors) > 0 {
			reason = e.Errors[0].Reason
			for _, r := range allowedReasons {
				if reason == r {
					return true
				}
			}
		}
		for _, code := range retry5xxCodes {
			if e.Code == code {
				return true
			}
		}
	case *url.Error:
		retryable := []string{"connection refused", "connection reset"}
		for _, s := range retryable {
			if strings.Contains(e.Error(), s) {
				return true
			}
		}
	case interface{ Temporary() bool }:
		if e.Temporary() {
			return true
		}
	}
	// Check wrapped error.
	return retryableError(errors.Unwrap(err), allowedReasons)
}
