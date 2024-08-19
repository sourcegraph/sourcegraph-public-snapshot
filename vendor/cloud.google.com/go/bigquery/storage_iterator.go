// Copyright 2023 Google LLC
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
	"sync"
	"sync/atomic"
	"time"

	"cloud.google.com/go/bigquery/internal/query"
	"cloud.google.com/go/bigquery/storage/apiv1/storagepb"
	"github.com/googleapis/gax-go/v2"
	"golang.org/x/sync/semaphore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// storageArrowIterator is a raw interface for getting data from Storage Read API
type storageArrowIterator struct {
	done        uint32 // atomic flag
	initialized bool
	errs        chan error
	ctx         context.Context

	schema    Schema
	rawSchema []byte
	records   chan *ArrowRecordBatch

	session *readSession
}

var _ ArrowIterator = &storageArrowIterator{}

func newStorageRowIteratorFromTable(ctx context.Context, table *Table, ordered bool) (*RowIterator, error) {
	md, err := table.Metadata(ctx)
	if err != nil {
		return nil, err
	}
	rs, err := table.c.rc.sessionForTable(ctx, table, ordered)
	if err != nil {
		return nil, err
	}
	it, err := newStorageRowIterator(rs, md.Schema)
	if err != nil {
		return nil, err
	}
	if rs.bqSession == nil {
		return nil, errors.New("read session not initialized")
	}
	arrowSerializedSchema := rs.bqSession.GetArrowSchema().GetSerializedSchema()
	dec, err := newArrowDecoder(arrowSerializedSchema, md.Schema)
	if err != nil {
		return nil, err
	}
	it.arrowDecoder = dec
	it.Schema = md.Schema
	return it, nil
}

func newStorageRowIteratorFromJob(ctx context.Context, j *Job) (*RowIterator, error) {
	// Needed to fetch destination table
	job, err := j.c.JobFromProject(ctx, j.projectID, j.jobID, j.location)
	if err != nil {
		return nil, err
	}
	cfg, err := job.Config()
	if err != nil {
		return nil, err
	}
	qcfg := cfg.(*QueryConfig)
	if qcfg.Dst == nil {
		if !job.isScript() {
			return nil, fmt.Errorf("job has no destination table to read")
		}
		lastJob, err := resolveLastChildSelectJob(ctx, job)
		if err != nil {
			return nil, err
		}
		return newStorageRowIteratorFromJob(ctx, lastJob)
	}
	ordered := query.HasOrderedResults(qcfg.Q)
	return newStorageRowIteratorFromTable(ctx, qcfg.Dst, ordered)
}

func resolveLastChildSelectJob(ctx context.Context, job *Job) (*Job, error) {
	childJobs := []*Job{}
	it := job.Children(ctx)
	for {
		job, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to resolve table for script job: %w", err)
		}
		if !job.isSelectQuery() {
			continue
		}
		childJobs = append(childJobs, job)
	}
	if len(childJobs) == 0 {
		return nil, fmt.Errorf("failed to resolve table for script job: no child jobs found")
	}
	return childJobs[0], nil
}

func newRawStorageRowIterator(rs *readSession, schema Schema) (*storageArrowIterator, error) {
	arrowIt := &storageArrowIterator{
		ctx:     rs.ctx,
		session: rs,
		schema:  schema,
		records: make(chan *ArrowRecordBatch, rs.settings.maxWorkerCount+1),
		errs:    make(chan error, rs.settings.maxWorkerCount+1),
	}
	if rs.bqSession == nil {
		err := rs.start()
		if err != nil {
			return nil, err
		}
	}
	arrowIt.rawSchema = rs.bqSession.GetArrowSchema().GetSerializedSchema()
	return arrowIt, nil
}

func newStorageRowIterator(rs *readSession, schema Schema) (*RowIterator, error) {
	arrowIt, err := newRawStorageRowIterator(rs, schema)
	if err != nil {
		return nil, err
	}
	totalRows := arrowIt.session.bqSession.EstimatedRowCount
	it := &RowIterator{
		ctx:           rs.ctx,
		arrowIterator: arrowIt,
		TotalRows:     uint64(totalRows),
		rows:          [][]Value{},
	}
	it.nextFunc = nextFuncForStorageIterator(it)
	it.pageInfo = &iterator.PageInfo{
		Token:   "",
		MaxSize: int(totalRows),
	}
	return it, nil
}

func nextFuncForStorageIterator(it *RowIterator) func() error {
	return func() error {
		if len(it.rows) > 0 {
			return nil
		}
		record, err := it.arrowIterator.Next()
		if err == iterator.Done {
			if len(it.rows) == 0 {
				return iterator.Done
			}
			return nil
		}
		if err != nil {
			return err
		}
		if it.Schema == nil {
			it.Schema = it.arrowIterator.Schema()
		}
		rows, err := it.arrowDecoder.decodeArrowRecords(record)
		if err != nil {
			return err
		}
		it.rows = rows
		return nil
	}
}

func (it *storageArrowIterator) init() error {
	if it.initialized {
		return nil
	}

	bqSession := it.session.bqSession
	if bqSession == nil {
		return errors.New("read session not initialized")
	}

	streams := bqSession.Streams
	if len(streams) == 0 {
		return iterator.Done
	}

	wg := sync.WaitGroup{}
	wg.Add(len(streams))
	sem := semaphore.NewWeighted(int64(it.session.settings.maxWorkerCount))
	go func() {
		wg.Wait()
		close(it.records)
		close(it.errs)
		it.markDone()
	}()

	go func() {
		for _, readStream := range streams {
			err := sem.Acquire(it.ctx, 1)
			if err != nil {
				wg.Done()
				continue
			}
			go func(readStreamName string) {
				it.processStream(readStreamName)
				sem.Release(1)
				wg.Done()
			}(readStream.Name)
		}
	}()
	it.initialized = true
	return nil
}

func (it *storageArrowIterator) markDone() {
	atomic.StoreUint32(&it.done, 1)
}

func (it *storageArrowIterator) isDone() bool {
	return atomic.LoadUint32(&it.done) != 0
}

func (it *storageArrowIterator) processStream(readStream string) {
	bo := gax.Backoff{}
	var offset int64
	for {
		rowStream, err := it.session.readRows(&storagepb.ReadRowsRequest{
			ReadStream: readStream,
			Offset:     offset,
		})
		if err != nil {
			if it.session.ctx.Err() != nil { // context cancelled, don't try again
				return
			}
			backoff, shouldRetry := retryReadRows(bo, err)
			if shouldRetry {
				if err := gax.Sleep(it.ctx, backoff); err != nil {
					return // context cancelled
				}
				continue
			}
			it.errs <- fmt.Errorf("failed to read rows on stream %s: %w", readStream, err)
			continue
		}
		offset, err = it.consumeRowStream(readStream, rowStream, offset)
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			if it.session.ctx.Err() != nil { // context cancelled, don't queue error
				return
			}
			backoff, shouldRetry := retryReadRows(bo, err)
			if shouldRetry {
				if err := gax.Sleep(it.ctx, backoff); err != nil {
					return // context cancelled
				}
				continue
			}
			it.errs <- fmt.Errorf("failed to read rows on stream %s: %w", readStream, err)
			// try to re-open row stream with updated offset
		}
	}
}

func retryReadRows(bo gax.Backoff, err error) (time.Duration, bool) {
	s, ok := status.FromError(err)
	if !ok {
		return bo.Pause(), false
	}
	switch s.Code() {
	case codes.Aborted,
		codes.Canceled,
		codes.DeadlineExceeded,
		codes.FailedPrecondition,
		codes.Internal,
		codes.Unavailable:
		return bo.Pause(), true
	}
	return bo.Pause(), false
}

func (it *storageArrowIterator) consumeRowStream(readStream string, rowStream storagepb.BigQueryRead_ReadRowsClient, offset int64) (int64, error) {
	for {
		r, err := rowStream.Recv()
		if err != nil {
			if err == io.EOF {
				return offset, err
			}
			return offset, fmt.Errorf("failed to consume rows on stream %s: %w", readStream, err)
		}
		if r.RowCount > 0 {
			offset += r.RowCount
			recordBatch := r.GetArrowRecordBatch()
			it.records <- &ArrowRecordBatch{
				PartitionID: readStream,
				Schema:      it.rawSchema,
				Data:        recordBatch.SerializedRecordBatch,
			}
		}
	}
}

// next return the next batch of rows as an arrow.Record.
// Accessing Arrow Records directly has the drawnback of having to deal
// with memory management.
func (it *storageArrowIterator) Next() (*ArrowRecordBatch, error) {
	if err := it.init(); err != nil {
		return nil, err
	}
	if len(it.records) > 0 {
		return <-it.records, nil
	}
	if it.isDone() {
		return nil, iterator.Done
	}
	select {
	case record := <-it.records:
		if record == nil {
			return nil, iterator.Done
		}
		return record, nil
	case err := <-it.errs:
		return nil, err
	case <-it.ctx.Done():
		return nil, it.ctx.Err()
	}
}

func (it *storageArrowIterator) SerializedArrowSchema() []byte {
	return it.rawSchema
}

func (it *storageArrowIterator) Schema() Schema {
	return it.schema
}

// IsAccelerated check if the current RowIterator is
// being accelerated by Storage API.
func (it *RowIterator) IsAccelerated() bool {
	return it.arrowIterator != nil
}

// ArrowIterator gives access to the raw Arrow Record Batch stream to be consumed directly.
// Experimental: this interface is experimental and may be modified or removed in future versions,
// regardless of any other documented package stability guarantees.
// Don't try to mix RowIterator.Next and ArrowIterator.Next calls.
func (it *RowIterator) ArrowIterator() (ArrowIterator, error) {
	if !it.IsAccelerated() {
		// TODO: can we convert plain RowIterator based on JSON API to an Arrow Stream ?
		return nil, errors.New("bigquery: require storage read API to be enabled")
	}
	return it.arrowIterator, nil
}
