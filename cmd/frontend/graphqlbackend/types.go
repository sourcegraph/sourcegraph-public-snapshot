package graphqlbackend

import (
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// Executor describes an executor instance that has recently connected to Sourcegraph.
type Executor = types.Executor

type ExecutorCompatibility string

const (
	ExecutorCompatibilityOutdated     ExecutorCompatibility = "OUTDATED"
	ExecutorCompatibilityUpToDate     ExecutorCompatibility = "UP_TO_DATE"
	ExecutorCompatibilityVersionAhead ExecutorCompatibility = "VERSION_AHEAD"
)

// ToGraphQL returns the GraphQL representation of the state.
func (c ExecutorCompatibility) ToGraphQL() *string {
	s := string(c)
	return &s
}
