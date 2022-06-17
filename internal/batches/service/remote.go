package service

import (
	"context"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ErrServerSideBatchChangesUnsupported = errors.New("server side batch changes are not available on this Sourcegraph instance")

func (svc *Service) areServerSideBatchChangesSupported() error {
	if !svc.features.ServerSideBatchChanges {
		return ErrServerSideBatchChangesUnsupported
	}
	return nil
}

const upsertEmptyBatchChangeQuery = `
mutation UpsertEmptyBatchChange(
	$name: String!
	$namespace: ID!
) {
	upsertEmptyBatchChange(
		name: $name,
		namespace: $namespace
	) {
		name
	}
}
`

func (svc *Service) UpsertBatchChange(
	ctx context.Context,
	name string,
	namespaceID string,
) (string, error) {
	if err := svc.areServerSideBatchChangesSupported(); err != nil {
		return "", err
	}

	var resp struct {
		UpsertEmptyBatchChange struct {
			Name string `json:"name"`
		} `json:"upsertEmptyBatchChange"`
	}

	if ok, err := svc.client.NewRequest(upsertEmptyBatchChangeQuery, map[string]interface{}{
		"name":      name,
		"namespace": namespaceID,
	}).Do(ctx, &resp); err != nil || !ok {
		return "", err
	}

	return resp.UpsertEmptyBatchChange.Name, nil
}

const createBatchSpecFromRawQuery = `
mutation CreateBatchSpecFromRaw(
    $batchSpec: String!,
    $namespace: ID!,
    $allowIgnored: Boolean!,
    $allowUnsupported: Boolean!,
    $noCache: Boolean!,
) {
    createBatchSpecFromRaw(
        batchSpec: $batchSpec,
        namespace: $namespace,
        allowIgnored: $allowIgnored,
        allowUnsupported: $allowUnsupported,
        noCache: $noCache,
    ) {
        id
    }
}
`

func (svc *Service) CreateBatchSpecFromRaw(
	ctx context.Context,
	batchSpec string,
	namespaceID string,
	allowIgnored bool,
	allowUnsupported bool,
	noCache bool,
) (string, error) {
	if err := svc.areServerSideBatchChangesSupported(); err != nil {
		return "", err
	}

	var resp struct {
		CreateBatchSpecFromRaw struct {
			ID string `json:"id"`
		} `json:"createBatchSpecFromRaw"`
	}

	if ok, err := svc.client.NewRequest(createBatchSpecFromRawQuery, map[string]interface{}{
		"batchSpec":        batchSpec,
		"namespace":        namespaceID,
		"allowIgnored":     allowIgnored,
		"allowUnsupported": allowUnsupported,
		"noCache":          noCache,
	}).Do(ctx, &resp); err != nil || !ok {
		return "", err
	}

	return resp.CreateBatchSpecFromRaw.ID, nil
}

const executeBatchSpecQuery = `
mutation ExecuteBatchSpec($batchSpec: ID!, $noCache: Boolean!) {
    executeBatchSpec(batchSpec: $batchSpec, noCache: $noCache) {
        id
    }
}
`

func (svc *Service) ExecuteBatchSpec(
	ctx context.Context,
	batchSpecID string,
	noCache bool,
) (string, error) {
	if err := svc.areServerSideBatchChangesSupported(); err != nil {
		return "", err
	}

	var resp struct {
		ExecuteBatchSpec struct {
			ID string `json:"id"`
		} `json:"executeBatchSpec"`
	}

	if ok, err := svc.client.NewRequest(executeBatchSpecQuery, map[string]interface{}{
		"batchSpec": batchSpecID,
		"noCache":   noCache,
	}).Do(ctx, &resp); err != nil || !ok {
		return "", err
	}

	return resp.ExecuteBatchSpec.ID, nil
}

const batchSpecWorkspaceResolutionQuery = `
query BatchSpecWorkspaceResolution($batchSpec: ID!) {
    node(id: $batchSpec) {
        ... on BatchSpec {
            workspaceResolution {
                failureMessage
                state
            }
        }
    }
}
`

type BatchSpecWorkspaceResolution struct {
	FailureMessage string `json:"failureMessage"`
	State          string `json:"state"`
}

func (svc *Service) GetBatchSpecWorkspaceResolution(ctx context.Context, id string) (*BatchSpecWorkspaceResolution, error) {
	if err := svc.areServerSideBatchChangesSupported(); err != nil {
		return nil, err
	}

	var resp struct {
		Node struct {
			WorkspaceResolution BatchSpecWorkspaceResolution `json:"workspaceResolution"`
		} `json:"node"`
	}

	if ok, err := svc.client.NewRequest(batchSpecWorkspaceResolutionQuery, map[string]interface{}{
		"batchSpec": id,
	}).Do(ctx, &resp); err != nil || !ok {
		return nil, err
	}

	return &resp.Node.WorkspaceResolution, nil
}
