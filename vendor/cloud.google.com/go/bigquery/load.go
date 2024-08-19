// Copyright 2016 Google LLC
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
	"io"
	"time"

	"cloud.google.com/go/internal/trace"
	bq "google.golang.org/api/bigquery/v2"
	"google.golang.org/api/googleapi"
)

// LoadConfig holds the configuration for a load job.
type LoadConfig struct {
	// Src is the source from which data will be loaded.
	Src LoadSource

	// Dst is the table into which the data will be loaded.
	Dst *Table

	// CreateDisposition specifies the circumstances under which the destination table will be created.
	// The default is CreateIfNeeded.
	CreateDisposition TableCreateDisposition

	// WriteDisposition specifies how existing data in the destination table is treated.
	// The default is WriteAppend.
	WriteDisposition TableWriteDisposition

	// The labels associated with this job.
	Labels map[string]string

	// If non-nil, the destination table is partitioned by time.
	TimePartitioning *TimePartitioning

	// If non-nil, the destination table is partitioned by integer range.
	RangePartitioning *RangePartitioning

	// Clustering specifies the data clustering configuration for the destination table.
	Clustering *Clustering

	// Custom encryption configuration (e.g., Cloud KMS keys).
	DestinationEncryptionConfig *EncryptionConfig

	// Allows the schema of the destination table to be updated as a side effect of
	// the load job.
	SchemaUpdateOptions []string

	// For Avro-based loads, controls whether logical type annotations are used.
	// See https://cloud.google.com/bigquery/docs/loading-data-cloud-storage-avro#logical_types
	// for additional information.
	UseAvroLogicalTypes bool

	// For ingestion from datastore backups, ProjectionFields governs which fields
	// are projected from the backup.  The default behavior projects all fields.
	ProjectionFields []string

	// HivePartitioningOptions allows use of Hive partitioning based on the
	// layout of objects in Cloud Storage.
	HivePartitioningOptions *HivePartitioningOptions

	// DecimalTargetTypes allows selection of how decimal values are converted when
	// processed in bigquery, subject to the value type having sufficient precision/scale
	// to support the values.  In the order of NUMERIC, BIGNUMERIC, and STRING, a type is
	// selected if is present in the list and if supports the necessary precision and scale.
	//
	// StringTargetType supports all precision and scale values.
	DecimalTargetTypes []DecimalTargetType

	// Sets a best-effort deadline on a specific job.  If job execution exceeds this
	// timeout, BigQuery may attempt to cancel this work automatically.
	//
	// This deadline cannot be adjusted or removed once the job is created.  Consider
	// using Job.Cancel in situations where you need more dynamic behavior.
	//
	// Experimental: this option is experimental and may be modified or removed in future versions,
	// regardless of any other documented package stability guarantees.
	JobTimeout time.Duration

	// When loading a table with external data, the user can provide a reference file with the table schema.
	// This is enabled for the following formats: AVRO, PARQUET, ORC.
	ReferenceFileSchemaURI string

	// If true, creates a new session, where session id will
	// be a server generated random id. If false, runs query with an
	// existing session_id passed in ConnectionProperty, otherwise runs the
	// load job in non-session mode.
	CreateSession bool

	// ConnectionProperties are optional key-values settings.
	ConnectionProperties []*ConnectionProperty

	// MediaOptions stores options for customizing media upload.
	MediaOptions []googleapi.MediaOption
}

func (l *LoadConfig) toBQ() (*bq.JobConfiguration, io.Reader) {
	config := &bq.JobConfiguration{
		Labels: l.Labels,
		Load: &bq.JobConfigurationLoad{
			CreateDisposition:                  string(l.CreateDisposition),
			WriteDisposition:                   string(l.WriteDisposition),
			DestinationTable:                   l.Dst.toBQ(),
			TimePartitioning:                   l.TimePartitioning.toBQ(),
			RangePartitioning:                  l.RangePartitioning.toBQ(),
			Clustering:                         l.Clustering.toBQ(),
			DestinationEncryptionConfiguration: l.DestinationEncryptionConfig.toBQ(),
			SchemaUpdateOptions:                l.SchemaUpdateOptions,
			UseAvroLogicalTypes:                l.UseAvroLogicalTypes,
			ProjectionFields:                   l.ProjectionFields,
			HivePartitioningOptions:            l.HivePartitioningOptions.toBQ(),
			ReferenceFileSchemaUri:             l.ReferenceFileSchemaURI,
			CreateSession:                      l.CreateSession,
		},
		JobTimeoutMs: l.JobTimeout.Milliseconds(),
	}
	for _, v := range l.DecimalTargetTypes {
		config.Load.DecimalTargetTypes = append(config.Load.DecimalTargetTypes, string(v))
	}
	for _, v := range l.ConnectionProperties {
		config.Load.ConnectionProperties = append(config.Load.ConnectionProperties, v.toBQ())
	}
	media := l.Src.populateLoadConfig(config.Load)
	return config, media
}

func bqToLoadConfig(q *bq.JobConfiguration, c *Client) *LoadConfig {
	lc := &LoadConfig{
		Labels:                      q.Labels,
		CreateDisposition:           TableCreateDisposition(q.Load.CreateDisposition),
		WriteDisposition:            TableWriteDisposition(q.Load.WriteDisposition),
		Dst:                         bqToTable(q.Load.DestinationTable, c),
		TimePartitioning:            bqToTimePartitioning(q.Load.TimePartitioning),
		RangePartitioning:           bqToRangePartitioning(q.Load.RangePartitioning),
		Clustering:                  bqToClustering(q.Load.Clustering),
		DestinationEncryptionConfig: bqToEncryptionConfig(q.Load.DestinationEncryptionConfiguration),
		SchemaUpdateOptions:         q.Load.SchemaUpdateOptions,
		UseAvroLogicalTypes:         q.Load.UseAvroLogicalTypes,
		ProjectionFields:            q.Load.ProjectionFields,
		HivePartitioningOptions:     bqToHivePartitioningOptions(q.Load.HivePartitioningOptions),
		ReferenceFileSchemaURI:      q.Load.ReferenceFileSchemaUri,
		CreateSession:               q.Load.CreateSession,
	}
	if q.JobTimeoutMs > 0 {
		lc.JobTimeout = time.Duration(q.JobTimeoutMs) * time.Millisecond
	}
	for _, v := range q.Load.DecimalTargetTypes {
		lc.DecimalTargetTypes = append(lc.DecimalTargetTypes, DecimalTargetType(v))
	}
	for _, v := range q.Load.ConnectionProperties {
		lc.ConnectionProperties = append(lc.ConnectionProperties, bqToConnectionProperty(v))
	}
	var fc *FileConfig
	if len(q.Load.SourceUris) == 0 {
		s := NewReaderSource(nil)
		fc = &s.FileConfig
		lc.Src = s
	} else {
		s := NewGCSReference(q.Load.SourceUris...)
		fc = &s.FileConfig
		lc.Src = s
	}
	bqPopulateFileConfig(q.Load, fc)
	return lc
}

// A Loader loads data from Google Cloud Storage into a BigQuery table.
type Loader struct {
	JobIDConfig
	LoadConfig
	c *Client
}

// A LoadSource represents a source of data that can be loaded into
// a BigQuery table.
//
// This package defines two LoadSources: GCSReference, for Google Cloud Storage
// objects, and ReaderSource, for data read from an io.Reader.
type LoadSource interface {
	// populates config, returns media
	populateLoadConfig(*bq.JobConfigurationLoad) io.Reader
}

// LoaderFrom returns a Loader which can be used to load data into a BigQuery table.
// The returned Loader may optionally be further configured before its Run method is called.
// See GCSReference and ReaderSource for additional configuration options that
// affect loading.
func (t *Table) LoaderFrom(src LoadSource) *Loader {
	return &Loader{
		c: t.c,
		LoadConfig: LoadConfig{
			Src: src,
			Dst: t,
		},
	}
}

// Run initiates a load job.
func (l *Loader) Run(ctx context.Context) (j *Job, err error) {
	ctx = trace.StartSpan(ctx, "cloud.google.com/go/bigquery.Load.Run")
	defer func() { trace.EndSpan(ctx, err) }()

	job, media := l.newJob()
	return l.c.insertJob(ctx, job, media, l.LoadConfig.MediaOptions...)
}

func (l *Loader) newJob() (*bq.Job, io.Reader) {
	config, media := l.LoadConfig.toBQ()
	return &bq.Job{
		JobReference:  l.JobIDConfig.createJobRef(l.c),
		Configuration: config,
	}, media
}

// DecimalTargetType is used to express preference ordering for converting values from external formats.
type DecimalTargetType string

var (
	// NumericTargetType indicates the preferred type is NUMERIC when supported.
	NumericTargetType DecimalTargetType = "NUMERIC"

	// BigNumericTargetType indicates the preferred type is BIGNUMERIC when supported.
	BigNumericTargetType DecimalTargetType = "BIGNUMERIC"

	// StringTargetType indicates the preferred type is STRING when supported.
	StringTargetType DecimalTargetType = "STRING"
)
