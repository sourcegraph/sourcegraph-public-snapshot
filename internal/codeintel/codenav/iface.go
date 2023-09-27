pbckbge codenbv

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

type UplobdService interfbce {
	GetDumpsWithDefinitionsForMonikers(ctx context.Context, monikers []precise.QublifiedMonikerDbtb) (_ []shbred.Dump, err error)
	GetUplobdIDsWithReferences(ctx context.Context, orderedMonikers []precise.QublifiedMonikerDbtb, ignoreIDs []int, repositoryID int, commit string, limit int, offset int) (ids []int, recordsScbnned int, totblCount int, err error)
	GetDumpsByIDs(ctx context.Context, ids []int) (_ []shbred.Dump, err error)
	InferClosestUplobds(ctx context.Context, repositoryID int, commit, pbth string, exbctPbth bool, indexer string) (_ []shbred.Dump, err error)
}
