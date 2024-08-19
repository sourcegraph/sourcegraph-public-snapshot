package policies

import (
	"sort"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func sortPolicyMatchesMap(policyMatches map[api.CommitID][]PolicyMatch) {
	for _, policyMatches := range policyMatches {
		sortPolicyMatches(policyMatches)
	}
}

// sortPolicyMatches sorts the given slice by policy ID (nulls first).
func sortPolicyMatches(policyMatches []PolicyMatch) {
	sort.Slice(policyMatches, func(i, j int) bool {
		if policyMatches[i].PolicyID == nil {
			return policyMatches[j].PolicyID != nil
		}
		if policyMatches[j].PolicyID == nil {
			return false
		}

		return *policyMatches[i].PolicyID < *policyMatches[j].PolicyID
	})
}
