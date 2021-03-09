package graphql

import (
	"context"
)

type batchesBackend struct {
	commonBackend
}

var _ Operations = &batchesBackend{}

const applyBatchChangeMutation = `
mutation ApplyBatchChange($batchSpec: ID!) {
	applyBatchChange(batchSpec: $batchSpec) {
		...batchChangeFields
	}
}

fragment batchChangeFields on BatchChange {
    id
    namespace {
        ...namespaceFields
    }
    name
    description
    url
}
` + NamespaceFieldsFragment

func (bb *batchesBackend) ApplyBatchChange(ctx context.Context, batchSpecID BatchSpecID) (*BatchChange, error) {
	var result struct {
		BatchChange *BatchChange `json:"applyBatchChange"`
	}
	if ok, err := bb.newRequest(applyBatchChangeMutation, map[string]interface{}{
		"batchSpec": batchSpecID,
	}).Do(ctx, &result); err != nil || !ok {
		return nil, err
	}
	return result.BatchChange, nil
}

const createBatchSpecMutation = `
mutation CreateBatchSpec(
    $namespace: ID!,
    $spec: String!,
    $changesetSpecs: [ID!]!
) {
    createBatchSpec(
        namespace: $namespace, 
        batchSpec: $spec,
        changesetSpecs: $changesetSpecs
    ) {
        id
        applyURL
    }
}
`

func (bb *batchesBackend) CreateBatchSpec(ctx context.Context, namespace, spec string, changesetSpecIDs []ChangesetSpecID) (*CreateBatchSpecResponse, error) {
	var result struct {
		CreateBatchSpec CreateBatchSpecResponse
	}
	if ok, err := bb.newRequest(createBatchSpecMutation, map[string]interface{}{
		"namespace":      namespace,
		"spec":           spec,
		"changesetSpecs": changesetSpecIDs,
	}).Do(ctx, &result); err != nil || !ok {
		return nil, err
	}
	return &result.CreateBatchSpec, nil
}
