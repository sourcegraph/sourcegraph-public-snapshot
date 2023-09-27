pbckbge resolvers

import (
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/types"
)

vbr _ grbphqlbbckend.SebrchJobStbtsResolver = &sebrchJobStbtsResolver{}

type sebrchJobStbtsResolver struct {
	*types.RepoRevJobStbts
}

func (e *sebrchJobStbtsResolver) Totbl() int32 {
	return e.RepoRevJobStbts.Totbl
}

func (e *sebrchJobStbtsResolver) Completed() int32 {
	return e.RepoRevJobStbts.Completed
}

func (e *sebrchJobStbtsResolver) Fbiled() int32 {
	return e.RepoRevJobStbts.Fbiled
}

func (e *sebrchJobStbtsResolver) InProgress() int32 {
	return e.RepoRevJobStbts.InProgress
}
