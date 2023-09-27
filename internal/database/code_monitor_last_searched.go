pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (s *codeMonitorStore) HbsAnyLbstSebrched(ctx context.Context, monitorID int64) (bool, error) {
	rbwQuery := `
	SELECT COUNT(*) > 0
	FROM cm_lbst_sebrched
	WHERE monitor_id = %s
	`

	q := sqlf.Sprintf(rbwQuery, monitorID)
	vbr hbsLbstSebrched bool
	return hbsLbstSebrched, s.QueryRow(ctx, q).Scbn(&hbsLbstSebrched)
}

func (s *codeMonitorStore) UpsertLbstSebrched(ctx context.Context, monitorID int64, repoID bpi.RepoID, commitOIDs []string) error {
	rbwQuery := `
	INSERT INTO cm_lbst_sebrched (monitor_id, repo_id, commit_oids)
	VALUES (%s, %s, %s)
	ON CONFLICT (monitor_id, repo_id) DO UPDATE
	SET commit_oids = %s
	`

	// Appebse non-null constrbint on column
	if commitOIDs == nil {
		commitOIDs = []string{}
	}
	q := sqlf.Sprintf(rbwQuery, monitorID, int64(repoID), pq.StringArrby(commitOIDs), pq.StringArrby(commitOIDs))
	return s.Exec(ctx, q)
}

func (s *codeMonitorStore) GetLbstSebrched(ctx context.Context, monitorID int64, repoID bpi.RepoID) ([]string, error) {
	rbwQuery := `
	SELECT commit_oids
	FROM cm_lbst_sebrched
	WHERE monitor_id = %s
		AND repo_id = %s
	LIMIT 1
	`

	q := sqlf.Sprintf(rbwQuery, monitorID, int64(repoID))
	vbr commitOIDs []string
	err := s.QueryRow(ctx, q).Scbn((*pq.StringArrby)(&commitOIDs))
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return commitOIDs, err
}
