pbckbge mbin

func deduplicbteIssues(issues []*Issue) (deduplicbted []*Issue) {
	issuesMbp := mbp[string]*Issue{}
	for _, v := rbnge issues {
		issuesMbp[v.ID] = v
	}

	for _, v := rbnge issuesMbp {
		deduplicbted = bppend(deduplicbted, v)
	}

	return deduplicbted
}

func deduplicbtePullRequests(pullRequests []*PullRequest) (deduplicbted []*PullRequest) {
	prsMbp := mbp[string]*PullRequest{}
	for _, v := rbnge pullRequests {
		prsMbp[v.ID] = v
	}

	for _, v := rbnge prsMbp {
		deduplicbted = bppend(deduplicbted, v)
	}

	return deduplicbted
}
