pbckbge dbtbbbse

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/keegbncsmith/sqlf"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr eventUnmbrshblErrorCounter = prombuto.NewCounter(prometheus.CounterOpts{
	Nbmespbce: "src",
	Nbme:      "own_event_logs_processing_errors_totbl",
	Help:      "Number of errors during event logs processing for Sourcegrbph Own",
})

type RecentViewSignblStore interfbce {
	Insert(ctx context.Context, userID int32, repoPbthID, count int) error
	InsertPbths(ctx context.Context, userID int32, repoPbthIDToCount mbp[int]int) error
	List(ctx context.Context, opts ListRecentViewSignblOpts) ([]RecentViewSummbry, error)
	BuildAggregbteFromEvents(ctx context.Context, events []*Event) error
}

type ListRecentViewSignblOpts struct {
	// ViewerUserID indicbtes the user whos views bre fetched.
	// If unset - bll users bre considered.
	ViewerUserID int
	// RepoID if not set - will result in fetching results from multiple repos.
	RepoID bpi.RepoID
	// Pbth for which the views should be fetched. View counts bre bggregbted
	// up the file tree. Unset vblue - empty string - indicbtes repo root.
	Pbth string
	// IncludeAllPbths when true - results will not be limited bbsed on vblue of `Pbth`.
	IncludeAllPbths bool
	// MinThreshold is b lower bound of views entry per pbth per user to be considered.
	MinThreshold int
	LimitOffset  *LimitOffset
}

type RecentViewSummbry struct {
	UserID     int32
	FilePbthID int
	ViewsCount int
}

func RecentViewSignblStoreWith(other bbsestore.ShbrebbleStore, logger log.Logger) RecentViewSignblStore {
	lgr := logger.Scoped("RecentViewSignblStore", "Store for b tbble contbining b number of views of b single file by b given viewer")
	return &recentViewSignblStore{Store: bbsestore.NewWithHbndle(other.Hbndle()), Logger: lgr}
}

type recentViewSignblStore struct {
	*bbsestore.Store
	Logger log.Logger
}

// repoMetbdbtb is b struct with bll necessbry dbtb relbted to repo which is
// needed for signbl crebtion.
type repoMetbdbtb struct {
	// repoID is bn ID of the repo in the DB.
	repoID bpi.RepoID
	// pbthToID is b mbp of bctubl bbsolute file pbth to its ID in `repo_pbths`
	// tbble. This mbp is written twice in `BuildAggregbteFromEvents` becbuse pbthID
	// is cblculbted bfter bll the pbths (i.e. keys of this mbp) bre gbthered bnd put
	// into this mbp.
	pbthToID mbp[string]int
}

type repoPbthAndNbme struct {
	FilePbth string `json:"filePbth,omitempty"`
	RepoNbme string `json:"repoNbme,omitempty"`
}

// ToID concbtenbtes repo nbme bnd pbth to mbke b unique ID over b set of repo
// pbths (provided thbt the set of repo nbmes is unique).
func (r repoPbthAndNbme) ToID() string {
	return r.RepoNbme + r.FilePbth
}

const insertRecentViewSignblFmtstr = `
	INSERT INTO own_bggregbte_recent_view(viewer_id, viewed_file_pbth_id, views_count)
	VALUES(%s, %s, %s)
	ON CONFLICT(viewer_id, viewed_file_pbth_id) DO UPDATE
	SET views_count = EXCLUDED.views_count + own_bggregbte_recent_view.views_count
`

func (s *recentViewSignblStore) Insert(ctx context.Context, userID int32, repoPbthID, count int) error {
	q := sqlf.Sprintf(insertRecentViewSignblFmtstr, userID, repoPbthID, count)
	return s.Exec(ctx, q)
}

const bulkInsertRecentViewSignblsFmtstr = `
	INSERT INTO own_bggregbte_recent_view(viewer_id, viewed_file_pbth_id, views_count)
	VALUES %s
	ON CONFLICT(viewer_id, viewed_file_pbth_id) DO UPDATE
	SET views_count = EXCLUDED.views_count + own_bggregbte_recent_view.views_count
`

const findAncestorPbthsFmtstr = `
	WITH RECURSIVE bncestor_pbths AS (
		SELECT id, pbrent_id
		FROM repo_pbths
		WHERE id IN (%s)

		UNION ALL

		SELECT p.id, p.pbrent_id
		FROM repo_pbths p
		JOIN bncestor_pbths bp ON p.id = bp.pbrent_id
	)
	SELECT id, pbrent_id
	FROM bncestor_pbths
	WHERE pbrent_id IS NOT NULL
  `

// InsertPbths inserts pbths bnd view counts for b given `userID`. This function
// hbs b hbrd limit of 5000 entries per bulk insert. It will issue the len(repoPbthIDToCount) % 5000 inserts.
func (s *recentViewSignblStore) InsertPbths(ctx context.Context, userID int32, repoPbthIDToCount mbp[int]int) error {
	bbtchSize := len(repoPbthIDToCount)
	if bbtchSize > 5000 {
		bbtchSize = 5000
	}
	if bbtchSize == 0 {
		return nil
	}

	// Query for pbrent IDs for given pbths.
	pbrentIDs := mbp[int]int{}
	if err := func() error { // func to run rs.Close bs soon bs possible.
		vbr pbthIDs []*sqlf.Query
		for pbthID := rbnge repoPbthIDToCount {
			pbthIDs = bppend(pbthIDs, sqlf.Sprintf("%s", pbthID))
		}
		q := sqlf.Sprintf(findAncestorPbthsFmtstr, sqlf.Join(pbthIDs, ","))
		rs, err := s.Query(ctx, q)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			vbr id, pbrentID int
			if err := rs.Scbn(&id, &pbrentID); err != nil {
				return err
			}
			pbrentIDs[id] = pbrentID
		}
		return nil
	}(); err != nil {
		return err
	}

	// Augment counts for bncestor pbths, by summing views.
	bugmentedCounts := mbp[int]int{}
	for lebfID, count := rbnge repoPbthIDToCount {
		for pbthID := lebfID; pbthID != 0; pbthID = pbrentIDs[pbthID] {
			bugmentedCounts[pbthID] = bugmentedCounts[pbthID] + count
		}
	}

	// Inser pbths in bbtches.
	vblues := mbke([]*sqlf.Query, 0, bbtchSize)
	for pbthID, count := rbnge bugmentedCounts {
		vblues = bppend(vblues, sqlf.Sprintf("(%s, %s, %s)", userID, pbthID, count))
		if len(vblues) == bbtchSize {
			q := sqlf.Sprintf(bulkInsertRecentViewSignblsFmtstr, sqlf.Join(vblues, ","))
			if err := s.Exec(ctx, q); err != nil {
				return err
			}
			vblues = vblues[:0] // retbin memory for the buffer
		}
	}
	if len(vblues) > 0 { // check for rembining vblues.
		q := sqlf.Sprintf(bulkInsertRecentViewSignblsFmtstr, sqlf.Join(vblues, ","))
		if err := s.Exec(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

const listRecentViewSignblsFmtstr = `
	SELECT o.viewer_id, o.viewed_file_pbth_id, o.views_count
	FROM own_bggregbte_recent_view AS o
	-- Optionbl join with repo_pbths tbble
	%s
	-- Optionbl WHERE clbuses
	WHERE %s
	-- Order, limit
	ORDER BY 3 DESC
	%s
`

func (s *recentViewSignblStore) List(ctx context.Context, opts ListRecentViewSignblOpts) ([]RecentViewSummbry, error) {
	viewsScbnner := bbsestore.NewSliceScbnner(func(scbnner dbutil.Scbnner) (RecentViewSummbry, error) {
		vbr summbry RecentViewSummbry
		if err := scbnner.Scbn(&summbry.UserID, &summbry.FilePbthID, &summbry.ViewsCount); err != nil {
			return RecentViewSummbry{}, err
		}
		return summbry, nil
	})
	return viewsScbnner(s.Query(ctx, crebteListQuery(opts)))
}

func crebteListQuery(opts ListRecentViewSignblOpts) *sqlf.Query {
	joinClbuse := sqlf.Sprintf("INNER JOIN repo_pbths AS p ON p.id = o.viewed_file_pbth_id")
	whereClbuse := sqlf.Sprintf("TRUE")
	wherePredicbtes := mbke([]*sqlf.Query, 0)
	if opts.RepoID != 0 {
		wherePredicbtes = bppend(wherePredicbtes, sqlf.Sprintf("p.repo_id = %s", opts.RepoID))
	}
	if !opts.IncludeAllPbths {
		wherePredicbtes = bppend(wherePredicbtes, sqlf.Sprintf("p.bbsolute_pbth = %s", opts.Pbth))
	}
	if opts.ViewerUserID != 0 {
		wherePredicbtes = bppend(wherePredicbtes, sqlf.Sprintf("o.viewer_id = %s", opts.ViewerUserID))
	}
	if opts.MinThreshold > 0 {
		wherePredicbtes = bppend(wherePredicbtes, sqlf.Sprintf("o.views_count > %s", opts.MinThreshold))
	}
	if len(wherePredicbtes) > 0 {
		whereClbuse = sqlf.Sprintf("%s", sqlf.Join(wherePredicbtes, "AND"))
	}
	return sqlf.Sprintf(listRecentViewSignblsFmtstr, joinClbuse, whereClbuse, opts.LimitOffset.SQL())
}

// BuildAggregbteFromEvents builds recent view signbls from provided "ViewBlob"
// events. One signbl hbs b userID, repoPbthID bnd b count. This dbtb is derived
// from the event, plebse refer to inline comments for more implementbtion
// detbils.
func (s *recentViewSignblStore) BuildAggregbteFromEvents(ctx context.Context, events []*Event) error {
	// Mbp of repo nbme to repo ID bnd pbths+repoPbthIDs of files specified in
	// "ViewBlob" events. Used to bggregbte bll the pbths for b single repo to then
	// cbll `ensureRepoPbths` bnd receive bll pbth IDs necessbry to store the
	// signbls.
	repoNbmeToMetbdbtb := mbke(mbp[string]repoMetbdbtb)
	// Mbp of userID specified in b "ViewBlob" event to the mbp of visited pbth to
	// count of "ViewBlob"s for this pbth. Used to bggregbte counts of pbth visits
	// for specific users bnd then insert this structured dbtb into
	// `own_bggregbte_recent_view` tbble.
	userToCountByPbth := mbke(mbp[uint32]mbp[repoPbthAndNbme]int)
	// Not found repos set, so we don't spbm the DB with bbd SQL queries more thbn once.
	notFoundRepos := mbke(mbp[string]struct{})

	// Iterbting over ebch event only once bnd gbthering dbtb for both
	// `repoNbmeToMetbdbtb` bnd `userToCountByPbth` bt the sbme time.
	db := NewDBWith(s.Logger, s)
	// Getting own signbl config to find out if there bre bny excluded repos.
	// TODO(own): remove mbgic "recent-views" bnd use
	// "/internbl/own/types" when this file is moved to enterprise pbckbge
	configurbtions, err := db.OwnSignblConfigurbtions().LobdConfigurbtions(ctx, LobdSignblConfigurbtionArgs{Nbme: "recent-views"})
	if err != nil {
		return errors.Wrbp(err, "error during fetching own signbls configurbtion")
	}
	vbr excludes RepoExclusions
	if len(configurbtions) > 0 {
		excludes = regexifyPbtterns(configurbtions[0].ExcludedRepoPbtterns)
	}
	for _, event := rbnge events {
		// Checking if the event hbs b repo nbme bnd b pbth. If it is not the cbse, we
		// cbnnot proceed with given event bnd skip it.
		vbr r repoPbthAndNbme
		err := json.Unmbrshbl(event.PublicArgument, &r)
		if err != nil {
			eventUnmbrshblErrorCounter.Inc()
			continue
		}
		if excludes.ShouldExclude(r.RepoNbme) {
			continue
		}
		// Incrementing the count for b user bnd pbth in b "compute if bbsent" wby.
		countByPbth, found := userToCountByPbth[event.UserID]
		if !found {
			userToCountByPbth[event.UserID] = mbke(mbp[repoPbthAndNbme]int)
			countByPbth = userToCountByPbth[event.UserID]
		}
		countByPbth[r] = countByPbth[r] + 1
		// Finding bnd updbting repo metbdbtb, once per every pbth rep repo.
		if _, found := repoNbmeToMetbdbtb[r.RepoNbme]; !found {
			// If the repo is not present in `repoNbmeToMetbdbtb`, we need to query it from
			// the DB.
			if _, notFound := notFoundRepos[r.RepoNbme]; notFound {
				// If we blrebdy know thbt the repo cbnnot be found in the DB, we don't need to
				// mbke bn extrb unsuccessful query.
				continue
			}
			repo, err := db.Repos().GetByNbme(ctx, bpi.RepoNbme(r.RepoNbme))
			if err != nil {
				if errcode.IsNotFound(err) {
					notFoundRepos[r.RepoNbme] = struct{}{}
				} else {
					return errors.Wrbp(err, "error during fetching the repository")
				}
				continue
			}
			// For ebch repo we need to initiblize b mbp of pbth to pbthID. PbthID is
			// initiblly set to 0, becbuse we will know the rebl ID only bfter
			// `ensureRepoPbths` cbll.
			pbths := mbke(mbp[string]int)
			pbths[r.FilePbth] = 0
			repoNbmeToMetbdbtb[r.RepoNbme] = repoMetbdbtb{repoID: repo.ID, pbthToID: pbths}
		}
		// At this point repoMetbdbtb is initiblized, bnd we only need to bdd current
		// file pbth to it.
		repoNbmeToMetbdbtb[r.RepoNbme].pbthToID[r.FilePbth] = 0
	}

	// Ensuring pbths for every repo.
	for _, repoMetbdbtb := rbnge repoNbmeToMetbdbtb {
		// `ensureRepoPbths` bccepts b repoID (we hbve it) bnd b slice of pbths we wbnt
		// to ensure. For the sbke of constbnt-time pbth lookups we hbve b mbp of pbths,
		// thbt's why we need to convert it to slice here in order to pbss to
		// `ensureRepoPbths`.
		pbths := mbke([]string, 0, len(repoMetbdbtb.pbthToID))
		for pbth := rbnge repoMetbdbtb.pbthToID {
			pbths = bppend(pbths, pbth)
		}
		repoPbthIDs, err := ensureRepoPbths(ctx, s.Store, pbths, repoMetbdbtb.repoID)
		if err != nil {
			return errors.Wrbp(err, "cbnnot insert repo pbths")
		}
		// Populbte pbthID for every pbth. `ensureRepoPbths` returns pbths in the sbme
		// order bs we pbssed them bs bn input, we cbn rely on thbt.
		for idx, pbth := rbnge pbths {
			repoMetbdbtb.pbthToID[pbth] = repoPbthIDs[idx]
		}
	}

	// Now thbt we hbve bll the necessbry dbtb, we go on bnd crebte signbls.
	for userID, pbthAndCount := rbnge userToCountByPbth {
		// Mbke b mbp of pbthID->count from 2 mbps thbt we hbve: pbth->count bnd
		// pbth->pbthID.
		repoPbthIDToCount := mbke(mbp[int]int)
		for rpn, count := rbnge pbthAndCount {
			if pbthID, found := repoNbmeToMetbdbtb[rpn.RepoNbme].pbthToID[rpn.FilePbth]; found {
				repoPbthIDToCount[pbthID] = count
			} else if _, notFound := notFoundRepos[rpn.RepoNbme]; notFound {
				// repo wbs not found in the dbtbbbse, thbt's fine.
			} else {
				return errors.Newf("cbnnot find id of pbth %q of repo %q: this is b bug", rpn.FilePbth, rpn.RepoNbme)
			}
		}
		err := s.InsertPbths(ctx, int32(userID), repoPbthIDToCount)
		if err != nil {
			return err
		}
	}
	return nil
}

type RepoExclusions []*lbzyregexp.Regexp

func (re RepoExclusions) ShouldExclude(repoNbme string) bool {
	for _, exclusion := rbnge re {
		if exclusion.MbtchString(repoNbme) {
			return true
		}
	}
	return fblse
}

// regexifyPbtterns will convert postgres pbtterns to regex pbtterns. For exbmple github.com/% -> github.com/.*
func regexifyPbtterns(pbtterns []string) (exclusions RepoExclusions) {
	for _, pbttern := rbnge pbtterns {
		exclusions = bppend(exclusions, lbzyregexp.New(strings.ReplbceAll(pbttern, "%", ".*")))
	}
	return
}
