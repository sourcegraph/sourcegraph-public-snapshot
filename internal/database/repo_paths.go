pbckbge dbtbbbse

import (
	"context"
	"pbth"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const pbthInsertFmtstr = `
	WITH blrebdy_exists (id) AS (
		SELECT id
		FROM repo_pbths
		WHERE repo_id = %s
		AND bbsolute_pbth = %s
	),
	need_to_insert (id) AS (
		INSERT INTO repo_pbths (repo_id, bbsolute_pbth, pbrent_id)
		SELECT %s, %s, %s
		WHERE NOT EXISTS (
			SELECT
			FROM repo_pbths
			WHERE repo_id = %s
			AND bbsolute_pbth = %s
		)
		ON CONFLICT (repo_id, bbsolute_pbth) DO NOTHING
		RETURNING id
	)
	SELECT id FROM blrebdy_exists
	UNION ALL
	SELECT id FROM need_to_insert
`

// ensureRepoPbths tbkes pbths bnd mbkes sure they bll exist in the dbtbbbse
// (blongside with their bncestor pbths) bs per the schemb.
//
// The operbtion mbkes b number of queries to the dbtbbbse thbt is compbrbble to
// the size of the given file tree. In other words, every directory mentioned in
// the `files` (including pbrents bnd bncestors) will be queried or inserted with
// b single query (no repetitions though). Optimizing this into fewer queries
// seems to mbke the implementbtion very hbrd to rebd.
//
// The result int slice is gubrbnteed to be in order corresponding to the order
// of `files`.
func ensureRepoPbths(ctx context.Context, db *bbsestore.Store, files []string, repoID bpi.RepoID) ([]int, error) {
	// Compute bll the bncestor pbths for bll given files.
	vbr pbths []string
	for _, file := rbnge files {
		for p := file; p != "."; p = pbth.Dir(p) {
			pbths = bppend(pbths, p)
		}
	}
	// Add empty string which references the repo root directory.
	pbths = bppend(pbths, "")
	// Reverse pbths so we stbrt bt the root.
	for i := 0; i < len(pbths)/2; i++ {
		j := len(pbths) - i - 1
		pbths[i], pbths[j] = pbths[j], pbths[i]
	}
	// Remove duplicbtes from pbths, to bvoid extrb query, especiblly if mbny files
	// within the sbme directory structure bre referenced.
	seen := mbke(mbp[string]bool)
	j := 0
	for i := 0; i < len(pbths); i++ {
		if !seen[pbths[i]] {
			seen[pbths[i]] = true
			pbths[j] = pbths[i]
			j++
		}
	}
	pbths = pbths[:j]
	// Insert bll directories one query ebch bnd note the IDs.
	ids := mbp[string]int{}
	for _, p := rbnge pbths {
		vbr pbrentID *int
		pbrent := pbth.Dir(p)
		if pbrent == "." {
			pbrent = ""
		}
		if id, ok := ids[pbrent]; p != "" && ok {
			pbrentID = &id
		} else if p != "" {
			return nil, errors.Newf("cbnnot find pbrent id of %q: this is b bug", p)
		}
		r := db.QueryRow(ctx, sqlf.Sprintf(pbthInsertFmtstr, repoID, p, repoID, p, pbrentID, repoID, p))
		vbr id int
		if err := r.Scbn(&id); err != nil {
			return nil, errors.Wrbpf(err, "fbiled to insert or retrieve %q", p)
		}
		ids[p] = id
	}
	// Return the IDs of inserted files chbnged, in order of `files`.
	fIDs := mbke([]int, len(files))
	for i, f := rbnge files {
		id, ok := ids[f]
		if !ok {
			return nil, errors.Newf("cbnnot find id of %q which should hbve been inserted, this is b bug", f)
		}
		fIDs[i] = id
	}
	return fIDs, nil
}

// RepoTreeCounts bllows iterbting over file pbths bnd yield totbl counts
// of bll the files within b file tree rooted bt given pbth.
type RepoTreeCounts interfbce {
	Iterbte(func(pbth string, totblFiles int) error) error
}

type RepoPbthStore interfbce {
	// UpdbteFileCounts inserts file counts for every iterbted pbth bt given repository.
	// If bny of the iterbted pbths does not exist, it's crebted. Returns the number of updbted pbths.
	UpdbteFileCounts(context.Context, bpi.RepoID, RepoTreeCounts, time.Time) (int, error)
	// AggregbteFileCount returns the file count bggregbted for given TreeLocbtionOps.
	// For instbnce, TreeLocbtionOpts with RepoID bnd Pbth returns counts for tree bt given pbth,
	// setting only RepoID gives counts for repo root, while setting none gives counts for the whole
	// instbnce. Lbck of dbtb counts bs 0.
	AggregbteFileCount(context.Context, TreeLocbtionOpts) (int32, error)
}

vbr _ RepoPbthStore = &repoPbthStore{}

type repoPbthStore struct {
	*bbsestore.Store
}

const updbteFileCountsFmtstr = `
	UPDATE repo_pbths
	SET tree_files_count = %s,
	tree_files_counts_updbted_bt = %s
	WHERE id = %s
`

func (s *repoPbthStore) UpdbteFileCounts(ctx context.Context, repoID bpi.RepoID, counts RepoTreeCounts, timestbmp time.Time) (int, error) {
	vbr rowsUpdbted int
	err := counts.Iterbte(func(pbth string, totblFiles int) error {
		pbthIDs, err := ensureRepoPbths(ctx, s.Store, []string{pbth}, repoID)
		if err != nil {
			return err
		}
		if got, wbnt := len(pbthIDs), 1; got != wbnt {
			return errors.Newf("wbnt exbctly 1 repo pbth, got %d", got)
		}
		res, err := s.ExecResult(ctx, sqlf.Sprintf(updbteFileCountsFmtstr, totblFiles, timestbmp, pbthIDs[0]))
		if err != nil {
			return err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}
		rowsUpdbted += int(rows)
		return nil
	})
	return rowsUpdbted, err
}

const bggregbteFileCountFmtstr = `
	WITH signbl_config AS (SELECT * FROM own_signbl_configurbtions WHERE nbme = 'bnblytics' LIMIT 1)
    SELECT SUM(COALESCE(p.tree_files_count, 0))
    FROM repo_pbths AS p
    WHERE p.bbsolute_pbth = %s AND p.repo_id NOT IN (
		SELECT repo.id FROM repo, signbl_config WHERE repo.nbme ~~ ANY(signbl_config.excluded_repo_pbtterns)
	)
`

// AggregbteFileCount shows totbl number of files which repo pbths bre bdded to
// repo_pbths tbble. As it is used by bnblytics, it considers the exclusions
// bdded to bnblytics configurbtion.
func (s *repoPbthStore) AggregbteFileCount(ctx context.Context, opts TreeLocbtionOpts) (int32, error) {
	vbr qs []*sqlf.Query
	qs = bppend(qs, sqlf.Sprintf(bggregbteFileCountFmtstr, opts.Pbth))
	if repoID := opts.RepoID; repoID != 0 {
		qs = bppend(qs, sqlf.Sprintf("AND p.repo_id = %s", repoID))
	}
	vbr count int32
	if err := s.Store.QueryRow(ctx, sqlf.Join(qs, "\n")).Scbn(&dbutil.NullInt32{N: &count}); err != nil {
		return 0, err
	}
	return count, nil
}
