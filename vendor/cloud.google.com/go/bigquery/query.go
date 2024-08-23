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
	"time"

	"cloud.google.com/go/internal/trace"
	"cloud.google.com/go/internal/uid"
	bq "google.golang.org/api/bigquery/v2"
)

// QueryConfig holds the configuration for a query job.
type QueryConfig struct {
	// Dst is the table into which the results of the query will be written.
	// If this field is nil, a temporary table will be created.
	Dst *Table

	// The query to execute. See https://cloud.google.com/bigquery/query-reference for details.
	Q string

	// DefaultProjectID and DefaultDatasetID specify the dataset to use for unqualified table names in the query.
	// If DefaultProjectID is set, DefaultDatasetID must also be set.
	DefaultProjectID string
	DefaultDatasetID string

	// TableDefinitions describes data sources outside of BigQuery.
	// The map keys may be used as table names in the query string.
	//
	// When a QueryConfig is returned from Job.Config, the map values
	// are always of type *ExternalDataConfig.
	TableDefinitions map[string]ExternalData

	// CreateDisposition specifies the circumstances under which the destination table will be created.
	// The default is CreateIfNeeded.
	CreateDisposition TableCreateDisposition

	// WriteDisposition specifies how existing data in the destination table is treated.
	// The default is WriteEmpty.
	WriteDisposition TableWriteDisposition

	// DisableQueryCache prevents results being fetched from the query cache.
	// If this field is false, results are fetched from the cache if they are available.
	// The query cache is a best-effort cache that is flushed whenever tables in the query are modified.
	// Cached results are only available when TableID is unspecified in the query's destination Table.
	// For more information, see https://cloud.google.com/bigquery/querying-data#querycaching
	DisableQueryCache bool

	// DisableFlattenedResults prevents results being flattened.
	// If this field is false, results from nested and repeated fields are flattened.
	// DisableFlattenedResults implies AllowLargeResults
	// For more information, see https://cloud.google.com/bigquery/docs/data#nested
	DisableFlattenedResults bool

	// AllowLargeResults allows the query to produce arbitrarily large result tables.
	// The destination must be a table.
	// When using this option, queries will take longer to execute, even if the result set is small.
	// For additional limitations, see https://cloud.google.com/bigquery/querying-data#largequeryresults
	AllowLargeResults bool

	// Priority specifies the priority with which to schedule the query.
	// The default priority is InteractivePriority.
	// For more information, see https://cloud.google.com/bigquery/querying-data#batchqueries
	Priority QueryPriority

	// MaxBillingTier sets the maximum billing tier for a Query.
	// Queries that have resource usage beyond this tier will fail (without
	// incurring a charge). If this field is zero, the project default will be used.
	MaxBillingTier int

	// MaxBytesBilled limits the number of bytes billed for
	// this job.  Queries that would exceed this limit will fail (without incurring
	// a charge).
	// If this field is less than 1, the project default will be
	// used.
	MaxBytesBilled int64

	// UseStandardSQL causes the query to use standard SQL. The default.
	// Deprecated: use UseLegacySQL.
	UseStandardSQL bool

	// UseLegacySQL causes the query to use legacy SQL.
	UseLegacySQL bool

	// Parameters is a list of query parameters. The presence of parameters
	// implies the use of standard SQL.
	// If the query uses positional syntax ("?"), then no parameter may have a name.
	// If the query uses named syntax ("@p"), then all parameters must have names.
	// It is illegal to mix positional and named syntax.
	Parameters []QueryParameter

	// TimePartitioning specifies time-based partitioning
	// for the destination table.
	TimePartitioning *TimePartitioning

	// RangePartitioning specifies integer range-based partitioning
	// for the destination table.
	RangePartitioning *RangePartitioning

	// Clustering specifies the data clustering configuration for the destination table.
	Clustering *Clustering

	// The labels associated with this job.
	Labels map[string]string

	// If true, don't actually run this job. A valid query will return a mostly
	// empty response with some processing statistics, while an invalid query will
	// return the same error it would if it wasn't a dry run.
	//
	// Query.Read will fail with dry-run queries. Call Query.Run instead, and then
	// call LastStatus on the returned job to get statistics. Calling Status on a
	// dry-run job will fail.
	DryRun bool

	// Custom encryption configuration (e.g., Cloud KMS keys).
	DestinationEncryptionConfig *EncryptionConfig

	// Allows the schema of the destination table to be updated as a side effect of
	// the query job.
	SchemaUpdateOptions []string

	// CreateSession will trigger creation of a new session when true.
	CreateSession bool

	// ConnectionProperties are optional key-values settings.
	ConnectionProperties []*ConnectionProperty

	// Sets a best-effort deadline on a specific job.  If job execution exceeds this
	// timeout, BigQuery may attempt to cancel this work automatically.
	//
	// This deadline cannot be adjusted or removed once the job is created.  Consider
	// using Job.Cancel in situations where you need more dynamic behavior.
	//
	// Experimental: this option is experimental and may be modified or removed in future versions,
	// regardless of any other documented package stability guarantees.
	JobTimeout time.Duration

	// Force usage of Storage API if client is available. For test scenarios
	forceStorageAPI bool
}

func (qc *QueryConfig) toBQ() (*bq.JobConfiguration, error) {
	qconf := &bq.JobConfigurationQuery{
		Query:                              qc.Q,
		CreateDisposition:                  string(qc.CreateDisposition),
		WriteDisposition:                   string(qc.WriteDisposition),
		AllowLargeResults:                  qc.AllowLargeResults,
		Priority:                           string(qc.Priority),
		MaximumBytesBilled:                 qc.MaxBytesBilled,
		TimePartitioning:                   qc.TimePartitioning.toBQ(),
		RangePartitioning:                  qc.RangePartitioning.toBQ(),
		Clustering:                         qc.Clustering.toBQ(),
		DestinationEncryptionConfiguration: qc.DestinationEncryptionConfig.toBQ(),
		SchemaUpdateOptions:                qc.SchemaUpdateOptions,
		CreateSession:                      qc.CreateSession,
	}
	if len(qc.TableDefinitions) > 0 {
		qconf.TableDefinitions = make(map[string]bq.ExternalDataConfiguration)
	}
	for name, data := range qc.TableDefinitions {
		qconf.TableDefinitions[name] = data.toBQ()
	}
	if qc.DefaultProjectID != "" || qc.DefaultDatasetID != "" {
		qconf.DefaultDataset = &bq.DatasetReference{
			DatasetId: qc.DefaultDatasetID,
			ProjectId: qc.DefaultProjectID,
		}
	}
	if tier := int64(qc.MaxBillingTier); tier > 0 {
		qconf.MaximumBillingTier = &tier
	}
	f := false
	if qc.DisableQueryCache {
		qconf.UseQueryCache = &f
	}
	if qc.DisableFlattenedResults {
		qconf.FlattenResults = &f
		// DisableFlattenResults implies AllowLargeResults.
		qconf.AllowLargeResults = true
	}
	if qc.UseStandardSQL && qc.UseLegacySQL {
		return nil, errors.New("bigquery: cannot provide both UseStandardSQL and UseLegacySQL")
	}
	if len(qc.Parameters) > 0 && qc.UseLegacySQL {
		return nil, errors.New("bigquery: cannot provide both Parameters (implying standard SQL) and UseLegacySQL")
	}
	ptrue := true
	pfalse := false
	if qc.UseLegacySQL {
		qconf.UseLegacySql = &ptrue
	} else {
		qconf.UseLegacySql = &pfalse
	}
	if qc.Dst != nil && !qc.Dst.implicitTable() {
		qconf.DestinationTable = qc.Dst.toBQ()
	}
	for _, p := range qc.Parameters {
		qp, err := p.toBQ()
		if err != nil {
			return nil, err
		}
		qconf.QueryParameters = append(qconf.QueryParameters, qp)
	}
	if len(qc.ConnectionProperties) > 0 {
		bqcp := make([]*bq.ConnectionProperty, len(qc.ConnectionProperties))
		for k, v := range qc.ConnectionProperties {
			bqcp[k] = v.toBQ()
		}
		qconf.ConnectionProperties = bqcp
	}
	jc := &bq.JobConfiguration{
		Labels: qc.Labels,
		DryRun: qc.DryRun,
		Query:  qconf,
	}
	if qc.JobTimeout > 0 {
		jc.JobTimeoutMs = qc.JobTimeout.Milliseconds()
	}
	return jc, nil
}

func bqToQueryConfig(q *bq.JobConfiguration, c *Client) (*QueryConfig, error) {
	qq := q.Query
	qc := &QueryConfig{
		Labels:                      q.Labels,
		DryRun:                      q.DryRun,
		JobTimeout:                  time.Duration(q.JobTimeoutMs) * time.Millisecond,
		Q:                           qq.Query,
		CreateDisposition:           TableCreateDisposition(qq.CreateDisposition),
		WriteDisposition:            TableWriteDisposition(qq.WriteDisposition),
		AllowLargeResults:           qq.AllowLargeResults,
		Priority:                    QueryPriority(qq.Priority),
		MaxBytesBilled:              qq.MaximumBytesBilled,
		UseLegacySQL:                qq.UseLegacySql == nil || *qq.UseLegacySql,
		TimePartitioning:            bqToTimePartitioning(qq.TimePartitioning),
		RangePartitioning:           bqToRangePartitioning(qq.RangePartitioning),
		Clustering:                  bqToClustering(qq.Clustering),
		DestinationEncryptionConfig: bqToEncryptionConfig(qq.DestinationEncryptionConfiguration),
		SchemaUpdateOptions:         qq.SchemaUpdateOptions,
		CreateSession:               qq.CreateSession,
	}
	qc.UseStandardSQL = !qc.UseLegacySQL

	if len(qq.TableDefinitions) > 0 {
		qc.TableDefinitions = make(map[string]ExternalData)
	}
	for name, qedc := range qq.TableDefinitions {
		edc, err := bqToExternalDataConfig(&qedc)
		if err != nil {
			return nil, err
		}
		qc.TableDefinitions[name] = edc
	}
	if qq.DefaultDataset != nil {
		qc.DefaultProjectID = qq.DefaultDataset.ProjectId
		qc.DefaultDatasetID = qq.DefaultDataset.DatasetId
	}
	if qq.MaximumBillingTier != nil {
		qc.MaxBillingTier = int(*qq.MaximumBillingTier)
	}
	if qq.UseQueryCache != nil && !*qq.UseQueryCache {
		qc.DisableQueryCache = true
	}
	if qq.FlattenResults != nil && !*qq.FlattenResults {
		qc.DisableFlattenedResults = true
	}
	if qq.DestinationTable != nil {
		qc.Dst = bqToTable(qq.DestinationTable, c)
	}
	for _, qp := range qq.QueryParameters {
		p, err := bqToQueryParameter(qp)
		if err != nil {
			return nil, err
		}
		qc.Parameters = append(qc.Parameters, p)
	}
	if len(qq.ConnectionProperties) > 0 {
		props := make([]*ConnectionProperty, len(qq.ConnectionProperties))
		for k, v := range qq.ConnectionProperties {
			props[k] = bqToConnectionProperty(v)
		}
		qc.ConnectionProperties = props
	}
	return qc, nil
}

// QueryPriority specifies a priority with which a query is to be executed.
type QueryPriority string

const (
	// BatchPriority specifies that the query should be scheduled with the
	// batch priority.  BigQuery queues each batch query on your behalf, and
	// starts the query as soon as idle resources are available, usually within
	// a few minutes. If BigQuery hasn't started the query within 24 hours,
	// BigQuery changes the job priority to interactive. Batch queries don't
	// count towards your concurrent rate limit, which can make it easier to
	// start many queries at once.
	//
	// More information can be found at https://cloud.google.com/bigquery/docs/running-queries#batchqueries.
	BatchPriority QueryPriority = "BATCH"
	// InteractivePriority specifies that the query should be scheduled with
	// interactive priority, which means that the query is executed as soon as
	// possible. Interactive queries count towards your concurrent rate limit
	// and your daily limit. It is the default priority with which queries get
	// executed.
	//
	// More information can be found at https://cloud.google.com/bigquery/docs/running-queries#queries.
	InteractivePriority QueryPriority = "INTERACTIVE"
)

// A Query queries data from a BigQuery table. Use Client.Query to create a Query.
type Query struct {
	JobIDConfig
	QueryConfig
	client *Client
}

// Query creates a query with string q.
// The returned Query may optionally be further configured before its Run method is called.
func (c *Client) Query(q string) *Query {
	return &Query{
		client:      c,
		QueryConfig: QueryConfig{Q: q},
	}
}

// Run initiates a query job.
func (q *Query) Run(ctx context.Context) (j *Job, err error) {
	ctx = trace.StartSpan(ctx, "cloud.google.com/go/bigquery.Query.Run")
	defer func() { trace.EndSpan(ctx, err) }()

	job, err := q.newJob()
	if err != nil {
		return nil, err
	}
	j, err = q.client.insertJob(ctx, job, nil)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (q *Query) newJob() (*bq.Job, error) {
	config, err := q.QueryConfig.toBQ()
	if err != nil {
		return nil, err
	}
	return &bq.Job{
		JobReference:  q.JobIDConfig.createJobRef(q.client),
		Configuration: config,
	}, nil
}

// Read submits a query for execution and returns the results via a RowIterator.
// If the request can be satisfied by running using the optimized query path, it
// is used in place of the jobs.insert path as this path does not expose a job
// object.
func (q *Query) Read(ctx context.Context) (it *RowIterator, err error) {
	if q.QueryConfig.DryRun {
		return nil, errors.New("bigquery: cannot evaluate Query.Read() for dry-run queries")
	}
	ctx = trace.StartSpan(ctx, "cloud.google.com/go/bigquery.Query.Run")
	defer func() { trace.EndSpan(ctx, err) }()
	queryRequest, err := q.probeFastPath()
	if err != nil {
		// Any error means we fallback to the older mechanism.
		job, err := q.Run(ctx)
		if err != nil {
			return nil, err
		}
		return job.Read(ctx)
	}

	// we have a config, run on fastPath.
	resp, err := q.client.runQuery(ctx, queryRequest)
	if err != nil {
		return nil, err
	}

	// construct a minimal job for backing the row iterator.
	var minimalJob *Job
	if resp.JobReference != nil {
		minimalJob = &Job{
			c:         q.client,
			jobID:     resp.JobReference.JobId,
			location:  resp.JobReference.Location,
			projectID: resp.JobReference.ProjectId,
		}
	}

	if resp.JobComplete {
		// If more pages are available, discard and use the Storage API instead
		if resp.PageToken != "" && q.client.isStorageReadAvailable() {
			it, err = newStorageRowIteratorFromJob(ctx, minimalJob)
			if err == nil {
				return it, nil
			}
		}
		rowSource := &rowSource{
			j:       minimalJob,
			queryID: resp.QueryId,
			// RowIterator can precache results from the iterator to save a lookup.
			cachedRows:      resp.Rows,
			cachedSchema:    resp.Schema,
			cachedNextToken: resp.PageToken,
		}
		return newRowIterator(ctx, rowSource, fetchPage), nil
	}
	// We're on the fastPath, but we need to poll because the job is incomplete.
	// Fallback to job-based Read().
	//
	// (Issue 2937) In order to satisfy basic probing of the job in classic path,
	// we need to supply additional config which is probed for presence, not contents.
	//
	minimalJob.config = &bq.JobConfiguration{
		Query: &bq.JobConfigurationQuery{},
	}

	return minimalJob.Read(ctx)
}

// probeFastPath is used to attempt configuring a jobs.Query request based on a
// user's Query configuration.  If all the options set on the job are supported on the
// faster query path, this method returns a QueryRequest suitable for execution.
func (q *Query) probeFastPath() (*bq.QueryRequest, error) {
	if q.forceStorageAPI && q.client.isStorageReadAvailable() {
		return nil, fmt.Errorf("force Storage API usage")
	}
	// This is a denylist of settings which prevent us from composing an equivalent
	// bq.QueryRequest due to differences between configuration parameters accepted
	// by jobs.insert vs jobs.query.
	if q.QueryConfig.Dst != nil ||
		q.QueryConfig.TableDefinitions != nil ||
		q.QueryConfig.CreateDisposition != "" ||
		q.QueryConfig.WriteDisposition != "" ||
		!(q.QueryConfig.Priority == "" || q.QueryConfig.Priority == InteractivePriority) ||
		q.QueryConfig.UseLegacySQL ||
		q.QueryConfig.MaxBillingTier != 0 ||
		q.QueryConfig.TimePartitioning != nil ||
		q.QueryConfig.RangePartitioning != nil ||
		q.QueryConfig.Clustering != nil ||
		q.QueryConfig.DestinationEncryptionConfig != nil ||
		q.QueryConfig.SchemaUpdateOptions != nil ||
		q.QueryConfig.JobTimeout != 0 ||
		// User has defined the jobID generation behavior
		q.JobIDConfig.JobID != "" {
		return nil, fmt.Errorf("QueryConfig incompatible with fastPath")
	}
	pfalse := false
	qRequest := &bq.QueryRequest{
		Query:              q.QueryConfig.Q,
		CreateSession:      q.CreateSession,
		Location:           q.Location,
		UseLegacySql:       &pfalse,
		MaximumBytesBilled: q.QueryConfig.MaxBytesBilled,
		RequestId:          uid.NewSpace("request", nil).New(),
		Labels:             q.Labels,
		FormatOptions: &bq.DataFormatOptions{
			UseInt64Timestamp: true,
		},
	}
	if q.QueryConfig.DisableQueryCache {
		qRequest.UseQueryCache = &pfalse
	}
	// Convert query parameters
	for _, p := range q.QueryConfig.Parameters {
		qp, err := p.toBQ()
		if err != nil {
			return nil, err
		}
		qRequest.QueryParameters = append(qRequest.QueryParameters, qp)
	}
	if q.QueryConfig.DefaultDatasetID != "" {
		qRequest.DefaultDataset = &bq.DatasetReference{
			ProjectId: q.QueryConfig.DefaultProjectID,
			DatasetId: q.QueryConfig.DefaultDatasetID,
		}
	}
	if q.client.enableQueryPreview {
		qRequest.JobCreationMode = "JOB_CREATION_OPTIONAL"
	}
	return qRequest, nil
}

// ConnectionProperty represents a single key and value pair that can be sent alongside a query request or load job.
type ConnectionProperty struct {
	// Name of the connection property to set.
	Key string
	// Value of the connection property.
	Value string
}

func (cp *ConnectionProperty) toBQ() *bq.ConnectionProperty {
	if cp == nil {
		return nil
	}
	return &bq.ConnectionProperty{
		Key:   cp.Key,
		Value: cp.Value,
	}
}

func bqToConnectionProperty(in *bq.ConnectionProperty) *ConnectionProperty {
	if in == nil {
		return nil
	}
	return &ConnectionProperty{
		Key:   in.Key,
		Value: in.Value,
	}
}
