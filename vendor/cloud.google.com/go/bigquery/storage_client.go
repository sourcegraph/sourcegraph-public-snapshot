// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bigquery

import (
	"context"
	"fmt"
	"runtime"

	"cloud.google.com/go/bigquery/internal"
	storage "cloud.google.com/go/bigquery/storage/apiv1"
	"cloud.google.com/go/bigquery/storage/apiv1/storagepb"
	"cloud.google.com/go/internal/detect"
	gax "github.com/googleapis/gax-go/v2"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

// readClient is a managed BigQuery Storage read client scoped to a single project.
type readClient struct {
	rawClient *storage.BigQueryReadClient
	projectID string

	settings readClientSettings
}

type readClientSettings struct {
	maxStreamCount int
	maxWorkerCount int
}

func defaultReadClientSettings() readClientSettings {
	maxWorkerCount := runtime.GOMAXPROCS(0)
	return readClientSettings{
		// with zero, the server will provide a value of streams so as to produce reasonable throughput
		maxStreamCount: 0,
		maxWorkerCount: maxWorkerCount,
	}
}

// newReadClient instantiates a new storage read client.
func newReadClient(ctx context.Context, projectID string, opts ...option.ClientOption) (c *readClient, err error) {
	numConns := runtime.GOMAXPROCS(0)
	if numConns > 4 {
		numConns = 4
	}
	o := []option.ClientOption{
		option.WithGRPCConnectionPool(numConns),
		option.WithUserAgent(fmt.Sprintf("%s/%s", userAgentPrefix, internal.Version)),
	}
	o = append(o, opts...)

	rawClient, err := storage.NewBigQueryReadClient(ctx, o...)
	if err != nil {
		return nil, err
	}
	rawClient.SetGoogleClientInfo("gccl", internal.Version)

	// Handle project autodetection.
	projectID, err = detect.ProjectID(ctx, projectID, "", opts...)
	if err != nil {
		return nil, err
	}

	settings := defaultReadClientSettings()
	rc := &readClient{
		rawClient: rawClient,
		projectID: projectID,
		settings:  settings,
	}

	return rc, nil
}

// close releases resources held by the client.
func (c *readClient) close() error {
	if c.rawClient == nil {
		return fmt.Errorf("already closed")
	}
	c.rawClient.Close()
	c.rawClient = nil
	return nil
}

// sessionForTable establishes a new session to fetch from a table using the Storage API
func (c *readClient) sessionForTable(ctx context.Context, table *Table, ordered bool) (*readSession, error) {
	tableID, err := table.Identifier(StorageAPIResourceID)
	if err != nil {
		return nil, err
	}

	// copy settings for a given session, to avoid overrides for all sessions
	settings := c.settings
	if ordered {
		settings.maxStreamCount = 1
	}

	rs := &readSession{
		ctx:                   ctx,
		table:                 table,
		tableID:               tableID,
		settings:              settings,
		readRowsFunc:          c.rawClient.ReadRows,
		createReadSessionFunc: c.rawClient.CreateReadSession,
	}
	return rs, nil
}

// ReadSession is the abstraction over a storage API read session.
type readSession struct {
	settings readClientSettings

	ctx     context.Context
	table   *Table
	tableID string

	bqSession *storagepb.ReadSession

	// decouple from readClient to enable testing
	createReadSessionFunc func(context.Context, *storagepb.CreateReadSessionRequest, ...gax.CallOption) (*storagepb.ReadSession, error)
	readRowsFunc          func(context.Context, *storagepb.ReadRowsRequest, ...gax.CallOption) (storagepb.BigQueryRead_ReadRowsClient, error)
}

// Start initiates a read session
func (rs *readSession) start() error {
	var preferredMinStreamCount int32
	maxStreamCount := int32(rs.settings.maxStreamCount)
	if maxStreamCount == 0 {
		preferredMinStreamCount = int32(rs.settings.maxWorkerCount)
	}
	createReadSessionRequest := &storagepb.CreateReadSessionRequest{
		Parent: fmt.Sprintf("projects/%s", rs.table.ProjectID),
		ReadSession: &storagepb.ReadSession{
			Table:      rs.tableID,
			DataFormat: storagepb.DataFormat_ARROW,
		},
		MaxStreamCount:          maxStreamCount,
		PreferredMinStreamCount: preferredMinStreamCount,
	}
	rpcOpts := gax.WithGRPCOptions(
		// Read API can send batches up to 128MB
		// https://cloud.google.com/bigquery/quotas#storage-limits
		grpc.MaxCallRecvMsgSize(1024 * 1024 * 129),
	)
	session, err := rs.createReadSessionFunc(rs.ctx, createReadSessionRequest, rpcOpts)
	if err != nil {
		return err
	}
	rs.bqSession = session
	return nil
}

// readRows returns a more direct iterators to the underlying Storage API row stream.
func (rs *readSession) readRows(req *storagepb.ReadRowsRequest) (storagepb.BigQueryRead_ReadRowsClient, error) {
	if rs.bqSession == nil {
		err := rs.start()
		if err != nil {
			return nil, err
		}
	}
	return rs.readRowsFunc(rs.ctx, req)
}
