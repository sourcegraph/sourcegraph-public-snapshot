package graphql

import (
	"context"

	"github.com/sourcegraph/src-cli/internal/api"
)

// Operations defines queries and mutations that are used by src-cli for Batch
// Change operations, and that vary between the old Campaigns world and the new
// Batch Changes world.
//
// TODO(campaigns-deprecation): this can be removed in Sourcegraph 4.0.
type Operations interface {
	ApplyBatchChange(ctx context.Context, batchSpecID BatchSpecID) (*BatchChange, error)
	CreateBatchSpec(ctx context.Context, namespace, spec string, changesetSpecIDs []ChangesetSpecID) (*CreateBatchSpecResponse, error)
}

type BatchSpecID string
type ChangesetSpecID string

type CreateBatchSpecResponse struct {
	ID       BatchSpecID
	ApplyURL string
}

func NewOperations(client api.Client, batchChanges, useGzipCompression bool) Operations {
	backend := commonBackend{
		client:             client,
		useGzipCompression: useGzipCompression,
	}

	if batchChanges {
		return &batchesBackend{backend}
	}

	return &campaignsBackend{backend}

}
