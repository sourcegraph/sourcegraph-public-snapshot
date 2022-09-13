package graphql

import (
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// Executor describes an executor instance that has recently connected to Sourcegraph.
type Executor = types.Executor

type ExecutorCompaitibility string

const (
	OutdatedCompatibilty      ExecutorCompaitibility = "OUTDATED"
	UpToDateCompatibility     ExecutorCompaitibility = "UOTODATE"
	VersionAheadCompatibility ExecutorCompaitibility = "VERSION_AHEAD"
)

// ToGraphQL returns the GraphQL representation of the state.
func (c ExecutorCompaitibility) ToGraphQL() string { return string(c) }
