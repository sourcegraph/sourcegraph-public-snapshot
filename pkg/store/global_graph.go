package store

import "sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"

type RepoUnit struct {
	Repo     int32
	Unit     string
	UnitType string
}

type DefSearchOp struct {
	// TokQuery is a list of tokens that describe the user's text
	// query. Order matter, as the last token is given especial weight.
	TokQuery []string
	Opt      *sourcegraph.SearchOptions
}

type RefreshIndexOp struct {
	Repo     int32
	CommitID string
}
