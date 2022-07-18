package gqltestutil

import "github.com/sourcegraph/sourcegraph/lib/errors"

type CreatePolicyInput struct {
	Name               string   `json:"name"`
	RepositoryPatterns []string `json:"repositoryPatterns"`

	Type    string `json:"type"`
	Pattern string `json:"pattern"`

	RetentionEnabled          bool `json:"retentionEnabled"`
	RetentionDurationHours    *int `json:"retentionDurationHours"`
	RetainIntermediateCommits bool `json:"retainIntermediateCommits"`

	IndexingEnabled          bool `json:"indexingEnabled"`
	IndexCommitMaxAgeHours   *int `json:"indexCommitMaxAgeHours"`
	IndexIntermediateCommits bool `json:"indexIntermediateCommits"`

	LockfileIndexingEnabled bool `json:"lockfileIndexingEnabled"`
}

// CreatePolicy creates a new code intelligence policy with the given input.
// It returns the GraphQL node ID of the newly created policy.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) CreatePolicy(input CreatePolicyInput) (string, error) {
	const query = `
mutation CreateCodeIntelligenceConfigurationPolicy($repositoryId: ID, $repositoryPatterns: [String!], $name: String!, $type: GitObjectType!, $pattern: String!, $retentionEnabled: Boolean!, $retentionDurationHours: Int, $retainIntermediateCommits: Boolean!, $indexingEnabled: Boolean!, $indexCommitMaxAgeHours: Int, $indexIntermediateCommits: Boolean!, $lockfileIndexingEnabled: Boolean!) {
  createCodeIntelligenceConfigurationPolicy(
    repository: $repositoryId
    repositoryPatterns: $repositoryPatterns
    name: $name
    type: $type
    pattern: $pattern
    retentionEnabled: $retentionEnabled
    retentionDurationHours: $retentionDurationHours
    retainIntermediateCommits: $retainIntermediateCommits
    indexingEnabled: $indexingEnabled
    indexCommitMaxAgeHours: $indexCommitMaxAgeHours
    indexIntermediateCommits: $indexIntermediateCommits
    lockfileIndexingEnabled: $lockfileIndexingEnabled
  ) {
    id
    __typename
  }
}
`
	variables := map[string]any{
		"name":               input.Name,
		"repositoryPatterns": input.RepositoryPatterns,

		"type":    input.Type,
		"pattern": input.Pattern,

		"retentionEnabled":          input.RetentionEnabled,
		"retentionDurationHours":    input.RetentionDurationHours,
		"retainIntermediateCommits": input.RetainIntermediateCommits,

		"indexingEnabled":          input.IndexingEnabled,
		"indexCommitMaxAgeHours":   input.IndexCommitMaxAgeHours,
		"indexIntermediateCommits": input.IndexIntermediateCommits,

		"lockfileIndexingEnabled": input.LockfileIndexingEnabled,
	}

	var resp struct {
		Data struct {
			CreateCodeIntelligenceConfigurationPolicy struct {
				ID string `json:"id"`
			} `json:"createCodeIntelligenceConfigurationPolicy"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return "", errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.CreateCodeIntelligenceConfigurationPolicy.ID, nil
}
