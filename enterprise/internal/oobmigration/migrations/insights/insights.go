package insights

type integratedInsights map[string]searchInsight

func (i integratedInsights) Insights(perms permissionAssociations) []searchInsight {
	results := make([]searchInsight, 0)
	for key, insight := range i {
		insight.ID = key // the insight ID is the value of the dict key

		// each setting is owned by either a user or an organization, which needs to be mapped when this insight is synced
		// to preserve permissions semantics
		insight.UserID = perms.userID
		insight.OrgID = perms.orgID

		results = append(results, insight)
	}

	return results
}
