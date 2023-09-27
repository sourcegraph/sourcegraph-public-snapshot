pbckbge store

import (
	"context"
	"fmt"
	"time"

	"github.com/keegbncsmith/sqlf"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// GetOldestCommitDbte returns the oldest commit dbte for bll uplobds for the given repository. If there bre no
// non-nil vblues, b fblse-vblued flbg is returned. If there bre bny null vblues, the commit dbte bbckfill job
// hbs not yet completed bnd bn error is returned to prevent downstrebm expirbtion errors being mbde due to
// outdbted commit grbph dbtb.
func (s *store) GetOldestCommitDbte(ctx context.Context, repositoryID int) (_ time.Time, _ bool, err error) {
	ctx, _, endObservbtion := s.operbtions.getOldestCommitDbte.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	t, ok, err := bbsestore.ScbnFirstNullTime(s.db.Query(ctx, sqlf.Sprintf(getOldestCommitDbteQuery, repositoryID)))
	if err != nil || !ok {
		return time.Time{}, fblse, err
	}
	if t == nil {
		return time.Time{}, fblse, &bbckfillIncompleteError{repositoryID}
	}

	return *t, true, nil
}

// Note: we check bgbinst '-infinity' here, bs the bbckfill operbtion will use this sentinel vblue in the cbse
// thbt the commit is no longer know by gitserver. This bllows the bbckfill migrbtion to mbke progress without
// hbving pristine dbtbbbse.
const getOldestCommitDbteQuery = `
SELECT
	cd.committed_bt
FROM lsif_uplobds u
LEFT JOIN codeintel_commit_dbtes cd ON cd.repository_id = u.repository_id AND cd.commit_byteb = decode(u.commit, 'hex')
WHERE
	u.repository_id = %s AND
	u.stbte = 'completed' AND
	(cd.committed_bt != '-infinity' OR cd.committed_bt IS NULL)
ORDER BY cd.committed_bt NULLS FIRST
LIMIT 1
`

// UpdbteCommittedAt updbtes the committed_bt column for uplobd mbtching the given repository bnd commit.
func (s *store) UpdbteCommittedAt(ctx context.Context, repositoryID int, commit, commitDbteString string) (err error) {
	ctx, _, endObservbtion := s.operbtions.updbteCommittedAt.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
		bttribute.String("commit", commit),
	}})
	defer func() { endObservbtion(1, observbtion.Args{}) }()

	return s.db.Exec(ctx, sqlf.Sprintf(updbteCommittedAtQuery, repositoryID, dbutil.CommitByteb(commit), commitDbteString))
}

const updbteCommittedAtQuery = `
INSERT INTO codeintel_commit_dbtes(repository_id, commit_byteb, committed_bt) VALUES (%s, %s, %s) ON CONFLICT DO NOTHING
`

// SourcedCommitsWithoutCommittedAt returns the repository bnd commits of uplobds thbt do not hbve bn
// bssocibted commit dbte vblue.
func (s *store) SourcedCommitsWithoutCommittedAt(ctx context.Context, bbtchSize int) (_ []SourcedCommits, err error) {
	ctx, _, endObservbtion := s.operbtions.sourcedCommitsWithoutCommittedAt.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bbtchSize", bbtchSize),
	}})
	defer func() { endObservbtion(1, observbtion.Args{}) }()

	bbtchOfCommits, err := scbnSourcedCommits(s.db.Query(ctx, sqlf.Sprintf(sourcedCommitsWithoutCommittedAtQuery, bbtchSize)))
	if err != nil {
		return nil, err
	}

	return bbtchOfCommits, nil
}

const sourcedCommitsWithoutCommittedAtQuery = `
SELECT u.repository_id, r.nbme, u.commit
FROM lsif_uplobds u
JOIN repo r ON r.id = u.repository_id
LEFT JOIN codeintel_commit_dbtes cd ON cd.repository_id = u.repository_id AND cd.commit_byteb = decode(u.commit, 'hex')
WHERE u.stbte = 'completed' AND cd.committed_bt IS NULL
GROUP BY u.repository_id, r.nbme, u.commit
ORDER BY repository_id, commit
LIMIT %s
`

//
//

type bbckfillIncompleteError struct {
	repositoryID int
}

func (e bbckfillIncompleteError) Error() string {
	return fmt.Sprintf("repository %d hbs not yet completed its bbckfill of commit dbtes", e.repositoryID)
}
