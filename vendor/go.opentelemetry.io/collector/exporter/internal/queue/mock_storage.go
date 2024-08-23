// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package queue // import "go.opentelemetry.io/collector/exporter/internal/queue"

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"syscall"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension/experimental/storage"
)

type mockStorageExtension struct {
	component.StartFunc
	component.ShutdownFunc
	st             sync.Map
	getClientError error
}

func (m *mockStorageExtension) GetClient(context.Context, component.Kind, component.ID, string) (storage.Client, error) {
	if m.getClientError != nil {
		return nil, m.getClientError
	}
	return &mockStorageClient{st: &m.st, closed: &atomic.Bool{}}, nil
}

func NewMockStorageExtension(getClientError error) storage.Extension {
	return &mockStorageExtension{getClientError: getClientError}
}

type mockStorageClient struct {
	st     *sync.Map
	closed *atomic.Bool
}

func (m *mockStorageClient) Get(ctx context.Context, s string) ([]byte, error) {
	getOp := storage.GetOperation(s)
	err := m.Batch(ctx, getOp)
	return getOp.Value, err
}

func (m *mockStorageClient) Set(ctx context.Context, s string, bytes []byte) error {
	return m.Batch(ctx, storage.SetOperation(s, bytes))
}

func (m *mockStorageClient) Delete(ctx context.Context, s string) error {
	return m.Batch(ctx, storage.DeleteOperation(s))
}

func (m *mockStorageClient) Close(context.Context) error {
	m.closed.Store(true)
	return nil
}

func (m *mockStorageClient) Batch(_ context.Context, ops ...storage.Operation) error {
	if m.isClosed() {
		panic("client already closed")
	}
	for _, op := range ops {
		switch op.Type {
		case storage.Get:
			val, found := m.st.Load(op.Key)
			if !found {
				break
			}
			op.Value = val.([]byte)
		case storage.Set:
			m.st.Store(op.Key, op.Value)
		case storage.Delete:
			m.st.Delete(op.Key)
		default:
			return errors.New("wrong operation type")
		}
	}

	return nil
}

func (m *mockStorageClient) isClosed() bool {
	return m.closed.Load()
}

func newFakeBoundedStorageClient(maxSizeInBytes int) *fakeBoundedStorageClient {
	return &fakeBoundedStorageClient{
		st:             map[string][]byte{},
		MaxSizeInBytes: maxSizeInBytes,
	}
}

// this storage client mimics the behavior of actual storage engines with limited storage space available
// in general, real storage engines often have a per-write-transaction storage overhead, needing to keep
// both the old and the new value stored until the transaction is committed
// this is useful for testing the persistent queue queue behavior with a full disk
type fakeBoundedStorageClient struct {
	MaxSizeInBytes int
	st             map[string][]byte
	sizeInBytes    int
	mux            sync.Mutex
}

func (m *fakeBoundedStorageClient) Get(ctx context.Context, key string) ([]byte, error) {
	op := storage.GetOperation(key)
	if err := m.Batch(ctx, op); err != nil {
		return nil, err
	}

	return op.Value, nil
}

func (m *fakeBoundedStorageClient) Set(ctx context.Context, key string, value []byte) error {
	return m.Batch(ctx, storage.SetOperation(key, value))
}

func (m *fakeBoundedStorageClient) Delete(ctx context.Context, key string) error {
	return m.Batch(ctx, storage.DeleteOperation(key))
}

func (m *fakeBoundedStorageClient) Close(context.Context) error {
	return nil
}

func (m *fakeBoundedStorageClient) Batch(_ context.Context, ops ...storage.Operation) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	totalAdded, totalRemoved := m.getTotalSizeChange(ops)

	// the assumption here is that the new data needs to coexist with the old data on disk
	// for the transaction to succeed
	// this seems to be true for the file storage extension at least
	if m.sizeInBytes+totalAdded > m.MaxSizeInBytes {
		return fmt.Errorf("insufficient space available: %w", syscall.ENOSPC)
	}

	for _, op := range ops {
		switch op.Type {
		case storage.Get:
			op.Value = m.st[op.Key]
		case storage.Set:
			m.st[op.Key] = op.Value
		case storage.Delete:
			delete(m.st, op.Key)
		default:
			return errors.New("wrong operation type")
		}
	}

	m.sizeInBytes += totalAdded - totalRemoved

	return nil
}

func (m *fakeBoundedStorageClient) SetMaxSizeInBytes(newMaxSize int) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.MaxSizeInBytes = newMaxSize
}

func (m *fakeBoundedStorageClient) GetSizeInBytes() int {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.sizeInBytes
}

func (m *fakeBoundedStorageClient) getTotalSizeChange(ops []storage.Operation) (totalAdded int, totalRemoved int) {
	totalAdded, totalRemoved = 0, 0
	for _, op := range ops {
		switch op.Type {
		case storage.Set:
			if oldValue, ok := m.st[op.Key]; ok {
				totalRemoved += len(oldValue)
			} else {
				totalAdded += len(op.Key)
			}
			totalAdded += len(op.Value)
		case storage.Delete:
			if value, ok := m.st[op.Key]; ok {
				totalRemoved += len(op.Key)
				totalRemoved += len(value)
			}
		default:
		}
	}
	return totalAdded, totalRemoved
}

func newFakeStorageClientWithErrors(errors []error) *fakeStorageClientWithErrors {
	return &fakeStorageClientWithErrors{
		errors: errors,
	}
}

// this storage client just returns errors from a list in order
// used for testing error handling
type fakeStorageClientWithErrors struct {
	errors         []error
	nextErrorIndex int
	mux            sync.Mutex
}

func (m *fakeStorageClientWithErrors) Get(ctx context.Context, key string) ([]byte, error) {
	op := storage.GetOperation(key)
	err := m.Batch(ctx, op)
	if err != nil {
		return nil, err
	}

	return op.Value, nil
}

func (m *fakeStorageClientWithErrors) Set(ctx context.Context, key string, value []byte) error {
	return m.Batch(ctx, storage.SetOperation(key, value))
}

func (m *fakeStorageClientWithErrors) Delete(ctx context.Context, key string) error {
	return m.Batch(ctx, storage.DeleteOperation(key))
}

func (m *fakeStorageClientWithErrors) Close(context.Context) error {
	return nil
}

func (m *fakeStorageClientWithErrors) Batch(context.Context, ...storage.Operation) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if m.nextErrorIndex >= len(m.errors) {
		return nil
	}

	m.nextErrorIndex++
	return m.errors[m.nextErrorIndex-1]
}

func (m *fakeStorageClientWithErrors) Reset() {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.nextErrorIndex = 0
}
