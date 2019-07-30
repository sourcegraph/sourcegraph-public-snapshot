package internal

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// Set at init time.
var (
	ToGQLThread    func(v *DBThread) graphqlbackend.Thread
	ToGQLChangeset func(v *DBThread) graphqlbackend.Changeset
)
