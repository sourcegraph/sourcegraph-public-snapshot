// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package storage // import "go.opentelemetry.io/collector/extension/experimental/storage"

import "context"

type nopClient struct{}

var nopClientInstance Client = &nopClient{}

// NewNopClient returns a nop client
func NewNopClient() Client {
	return nopClientInstance
}

// Get does nothing, and returns nil, nil
func (c nopClient) Get(context.Context, string) ([]byte, error) {
	return nil, nil // no result, but no problem
}

// Set does nothing and returns nil
func (c nopClient) Set(context.Context, string, []byte) error {
	return nil // no problem
}

// Delete does nothing and returns nil
func (c nopClient) Delete(context.Context, string) error {
	return nil // no problem
}

// Close does nothing and returns nil
func (c nopClient) Close(context.Context) error {
	return nil
}

// Batch does nothing, and returns nil, nil
func (c nopClient) Batch(context.Context, ...Operation) error {
	return nil // no result, but no problem
}
