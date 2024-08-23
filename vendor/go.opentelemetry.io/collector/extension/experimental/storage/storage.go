// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package storage // import "go.opentelemetry.io/collector/extension/experimental/storage"

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
)

// Extension is the interface that storage extensions must implement
type Extension interface {
	extension.Extension

	// GetClient will create a client for use by the specified component.
	// Each component can have multiple storages (e.g. one for each signal),
	// which can be identified using storageName parameter.
	// The component can use the client to manage state
	GetClient(ctx context.Context, kind component.Kind, id component.ID, storageName string) (Client, error)
}

// Client is the interface that storage clients must implement
// All methods should return error only if a problem occurred.
// This mirrors the behavior of a golang map:
//   - Set doesn't error if a key already exists - it just overwrites the value.
//   - Get doesn't error if a key is not found - it just returns nil.
//   - Delete doesn't error if the key doesn't exist - it just no-ops.
//
// Similarly:
//   - Batch doesn't error if any of the above happens for either retrieved or updated keys
//
// This also provides a way to differentiate data operations
//
//	[overwrite | not-found | no-op] from "real" problems
type Client interface {

	// Get will retrieve data from storage that corresponds to the
	// specified key. It should return (nil, nil) if not found
	Get(ctx context.Context, key string) ([]byte, error)

	// Set will store data. The data can be retrieved by the same
	// component after a process restart, using the same key
	Set(ctx context.Context, key string, value []byte) error

	// Delete will delete data associated with the specified key
	Delete(ctx context.Context, key string) error

	// Batch handles specified operations in batch. Get operation results are put in-place
	Batch(ctx context.Context, ops ...Operation) error

	// Close will release any resources held by the client
	Close(ctx context.Context) error
}

type opType int

const (
	Get opType = iota
	Set
	Delete
)

type operation struct {
	// Key specifies key which is going to be get/set/deleted
	Key string
	// Value specifies value that is going to be set or holds result of get operation
	Value []byte
	// Type describes the operation type
	Type opType
}

type Operation *operation

func SetOperation(key string, value []byte) Operation {
	return &operation{
		Key:   key,
		Value: value,
		Type:  Set,
	}
}

func GetOperation(key string) Operation {
	return &operation{
		Key:  key,
		Type: Get,
	}
}

func DeleteOperation(key string) Operation {
	return &operation{
		Key:  key,
		Type: Delete,
	}
}
