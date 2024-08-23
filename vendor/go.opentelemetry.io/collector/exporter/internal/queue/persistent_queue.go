// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package queue // import "go.opentelemetry.io/collector/exporter/internal/queue"

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"go.uber.org/multierr"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/internal/experr"
	"go.opentelemetry.io/collector/extension/experimental/storage"
)

// persistentQueue provides a persistent queue implementation backed by file storage extension
//
// Write index describes the position at which next item is going to be stored.
// Read index describes which item needs to be read next.
// When Write index = Read index, no elements are in the queue.
//
// The items currently dispatched by consumers are not deleted until the processing is finished.
// Their list is stored under a separate key.
//
//	┌───────file extension-backed queue───────┐
//	│                                         │
//	│     ┌───┐     ┌───┐ ┌───┐ ┌───┐ ┌───┐   │
//	│ n+1 │ n │ ... │ 4 │ │ 3 │ │ 2 │ │ 1 │   │
//	│     └───┘     └───┘ └─x─┘ └─|─┘ └─x─┘   │
//	│                       x     |     x     │
//	└───────────────────────x─────|─────x─────┘
//	   ▲              ▲     x     |     x
//	   │              │     x     |     xxxx deleted
//	   │              │     x     |
//	 write          read    x     └── currently dispatched item
//	 index          index   x
//	                        xxxx deleted
type persistentQueue[T any] struct {
	// sizedChannel is used by the persistent queue for two purposes:
	// 1. a communication channel notifying the consumer that a new item is available.
	// 2. capacity control based on the size of the items.
	*sizedChannel[permanentQueueEl]

	set    PersistentQueueSettings[T]
	logger *zap.Logger
	client storage.Client

	// isRequestSized indicates whether the queue is sized by the number of requests.
	isRequestSized bool

	// mu guards everything declared below.
	mu                       sync.Mutex
	readIndex                uint64
	writeIndex               uint64
	currentlyDispatchedItems []uint64
	refClient                int64
	stopped                  bool
}

const (
	zapKey           = "key"
	zapErrorCount    = "errorCount"
	zapNumberOfItems = "numberOfItems"

	readIndexKey                = "ri"
	writeIndexKey               = "wi"
	currentlyDispatchedItemsKey = "di"
	queueSizeKey                = "si"
)

var (
	errValueNotSet        = errors.New("value not set")
	errInvalidValue       = errors.New("invalid value")
	errNoStorageClient    = errors.New("no storage client extension found")
	errWrongExtensionType = errors.New("requested extension is not a storage extension")
)

type PersistentQueueSettings[T any] struct {
	Sizer            Sizer[T]
	Capacity         int64
	DataType         component.DataType
	StorageID        component.ID
	Marshaler        func(req T) ([]byte, error)
	Unmarshaler      func([]byte) (T, error)
	ExporterSettings exporter.Settings
}

// NewPersistentQueue creates a new queue backed by file storage; name and signal must be a unique combination that identifies the queue storage
func NewPersistentQueue[T any](set PersistentQueueSettings[T]) Queue[T] {
	_, isRequestSized := set.Sizer.(*RequestSizer[T])
	return &persistentQueue[T]{
		set:            set,
		logger:         set.ExporterSettings.Logger,
		isRequestSized: isRequestSized,
	}
}

// Start starts the persistentQueue with the given number of consumers.
func (pq *persistentQueue[T]) Start(ctx context.Context, host component.Host) error {
	storageClient, err := toStorageClient(ctx, pq.set.StorageID, host, pq.set.ExporterSettings.ID, pq.set.DataType)
	if err != nil {
		return err
	}
	pq.initClient(ctx, storageClient)
	return nil
}

func (pq *persistentQueue[T]) initClient(ctx context.Context, client storage.Client) {
	pq.client = client
	// Start with a reference 1 which is the reference we use for the producer goroutines and initialization.
	pq.refClient = 1
	pq.initPersistentContiguousStorage(ctx)
	// Make sure the leftover requests are handled
	pq.retrieveAndEnqueueNotDispatchedReqs(ctx)
}

func (pq *persistentQueue[T]) initPersistentContiguousStorage(ctx context.Context) {
	riOp := storage.GetOperation(readIndexKey)
	wiOp := storage.GetOperation(writeIndexKey)

	err := pq.client.Batch(ctx, riOp, wiOp)
	if err == nil {
		pq.readIndex, err = bytesToItemIndex(riOp.Value)
	}

	if err == nil {
		pq.writeIndex, err = bytesToItemIndex(wiOp.Value)
	}

	if err != nil {
		if errors.Is(err, errValueNotSet) {
			pq.logger.Info("Initializing new persistent queue")
		} else {
			pq.logger.Error("Failed getting read/write index, starting with new ones", zap.Error(err))
		}
		pq.readIndex = 0
		pq.writeIndex = 0
	}

	initIndexSize := pq.writeIndex - pq.readIndex

	var (
		initEls       []permanentQueueEl
		initQueueSize uint64
	)

	// Pre-allocate the communication channel with the size of the restored queue.
	if initIndexSize > 0 {
		initQueueSize = initIndexSize
		// If the queue is sized by the number of requests, no need to read the queue size from storage.
		if !pq.isRequestSized {
			if restoredQueueSize, err := pq.restoreQueueSizeFromStorage(ctx); err == nil {
				initQueueSize = restoredQueueSize
			}
		}

		// Ensure the communication channel filled with evenly sized elements up to the total restored queue size.
		initEls = make([]permanentQueueEl, initIndexSize)
	}

	pq.sizedChannel = newSizedChannel[permanentQueueEl](pq.set.Capacity, initEls, int64(initQueueSize))
}

// permanentQueueEl is the type of the elements passed to the sizedChannel by the persistentQueue.
type permanentQueueEl struct{}

// restoreQueueSizeFromStorage restores the queue size from storage.
func (pq *persistentQueue[T]) restoreQueueSizeFromStorage(ctx context.Context) (uint64, error) {
	val, err := pq.client.Get(ctx, queueSizeKey)
	if err != nil {
		if errors.Is(err, errValueNotSet) {
			pq.logger.Warn("Cannot read the queue size snapshot from storage. "+
				"The reported queue size will be inaccurate until the initial queue is drained. "+
				"It's expected when the items sized queue enabled for the first time", zap.Error(err))
		} else {
			pq.logger.Error("Failed to read the queue size snapshot from storage. "+
				"The reported queue size will be inaccurate until the initial queue is drained.", zap.Error(err))
		}
		return 0, err
	}
	return bytesToItemIndex(val)
}

// Consume applies the provided function on the head of queue.
// The call blocks until there is an item available or the queue is stopped.
// The function returns true when an item is consumed or false if the queue is stopped.
func (pq *persistentQueue[T]) Consume(consumeFunc func(context.Context, T) error) bool {
	for {
		var (
			req                  T
			onProcessingFinished func(error)
			consumed             bool
		)

		// If we are stopped we still process all the other events in the channel before, but we
		// return fast in the `getNextItem`, so we will free the channel fast and get to the stop.
		_, ok := pq.sizedChannel.pop(func(permanentQueueEl) int64 {
			req, onProcessingFinished, consumed = pq.getNextItem(context.Background())
			if !consumed {
				return 0
			}
			return pq.set.Sizer.Sizeof(req)
		})
		if !ok {
			return false
		}
		if consumed {
			onProcessingFinished(consumeFunc(context.Background(), req))
			return true
		}
	}
}

func (pq *persistentQueue[T]) Shutdown(ctx context.Context) error {
	// If the queue is not initialized, there is nothing to shut down.
	if pq.client == nil {
		return nil
	}

	pq.mu.Lock()
	defer pq.mu.Unlock()
	backupErr := pq.backupQueueSize(ctx)
	pq.sizedChannel.shutdown()
	// Mark this queue as stopped, so consumer don't start any more work.
	pq.stopped = true
	return multierr.Combine(backupErr, pq.unrefClient(ctx))
}

// backupQueueSize writes the current queue size to storage. The value is used to recover the queue size
// in case if the collector is killed.
func (pq *persistentQueue[T]) backupQueueSize(ctx context.Context) error {
	// No need to write the queue size if the queue is sized by the number of requests.
	// That information is already stored as difference between read and write indexes.
	if pq.isRequestSized {
		return nil
	}

	return pq.client.Set(ctx, queueSizeKey, itemIndexToBytes(uint64(pq.Size())))
}

// unrefClient unrefs the client, and closes if no more references. Callers MUST hold the mutex.
// This is needed because consumers of the queue may still process the requests while the queue is shutting down or immediately after.
func (pq *persistentQueue[T]) unrefClient(ctx context.Context) error {
	pq.refClient--
	if pq.refClient == 0 {
		return pq.client.Close(ctx)
	}
	return nil
}

// Offer inserts the specified element into this queue if it is possible to do so immediately
// without violating capacity restrictions. If success returns no error.
// It returns ErrQueueIsFull if no space is currently available.
func (pq *persistentQueue[T]) Offer(ctx context.Context, req T) error {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	return pq.putInternal(ctx, req)
}

// putInternal is the internal version that requires caller to hold the mutex lock.
func (pq *persistentQueue[T]) putInternal(ctx context.Context, req T) error {
	err := pq.sizedChannel.push(permanentQueueEl{}, pq.set.Sizer.Sizeof(req), func() error {
		itemKey := getItemKey(pq.writeIndex)
		newIndex := pq.writeIndex + 1

		reqBuf, err := pq.set.Marshaler(req)
		if err != nil {
			return err
		}

		// Carry out a transaction where we both add the item and update the write index
		ops := []storage.Operation{
			storage.SetOperation(writeIndexKey, itemIndexToBytes(newIndex)),
			storage.SetOperation(itemKey, reqBuf),
		}
		if storageErr := pq.client.Batch(ctx, ops...); storageErr != nil {
			return storageErr
		}

		pq.writeIndex = newIndex
		return nil
	})
	if err != nil {
		return err
	}

	// Back up the queue size to storage every 10 writes. The stored value is used to recover the queue size
	// in case if the collector is killed. The recovered queue size is allowed to be inaccurate.
	if (pq.writeIndex % 10) == 5 {
		if err := pq.backupQueueSize(ctx); err != nil {
			pq.logger.Error("Error writing queue size to storage", zap.Error(err))
		}
	}

	return nil
}

// getNextItem pulls the next available item from the persistent storage along with a callback function that should be
// called after the item is processed to clean up the storage. If no new item is available, returns false.
func (pq *persistentQueue[T]) getNextItem(ctx context.Context) (T, func(error), bool) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	var request T

	if pq.stopped {
		return request, nil, false
	}

	if pq.readIndex == pq.writeIndex {
		return request, nil, false
	}

	index := pq.readIndex
	// Increase here, so even if errors happen below, it always iterates
	pq.readIndex++
	pq.currentlyDispatchedItems = append(pq.currentlyDispatchedItems, index)
	getOp := storage.GetOperation(getItemKey(index))
	err := pq.client.Batch(ctx,
		storage.SetOperation(readIndexKey, itemIndexToBytes(pq.readIndex)),
		storage.SetOperation(currentlyDispatchedItemsKey, itemIndexArrayToBytes(pq.currentlyDispatchedItems)),
		getOp)

	if err == nil {
		request, err = pq.set.Unmarshaler(getOp.Value)
	}

	if err != nil {
		pq.logger.Debug("Failed to dispatch item", zap.Error(err))
		// We need to make sure that currently dispatched items list is cleaned
		if err = pq.itemDispatchingFinish(ctx, index); err != nil {
			pq.logger.Error("Error deleting item from queue", zap.Error(err))
		}

		return request, nil, false
	}

	// Increase the reference count, so the client is not closed while the request is being processed.
	// The client cannot be closed because we hold the lock since last we checked `stopped`.
	pq.refClient++
	return request, func(consumeErr error) {
		// Delete the item from the persistent storage after it was processed.
		pq.mu.Lock()
		// Always unref client even if the consumer is shutdown because we always ref it for every valid request.
		defer func() {
			if err = pq.unrefClient(ctx); err != nil {
				pq.logger.Error("Error closing the storage client", zap.Error(err))
			}
			pq.mu.Unlock()
		}()

		if experr.IsShutdownErr(consumeErr) {
			// The queue is shutting down, don't mark the item as dispatched, so it's picked up again after restart.
			// TODO: Handle partially delivered requests by updating their values in the storage.
			return
		}

		if err = pq.itemDispatchingFinish(ctx, index); err != nil {
			pq.logger.Error("Error deleting item from queue", zap.Error(err))
		}

		// Back up the queue size to storage on every 10 reads. The stored value is used to recover the queue size
		// in case if the collector is killed. The recovered queue size is allowed to be inaccurate.
		if (pq.readIndex % 10) == 0 {
			if qsErr := pq.backupQueueSize(ctx); qsErr != nil {
				pq.logger.Error("Error writing queue size to storage", zap.Error(err))
			}
		}

		// Ensure the used size and the channel size are in sync.
		pq.sizedChannel.syncSize()

	}, true
}

// retrieveAndEnqueueNotDispatchedReqs gets the items for which sending was not finished, cleans the storage
// and moves the items at the back of the queue.
func (pq *persistentQueue[T]) retrieveAndEnqueueNotDispatchedReqs(ctx context.Context) {
	var dispatchedItems []uint64

	pq.mu.Lock()
	defer pq.mu.Unlock()
	pq.logger.Debug("Checking if there are items left for dispatch by consumers")
	itemKeysBuf, err := pq.client.Get(ctx, currentlyDispatchedItemsKey)
	if err == nil {
		dispatchedItems, err = bytesToItemIndexArray(itemKeysBuf)
	}
	if err != nil {
		pq.logger.Error("Could not fetch items left for dispatch by consumers", zap.Error(err))
		return
	}

	if len(dispatchedItems) == 0 {
		pq.logger.Debug("No items left for dispatch by consumers")
		return
	}

	pq.logger.Info("Fetching items left for dispatch by consumers", zap.Int(zapNumberOfItems,
		len(dispatchedItems)))
	retrieveBatch := make([]storage.Operation, len(dispatchedItems))
	cleanupBatch := make([]storage.Operation, len(dispatchedItems))
	for i, it := range dispatchedItems {
		key := getItemKey(it)
		retrieveBatch[i] = storage.GetOperation(key)
		cleanupBatch[i] = storage.DeleteOperation(key)
	}
	retrieveErr := pq.client.Batch(ctx, retrieveBatch...)
	cleanupErr := pq.client.Batch(ctx, cleanupBatch...)

	if cleanupErr != nil {
		pq.logger.Debug("Failed cleaning items left by consumers", zap.Error(cleanupErr))
	}

	if retrieveErr != nil {
		pq.logger.Warn("Failed retrieving items left by consumers", zap.Error(retrieveErr))
		return
	}

	errCount := 0
	for _, op := range retrieveBatch {
		if op.Value == nil {
			pq.logger.Warn("Failed retrieving item", zap.String(zapKey, op.Key), zap.Error(errValueNotSet))
			continue
		}
		req, err := pq.set.Unmarshaler(op.Value)
		// If error happened or item is nil, it will be efficiently ignored
		if err != nil {
			pq.logger.Warn("Failed unmarshalling item", zap.String(zapKey, op.Key), zap.Error(err))
			continue
		}
		if pq.putInternal(ctx, req) != nil {
			errCount++
		}
	}

	if errCount > 0 {
		pq.logger.Error("Errors occurred while moving items for dispatching back to queue",
			zap.Int(zapNumberOfItems, len(retrieveBatch)), zap.Int(zapErrorCount, errCount))
	} else {
		pq.logger.Info("Moved items for dispatching back to queue",
			zap.Int(zapNumberOfItems, len(retrieveBatch)))
	}
}

// itemDispatchingFinish removes the item from the list of currently dispatched items and deletes it from the persistent queue
func (pq *persistentQueue[T]) itemDispatchingFinish(ctx context.Context, index uint64) error {
	lenCDI := len(pq.currentlyDispatchedItems)
	for i := 0; i < lenCDI; i++ {
		if pq.currentlyDispatchedItems[i] == index {
			pq.currentlyDispatchedItems[i] = pq.currentlyDispatchedItems[lenCDI-1]
			pq.currentlyDispatchedItems = pq.currentlyDispatchedItems[:lenCDI-1]
			break
		}
	}

	setOp := storage.SetOperation(currentlyDispatchedItemsKey, itemIndexArrayToBytes(pq.currentlyDispatchedItems))
	deleteOp := storage.DeleteOperation(getItemKey(index))
	if err := pq.client.Batch(ctx, setOp, deleteOp); err != nil {
		// got an error, try to gracefully handle it
		pq.logger.Warn("Failed updating currently dispatched items, trying to delete the item first",
			zap.Error(err))
	} else {
		// Everything ok, exit
		return nil
	}

	if err := pq.client.Batch(ctx, deleteOp); err != nil {
		// Return an error here, as this indicates an issue with the underlying storage medium
		return fmt.Errorf("failed deleting item from queue, got error from storage: %w", err)
	}

	if err := pq.client.Batch(ctx, setOp); err != nil {
		// even if this fails, we still have the right dispatched items in memory
		// at worst, we'll have the wrong list in storage, and we'll discard the nonexistent items during startup
		return fmt.Errorf("failed updating currently dispatched items, but deleted item successfully: %w", err)
	}

	return nil
}

func toStorageClient(ctx context.Context, storageID component.ID, host component.Host, ownerID component.ID, signal component.DataType) (storage.Client, error) {
	ext, found := host.GetExtensions()[storageID]
	if !found {
		return nil, errNoStorageClient
	}

	storageExt, ok := ext.(storage.Extension)
	if !ok {
		return nil, errWrongExtensionType
	}

	return storageExt.GetClient(ctx, component.KindExporter, ownerID, signal.String())
}

func getItemKey(index uint64) string {
	return strconv.FormatUint(index, 10)
}

func itemIndexToBytes(value uint64) []byte {
	return binary.LittleEndian.AppendUint64([]byte{}, value)
}

func bytesToItemIndex(buf []byte) (uint64, error) {
	if buf == nil {
		return uint64(0), errValueNotSet
	}
	// The sizeof uint64 in binary is 8.
	if len(buf) < 8 {
		return 0, errInvalidValue
	}
	return binary.LittleEndian.Uint64(buf), nil
}

func itemIndexArrayToBytes(arr []uint64) []byte {
	size := len(arr)
	buf := make([]byte, 0, 4+size*8)
	buf = binary.LittleEndian.AppendUint32(buf, uint32(size))
	for _, item := range arr {
		buf = binary.LittleEndian.AppendUint64(buf, item)
	}
	return buf
}

func bytesToItemIndexArray(buf []byte) ([]uint64, error) {
	if len(buf) == 0 {
		return nil, nil
	}

	// The sizeof uint32 in binary is 4.
	if len(buf) < 4 {
		return nil, errInvalidValue
	}
	size := int(binary.LittleEndian.Uint32(buf))
	if size == 0 {
		return nil, nil
	}

	buf = buf[4:]
	// The sizeof uint64 in binary is 8, so we need to have size*8 bytes.
	if len(buf) < size*8 {
		return nil, errInvalidValue
	}

	val := make([]uint64, size)
	for i := 0; i < size; i++ {
		val[i] = binary.LittleEndian.Uint64(buf)
		buf = buf[8:]
	}
	return val, nil
}
