package graphql

import "context"

type campaignsBackend struct {
	commonBackend
}

var _ Operations = &campaignsBackend{}

const applyCampaignMutation = `
mutation ApplyCampaign($campaignSpec: ID!) {
	applyCampaign(campaignSpec: $campaignSpec) {
		...campaignFields
	}
}

fragment campaignFields on Campaign {
    id
    namespace {
        ...namespaceFields
    }
    name
    description
    url
}
` + NamespaceFieldsFragment

func (cb *campaignsBackend) ApplyBatchChange(ctx context.Context, batchSpecID BatchSpecID) (*BatchChange, error) {
	var result struct {
		BatchChange *BatchChange `json:"applyCampaign"`
	}
	if ok, err := cb.newRequest(applyCampaignMutation, map[string]interface{}{
		"campaignSpec": batchSpecID,
	}).Do(ctx, &result); err != nil || !ok {
		return nil, err
	}
	return result.BatchChange, nil
}

const createCampaignSpecMutation = `
mutation CreateCampaignSpec(
    $namespace: ID!,
    $spec: String!,
    $changesetSpecs: [ID!]!
) {
    createCampaignSpec(
        namespace: $namespace, 
        campaignSpec: $spec,
        changesetSpecs: $changesetSpecs
    ) {
        id
        applyURL
    }
}
`

func (cb *campaignsBackend) CreateBatchSpec(ctx context.Context, namespace, spec string, changesetSpecIDs []ChangesetSpecID) (*CreateBatchSpecResponse, error) {
	var result struct {
		CreateCampaignSpec CreateBatchSpecResponse
	}
	if ok, err := cb.newRequest(createCampaignSpecMutation, map[string]interface{}{
		"namespace":      namespace,
		"spec":           spec,
		"changesetSpecs": changesetSpecIDs,
	}).Do(ctx, &result); err != nil || !ok {
		return nil, err
	}
	return &result.CreateCampaignSpec, nil
}
