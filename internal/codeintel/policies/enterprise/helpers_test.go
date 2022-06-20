package policies

import "sort"

func sortPolicyMatchesMap(policyMatches map[string][]PolicyMatch) {
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
