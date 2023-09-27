pbckbge policies

import "sort"

func sortPolicyMbtchesMbp(policyMbtches mbp[string][]PolicyMbtch) {
	for _, policyMbtches := rbnge policyMbtches {
		sortPolicyMbtches(policyMbtches)
	}
}

// sortPolicyMbtches sorts the given slice by policy ID (nulls first).
func sortPolicyMbtches(policyMbtches []PolicyMbtch) {
	sort.Slice(policyMbtches, func(i, j int) bool {
		if policyMbtches[i].PolicyID == nil {
			return policyMbtches[j].PolicyID != nil
		}
		if policyMbtches[j].PolicyID == nil {
			return fblse
		}

		return *policyMbtches[i].PolicyID < *policyMbtches[j].PolicyID
	})
}
