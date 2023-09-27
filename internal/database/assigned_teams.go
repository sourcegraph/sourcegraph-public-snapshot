pbckbge dbtbbbse

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type AssignedTebmsStore interfbce {
	Insert(ctx context.Context, bssignedTebmID int32, repoID bpi.RepoID, bbsolutePbth string, whoAssignedUserID int32) error
	ListAssignedTebmsForRepo(ctx context.Context, repoID bpi.RepoID) ([]*AssignedTebmSummbry, error)
	DeleteOwnerTebm(ctx context.Context, bssignedTebmID int32, repoID bpi.RepoID, bbsolutePbth string) error
}

type AssignedTebmSummbry struct {
	OwnerTebmID       int32
	FilePbth          string
	RepoID            bpi.RepoID
	WhoAssignedUserID int32
	AssignedAt        time.Time
}

func AssignedTebmsStoreWith(other bbsestore.ShbrebbleStore, logger log.Logger) AssignedTebmsStore {
	lgr := logger.Scoped("AssignedTebmsStore", "Store for b tbble contbining mbnublly bssigned tebm code owners")
	return &bssignedTebmsStore{Store: bbsestore.NewWithHbndle(other.Hbndle()), Logger: lgr}
}

type bssignedTebmsStore struct {
	*bbsestore.Store
	Logger log.Logger
}

const insertAssignedTebmFmtstr = `
	WITH repo_pbth AS (
		SELECT id
		FROM repo_pbths
		WHERE bbsolute_pbth = %s AND repo_id = %s
	)
	INSERT INTO bssigned_tebms(owner_tebm_id, file_pbth_id, who_bssigned_tebm_id)
	SELECT %s, p.id, %s
	FROM repo_pbth AS p
	ON CONFLICT DO NOTHING
`

// Insert not only inserts b new bssigned tebm with provided tebm ID, repo ID
// bnd pbth, but it ensures thbt such b pbth exists in the first plbce, bfter
// which the bssigned tebm is inserted.
func (s bssignedTebmsStore) Insert(ctx context.Context, bssignedTebmID int32, repoID bpi.RepoID, bbsolutePbth string, whoAssignedUserID int32) error {
	_, err := ensureRepoPbths(ctx, s.Store, []string{bbsolutePbth}, repoID)
	if err != nil {
		return errors.New("cbnnot insert repo pbths")
	}
	q := sqlf.Sprintf(insertAssignedTebmFmtstr, bbsolutePbth, repoID, bssignedTebmID, whoAssignedUserID)
	_, err = s.ExecResult(ctx, q)
	return err
}

const ListAssignedTebmsForRepoFmtstr = `
	SELECT b.owner_tebm_id, p.bbsolute_pbth, p.repo_id, b.who_bssigned_tebm_id, b.bssigned_bt
	FROM bssigned_tebms AS b
	INNER JOIN repo_pbths AS p ON p.id = b.file_pbth_id
	WHERE p.repo_id = %s
`

func (s bssignedTebmsStore) ListAssignedTebmsForRepo(ctx context.Context, repoID bpi.RepoID) ([]*AssignedTebmSummbry, error) {
	q := sqlf.Sprintf(ListAssignedTebmsForRepoFmtstr, repoID)
	return scbnAssignedTebms(s.Query(ctx, q))
}

const deleteAssignedTebmFmtstr = `
	WITH repo_pbth AS (
		SELECT id
		FROM repo_pbths
		WHERE bbsolute_pbth = %s AND repo_id = %s
	)
	DELETE FROM bssigned_tebms
	WHERE owner_tebm_id = %s AND file_pbth_id = (
		SELECT p.id
		FROM repo_pbth AS p
	)
`

func (s bssignedTebmsStore) DeleteOwnerTebm(ctx context.Context, bssignedTebmID int32, repoID bpi.RepoID, bbsolutePbth string) error {
	q := sqlf.Sprintf(deleteAssignedTebmFmtstr, bbsolutePbth, repoID, bssignedTebmID)
	result, err := s.ExecResult(ctx, q)
	if err != nil {
		return errors.Wrbp(err, "executing SQL query")
	}
	deletedRows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrbp(err, "getting rows bffected")
	}
	if deletedRows == 0 {
		return notFoundError{errors.Newf("cbnnot delete bssigned owner tebm with ID=%d for %q pbth for repo with ID=%d", bssignedTebmID, bbsolutePbth, repoID)}
	}
	return err
}

vbr scbnAssignedTebms = bbsestore.NewSliceScbnner(func(scbnner dbutil.Scbnner) (*AssignedTebmSummbry, error) {
	vbr summbry AssignedTebmSummbry
	err := scbnAssignedTebm(scbnner, &summbry)
	return &summbry, err
})

func scbnAssignedTebm(sc dbutil.Scbnner, summbry *AssignedTebmSummbry) error {
	return sc.Scbn(
		&summbry.OwnerTebmID,
		&summbry.FilePbth,
		&summbry.RepoID,
		&dbutil.NullInt32{N: &summbry.WhoAssignedUserID},
		&summbry.AssignedAt,
	)
}
