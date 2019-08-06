package internal

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// Set at init time.
var (
	ToGQLThread    func(v *DBThread) graphqlbackend.Thread
	ToGQLIssue     func(v *DBThread) graphqlbackend.Issue
	ToGQLChangeset func(v *DBThread) graphqlbackend.Changeset
)
