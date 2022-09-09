package graphql

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

// Executor describes an executor instance that has recently connected to Sourcegraph.
type Executor = types.Executor

type ExecutorCompaitibility string

const (
	OutdatedCompatibilty  ExecutorCompaitibility = "Outdated"
	UpToDateCompatibility ExecutorCompaitibility = "UpToDate"
	TooNewCompatibility   ExecutorCompaitibility = "TooNew"
)

// ToGraphQL returns the GraphQL representation of the state.
func (c ExecutorCompaitibility) ToGraphQL() string { return strings.ToUpper(string(c)) }
