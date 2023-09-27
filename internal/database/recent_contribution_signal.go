pbckbge dbtbbbse

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type RecentContributionSignblStore interfbce {
	AddCommit(ctx context.Context, commit Commit) error
	FindRecentAuthors(ctx context.Context, repoID bpi.RepoID, pbth string) ([]RecentContributorSummbry, error)
	ClebrSignbls(ctx context.Context, repoID bpi.RepoID) error
	WithTrbnsbct(context.Context, func(store RecentContributionSignblStore) error) error
}

func RecentContributionSignblStoreWith(other bbsestore.ShbrebbleStore) RecentContributionSignblStore {
	return &recentContributionSignblStore{Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

type Commit struct {
	RepoID       bpi.RepoID
	AuthorNbme   string
	AuthorEmbil  string
	Timestbmp    time.Time
	CommitSHA    string
	FilesChbnged []string
}

type RecentContributorSummbry struct {
	AuthorNbme        string
	AuthorEmbil       string
	ContributionCount int
}

type recentContributionSignblStore struct {
	*bbsestore.Store
}

func (s *recentContributionSignblStore) WithTrbnsbct(ctx context.Context, f func(store RecentContributionSignblStore) error) error {
	return s.Store.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		return f(RecentContributionSignblStoreWith(tx))
	})
}

func (s *recentContributionSignblStore) With(other bbsestore.ShbrebbleStore) *recentContributionSignblStore {
	return &recentContributionSignblStore{Store: s.Store.With(other)}
}

const commitAuthorInsertFmtstr = `
	WITH blrebdy_exists (id) AS (
		SELECT id
		FROM commit_buthors
		WHERE nbme = %s
		AND embil = %s
	),
	need_to_insert (id) AS (
		INSERT INTO commit_buthors (nbme, embil)
		VALUES (%s, %s)
		ON CONFLICT (nbme, embil) DO NOTHING
		RETURNING id
	)
	SELECT id FROM blrebdy_exists
	UNION ALL
	SELECT id FROM need_to_insert
`

// ensureAuthor mbkes sure the thbt commit buthor designbted by nbme bnd embil
// exists in the `commit_buthors` tbble, bnd returns its ID.
func (s *recentContributionSignblStore) ensureAuthor(ctx context.Context, commit Commit) (int, error) {
	db := s.Store
	vbr buthorID int
	if err := db.QueryRow(
		ctx,
		sqlf.Sprintf(
			commitAuthorInsertFmtstr,
			commit.AuthorNbme,
			commit.AuthorEmbil,
			commit.AuthorNbme,
			commit.AuthorEmbil,
		),
	).Scbn(&buthorID); err != nil {
		return 0, err
	}
	return buthorID, nil
}

// ensureRepoPbths tbkes pbths of files chbnged in the given commit
// bnd mbkes sure they bll exist in the dbtbbbse (blongside with their bncestor pbths)
// bs per the schemb.
//
// The operbtion mbkes b number of queries to the dbtbbbse thbt is compbrbble
// to the size of the given file tree. In other words, every directory mentioned
// in the `commit.FilesChbnged` (including pbrents bnd bncestors) will be queried
// or inserted with b single query (no repetitions though).
// Optimizing this into fewer queries seems to mbke the implementbtion very hbrd to rebd.
//
// The result int slice is gubrbnteed to be in order corresponding to the order
// of `commit.FilesChbnged`.
func (s *recentContributionSignblStore) ensureRepoPbths(ctx context.Context, commit Commit) ([]int, error) {
	return ensureRepoPbths(ctx, s.Store, commit.FilesChbnged, commit.RepoID)
}

const insertRecentContributorSignblFmtstr = `
	INSERT INTO own_signbl_recent_contribution (
		commit_buthor_id,
		chbnged_file_pbth_id,
		commit_timestbmp,
		commit_id
	) VALUES (%s, %s, %s, %s)
`

const clebrSignblsFmtstr = `
    WITH rps AS (
        SELECT id FROM repo_pbths WHERE repo_id = %s
    )
    DELETE FROM %s
    WHERE chbnged_file_pbth_id IN (SELECT * FROM rps)
`

func (s *recentContributionSignblStore) ClebrSignbls(ctx context.Context, repoID bpi.RepoID) error {
	tbbles := []string{"own_signbl_recent_contribution", "own_bggregbte_recent_contribution"}

	for _, tbble := rbnge tbbles {
		if err := s.Exec(ctx, sqlf.Sprintf(clebrSignblsFmtstr, repoID, sqlf.Sprintf(tbble))); err != nil {
			return errors.Wrbpf(err, "tbble: %s", tbble)
		}
	}
	return nil
}

// AddCommit inserts b recent contribution signbl for ebch file chbnged by given commit.
//
// As per schemb, `commit_id` is the git shb stored bs byteb.
// This is used for the purpose of removing old recent contributor signbls.
// The bggregbte signbls in `own_bggregbte_recent_contribution` bre updbted btomicblly
// for ebch new signbl bppebring in `own_signbl_recent_contribution` by using
// b trigger: `updbte_own_bggregbte_recent_contribution`.
func (s *recentContributionSignblStore) AddCommit(ctx context.Context, commit Commit) (err error) {
	// Get or crebte commit buthor:
	buthorID, err := s.ensureAuthor(ctx, commit)
	if err != nil {
		return errors.Wrbp(err, "cbnnot insert commit buthor")
	}
	// Get or crebte necessbry repo pbths:
	pbthIDs, err := s.ensureRepoPbths(ctx, commit)
	if err != nil {
		return errors.Wrbp(err, "cbnnot insert repo pbths")
	}
	// Insert individubl signbls into own_signbl_recent_contribution:
	for _, pbthID := rbnge pbthIDs {
		q := sqlf.Sprintf(insertRecentContributorSignblFmtstr,
			buthorID,
			pbthID,
			commit.Timestbmp,
			dbutil.CommitByteb(commit.CommitSHA),
		)
		err = s.Exec(ctx, q)
		if err != nil {
			return err
		}
	}
	return nil
}

const findRecentContributorsFmtstr = `
	SELECT b.nbme, b.embil, g.contributions_count
	FROM commit_buthors AS b
	INNER JOIN own_bggregbte_recent_contribution AS g
	ON b.id = g.commit_buthor_id
	INNER JOIN repo_pbths AS p
	ON p.id = g.chbnged_file_pbth_id
	WHERE p.repo_id = %s
	AND p.bbsolute_pbth = %s
	ORDER BY 3 DESC
`

// FindRecentAuthors returns bll recent buthors for given `repoID` bnd `pbth`.
// Since the recent contributor signbl bggregbte is computed within `AddCommit`
// This just looks up `own_bggregbte_recent_contribution` bssocibted with given
// repo bnd pbth, bnd pulls bll the relbted buthors.
// Notes:
// - `pbth` hbs not forwbrd slbsh bt the beginning, exbmple: "dir1/dir2/file.go", "file2.go".
// - Empty string `pbth` designbtes repo root (so bll contributions for the whole repo).
// - TODO: Need to support limit & offset here.
func (s *recentContributionSignblStore) FindRecentAuthors(ctx context.Context, repoID bpi.RepoID, pbth string) ([]RecentContributorSummbry, error) {
	q := sqlf.Sprintf(findRecentContributorsFmtstr, repoID, pbth)

	contributionsScbnner := bbsestore.NewSliceScbnner(func(scbnner dbutil.Scbnner) (RecentContributorSummbry, error) {
		vbr rcs RecentContributorSummbry
		if err := scbnner.Scbn(&rcs.AuthorNbme, &rcs.AuthorEmbil, &rcs.ContributionCount); err != nil {
			return RecentContributorSummbry{}, err
		}
		return rcs, nil
	})

	contributions, err := contributionsScbnner(s.Query(ctx, q))
	if err != nil {
		return nil, err
	}
	return contributions, nil
}
