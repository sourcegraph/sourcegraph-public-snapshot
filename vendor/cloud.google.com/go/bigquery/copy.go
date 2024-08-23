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
	"time"

	bq "google.golang.org/api/bigquery/v2"
)

// TableCopyOperationType is used to indicate the type of operation performed by a BigQuery
// copy job.
type TableCopyOperationType string

var (
	// CopyOperation indicates normal table to table copying.
	CopyOperation TableCopyOperationType = "COPY"
	// SnapshotOperation indicates creating a snapshot from a regular table, which
	// operates as an immutable copy.
	SnapshotOperation TableCopyOperationType = "SNAPSHOT"
	// RestoreOperation indicates creating/restoring a table from a snapshot.
	RestoreOperation TableCopyOperationType = "RESTORE"
	// CloneOperation indicates creating a table clone, which creates a writeable
	// copy of a base table that is billed based on difference from the base table.
	CloneOperation TableCopyOperationType = "CLONE"
)

// CopyConfig holds the configuration for a copy job.
type CopyConfig struct {
	// Srcs are the tables from which data will be copied.
	Srcs []*Table

	// Dst is the table into which the data will be copied.
	Dst *Table

	// CreateDisposition specifies the circumstances under which the destination table will be created.
	// The default is CreateIfNeeded.
	CreateDisposition TableCreateDisposition

	// WriteDisposition specifies how existing data in the destination table is treated.
	// The default is WriteEmpty.
	WriteDisposition TableWriteDisposition

	// The labels associated with this job.
	Labels map[string]string

	// Custom encryption configuration (e.g., Cloud KMS keys).
	DestinationEncryptionConfig *EncryptionConfig

	// One of the supported operation types when executing a Table Copy jobs.  By default this
	// copies tables, but can also be set to perform snapshot or restore operations.
	OperationType TableCopyOperationType

	// Sets a best-effort deadline on a specific job.  If job execution exceeds this
	// timeout, BigQuery may attempt to cancel this work automatically.
	//
	// This deadline cannot be adjusted or removed once the job is created.  Consider
	// using Job.Cancel in situations where you need more dynamic behavior.
	//
	// Experimental: this option is experimental and may be modified or removed in future versions,
	// regardless of any other documented package stability guarantees.
	JobTimeout time.Duration
}

func (c *CopyConfig) toBQ() *bq.JobConfiguration {
	var ts []*bq.TableReference
	for _, t := range c.Srcs {
		ts = append(ts, t.toBQ())
	}
	return &bq.JobConfiguration{
		Labels: c.Labels,
		Copy: &bq.JobConfigurationTableCopy{
			CreateDisposition:                  string(c.CreateDisposition),
			WriteDisposition:                   string(c.WriteDisposition),
			DestinationTable:                   c.Dst.toBQ(),
			DestinationEncryptionConfiguration: c.DestinationEncryptionConfig.toBQ(),
			SourceTables:                       ts,
			OperationType:                      string(c.OperationType),
		},
		JobTimeoutMs: c.JobTimeout.Milliseconds(),
	}
}

func bqToCopyConfig(q *bq.JobConfiguration, c *Client) *CopyConfig {
	cc := &CopyConfig{
		Labels:                      q.Labels,
		CreateDisposition:           TableCreateDisposition(q.Copy.CreateDisposition),
		WriteDisposition:            TableWriteDisposition(q.Copy.WriteDisposition),
		Dst:                         bqToTable(q.Copy.DestinationTable, c),
		DestinationEncryptionConfig: bqToEncryptionConfig(q.Copy.DestinationEncryptionConfiguration),
		OperationType:               TableCopyOperationType(q.Copy.OperationType),
		JobTimeout:                  time.Duration(q.JobTimeoutMs) * time.Millisecond,
	}
	for _, t := range q.Copy.SourceTables {
		cc.Srcs = append(cc.Srcs, bqToTable(t, c))
	}
	return cc
}

// A Copier copies data into a BigQuery table from one or more BigQuery tables.
type Copier struct {
	JobIDConfig
	CopyConfig
	c *Client
}

// CopierFrom returns a Copier which can be used to copy data into a
// BigQuery table from one or more BigQuery tables.
// The returned Copier may optionally be further configured before its Run method is called.
func (t *Table) CopierFrom(srcs ...*Table) *Copier {
	return &Copier{
		c: t.c,
		CopyConfig: CopyConfig{
			Srcs: srcs,
			Dst:  t,
		},
	}
}

// Run initiates a copy job.
func (c *Copier) Run(ctx context.Context) (*Job, error) {
	return c.c.insertJob(ctx, c.newJob(), nil)
}

func (c *Copier) newJob() *bq.Job {
	return &bq.Job{
		JobReference:  c.JobIDConfig.createJobRef(c.c),
		Configuration: c.CopyConfig.toBQ(),
	}
}
