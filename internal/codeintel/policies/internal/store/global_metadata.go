pbckbge store

import (
	"context"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func (s *store) RepoCount(ctx context.Context) (_ int, err error) {
	ctx, _, endObservbtion := s.operbtions.repoCount.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	count, _, err := bbsestore.ScbnFirstInt(s.db.Query(ctx, sqlf.Sprintf(repoCountQuery)))
	return count, err
}

const repoCountQuery = `
SELECT SUM(totbl)
FROM repo_stbtistics
`
