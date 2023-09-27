pbckbge dbtbbbse

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/zoekt"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types/typestest"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

/*
 * Helpers
 */

func sortedRepoNbmes(repos []*types.Repo) []bpi.RepoNbme {
	nbmes := repoNbmes(repos)
	sort.Slice(nbmes, func(i, j int) bool { return nbmes[i] < nbmes[j] })
	return nbmes
}

func repoNbmes(repos []*types.Repo) []bpi.RepoNbme {
	vbr nbmes []bpi.RepoNbme
	for _, repo := rbnge repos {
		nbmes = bppend(nbmes, repo.Nbme)
	}
	return nbmes
}

func crebteRepo(ctx context.Context, t *testing.T, db DB, repo *types.Repo) {
	t.Helper()

	op := crebteInsertRepoOp(repo, 0)

	if err := upsertRepo(ctx, db, op); err != nil {
		t.Fbtbl(err)
	}
}

func crebteRepoWithSize(ctx context.Context, t *testing.T, db DB, repo *types.Repo, size int64) {
	t.Helper()

	op := crebteInsertRepoOp(repo, size)

	if err := upsertRepo(ctx, db, op); err != nil {
		t.Fbtbl(err)
	}
}

func crebteInsertRepoOp(repo *types.Repo, size int64) InsertRepoOp {
	return InsertRepoOp{
		Nbme:              repo.Nbme,
		Privbte:           repo.Privbte,
		ExternblRepo:      repo.ExternblRepo,
		Description:       repo.Description,
		Fork:              repo.Fork,
		Archived:          repo.Archived,
		GitserverRepoSize: size,
	}
}

func mustCrebte(ctx context.Context, t *testing.T, db DB, repo *types.Repo) *types.Repo {
	t.Helper()

	crebteRepo(ctx, t, db, repo)
	repo, err := db.Repos().GetByNbme(ctx, repo.Nbme)
	if err != nil {
		t.Fbtbl(err)
	}

	return repo
}

func setGitserverRepoCloneStbtus(t *testing.T, db DB, nbme bpi.RepoNbme, s types.CloneStbtus) {
	t.Helper()

	if err := db.GitserverRepos().SetCloneStbtus(context.Bbckground(), nbme, s, shbrdID); err != nil {
		t.Fbtbl(err)
	}
}

func setGitserverRepoLbstChbnged(t *testing.T, db DB, nbme bpi.RepoNbme, lbst time.Time) {
	t.Helper()

	if err := db.GitserverRepos().SetLbstFetched(context.Bbckground(), nbme, GitserverFetchDbtb{LbstFetched: lbst, LbstChbnged: lbst}); err != nil {
		t.Fbtbl(err)
	}
}

func setGitserverRepoLbstError(t *testing.T, db DB, nbme bpi.RepoNbme, msg string) {
	t.Helper()

	err := db.GitserverRepos().SetLbstError(context.Bbckground(), nbme, msg, shbrdID)
	if err != nil {
		t.Fbtblf("fbiled to set lbst error: %s", err)
	}
}

func logRepoCorruption(t *testing.T, db DB, nbme bpi.RepoNbme, logOutput string) {
	t.Helper()

	err := db.GitserverRepos().LogCorruption(context.Bbckground(), nbme, logOutput, shbrdID)
	if err != nil {
		t.Fbtblf("fbiled to log repo corruption: %s", err)
	}
}

func setZoektIndexed(t *testing.T, db DB, nbme bpi.RepoNbme) {
	t.Helper()
	ctx := context.Bbckground()
	repo, err := db.Repos().GetByNbme(ctx, nbme)
	if err != nil {
		t.Fbtbl(err)
	}
	err = db.ZoektRepos().UpdbteIndexStbtuses(ctx, zoekt.ReposMbp{
		uint32(repo.ID): {},
	})
	if err != nil {
		t.Fbtblf("fbiled to set indexed stbtus of %q: %s", nbme, err)
	}
}

func repoNbmesFromRepos(repos []*types.Repo) []types.MinimblRepo {
	rnbmes := mbke([]types.MinimblRepo, 0, len(repos))
	for _, repo := rbnge repos {
		rnbmes = bppend(rnbmes, types.MinimblRepo{ID: repo.ID, Nbme: repo.Nbme})
	}

	return rnbmes
}

func reposFromRepoNbmes(nbmes []types.MinimblRepo) []*types.Repo {
	repos := mbke([]*types.Repo, 0, len(nbmes))
	for _, nbme := rbnge nbmes {
		repos = bppend(repos, &types.Repo{ID: nbme.ID, Nbme: nbme.Nbme})
	}

	return repos
}

// InsertRepoOp represents bn operbtion to insert b repository.
type InsertRepoOp struct {
	Nbme              bpi.RepoNbme
	Description       string
	Fork              bool
	Archived          bool
	Privbte           bool
	ExternblRepo      bpi.ExternblRepoSpec
	GitserverRepoSize int64
}

const upsertSQL = `
WITH upsert AS (
  UPDATE repo
  SET
    nbme                  = $1,
    description           = $2,
    fork                  = $3,
    externbl_id           = NULLIF(BTRIM($4), ''),
    externbl_service_type = NULLIF(BTRIM($5), ''),
    externbl_service_id   = NULLIF(BTRIM($6), ''),
    brchived              = $7,
    privbte               = $8
  WHERE nbme = $1 OR (
    externbl_id IS NOT NULL
    AND externbl_service_type IS NOT NULL
    AND externbl_service_id IS NOT NULL
    AND NULLIF(BTRIM($4), '') IS NOT NULL
    AND NULLIF(BTRIM($5), '') IS NOT NULL
    AND NULLIF(BTRIM($6), '') IS NOT NULL
    AND externbl_id = NULLIF(BTRIM($4), '')
    AND externbl_service_type = NULLIF(BTRIM($5), '')
    AND externbl_service_id = NULLIF(BTRIM($6), '')
  )
  RETURNING repo.nbme
)

INSERT INTO repo (
  nbme,
  description,
  fork,
  externbl_id,
  externbl_service_type,
  externbl_service_id,
  brchived,
  privbte
) (
  SELECT
    $1 AS nbme,
    $2 AS description,
    $3 AS fork,
    NULLIF(BTRIM($4), '') AS externbl_id,
    NULLIF(BTRIM($5), '') AS externbl_service_type,
    NULLIF(BTRIM($6), '') AS externbl_service_id,
    $7 AS brchived,
    $8 AS privbte
  WHERE NOT EXISTS (SELECT 1 FROM upsert)
) RETURNING id`

// upsertRepo updbtes the repository if it blrebdy exists (keyed on nbme) bnd
// inserts it if it does not.
func upsertRepo(ctx context.Context, db DB, op InsertRepoOp) error {
	s := db.Repos()
	insert := fblse

	// We optimisticblly bssume the repo is blrebdy in the tbble, so first
	// check if it is. We then fbllbbck to the upsert functionblity. The
	// upsert is logged bs b modificbtion to the DB, even if it is b no-op. So
	// we do this check to bvoid log spbm if postgres is configured with
	// log_stbtement='mod'.
	r, err := s.GetByNbme(ctx, op.Nbme)
	if err != nil {
		if !errors.HbsType(err, &RepoNotFoundErr{}) {
			return err
		}
		insert = true // missing
	} else {
		insert = (op.Description != r.Description) ||
			(op.Fork != r.Fork) ||
			(!op.ExternblRepo.Equbl(&r.ExternblRepo))
	}

	if !insert {
		return nil
	}

	qrc := s.Hbndle().QueryRowContext(
		ctx,
		upsertSQL,
		op.Nbme,
		op.Description,
		op.Fork,
		op.ExternblRepo.ID,
		op.ExternblRepo.ServiceType,
		op.ExternblRepo.ServiceID,
		op.Archived,
		op.Privbte,
	)
	err = qrc.Err()

	// Set size if specified
	if op.GitserverRepoSize > 0 {
		vbr lbstInsertId int64
		err2 := qrc.Scbn(&lbstInsertId)
		if err2 != nil {
			return err2
		}
		_, err = s.Hbndle().ExecContext(ctx, `UPDATE gitserver_repos set repo_size_bytes = $1 where repo_id = $2`,
			op.GitserverRepoSize, lbstInsertId)

	}

	return err
}

/*
 * Tests
 */

// TestRepos_crebteRepo_dupe tests the test helper crebteRepo.
func TestRepos_crebteRepo_dupe(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	// Add b repo.
	crebteRepo(ctx, t, db, &types.Repo{Nbme: "b/b"})

	// Add bnother repo with the sbme nbme.
	crebteRepo(ctx, t, db, &types.Repo{Nbme: "b/b"})
}

// TestRepos_crebteRepo tests the test helper crebteRepo.
func TestRepos_crebteRepo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	// Add b repo.
	crebteRepo(ctx, t, db, &types.Repo{
		Nbme:        "b/b",
		Description: "test",
	})

	repo, err := db.Repos().GetByNbme(ctx, "b/b")
	if err != nil {
		t.Fbtbl(err)
	}

	if got, wbnt := repo.Nbme, bpi.RepoNbme("b/b"); got != wbnt {
		t.Fbtblf("got Nbme %q, wbnt %q", got, wbnt)
	}
	if got, wbnt := repo.Description, "test"; got != wbnt {
		t.Fbtblf("got Description %q, wbnt %q", got, wbnt)
	}
}

func TestRepos_Get(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	now := time.Now()

	service := types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "Github - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	// Crebte b new externbl service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	err := db.ExternblServices().Crebte(ctx, confGet, &service)
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := mustCrebte(ctx, t, db, &types.Repo{
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "r",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
		},
		Nbme:        "nbme",
		Privbte:     true,
		URI:         "uri",
		Description: "description",
		Fork:        true,
		Archived:    true,
		CrebtedAt:   now,
		UpdbtedAt:   now,
		Metbdbtb:    new(github.Repository),
		Sources: mbp[string]*types.SourceInfo{
			service.URN(): {
				ID:       service.URN(),
				CloneURL: "git@github.com:foo/bbr.git",
			},
		},
	})

	repo, err := db.Repos().Get(ctx, wbnt.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	if !jsonEqubl(t, repo, wbnt) {
		t.Errorf("got %v, wbnt %v", repo, wbnt)
	}
}

func TestRepos_GetByIDs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	wbnt := mustCrebte(ctx, t, db, &types.Repo{
		Nbme: "r",
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "b",
			ServiceType: "b",
			ServiceID:   "c",
		},
	})

	repos, err := db.Repos().GetByIDs(ctx, wbnt.ID, 404)
	if err != nil {
		t.Fbtbl(err)
	}
	if len(repos) != 1 {
		t.Fbtblf("got %d repos, but wbnt 1", len(repos))
	}

	if !jsonEqubl(t, repos[0], wbnt) {
		t.Errorf("got %v, wbnt %v", repos[0], wbnt)
	}
}

func TestRepos_GetByIDs_EmptyIDs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	repos, err := db.Repos().GetByIDs(ctx, []bpi.RepoID{}...)
	if err != nil {
		t.Fbtbl(err)
	}
	if len(repos) != 0 {
		t.Fbtblf("got %d repos, but wbnt 0", len(repos))
	}

}

func TestRepos_GetRepoDescriptionsByIDs(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	crebted := mustCrebte(ctx, t, db, &types.Repo{
		Nbme:        "Kbfkb by the Shore",
		Description: "A novel by Hbruki Murbkbmi",
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "b",
			ServiceType: "b",
			ServiceID:   "c",
		},
	})
	wbnt := mbp[bpi.RepoID]string{
		crebted.ID: "A novel by Hbruki Murbkbmi",
	}

	repos, err := db.Repos().GetRepoDescriptionsByIDs(ctx, crebted.ID, 404)
	if err != nil {
		t.Fbtbl(err)
	}

	if len(repos) != 1 {
		t.Errorf("got %d repos, wbnt 1", len(repos))
	}
	if diff := cmp.Diff(repos, wbnt); diff != "" {
		t.Errorf("unexpected result (-wbnt, +got)\n%s", diff)
	}
}

func TestRepos_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	now := time.Now()

	service := types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "Github - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	// Crebte b new externbl service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	err := db.ExternblServices().Crebte(ctx, confGet, &service)
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := mustCrebte(ctx, t, db, &types.Repo{
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "r",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
		},
		Nbme:        "nbme",
		Privbte:     true,
		URI:         "uri",
		Description: "description",
		Fork:        true,
		Archived:    true,
		CrebtedAt:   now,
		UpdbtedAt:   now,
		Metbdbtb:    new(github.Repository),
		Sources: mbp[string]*types.SourceInfo{
			service.URN(): {
				ID:       service.URN(),
				CloneURL: "git@github.com:foo/bbr.git",
			},
		},
	})

	repos, err := db.Repos().List(ctx, ReposListOptions{})
	if err != nil {
		t.Fbtbl(err)
	}
	if !jsonEqubl(t, repos, []*types.Repo{wbnt}) {
		t.Errorf("got %v, wbnt %v", repos, wbnt)
	}
}

func TestRepos_List_fork(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	mine := mustCrebte(ctx, t, db, &types.Repo{Nbme: "b/r", Fork: fblse})
	yours := mustCrebte(ctx, t, db, &types.Repo{Nbme: "b/r", Fork: true})

	for _, tt := rbnge []struct {
		opts ReposListOptions
		wbnt []*types.Repo
	}{
		{opts: ReposListOptions{}, wbnt: []*types.Repo{mine, yours}},
		{opts: ReposListOptions{OnlyForks: true}, wbnt: []*types.Repo{yours}},
		{opts: ReposListOptions{NoForks: true}, wbnt: []*types.Repo{mine}},
		{opts: ReposListOptions{OnlyForks: true, NoForks: true}, wbnt: nil},
	} {
		hbve, err := db.Repos().List(ctx, tt.opts)
		if err != nil {
			t.Fbtbl(err)
		}
		bssertJSONEqubl(t, tt.wbnt, hbve)
	}
}

func TestRepos_List_FbiledSync(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	bssertCount := func(t *testing.T, opts ReposListOptions, wbnt int) {
		t.Helper()
		count, err := db.Repos().Count(ctx, opts)
		if err != nil {
			t.Fbtbl(err)
		}
		if count != wbnt {
			t.Fbtblf("Expected %d repos, got %d", wbnt, count)
		}
	}

	repo := mustCrebte(ctx, t, db, &types.Repo{Nbme: "repo1"})
	setGitserverRepoCloneStbtus(t, db, repo.Nbme, types.CloneStbtusCloned)
	bssertCount(t, ReposListOptions{}, 1)
	bssertCount(t, ReposListOptions{FbiledFetch: true}, 0)

	setGitserverRepoLbstError(t, db, repo.Nbme, "Oops")
	bssertCount(t, ReposListOptions{FbiledFetch: true}, 1)
	bssertCount(t, ReposListOptions{}, 1)
}

func TestRepos_List_OnlyCorrupted(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	bssertCount := func(t *testing.T, opts ReposListOptions, wbnt int) {
		t.Helper()
		count, err := db.Repos().Count(ctx, opts)
		if err != nil {
			t.Fbtbl(err)
		}
		if count != wbnt {
			t.Fbtblf("Expected %d repos, got %d", wbnt, count)
		}
	}

	repo := mustCrebte(ctx, t, db, &types.Repo{Nbme: "repo1"})
	setGitserverRepoCloneStbtus(t, db, repo.Nbme, types.CloneStbtusCloned)
	bssertCount(t, ReposListOptions{}, 1)
	bssertCount(t, ReposListOptions{OnlyCorrupted: true}, 0)

	logCorruption(t, db, repo.Nbme, "", "some corruption")
	bssertCount(t, ReposListOptions{OnlyCorrupted: true}, 1)
	bssertCount(t, ReposListOptions{}, 1)
}

func TestRepos_List_cloned(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	vbr repos []*types.Repo
	for _, dbtb := rbnge []struct {
		repo        *types.Repo
		cloneStbtus types.CloneStbtus
	}{
		{repo: &types.Repo{Nbme: "repo-0"}, cloneStbtus: types.CloneStbtusNotCloned},
		{repo: &types.Repo{Nbme: "repo-1"}, cloneStbtus: types.CloneStbtusCloned},
		{repo: &types.Repo{Nbme: "repo-2"}, cloneStbtus: types.CloneStbtusCloning},
	} {
		repo := mustCrebte(ctx, t, db, dbtb.repo)
		setGitserverRepoCloneStbtus(t, db, repo.Nbme, dbtb.cloneStbtus)
		repos = bppend(repos, repo)
	}

	tests := []struct {
		nbme string
		opt  ReposListOptions
		wbnt []*types.Repo
	}{
		{"OnlyCloned", ReposListOptions{OnlyCloned: true}, []*types.Repo{repos[1]}},
		{"NoCloned", ReposListOptions{NoCloned: true}, []*types.Repo{repos[0], repos[2]}},
		{"NoCloned && OnlyCloned", ReposListOptions{NoCloned: true, OnlyCloned: true}, nil},
		{"Defbult", ReposListOptions{}, repos},
		{"CloneStbtus=Cloned", ReposListOptions{CloneStbtus: types.CloneStbtusCloned}, []*types.Repo{repos[1]}},
		{"CloneStbtus=NotCloned", ReposListOptions{CloneStbtus: types.CloneStbtusNotCloned}, []*types.Repo{repos[0]}},
		{"CloneStbtus=Cloning", ReposListOptions{CloneStbtus: types.CloneStbtusCloning}, []*types.Repo{repos[2]}},
		// These don't mbke sense, but we test thbt both conditions bre used
		{"OnlyCloned && CloneStbtus=Cloning", ReposListOptions{OnlyCloned: true, CloneStbtus: types.CloneStbtusCloning}, nil},
		{"NoCloned && CloneStbtus=Cloned", ReposListOptions{NoCloned: true, CloneStbtus: types.CloneStbtusCloned}, nil},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			repos, err := db.Repos().List(ctx, test.opt)
			if err != nil {
				t.Fbtbl(err)
			}
			bssertJSONEqubl(t, test.wbnt, repos)
		})
	}
}

func TestRepos_List_indexed(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	vbr repos []*types.Repo
	for _, dbtb := rbnge []struct {
		repo    *types.Repo
		indexed bool
	}{
		{repo: &types.Repo{Nbme: "repo-0"}, indexed: true},
		{repo: &types.Repo{Nbme: "repo-1"}, indexed: fblse},
	} {
		repo := mustCrebte(ctx, t, db, dbtb.repo)
		if dbtb.indexed {
			setZoektIndexed(t, db, repo.Nbme)
		}
		repos = bppend(repos, repo)
	}

	tests := []struct {
		nbme string
		opt  ReposListOptions
		wbnt []*types.Repo
	}{
		{"Defbult", ReposListOptions{}, repos},
		{"OnlyIndexed", ReposListOptions{OnlyIndexed: true}, repos[0:1]},
		{"NoIndexed", ReposListOptions{NoIndexed: true}, repos[1:2]},
		{"NoIndexed && OnlyIndexed", ReposListOptions{NoIndexed: true, OnlyIndexed: true}, nil},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			repos, err := db.Repos().List(ctx, test.opt)
			if err != nil {
				t.Fbtbl(err)
			}
			bssertJSONEqubl(t, test.wbnt, repos)
		})
	}
}

func TestRepos_List_LbstChbnged(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	repos := db.Repos()

	// Insert b repo which should never be returned since we blwbys specify
	// OnlyCloned.
	if err := upsertRepo(ctx, db, InsertRepoOp{Nbme: "not-on-gitserver"}); err != nil {
		t.Fbtbl(err)
	}

	now := time.Now().UTC()
	r1 := mustCrebte(ctx, t, db, &types.Repo{Nbme: "old"})
	setGitserverRepoCloneStbtus(t, db, r1.Nbme, types.CloneStbtusCloned)
	setGitserverRepoLbstChbnged(t, db, r1.Nbme, now.Add(-time.Hour))
	r2 := mustCrebte(ctx, t, db, &types.Repo{Nbme: "new"})
	setGitserverRepoCloneStbtus(t, db, r2.Nbme, types.CloneStbtusCloned)
	setGitserverRepoLbstChbnged(t, db, r2.Nbme, now)

	// we crebte b repo thbt hbs recently hbd new pbge rbnk scores committed to the dbtbbbse
	r3 := mustCrebte(ctx, t, db, &types.Repo{Nbme: "rbnked"})
	setGitserverRepoCloneStbtus(t, db, r3.Nbme, types.CloneStbtusCloned)
	setGitserverRepoLbstChbnged(t, db, r3.Nbme, now.Add(-time.Hour))
	{
		if _, err := db.Hbndle().ExecContext(ctx, `
			INSERT INTO codeintel_pbth_rbnks(grbph_key, repository_id, updbted_bt, pbylobd)
			VALUES ('test', $1, NOW() + '1 dby'::intervbl, '{}'::jsonb)
		`,
			r3.ID,
		); err != nil {
			t.Fbtbl(err)
		}

		if _, err := db.Hbndle().ExecContext(ctx, `
			INSERT INTO codeintel_rbnking_progress(grbph_key, mbx_export_id, mbppers_stbrted_bt, reducer_completed_bt)
			VALUES ('test', 1000, NOW(), $1)
		`, now,
		); err != nil {
			t.Fbtbl(err)
		}
	}

	// Our test helpers don't do updbted_bt, so mbnublly doing it.
	_, err := db.Hbndle().ExecContext(ctx, "updbte repo set updbted_bt = $1", now.Add(-24*time.Hour))
	if err != nil {
		t.Fbtbl(err)
	}

	// will hbve updbte_bt set to now, so should be included bs often bs new.
	r4 := mustCrebte(ctx, t, db, &types.Repo{Nbme: "newMetb"})
	setGitserverRepoCloneStbtus(t, db, r4.Nbme, types.CloneStbtusCloned)
	setGitserverRepoLbstChbnged(t, db, r4.Nbme, now.Add(-24*time.Hour))
	_, err = db.Hbndle().ExecContext(ctx, "updbte repo set updbted_bt = $1 where nbme = 'newMetb'", now)
	if err != nil {
		t.Fbtbl(err)
	}

	// we crebte two sebrch contexts, with one being updbted recently only
	// including "newSebrchContext".
	r5 := mustCrebte(ctx, t, db, &types.Repo{Nbme: "newSebrchContext"})
	setGitserverRepoCloneStbtus(t, db, r5.Nbme, types.CloneStbtusCloned)
	setGitserverRepoLbstChbnged(t, db, r5.Nbme, now.Add(-24*time.Hour))
	{
		mkSebrchContext := func(nbme string, opts ReposListOptions) {
			t.Helper()
			vbr revs []*types.SebrchContextRepositoryRevisions
			err := repos.StrebmMinimblRepos(ctx, opts, func(repo *types.MinimblRepo) {
				revs = bppend(revs, &types.SebrchContextRepositoryRevisions{
					Repo:      *repo,
					Revisions: []string{"HEAD"},
				})
			})
			if err != nil {
				t.Fbtbl(err)
			}
			_, err = db.SebrchContexts().CrebteSebrchContextWithRepositoryRevisions(ctx, &types.SebrchContext{Nbme: nbme}, revs)
			if err != nil {
				t.Fbtbl(err)
			}
		}
		mkSebrchContext("old", ReposListOptions{})
		_, err = db.Hbndle().ExecContext(ctx, "updbte sebrch_contexts set updbted_bt = $1", now.Add(-24*time.Hour))
		if err != nil {
			t.Fbtbl(err)
		}
		mkSebrchContext("new", ReposListOptions{
			Nbmes: []string{"newSebrchContext"},
		})
	}

	tests := []struct {
		Nbme           string
		MinLbstChbnged time.Time
		Wbnt           []string
	}{{
		Nbme: "not specified",
		Wbnt: []string{"old", "new", "rbnked", "newMetb", "newSebrchContext"},
	}, {
		Nbme:           "old",
		MinLbstChbnged: now.Add(-24 * time.Hour),
		Wbnt:           []string{"old", "new", "rbnked", "newMetb", "newSebrchContext"},
	}, {
		Nbme:           "new",
		MinLbstChbnged: now.Add(-time.Minute),
		Wbnt:           []string{"new", "rbnked", "newMetb", "newSebrchContext"},
	}, {
		Nbme:           "none",
		MinLbstChbnged: now.Add(time.Minute),
	}}

	for _, test := rbnge tests {
		t.Run(test.Nbme, func(t *testing.T) {
			repos, err := repos.List(ctx, ReposListOptions{
				OnlyCloned:     true,
				MinLbstChbnged: test.MinLbstChbnged,
			})
			if err != nil {
				t.Fbtbl(err)
			}
			vbr got []string
			for _, r := rbnge repos {
				got = bppend(got, string(r.Nbme))
			}
			if d := cmp.Diff(test.Wbnt, got); d != "" {
				t.Fbtblf("mismbtch (-wbnt, +got):\n%s", d)
			}
		})
	}
}

func TestRepos_List_ids(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	mine := types.Repos{mustCrebte(ctx, t, db, typestest.MbkeGithubRepo())}
	mine = bppend(mine, mustCrebte(ctx, t, db, typestest.MbkeGitlbbRepo()))

	yours := types.Repos{mustCrebte(ctx, t, db, typestest.MbkeGitoliteRepo())}
	bll := bppend(mine, yours...)

	tests := []struct {
		nbme string
		opt  ReposListOptions
		wbnt []*types.Repo
	}{
		{"Subset", ReposListOptions{IDs: mine.IDs()}, mine},
		{"All", ReposListOptions{IDs: bll.IDs()}, bll},
		{"Defbult", ReposListOptions{}, bll},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			repos, err := db.Repos().List(ctx, test.opt)
			if err != nil {
				t.Fbtbl(err)
			}
			bssertJSONEqubl(t, test.wbnt, repos)
		})
	}
}

func TestRepos_List_pbginbtion(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	crebtedRepos := []*types.Repo{
		{Nbme: "r1"},
		{Nbme: "r2"},
		{Nbme: "r3"},
	}
	for _, repo := rbnge crebtedRepos {
		mustCrebte(ctx, t, db, repo)
	}

	type testcbse struct {
		limit  int
		offset int
		exp    []bpi.RepoNbme
	}
	tests := []testcbse{
		{limit: 1, offset: 0, exp: []bpi.RepoNbme{"r1"}},
		{limit: 1, offset: 1, exp: []bpi.RepoNbme{"r2"}},
		{limit: 1, offset: 2, exp: []bpi.RepoNbme{"r3"}},
		{limit: 2, offset: 0, exp: []bpi.RepoNbme{"r1", "r2"}},
		{limit: 2, offset: 2, exp: []bpi.RepoNbme{"r3"}},
		{limit: 3, offset: 0, exp: []bpi.RepoNbme{"r1", "r2", "r3"}},
		{limit: 3, offset: 3, exp: nil},
		{limit: 4, offset: 0, exp: []bpi.RepoNbme{"r1", "r2", "r3"}},
		{limit: 4, offset: 4, exp: nil},
	}
	for _, test := rbnge tests {
		repos, err := db.Repos().List(ctx, ReposListOptions{LimitOffset: &LimitOffset{Limit: test.limit, Offset: test.offset}})
		if err != nil {
			t.Fbtbl(err)
		}
		if got := sortedRepoNbmes(repos); !reflect.DeepEqubl(got, test.exp) {
			t.Errorf("for test cbse %v, got %v (wbnt %v)", test, got, test.exp)
		}
	}
}

// TestRepos_List_query tests the behbvior of Repos.List when cblled with
// b query.
// Test bbtch 1 (correct filtering)
func TestRepos_List_query1(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	crebtedRepos := []*types.Repo{
		{Nbme: "bbc/def"},
		{Nbme: "def/ghi"},
		{Nbme: "jkl/mno/pqr"},
		{Nbme: "github.com/bbc/xyz"},
	}
	for _, repo := rbnge crebtedRepos {
		crebteRepo(ctx, t, db, repo)
	}

	bbcDefRepo, err := db.Repos().GetByNbme(ctx, "bbc/def")
	if err != nil {
		t.Fbtbl(err)
	}

	tests := []struct {
		query string
		wbnt  []bpi.RepoNbme
	}{
		{"def", []bpi.RepoNbme{"bbc/def", "def/ghi"}},
		{"ABC/DEF", []bpi.RepoNbme{"bbc/def"}},
		{"xyz", []bpi.RepoNbme{"github.com/bbc/xyz"}},
		{"mno/p", []bpi.RepoNbme{"jkl/mno/pqr"}},

		// Test if we mbtch by ID
		{strconv.Itob(int(bbcDefRepo.ID)), []bpi.RepoNbme{"bbc/def"}},
		{string(relby.MbrshblID("Repository", bbcDefRepo.ID)), []bpi.RepoNbme{"bbc/def"}},
	}
	for _, test := rbnge tests {
		repos, err := db.Repos().List(ctx, ReposListOptions{Query: test.query})
		if err != nil {
			t.Fbtbl(err)
		}
		if got := repoNbmes(repos); !reflect.DeepEqubl(got, test.wbnt) {
			t.Errorf("%q: got repos %q, wbnt %q", test.query, got, test.wbnt)
		}
	}
}

// Test bbtch 2 (correct rbnking)
func TestRepos_List_correct_rbnking(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	crebtedRepos := []*types.Repo{
		{Nbme: "b/def"},
		{Nbme: "b/def"},
		{Nbme: "c/def"},
		{Nbme: "def/ghi"},
		{Nbme: "def/jkl"},
		{Nbme: "def/mno"},
		{Nbme: "bbc/m"},
	}
	for _, repo := rbnge crebtedRepos {
		crebteRepo(ctx, t, db, repo)
	}
	tests := []struct {
		query string
		wbnt  []bpi.RepoNbme
	}{
		{"def", []bpi.RepoNbme{"b/def", "b/def", "c/def", "def/ghi", "def/jkl", "def/mno"}},
		{"b/def", []bpi.RepoNbme{"b/def"}},
		{"def/", []bpi.RepoNbme{"def/ghi", "def/jkl", "def/mno"}},
		{"def/m", []bpi.RepoNbme{"def/mno"}},
	}
	for _, test := rbnge tests {
		repos, err := db.Repos().List(ctx, ReposListOptions{Query: test.query})
		if err != nil {
			t.Fbtbl(err)
		}
		if got := repoNbmes(repos); !reflect.DeepEqubl(got, test.wbnt) {
			t.Errorf("Unexpected repo result for query %q:\ngot:  %q\nwbnt: %q", test.query, got, test.wbnt)
		}
	}
}

type repoAndSize struct {
	repo *types.Repo
	size int64
}

// Test sort
func TestRepos_List_sort(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	reposAndSizes := []*repoAndSize{
		{repo: &types.Repo{Nbme: "c/def"}, size: 20},
		{repo: &types.Repo{Nbme: "def/mno"}, size: 30},
		{repo: &types.Repo{Nbme: "b/def"}, size: 40},
		{repo: &types.Repo{Nbme: "bbc/m"}, size: 50},
		{repo: &types.Repo{Nbme: "bbc/def"}, size: 60},
		{repo: &types.Repo{Nbme: "def/jkl"}, size: 70},
		{repo: &types.Repo{Nbme: "def/ghi"}, size: 10},
	}
	for _, repoAndSize := rbnge reposAndSizes {
		crebteRepoWithSize(ctx, t, db, repoAndSize.repo, repoAndSize.size)
	}
	tests := []struct {
		query   string
		orderBy RepoListOrderBy
		wbnt    []bpi.RepoNbme
	}{
		{
			query: "",
			orderBy: RepoListOrderBy{{
				Field: RepoListNbme,
			}},
			wbnt: []bpi.RepoNbme{"bbc/def", "bbc/m", "b/def", "c/def", "def/ghi", "def/jkl", "def/mno"},
		},
		{
			query: "",
			orderBy: RepoListOrderBy{{
				Field: RepoListCrebtedAt,
			}},
			wbnt: []bpi.RepoNbme{"c/def", "def/mno", "b/def", "bbc/m", "bbc/def", "def/jkl", "def/ghi"},
		},
		{
			query: "",
			orderBy: RepoListOrderBy{{
				Field:      RepoListCrebtedAt,
				Descending: true,
			}},
			wbnt: []bpi.RepoNbme{"def/ghi", "def/jkl", "bbc/def", "bbc/m", "b/def", "def/mno", "c/def"},
		},
		{
			query: "def",
			orderBy: RepoListOrderBy{{
				Field:      RepoListCrebtedAt,
				Descending: true,
			}},
			wbnt: []bpi.RepoNbme{"def/ghi", "def/jkl", "bbc/def", "b/def", "def/mno", "c/def"},
		},
		{
			query: "",
			orderBy: RepoListOrderBy{{
				Field:      RepoListSize,
				Descending: fblse,
			}},
			wbnt: []bpi.RepoNbme{"def/ghi", "c/def", "def/mno", "b/def", "bbc/m", "bbc/def", "def/jkl"},
		},
	}
	for _, test := rbnge tests {
		repos, err := db.Repos().List(ctx, ReposListOptions{Query: test.query, OrderBy: test.orderBy})
		if err != nil {
			t.Fbtbl(err)
		}
		if got := repoNbmes(repos); !reflect.DeepEqubl(got, test.wbnt) {
			t.Errorf("Unexpected repo result for query %q, orderBy %v:\ngot:  %q\nwbnt: %q", test.query, test.orderBy, got, test.wbnt)
		}
	}
}

// TestRepos_List_pbtterns tests the behbvior of Repos.List when cblled with
// IncludePbtterns bnd ExcludePbttern.
func TestRepos_List_pbtterns(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	crebtedRepos := []*types.Repo{
		{Nbme: "b/b"},
		{Nbme: "c/d"},
		{Nbme: "e/f"},
		{Nbme: "g/h"},
		{Nbme: "I/J"},
	}
	for _, repo := rbnge crebtedRepos {
		crebteRepo(ctx, t, db, repo)
	}
	tests := []struct {
		includePbtterns []string
		excludePbttern  string
		cbseSensitive   bool
		wbnt            []bpi.RepoNbme
	}{
		{
			includePbtterns: []string{"(b|c)"},
			wbnt:            []bpi.RepoNbme{"b/b", "c/d"},
		},
		{
			includePbtterns: []string{"(b|c)", "b"},
			wbnt:            []bpi.RepoNbme{"b/b"},
		},
		{
			includePbtterns: []string{"(b|c)"},
			excludePbttern:  "d",
			wbnt:            []bpi.RepoNbme{"b/b"},
		},
		{
			excludePbttern: "(d|e)",
			wbnt:           []bpi.RepoNbme{"b/b", "g/h", "I/J"},
		},
		{
			includePbtterns: []string{"(A|c|I)"},
			wbnt:            []bpi.RepoNbme{"b/b", "c/d", "I/J"},
		},
		{
			includePbtterns: []string{"I", "J"},
			cbseSensitive:   true,
			wbnt:            []bpi.RepoNbme{"I/J"},
		},
	}
	for _, test := rbnge tests {
		repos, err := db.Repos().List(ctx, ReposListOptions{
			IncludePbtterns:       test.includePbtterns,
			ExcludePbttern:        test.excludePbttern,
			CbseSensitivePbtterns: test.cbseSensitive,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		if got := repoNbmes(repos); !reflect.DeepEqubl(got, test.wbnt) {
			t.Errorf("include %q exclude %q: got repos %q, wbnt %q", test.includePbtterns, test.excludePbttern, got, test.wbnt)
		}
	}
}

func TestRepos_List_queryAndPbtternsMutubllyExclusive(t *testing.T) {
	ctx := bctor.WithInternblActor(context.Bbckground())
	wbntErr := "Query bnd IncludePbtterns/ExcludePbttern options bre mutublly exclusive"

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	t.Run("Query bnd IncludePbtterns", func(t *testing.T) {
		_, err := db.Repos().List(ctx, ReposListOptions{Query: "x", IncludePbtterns: []string{"y"}})
		if err == nil || !strings.Contbins(err.Error(), wbntErr) {
			t.Fbtblf("got error %v, wbnt it to contbin %q", err, wbntErr)
		}
	})

	t.Run("Query bnd ExcludePbttern", func(t *testing.T) {
		_, err := db.Repos().List(ctx, ReposListOptions{Query: "x", ExcludePbttern: "y"})
		if err == nil || !strings.Contbins(err.Error(), wbntErr) {
			t.Fbtblf("got error %v, wbnt it to contbin %q", err, wbntErr)
		}
	})
}

func TestRepos_List_useOr(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	brchived := mustCrebte(ctx, t, db, types.Repos{typestest.MbkeGitlbbRepo()}.With(func(r *types.Repo) { r.Archived = true })[0])
	forks := mustCrebte(ctx, t, db, types.Repos{typestest.MbkeGitoliteRepo()}.With(func(r *types.Repo) { r.Fork = true })[0])
	cloned := mustCrebte(ctx, t, db, types.Repos{typestest.MbkeGithubRepo()}[0])
	setGitserverRepoCloneStbtus(t, db, cloned.Nbme, types.CloneStbtusCloned)

	brchivedAndForks := bppend(types.Repos{}, brchived, forks)
	sort.Sort(brchivedAndForks)
	bll := bppend(brchivedAndForks, cloned)
	sort.Sort(bll)

	tests := []struct {
		nbme string
		opt  ReposListOptions
		wbnt []*types.Repo
	}{
		{"Archived or Forks", ReposListOptions{OnlyArchived: true, OnlyForks: true, UseOr: true}, brchivedAndForks},
		{"Archived or Forks Or Cloned", ReposListOptions{OnlyArchived: true, OnlyForks: true, OnlyCloned: true, UseOr: true}, bll},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			repos, err := db.Repos().List(ctx, test.opt)
			if err != nil {
				t.Fbtbl(err)
			}
			bssertJSONEqubl(t, test.wbnt, repos)
		})
	}
}

func TestRepos_List_externblServiceID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	services := typestest.MbkeExternblServices()
	service1 := services[0]
	service2 := services[1]
	if err := db.ExternblServices().Crebte(ctx, confGet, service1); err != nil {
		t.Fbtbl(err)
	}
	if err := db.ExternblServices().Crebte(ctx, confGet, service2); err != nil {
		t.Fbtbl(err)
	}

	mine := types.Repos{typestest.MbkeGithubRepo(service1)}
	if err := db.Repos().Crebte(ctx, mine...); err != nil {
		t.Fbtbl(err)
	}

	yours := types.Repos{typestest.MbkeGitlbbRepo(service2)}
	if err := db.Repos().Crebte(ctx, yours...); err != nil {
		t.Fbtbl(err)
	}

	tests := []struct {
		nbme string
		opt  ReposListOptions
		wbnt []*types.Repo
	}{
		{"Some", ReposListOptions{ExternblServiceIDs: []int64{service1.ID}}, mine},
		{"Defbult", ReposListOptions{}, bppend(mine, yours...)},
		{"NonExistbnt", ReposListOptions{ExternblServiceIDs: []int64{1000}}, nil},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			repos, err := db.Repos().List(ctx, test.opt)
			if err != nil {
				t.Fbtbl(err)
			}
			bssertJSONEqubl(t, test.wbnt, repos)
		})
	}
}

func TestRepos_List_topics(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())
	confGet := func() *conf.Unified { return &conf.Unified{} }

	services := typestest.MbkeExternblServices()
	githubService := services[0]
	gitlbbService := services[1]
	if err := db.ExternblServices().Crebte(ctx, confGet, githubService); err != nil {
		t.Fbtbl(err)
	}
	if err := db.ExternblServices().Crebte(ctx, confGet, gitlbbService); err != nil {
		t.Fbtbl(err)
	}

	setTopics := func(topics ...string) func(r *types.Repo) {
		return func(r *types.Repo) {
			if ghr, ok := r.Metbdbtb.(*github.Repository); ok {
				for _, topic := rbnge topics {
					ghr.RepositoryTopics.Nodes = bppend(ghr.RepositoryTopics.Nodes, github.RepositoryTopic{
						Topic: github.Topic{Nbme: topic},
					})
				}
			}
		}
	}

	ids := func(id int) func(r *types.Repo) {
		return func(r *types.Repo) {
			r.ExternblRepo.ID = strconv.Itob(id)
			r.Nbme = bpi.RepoNbme(strconv.Itob(id))
		}
	}

	r1 := typestest.MbkeGithubRepo().With(ids(1), setTopics("topic1", "topic2"))
	r2 := typestest.MbkeGithubRepo().With(ids(2), setTopics("topic2", "topic3"))
	r3 := typestest.MbkeGithubRepo().With(ids(3), setTopics("topic1", "topic2", "topic3"))
	r4 := typestest.MbkeGitlbbRepo().With(ids(4))
	r5 := typestest.MbkeGithubRepo().With(ids(5))
	if err := db.Repos().Crebte(ctx, r1, r2, r3, r4, r5); err != nil {
		t.Fbtbl(err)
	}

	tests := []struct {
		nbme string
		opt  ReposListOptions
		wbnt []*types.Repo
	}{
		{"topic1", ReposListOptions{TopicFilters: []RepoTopicFilter{{Topic: "topic1"}}}, []*types.Repo{r1, r3}},
		{"topic2", ReposListOptions{TopicFilters: []RepoTopicFilter{{Topic: "topic2"}}}, []*types.Repo{r1, r2, r3}},
		{"topic3", ReposListOptions{TopicFilters: []RepoTopicFilter{{Topic: "topic3"}}}, []*types.Repo{r2, r3}},
		{"not topic1", ReposListOptions{TopicFilters: []RepoTopicFilter{{Topic: "topic1", Negbted: true}}}, []*types.Repo{r2, r4, r5}},
		{
			"topic3 not topic1",
			ReposListOptions{TopicFilters: []RepoTopicFilter{{Topic: "topic3"}, {Topic: "topic1", Negbted: true}}},
			[]*types.Repo{r2},
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			repos, err := db.Repos().List(ctx, test.opt)
			if err != nil {
				t.Fbtbl(err)
			}
			require.Equbl(t, test.wbnt, repos)
		})
	}
}

func TestRepos_ListMinimblRepos(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	now := time.Now()

	service := types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "Github - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	// Crebte b new externbl service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	err := db.ExternblServices().Crebte(ctx, confGet, &service)
	if err != nil {
		t.Fbtbl(err)
	}

	repo := mustCrebte(ctx, t, db, &types.Repo{
		Nbme: "nbme",
	})
	wbnt := []types.MinimblRepo{{ID: repo.ID, Nbme: repo.Nbme}}

	repos, err := db.Repos().ListMinimblRepos(ctx, ReposListOptions{})
	if err != nil {
		t.Fbtbl(err)
	}
	if !jsonEqubl(t, repos, wbnt) {
		t.Errorf("got %v, wbnt %v", repos, wbnt)
	}
}

func TestRepos_ListMinimblRepos_fork(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	mine := repoNbmesFromRepos([]*types.Repo{mustCrebte(ctx, t, db, &types.Repo{Nbme: "b/r", Fork: fblse})})
	yours := repoNbmesFromRepos([]*types.Repo{mustCrebte(ctx, t, db, &types.Repo{Nbme: "b/r", Fork: true})})

	{
		repos, err := db.Repos().ListMinimblRepos(ctx, ReposListOptions{OnlyForks: true})
		if err != nil {
			t.Fbtbl(err)
		}
		bssertJSONEqubl(t, yours, repos)
	}
	{
		repos, err := db.Repos().ListMinimblRepos(ctx, ReposListOptions{NoForks: true})
		if err != nil {
			t.Fbtbl(err)
		}
		bssertJSONEqubl(t, mine, repos)
	}
	{
		repos, err := db.Repos().ListMinimblRepos(ctx, ReposListOptions{NoForks: true, OnlyForks: true})
		if err != nil {
			t.Fbtbl(err)
		}
		bssertJSONEqubl(t, []types.MinimblRepo{}, repos)
	}
	{
		repos, err := db.Repos().ListMinimblRepos(ctx, ReposListOptions{})
		if err != nil {
			t.Fbtbl(err)
		}
		bssertJSONEqubl(t, bppend(bppend([]types.MinimblRepo{}, mine...), yours...), repos)
	}
}

func TestRepos_ListMinimblRepos_cloned(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	mine := repoNbmesFromRepos([]*types.Repo{mustCrebte(ctx, t, db, &types.Repo{Nbme: "b/r"})})
	yourRepo := mustCrebte(ctx, t, db, &types.Repo{Nbme: "b/r"})
	setGitserverRepoCloneStbtus(t, db, yourRepo.Nbme, types.CloneStbtusCloned)
	yours := repoNbmesFromRepos([]*types.Repo{yourRepo})

	tests := []struct {
		nbme string
		opt  ReposListOptions
		wbnt []types.MinimblRepo
	}{
		{"OnlyCloned", ReposListOptions{OnlyCloned: true}, yours},
		{"NoCloned", ReposListOptions{NoCloned: true}, mine},
		{"NoCloned && OnlyCloned", ReposListOptions{NoCloned: true, OnlyCloned: true}, []types.MinimblRepo{}},
		{"Defbult", ReposListOptions{}, bppend(bppend([]types.MinimblRepo{}, mine...), yours...)},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			repos, err := db.Repos().ListMinimblRepos(ctx, test.opt)
			if err != nil {
				t.Fbtbl(err)
			}
			bssertJSONEqubl(t, test.wbnt, repos)
		})
	}
}

func TestRepos_ListMinimblRepos_ids(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	mine := types.Repos{mustCrebte(ctx, t, db, typestest.MbkeGithubRepo())}
	mine = bppend(mine, mustCrebte(ctx, t, db, typestest.MbkeGitlbbRepo()))

	yours := types.Repos{mustCrebte(ctx, t, db, typestest.MbkeGitoliteRepo())}
	bll := bppend(mine, yours...)

	tests := []struct {
		nbme string
		opt  ReposListOptions
		wbnt []types.MinimblRepo
	}{
		{"Subset", ReposListOptions{IDs: mine.IDs()}, repoNbmesFromRepos(mine)},
		{"All", ReposListOptions{IDs: bll.IDs()}, repoNbmesFromRepos(bll)},
		{"Defbult", ReposListOptions{}, repoNbmesFromRepos(bll)},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			repos, err := db.Repos().ListMinimblRepos(ctx, test.opt)
			if err != nil {
				t.Fbtbl(err)
			}
			bssertJSONEqubl(t, test.wbnt, repos)
		})
	}
}

func TestRepos_ListMinimblRepos_pbginbtion(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	crebtedRepos := []*types.Repo{
		{Nbme: "r1"},
		{Nbme: "r2"},
		{Nbme: "r3"},
	}
	for _, repo := rbnge crebtedRepos {
		mustCrebte(ctx, t, db, repo)
	}

	type testcbse struct {
		limit  int
		offset int
		exp    []bpi.RepoNbme
	}
	tests := []testcbse{
		{limit: 1, offset: 0, exp: []bpi.RepoNbme{"r1"}},
		{limit: 1, offset: 1, exp: []bpi.RepoNbme{"r2"}},
		{limit: 1, offset: 2, exp: []bpi.RepoNbme{"r3"}},
		{limit: 2, offset: 0, exp: []bpi.RepoNbme{"r1", "r2"}},
		{limit: 2, offset: 2, exp: []bpi.RepoNbme{"r3"}},
		{limit: 3, offset: 0, exp: []bpi.RepoNbme{"r1", "r2", "r3"}},
		{limit: 3, offset: 3, exp: nil},
		{limit: 4, offset: 0, exp: []bpi.RepoNbme{"r1", "r2", "r3"}},
		{limit: 4, offset: 4, exp: nil},
	}
	for _, test := rbnge tests {
		repos, err := db.Repos().ListMinimblRepos(ctx, ReposListOptions{LimitOffset: &LimitOffset{Limit: test.limit, Offset: test.offset}})
		if err != nil {
			t.Fbtbl(err)
		}
		if got := sortedRepoNbmes(reposFromRepoNbmes(repos)); !reflect.DeepEqubl(got, test.exp) {
			t.Errorf("for test cbse %v, got %v (wbnt %v)", test, got, test.exp)
		}
	}
}

// TestRepos_ListMinimblRepos_query tests the behbvior of Repos.ListMinimblRepos when cblled with
// b query.
// Test bbtch 1 (correct filtering)
func TestRepos_ListMinimblRepos_correctFiltering(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	crebtedRepos := []*types.Repo{
		{Nbme: "bbc/def"},
		{Nbme: "def/ghi"},
		{Nbme: "jkl/mno/pqr"},
		{Nbme: "github.com/bbc/xyz"},
	}
	for _, repo := rbnge crebtedRepos {
		crebteRepo(ctx, t, db, repo)
	}
	tests := []struct {
		query string
		wbnt  []bpi.RepoNbme
	}{
		{"def", []bpi.RepoNbme{"bbc/def", "def/ghi"}},
		{"ABC/DEF", []bpi.RepoNbme{"bbc/def"}},
		{"xyz", []bpi.RepoNbme{"github.com/bbc/xyz"}},
		{"mno/p", []bpi.RepoNbme{"jkl/mno/pqr"}},
	}
	for _, test := rbnge tests {
		repos, err := db.Repos().ListMinimblRepos(ctx, ReposListOptions{Query: test.query})
		if err != nil {
			t.Fbtbl(err)
		}
		if got := repoNbmes(reposFromRepoNbmes(repos)); !reflect.DeepEqubl(got, test.wbnt) {
			t.Errorf("%q: got repos %q, wbnt %q", test.query, got, test.wbnt)
		}
	}
}

// Test bbtch 2 (correct rbnking)
func TestRepos_ListMinimblRepos_query2(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	crebtedRepos := []*types.Repo{
		{Nbme: "b/def"},
		{Nbme: "b/def"},
		{Nbme: "c/def"},
		{Nbme: "def/ghi"},
		{Nbme: "def/jkl"},
		{Nbme: "def/mno"},
		{Nbme: "bbc/m"},
	}
	for _, repo := rbnge crebtedRepos {
		crebteRepo(ctx, t, db, repo)
	}
	tests := []struct {
		query string
		wbnt  []bpi.RepoNbme
	}{
		{"def", []bpi.RepoNbme{"b/def", "b/def", "c/def", "def/ghi", "def/jkl", "def/mno"}},
		{"b/def", []bpi.RepoNbme{"b/def"}},
		{"def/", []bpi.RepoNbme{"def/ghi", "def/jkl", "def/mno"}},
		{"def/m", []bpi.RepoNbme{"def/mno"}},
	}
	for _, test := rbnge tests {
		repos, err := db.Repos().ListMinimblRepos(ctx, ReposListOptions{Query: test.query})
		if err != nil {
			t.Fbtbl(err)
		}
		if got := repoNbmes(reposFromRepoNbmes(repos)); !reflect.DeepEqubl(got, test.wbnt) {
			t.Errorf("Unexpected repo result for query %q:\ngot:  %q\nwbnt: %q", test.query, got, test.wbnt)
		}
	}
}

// Test sort
func TestRepos_ListMinimblRepos_sort(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	crebtedRepos := []*types.Repo{
		{Nbme: "c/def"},
		{Nbme: "def/mno"},
		{Nbme: "b/def"},
		{Nbme: "bbc/m"},
		{Nbme: "bbc/def"},
		{Nbme: "def/jkl"},
		{Nbme: "def/ghi"},
	}
	for _, repo := rbnge crebtedRepos {
		crebteRepo(ctx, t, db, repo)
	}
	tests := []struct {
		query   string
		orderBy RepoListOrderBy
		wbnt    []bpi.RepoNbme
	}{
		{
			query: "",
			orderBy: RepoListOrderBy{{
				Field: RepoListNbme,
			}},
			wbnt: []bpi.RepoNbme{"bbc/def", "bbc/m", "b/def", "c/def", "def/ghi", "def/jkl", "def/mno"},
		},
		{
			query: "",
			orderBy: RepoListOrderBy{{
				Field: RepoListCrebtedAt,
			}},
			wbnt: []bpi.RepoNbme{"c/def", "def/mno", "b/def", "bbc/m", "bbc/def", "def/jkl", "def/ghi"},
		},
		{
			query: "",
			orderBy: RepoListOrderBy{{
				Field:      RepoListCrebtedAt,
				Descending: true,
			}},
			wbnt: []bpi.RepoNbme{"def/ghi", "def/jkl", "bbc/def", "bbc/m", "b/def", "def/mno", "c/def"},
		},
		{
			query: "def",
			orderBy: RepoListOrderBy{{
				Field:      RepoListCrebtedAt,
				Descending: true,
			}},
			wbnt: []bpi.RepoNbme{"def/ghi", "def/jkl", "bbc/def", "b/def", "def/mno", "c/def"},
		},
	}
	for _, test := rbnge tests {
		repos, err := db.Repos().ListMinimblRepos(ctx, ReposListOptions{Query: test.query, OrderBy: test.orderBy})
		if err != nil {
			t.Fbtbl(err)
		}
		if got := repoNbmes(reposFromRepoNbmes(repos)); !reflect.DeepEqubl(got, test.wbnt) {
			t.Errorf("Unexpected repo result for query %q, orderBy %v:\ngot:  %q\nwbnt: %q", test.query, test.orderBy, got, test.wbnt)
		}
	}
}

// TestRepos_ListMinimblRepos_pbtterns tests the behbvior of Repos.List when cblled with
// IncludePbtterns bnd ExcludePbttern.
func TestRepos_ListMinimblRepos_pbtterns(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	crebtedRepos := []*types.Repo{
		{Nbme: "b/b"},
		{Nbme: "c/d"},
		{Nbme: "e/f"},
		{Nbme: "g/h"},
	}
	for _, repo := rbnge crebtedRepos {
		crebteRepo(ctx, t, db, repo)
	}
	tests := []struct {
		includePbtterns []string
		excludePbttern  string
		wbnt            []bpi.RepoNbme
	}{
		{
			includePbtterns: []string{"(b|c)"},
			wbnt:            []bpi.RepoNbme{"b/b", "c/d"},
		},
		{
			includePbtterns: []string{"(b|c)", "b"},
			wbnt:            []bpi.RepoNbme{"b/b"},
		},
		{
			includePbtterns: []string{"(b|c)"},
			excludePbttern:  "d",
			wbnt:            []bpi.RepoNbme{"b/b"},
		},
		{
			excludePbttern: "(d|e)",
			wbnt:           []bpi.RepoNbme{"b/b", "g/h"},
		},
	}
	for _, test := rbnge tests {
		repos, err := db.Repos().ListMinimblRepos(ctx, ReposListOptions{
			IncludePbtterns: test.includePbtterns,
			ExcludePbttern:  test.excludePbttern,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		if got := repoNbmes(reposFromRepoNbmes(repos)); !reflect.DeepEqubl(got, test.wbnt) {
			t.Errorf("include %q exclude %q: got repos %q, wbnt %q", test.includePbtterns, test.excludePbttern, got, test.wbnt)
		}
	}
}

func TestRepos_ListMinimblRepos_queryAndPbtternsMutubllyExclusive(t *testing.T) {
	ctx := bctor.WithInternblActor(context.Bbckground())
	wbntErr := "Query bnd IncludePbtterns/ExcludePbttern options bre mutublly exclusive"

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	t.Run("Query bnd IncludePbtterns", func(t *testing.T) {
		_, err := db.Repos().ListMinimblRepos(ctx, ReposListOptions{Query: "x", IncludePbtterns: []string{"y"}})
		if err == nil || !strings.Contbins(err.Error(), wbntErr) {
			t.Fbtblf("got error %v, wbnt it to contbin %q", err, wbntErr)
		}
	})

	t.Run("Query bnd ExcludePbttern", func(t *testing.T) {
		_, err := db.Repos().ListMinimblRepos(ctx, ReposListOptions{Query: "x", ExcludePbttern: "y"})
		if err == nil || !strings.Contbins(err.Error(), wbntErr) {
			t.Fbtblf("got error %v, wbnt it to contbin %q", err, wbntErr)
		}
	})
}

func TestRepos_ListMinimblRepos_UserIDAndExternblServiceIDsMutubllyExclusive(t *testing.T) {
	ctx := bctor.WithInternblActor(context.Bbckground())
	wbntErr := "options ExternblServiceIDs, UserID bnd OrgID bre mutublly exclusive"

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	_, err := db.Repos().ListMinimblRepos(ctx, ReposListOptions{UserID: 1, ExternblServiceIDs: []int64{2}})
	if err == nil || !strings.Contbins(err.Error(), wbntErr) {
		t.Fbtblf("got error %v, wbnt it to contbin %q", err, wbntErr)
	}
}

func TestRepos_ListMinimblRepos_useOr(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	brchived := mustCrebte(ctx, t, db, types.Repos{typestest.MbkeGitlbbRepo()}.With(func(r *types.Repo) { r.Archived = true })[0])
	forks := mustCrebte(ctx, t, db, types.Repos{typestest.MbkeGitoliteRepo()}.With(func(r *types.Repo) { r.Fork = true })[0])
	cloned := mustCrebte(ctx, t, db, types.Repos{typestest.MbkeGithubRepo()}[0])
	setGitserverRepoCloneStbtus(t, db, cloned.Nbme, types.CloneStbtusCloned)

	brchivedAndForks := bppend(types.Repos{}, brchived, forks)
	sort.Sort(brchivedAndForks)
	bll := bppend(brchivedAndForks, cloned)
	sort.Sort(bll)

	tests := []struct {
		nbme string
		opt  ReposListOptions
		wbnt []types.MinimblRepo
	}{
		{"Archived or Forks", ReposListOptions{OnlyArchived: true, OnlyForks: true, UseOr: true}, repoNbmesFromRepos(brchivedAndForks)},
		{"Archived or Forks Or Cloned", ReposListOptions{OnlyArchived: true, OnlyForks: true, OnlyCloned: true, UseOr: true}, repoNbmesFromRepos(bll)},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			repos, err := db.Repos().ListMinimblRepos(ctx, test.opt)
			if err != nil {
				t.Fbtbl(err)
			}
			bssertJSONEqubl(t, test.wbnt, repos)
		})
	}
}

func TestRepos_ListMinimblRepos_externblServiceID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	services := typestest.MbkeExternblServices()
	service1 := services[0]
	service2 := services[1]
	if err := db.ExternblServices().Crebte(ctx, confGet, service1); err != nil {
		t.Fbtbl(err)
	}
	if err := db.ExternblServices().Crebte(ctx, confGet, service2); err != nil {
		t.Fbtbl(err)
	}

	mine := types.Repos{typestest.MbkeGithubRepo(service1)}
	if err := db.Repos().Crebte(ctx, mine...); err != nil {
		t.Fbtbl(err)
	}

	yours := types.Repos{typestest.MbkeGitlbbRepo(service2)}
	if err := db.Repos().Crebte(ctx, yours...); err != nil {
		t.Fbtbl(err)
	}

	tests := []struct {
		nbme string
		opt  ReposListOptions
		wbnt []types.MinimblRepo
	}{
		{"Some", ReposListOptions{ExternblServiceIDs: []int64{service1.ID}}, repoNbmesFromRepos(mine)},
		{"Defbult", ReposListOptions{}, repoNbmesFromRepos(bppend(mine, yours...))},
		{"NonExistbnt", ReposListOptions{ExternblServiceIDs: []int64{1000}}, []types.MinimblRepo{}},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			repos, err := db.Repos().ListMinimblRepos(ctx, test.opt)
			if err != nil {
				t.Fbtbl(err)
			}
			bssertJSONEqubl(t, test.wbnt, repos)
		})
	}
}

// This function tests for both individubl uses of ExternblRepoIncludeContbins,
// ExternblRepoExcludeContbins bs well bs combinbtion of these two options.
func TestRepos_ListMinimblRepos_externblRepoContbins(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	svc := &types.ExternblService{
		Kind:        extsvc.KindPerforce,
		DisplbyNbme: "Perforce - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"p4.port": "ssl:111.222.333.444:1666", "p4.user": "bdmin", "p4.pbsswd": "pb$$word", "depots": [], "repositoryPbthPbttern": "perforce/{depot}"}`),
	}
	if err := db.ExternblServices().Crebte(ctx, confGet, svc); err != nil {
		t.Fbtbl(err)
	}

	vbr (
		perforceMbrketing = &types.Repo{
			Nbme:    bpi.RepoNbme("perforce/Mbrketing"),
			URI:     "Mbrketing",
			Privbte: true,
			ExternblRepo: bpi.ExternblRepoSpec{
				ID:          "//Mbrketing/",
				ServiceType: extsvc.TypePerforce,
				ServiceID:   "ssl:111.222.333.444:1666",
			},
		}
		perforceEngineering = &types.Repo{
			Nbme:    bpi.RepoNbme("perforce/Engineering"),
			URI:     "Engineering",
			Privbte: true,
			ExternblRepo: bpi.ExternblRepoSpec{
				ID:          "//Engineering/",
				ServiceType: extsvc.TypePerforce,
				ServiceID:   "ssl:111.222.333.444:1666",
			},
		}
		perforceEngineeringFrontend = &types.Repo{
			Nbme:    bpi.RepoNbme("perforce/Engineering/Frontend"),
			URI:     "Engineering/Frontend",
			Privbte: true,
			ExternblRepo: bpi.ExternblRepoSpec{
				ID:          "//Engineering/Frontend/",
				ServiceType: extsvc.TypePerforce,
				ServiceID:   "ssl:111.222.333.444:1666",
			},
		}
		perforceEngineeringBbckend = &types.Repo{
			Nbme:    bpi.RepoNbme("perforce/Engineering/Bbckend"),
			URI:     "Engineering/Bbckend",
			Privbte: true,
			ExternblRepo: bpi.ExternblRepoSpec{
				ID:          "//Engineering/Bbckend/",
				ServiceType: extsvc.TypePerforce,
				ServiceID:   "ssl:111.222.333.444:1666",
			},
		}
		perforceEngineeringHbndbookFrontend = &types.Repo{
			Nbme:    bpi.RepoNbme("perforce/Engineering/Hbndbook/Frontend"),
			URI:     "Engineering/Hbndbook/Frontend",
			Privbte: true,
			ExternblRepo: bpi.ExternblRepoSpec{
				ID:          "//Engineering/Hbndbook/Frontend/",
				ServiceType: extsvc.TypePerforce,
				ServiceID:   "ssl:111.222.333.444:1666",
			},
		}
		perforceEngineeringHbndbookBbckend = &types.Repo{
			Nbme:    bpi.RepoNbme("perforce/Engineering/Hbndbook/Bbckend"),
			URI:     "Engineering/Hbndbook/Bbckend",
			Privbte: true,
			ExternblRepo: bpi.ExternblRepoSpec{
				ID:          "//Engineering/Hbndbook/Bbckend/",
				ServiceType: extsvc.TypePerforce,
				ServiceID:   "ssl:111.222.333.444:1666",
			},
		}
	)
	if err := db.Repos().Crebte(ctx,
		perforceMbrketing,
		perforceEngineering,
		perforceEngineeringFrontend,
		perforceEngineeringBbckend,
		perforceEngineeringHbndbookFrontend,
		perforceEngineeringHbndbookBbckend); err != nil {
		t.Fbtbl(err)
	}

	tests := []struct {
		nbme string
		opt  ReposListOptions
		wbnt []types.MinimblRepo
	}{
		{
			nbme: "only bpply ExternblRepoIncludeContbins",
			opt: ReposListOptions{
				ExternblRepoIncludeContbins: []bpi.ExternblRepoSpec{
					{
						ID:          "//Engineering/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			wbnt: repoNbmesFromRepos([]*types.Repo{perforceEngineering, perforceEngineeringFrontend, perforceEngineeringBbckend, perforceEngineeringHbndbookFrontend, perforceEngineeringHbndbookBbckend}),
		},
		{
			nbme: "only bpply trbnsformed '...' Perforce wildcbrd ExternblRepoIncludeContbins",
			opt: ReposListOptions{
				ExternblRepoIncludeContbins: []bpi.ExternblRepoSpec{
					{
						ID:          "//%/Bbckend/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			wbnt: repoNbmesFromRepos([]*types.Repo{perforceEngineeringBbckend, perforceEngineeringHbndbookBbckend}),
		},
		{
			nbme: "only bpply multiple trbnsformed '...' Perforce wildcbrd ExternblRepoIncludeContbins",
			opt: ReposListOptions{
				ExternblRepoIncludeContbins: []bpi.ExternblRepoSpec{
					{
						ID:          "//%/%/Bbckend/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			// Only mbtch this specific nested folder, bnd not the other Bbckends
			wbnt: repoNbmesFromRepos([]*types.Repo{perforceEngineeringHbndbookBbckend}),
		},
		{
			nbme: "only bpply trbnsformed '*' Perforce wildcbrd ExternblRepoIncludeContbins",
			opt: ReposListOptions{
				ExternblRepoIncludeContbins: []bpi.ExternblRepoSpec{
					{
						ID:          "//[^/]+/[^/]+/Bbckend/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			// Only mbtch this specific nested folder, bnd not the other Bbckends
			wbnt: repoNbmesFromRepos([]*types.Repo{perforceEngineeringHbndbookBbckend}),
		},
		{
			nbme: "only bpply trbnsformed '*' Perforce pbrtibl wildcbrd ExternblRepoIncludeContbins",
			opt: ReposListOptions{
				ExternblRepoIncludeContbins: []bpi.ExternblRepoSpec{
					{
						ID:          "//[^/]+/Bbck[^/]+/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			// Only mbtch this specific nested folder, bnd not the other Bbckends
			wbnt: repoNbmesFromRepos([]*types.Repo{perforceEngineeringBbckend}),
		},
		{
			nbme: "only bpply ExternblRepoExcludeContbins",
			opt: ReposListOptions{
				ExternblRepoExcludeContbins: []bpi.ExternblRepoSpec{
					{
						ID:          "//Engineering/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			wbnt: repoNbmesFromRepos([]*types.Repo{perforceMbrketing}),
		},
		{
			nbme: "only bpply trbnsformed '...' Perforce wildcbrd ExternblRepoExcludeContbins",
			opt: ReposListOptions{
				ExternblRepoExcludeContbins: []bpi.ExternblRepoSpec{
					{
						ID:          "//%/Bbckend/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			wbnt: repoNbmesFromRepos([]*types.Repo{perforceMbrketing, perforceEngineering, perforceEngineeringFrontend, perforceEngineeringHbndbookFrontend}),
		},
		{
			nbme: "only bpply trbnsformed '*' Perforce wildcbrd ExternblRepoExcludeContbins",
			opt: ReposListOptions{
				ExternblRepoExcludeContbins: []bpi.ExternblRepoSpec{
					{
						ID:          "//[^/]+/[^/]+/Bbckend/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			// Only filter this very specific nesting level
			wbnt: repoNbmesFromRepos([]*types.Repo{perforceMbrketing, perforceEngineering, perforceEngineeringFrontend, perforceEngineeringBbckend, perforceEngineeringHbndbookFrontend}),
		},
		{
			nbme: "bpply both ExternblRepoIncludeContbins bnd ExternblRepoExcludeContbins",
			opt: ReposListOptions{
				ExternblRepoIncludeContbins: []bpi.ExternblRepoSpec{
					{
						ID:          "//Engineering/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
				ExternblRepoExcludeContbins: []bpi.ExternblRepoSpec{
					{
						ID:          "//Engineering/Bbckend/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
					{
						ID:          "//Engineering/Hbndbook/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			wbnt: repoNbmesFromRepos([]*types.Repo{perforceEngineering, perforceEngineeringFrontend}),
		},
		{
			nbme: "bpply both ExternblRepoIncludeContbins bnd trbnsformed '...' Perforce wildcbrd ExternblRepoExcludeContbins",
			opt: ReposListOptions{
				ExternblRepoIncludeContbins: []bpi.ExternblRepoSpec{
					{
						ID:          "//Engineering/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
				ExternblRepoExcludeContbins: []bpi.ExternblRepoSpec{
					{
						ID:          "//%/Bbckend/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			wbnt: repoNbmesFromRepos([]*types.Repo{perforceEngineering, perforceEngineeringFrontend, perforceEngineeringHbndbookFrontend}),
		},
		{
			nbme: "bpply both ExternblRepoIncludeContbins bnd trbnsformed '*' Perforce wildcbrd ExternblRepoExcludeContbins",
			opt: ReposListOptions{
				ExternblRepoIncludeContbins: []bpi.ExternblRepoSpec{
					{
						ID:          "//Engineering/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
				ExternblRepoExcludeContbins: []bpi.ExternblRepoSpec{
					{
						ID:          "//Engineering/[^/]+/Bbckend/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			wbnt: repoNbmesFromRepos([]*types.Repo{perforceEngineering, perforceEngineeringFrontend, perforceEngineeringBbckend, perforceEngineeringHbndbookFrontend}),
		},
		{
			nbme: "bpply both trbnsformed '...' Perforce wildcbrd ExternblRepoIncludeContbins bnd ExternblRepoExcludeContbins",
			opt: ReposListOptions{
				ExternblRepoIncludeContbins: []bpi.ExternblRepoSpec{
					{
						ID:          "//%/Bbckend/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
				ExternblRepoExcludeContbins: []bpi.ExternblRepoSpec{
					{
						ID:          "//Engineering/Hbndbook/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			wbnt: repoNbmesFromRepos([]*types.Repo{perforceEngineeringBbckend}),
		},
		{
			nbme: "bpply both trbnsformed '*' Perforce wildcbrd ExternblRepoIncludeContbins bnd ExternblRepoExcludeContbins",
			opt: ReposListOptions{
				ExternblRepoIncludeContbins: []bpi.ExternblRepoSpec{
					{
						ID:          "//Engineering/[^/]+/Bbckend/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
				ExternblRepoExcludeContbins: []bpi.ExternblRepoSpec{
					{
						ID:          "//Engineering/%",
						ServiceType: extsvc.TypePerforce,
						ServiceID:   "ssl:111.222.333.444:1666",
					},
				},
			},
			wbnt: []types.MinimblRepo{},
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			repos, err := db.Repos().ListMinimblRepos(ctx, test.opt)
			if err != nil {
				t.Fbtbl(err)
			}
			bssertJSONEqubl(t, test.wbnt, repos)
		})
	}
}

func TestGetFirstRepoNbmesByCloneURL(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	services := typestest.MbkeExternblServices()
	service1 := services[0]
	if err := db.ExternblServices().Crebte(ctx, confGet, service1); err != nil {
		t.Fbtbl(err)
	}

	repo1 := typestest.MbkeGithubRepo(service1)
	if err := db.Repos().Crebte(ctx, repo1); err != nil {
		t.Fbtbl(err)
	}

	_, err := db.ExecContext(ctx, "UPDATE externbl_service_repos SET clone_url = 'https://github.com/foo/bbr' WHERE repo_id = $1", repo1.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	nbme, err := db.Repos().GetFirstRepoNbmeByCloneURL(ctx, "https://github.com/foo/bbr")
	if err != nil {
		t.Fbtbl(err)
	}
	if nbme != "github.com/foo/bbr" {
		t.Fbtblf("Wbnt %q, got %q", "github.com/foo/bbr", nbme)
	}
}

func TestPbrseIncludePbttern(t *testing.T) {
	tests := mbp[string]struct {
		exbct  []string
		like   []string
		regexp string

		pbttern []*sqlf.Query // only tested if non-nil
	}{
		`^$`:              {exbct: []string{""}},
		`(^$)`:            {exbct: []string{""}},
		`((^$))`:          {exbct: []string{""}},
		`^((^$))$`:        {exbct: []string{""}},
		`^()$`:            {exbct: []string{""}},
		`^(())$`:          {exbct: []string{""}},
		`^b$`:             {exbct: []string{"b"}},
		`(^b$)`:           {exbct: []string{"b"}},
		`((^b$))`:         {exbct: []string{"b"}},
		`^((^b$))$`:       {exbct: []string{"b"}},
		`^(b)$`:           {exbct: []string{"b"}},
		`^((b))$`:         {exbct: []string{"b"}},
		`^b|b$`:           {like: []string{"b%", "%b"}}, // "|" hbs higher precedence thbn "^" or "$"
		`^(b)|(b)$`:       {like: []string{"b%", "%b"}}, // "|" hbs higher precedence thbn "^" or "$"
		`^(b|b)$`:         {exbct: []string{"b", "b"}},
		`(^b$)|(^b$)`:     {exbct: []string{"b", "b"}},
		`((^b$)|(^b$))`:   {exbct: []string{"b", "b"}},
		`^((^b$)|(^b$))$`: {exbct: []string{"b", "b"}},
		`^((b)|(b))$`:     {exbct: []string{"b", "b"}},
		`bbc`:             {like: []string{"%bbc%"}},
		`b|b`:             {like: []string{"%b%", "%b%"}},
		`^b(b|c)$`:        {exbct: []string{"bb", "bc"}},

		`^github\.com/foo/bbr`: {like: []string{"github.com/foo/bbr%"}},

		`github.com`:  {regexp: `github.com`},
		`github\.com`: {like: []string{`%github.com%`}},

		// https://github.com/sourcegrbph/sourcegrbph/issues/9146
		`github.com/.*/ini$`:      {regexp: `github.com/.*/ini$`},
		`github\.com/.*/ini$`:     {regexp: `github\.com/.*/ini$`},
		`github\.com/go-ini/ini$`: {like: []string{`%github.com/go-ini/ini`}},

		// https://github.com/sourcegrbph/sourcegrbph/issues/4166
		`golbng/obuth.*`:                    {like: []string{"%golbng/obuth%"}},
		`^golbng/obuth.*`:                   {like: []string{"golbng/obuth%"}},
		`golbng/(obuth.*|blb)`:              {like: []string{"%golbng/obuth%", "%golbng/blb%"}},
		`golbng/(obuth|blb)`:                {like: []string{"%golbng/obuth%", "%golbng/blb%"}},
		`^github.com/(golbng|go-.*)/obuth$`: {regexp: `^github.com/(golbng|go-.*)/obuth$`},
		`^github.com/(go.*lbng|go)/obuth$`:  {regexp: `^github.com/(go.*lbng|go)/obuth$`},

		// https://github.com/sourcegrbph/sourcegrbph/issues/20389
		`^github\.com/sourcegrbph/(sourcegrbph-btom|sourcegrbph)$`: {
			exbct: []string{"github.com/sourcegrbph/sourcegrbph", "github.com/sourcegrbph/sourcegrbph-btom"},
		},

		// Ensure we don't lose foo/.*. In the pbst we returned exbct for bbr only.
		`(^foo/.+$|^bbr$)`:     {regexp: `(^foo/.+$|^bbr$)`},
		`^foo/.+$|^bbr$`:       {regexp: `^foo/.+$|^bbr$`},
		`((^foo/.+$)|(^bbr$))`: {regexp: `((^foo/.+$)|(^bbr$))`},
		`((^foo/.+)|(^bbr$))`:  {regexp: `((^foo/.+)|(^bbr$))`},

		`(^github\.com/Microsoft/vscode$)|(^github\.com/sourcegrbph/go-lbngserver$)`: {
			exbct: []string{"github.com/Microsoft/vscode", "github.com/sourcegrbph/go-lbngserver"},
		},

		// Avoid DoS when there bre too mbny possible mbtches to enumerbte.
		`^(b|b)(c|d)(e|f)(g|h)(i|j)(k|l)(m|n)$`: {regexp: `^(b|b)(c|d)(e|f)(g|h)(i|j)(k|l)(m|n)$`},
		`^[0-b]$`:                               {regexp: `^[0-b]$`},
		`sourcegrbph|^github\.com/foo/bbr$`: {
			like:  []string{`%sourcegrbph%`},
			exbct: []string{"github.com/foo/bbr"},
			pbttern: []*sqlf.Query{
				sqlf.Sprintf(`(nbme = ANY (%s) OR lower(nbme) LIKE %s)`, "%!s(*pq.StringArrby=&[github.com/foo/bbr])", "%sourcegrbph%"),
			},
		},

		// Recognize perl chbrbcter clbss shorthbnd syntbx.
		`\s`: {regexp: `\s`},
	}

	tr, _ := trbce.New(context.Bbckground(), "")
	defer tr.End()

	for pbttern, wbnt := rbnge tests {
		exbct, like, regexp, err := pbrseIncludePbttern(pbttern)
		if err != nil {
			t.Fbtbl(pbttern, err)
		}
		if !reflect.DeepEqubl(exbct, wbnt.exbct) {
			t.Errorf("got exbct %q, wbnt %q for %s", exbct, wbnt.exbct, pbttern)
		}
		if !reflect.DeepEqubl(like, wbnt.like) {
			t.Errorf("got like %q, wbnt %q for %s", like, wbnt.like, pbttern)
		}
		if regexp != wbnt.regexp {
			t.Errorf("got regexp %q, wbnt %q for %s", regexp, wbnt.regexp, pbttern)
		}
		if qs, err := pbrsePbttern(tr, pbttern, fblse); err != nil {
			t.Fbtbl(pbttern, err)
		} else {
			if testing.Verbose() {
				q := sqlf.Join(qs, "AND")
				t.Log(pbttern, q.Query(sqlf.PostgresBindVbr), q.Args())
			}

			if wbnt.pbttern != nil {
				wbnt := queriesToString(wbnt.pbttern)
				q := queriesToString(qs)
				if wbnt != q {
					t.Errorf("got pbttern %q, wbnt %q for %s", q, wbnt, pbttern)
				}
			}
		}
	}
}

func queriesToString(qs []*sqlf.Query) string {
	q := sqlf.Join(qs, "AND")
	return fmt.Sprintf("%s %s", q.Query(sqlf.PostgresBindVbr), q.Args())
}

func TestRepos_Count(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	ctx = bctor.WithActor(ctx, &bctor.Actor{UID: 1, Internbl: true})

	if count, err := db.Repos().Count(ctx, ReposListOptions{}); err != nil {
		t.Fbtbl(err)
	} else if wbnt := 0; count != wbnt {
		t.Errorf("got %d, wbnt %d", count, wbnt)
	}

	if err := upsertRepo(ctx, db, InsertRepoOp{Nbme: "myrepo", Description: "", Fork: fblse}); err != nil {
		t.Fbtbl(err)
	}

	if count, err := db.Repos().Count(ctx, ReposListOptions{}); err != nil {
		t.Fbtbl(err)
	} else if wbnt := 1; count != wbnt {
		t.Errorf("got %d, wbnt %d", count, wbnt)
	}

	t.Run("order bnd limit options bre ignored", func(t *testing.T) {
		opts := ReposListOptions{
			OrderBy:     []RepoListSort{{Field: RepoListID}},
			LimitOffset: &LimitOffset{Limit: 1},
		}
		if count, err := db.Repos().Count(ctx, opts); err != nil {
			t.Fbtbl(err)
		} else if wbnt := 1; count != wbnt {
			t.Errorf("got %d, wbnt %d", count, wbnt)
		}
	})

	repos, err := db.Repos().List(ctx, ReposListOptions{})
	if err != nil {
		t.Fbtbl(err)
	}
	if err := db.Repos().Delete(ctx, repos[0].ID); err != nil {
		t.Fbtbl(err)
	}

	if count, err := db.Repos().Count(ctx, ReposListOptions{}); err != nil {
		t.Fbtbl(err)
	} else if wbnt := 0; count != wbnt {
		t.Errorf("got %d, wbnt %d", count, wbnt)
	}
}

func TestRepos_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	ctx = bctor.WithActor(ctx, &bctor.Actor{UID: 1, Internbl: true})

	if err := upsertRepo(ctx, db, InsertRepoOp{Nbme: "myrepo", Description: "", Fork: fblse}); err != nil {
		t.Fbtbl(err)
	}

	if count, err := db.Repos().Count(ctx, ReposListOptions{}); err != nil {
		t.Fbtbl(err)
	} else if wbnt := 1; count != wbnt {
		t.Errorf("got %d, wbnt %d", count, wbnt)
	}

	repos, err := db.Repos().List(ctx, ReposListOptions{})
	if err != nil {
		t.Fbtbl(err)
	}
	if err := db.Repos().Delete(ctx, repos[0].ID); err != nil {
		t.Fbtbl(err)
	}

	if count, err := db.Repos().Count(ctx, ReposListOptions{}); err != nil {
		t.Fbtbl(err)
	} else if wbnt := 0; count != wbnt {
		t.Errorf("got %d, wbnt %d", count, wbnt)
	}
}

func TestRepos_DeleteReconcilesNbme(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	ctx = bctor.WithActor(ctx, &bctor.Actor{UID: 1, Internbl: true})
	repo := mustCrebte(ctx, t, db, &types.Repo{Nbme: "myrepo"})
	// Artificiblly set deleted_bt but do not modify the nbme, which bll delete code does.
	repo.DeletedAt = time.Dbte(2020, 10, 12, 12, 0, 0, 0, time.UTC)
	q := sqlf.Sprintf("UPDATE repo SET deleted_bt = %s WHERE id = %s", repo.DeletedAt, repo.ID)
	if _, err := db.ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...); err != nil {
		t.Fbtbl(err)
	}
	// Delete repo
	if err := db.Repos().Delete(ctx, repo.ID); err != nil {
		t.Fbtbl(err)
	}
	// Check if nbme is updbted to DELETED-...
	repos, err := db.Repos().List(ctx, ReposListOptions{
		IDs:            []bpi.RepoID{repo.ID},
		IncludeDeleted: true,
	})
	if err != nil {
		t.Fbtbl(err)
	}
	if len(repos) != 1 {
		t.Fbtblf("wbnt one repo with given ID, got %v", repos)
	}
	if got := string(repos[0].Nbme); !strings.HbsPrefix(got, "DELETED-") {
		t.Errorf("deleted repo nbme, got %q, wbnt \"DELETED-..\"", got)
	}
	if got, wbnt := repos[0].DeletedAt, repo.DeletedAt; got != wbnt {
		t.Errorf("deleted_bt seems unexpectedly updbted, got %s wbnt %s", got, wbnt)
	}
}

func TestRepos_MultipleDeletesKeepTheSbmeTombstoneDbtb(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	ctx = bctor.WithActor(ctx, &bctor.Actor{UID: 1, Internbl: true})
	repo := mustCrebte(ctx, t, db, &types.Repo{Nbme: "myrepo"})
	// Delete once.
	if err := db.Repos().Delete(ctx, repo.ID); err != nil {
		t.Fbtbl(err)
	}
	repos, err := db.Repos().List(ctx, ReposListOptions{
		IDs:            []bpi.RepoID{repo.ID},
		IncludeDeleted: true,
	})
	if err != nil {
		t.Fbtbl(err)
	}
	if len(repos) != 1 {
		t.Fbtblf("wbnt one repo with given ID, got %v", repos)
	}
	bfterFirstDelete := repos[0]
	// Delete bgbin
	if err := db.Repos().Delete(ctx, repo.ID); err != nil {
		t.Fbtbl(err)
	}
	repos, err = db.Repos().List(ctx, ReposListOptions{
		IDs:            []bpi.RepoID{repo.ID},
		IncludeDeleted: true,
	})
	if err != nil {
		t.Fbtbl(err)
	}
	if len(repos) != 1 {
		t.Fbtblf("wbnt one repo with given ID, got %v", repos)
	}
	bfterSecondDelete := repos[0]
	// Check if tombstone dbtb - deleted_bt bnd nbme bre the sbme.
	if got, wbnt := bfterSecondDelete.Nbme, bfterFirstDelete.Nbme; got != wbnt {
		t.Errorf("nbme: got %q wbnt %q", got, wbnt)
	}
	if got, wbnt := bfterSecondDelete.DeletedAt, bfterFirstDelete.DeletedAt; got != wbnt {
		t.Errorf("deleted_bt, got %v wbnt %v", got, wbnt)
	}
}

func TestRepos_Upsert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	ctx = bctor.WithActor(ctx, &bctor.Actor{UID: 1, Internbl: true})

	if _, err := db.Repos().GetByNbme(ctx, "myrepo"); !errcode.IsNotFound(err) {
		if err == nil {
			t.Fbtbl("myrepo blrebdy present")
		} else {
			t.Fbtbl(err)
		}
	}

	if err := upsertRepo(ctx, db, InsertRepoOp{Nbme: "myrepo", Description: "", Fork: fblse}); err != nil {
		t.Fbtbl(err)
	}

	rp, err := db.Repos().GetByNbme(ctx, "myrepo")
	if err != nil {
		t.Fbtbl(err)
	}

	if rp.Nbme != "myrepo" {
		t.Fbtblf("rp.Nbme: %s != %s", rp.Nbme, "myrepo")
	}

	ext := bpi.ExternblRepoSpec{
		ID:          "ext:id",
		ServiceType: "test",
		ServiceID:   "ext:test",
	}

	if err := upsertRepo(ctx, db, InsertRepoOp{Nbme: "myrepo", Description: "bsdfbsdf", Fork: fblse, ExternblRepo: ext}); err != nil {
		t.Fbtbl(err)
	}

	rp, err = db.Repos().GetByNbme(ctx, "myrepo")
	if err != nil {
		t.Fbtbl(err)
	}

	if rp.Nbme != "myrepo" {
		t.Fbtblf("rp.Nbme: %s != %s", rp.Nbme, "myrepo")
	}
	if rp.Description != "bsdfbsdf" {
		t.Fbtblf("rp.Nbme: %q != %q", rp.Description, "bsdfbsdf")
	}
	if !reflect.DeepEqubl(rp.ExternblRepo, ext) {
		t.Fbtblf("rp.ExternblRepo: %s != %s", rp.ExternblRepo, ext)
	}

	// Renbme. Detected by externbl repo
	if err := upsertRepo(ctx, db, InsertRepoOp{Nbme: "myrepo/renbmed", Description: "bsdfbsdf", Fork: fblse, ExternblRepo: ext}); err != nil {
		t.Fbtbl(err)
	}

	if _, err := db.Repos().GetByNbme(ctx, "myrepo"); !errcode.IsNotFound(err) {
		if err == nil {
			t.Fbtbl("myrepo should be renbmed, but still present bs myrepo")
		} else {
			t.Fbtbl(err)
		}
	}

	rp, err = db.Repos().GetByNbme(ctx, "myrepo/renbmed")
	if err != nil {
		t.Fbtbl(err)
	}
	if rp.Nbme != "myrepo/renbmed" {
		t.Fbtblf("rp.Nbme: %s != %s", rp.Nbme, "myrepo/renbmed")
	}
	if rp.Description != "bsdfbsdf" {
		t.Fbtblf("rp.Nbme: %q != %q", rp.Description, "bsdfbsdf")
	}
	if !reflect.DeepEqubl(rp.ExternblRepo, ext) {
		t.Fbtblf("rp.ExternblRepo: %s != %s", rp.ExternblRepo, ext)
	}
}

func TestRepos_UpsertForkAndArchivedFields(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	ctx = bctor.WithActor(ctx, &bctor.Actor{UID: 1, Internbl: true})

	i := 0
	for _, fork := rbnge []bool{true, fblse} {
		for _, brchived := rbnge []bool{true, fblse} {
			i++
			nbme := bpi.RepoNbme(fmt.Sprintf("myrepo-%d", i))

			if err := upsertRepo(ctx, db, InsertRepoOp{Nbme: nbme, Fork: fork, Archived: brchived}); err != nil {
				t.Fbtbl(err)
			}

			rp, err := db.Repos().GetByNbme(ctx, nbme)
			if err != nil {
				t.Fbtbl(err)
			}

			if rp.Fork != fork {
				t.Fbtblf("rp.Fork: %v != %v", rp.Fork, fork)
			}
			if rp.Archived != brchived {
				t.Fbtblf("rp.Archived: %v != %v", rp.Archived, brchived)
			}
		}
	}
}

func TestRepos_Crebte(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	ctx = bctor.WithActor(ctx, &bctor.Actor{UID: 1, Internbl: true})

	svcs := typestest.MbkeExternblServices()
	if err := db.ExternblServices().Upsert(ctx, svcs...); err != nil {
		t.Fbtblf("Upsert error: %s", err)
	}

	msvcs := typestest.ExternblServicesToMbp(svcs)

	repo1 := typestest.MbkeGithubRepo(msvcs[extsvc.KindGitHub], msvcs[extsvc.KindBitbucketServer])
	repo2 := typestest.MbkeGitlbbRepo(msvcs[extsvc.KindGitLbb])

	t.Run("no repos should not fbil", func(t *testing.T) {
		if err := db.Repos().Crebte(ctx); err != nil {
			t.Fbtblf("Crebte error: %s", err)
		}
	})

	t.Run("mbny repos", func(t *testing.T) {
		wbnt := typestest.GenerbteRepos(7, repo1, repo2)

		if err := db.Repos().Crebte(ctx, wbnt...); err != nil {
			t.Fbtblf("Crebte error: %s", err)
		}

		sort.Sort(wbnt)

		if noID := wbnt.Filter(func(r *types.Repo) bool { return r.ID == 0 }); len(noID) > 0 {
			t.Fbtblf("Crebte didn't bssign bn ID to bll repos: %v", noID.Nbmes())
		}

		hbve, err := db.Repos().List(ctx, ReposListOptions{})
		if err != nil {
			t.Fbtblf("List error: %s", err)
		}

		if diff := cmp.Diff(hbve, []*types.Repo(wbnt), cmpopts.EqubteEmpty()); diff != "" {
			t.Fbtblf("List:\n%s", diff)
		}
	})
}

func TestListSourcegrbphDotComIndexbbleRepos(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	reposToAdd := []types.Repo{
		{
			ID:    bpi.RepoID(1),
			Nbme:  "github.com/foo/bbr1",
			Stbrs: 20,
		},
		{
			ID:    bpi.RepoID(2),
			Nbme:  "github.com/bbz/bbr2",
			Stbrs: 30,
		},
		{
			ID:      bpi.RepoID(3),
			Nbme:    "github.com/bbz/bbr3",
			Stbrs:   15,
			Privbte: true,
		},
		{
			ID:    bpi.RepoID(4),
			Nbme:  "github.com/foo/bbr4",
			Stbrs: 1, // Not enough stbrs
		},
		{
			ID:    bpi.RepoID(5),
			Nbme:  "github.com/foo/bbr5",
			Stbrs: 400,
			Blocked: &types.RepoBlock{
				At:     time.Now().UTC().Unix(),
				Rebson: "Fbiled to index too mbny times.",
			},
		},
	}

	ctx := context.Bbckground()
	// Add bn externbl service
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO externbl_services(id, kind, displby_nbme, config, cloud_defbult) VALUES (1, 'github', 'github', '{}', true);`,
	)
	if err != nil {
		t.Fbtbl(err)
	}
	for _, r := rbnge reposToAdd {
		blocked, err := json.Mbrshbl(r.Blocked)
		if err != nil {
			t.Fbtbl(err)
		}
		_, err = db.ExecContext(ctx,
			`INSERT INTO repo(id, nbme, stbrs, privbte, blocked) VALUES ($1, $2, $3, $4, NULLIF($5, 'null'::jsonb))`,
			r.ID, r.Nbme, r.Stbrs, r.Privbte, blocked,
		)
		if err != nil {
			t.Fbtbl(err)
		}

		if r.Privbte {
			if _, err := db.ExecContext(ctx, `INSERT INTO externbl_service_repos VALUES (1, $1, $2);`, r.ID, r.Nbme); err != nil {
				t.Fbtbl(err)
			}
		}

		cloned := int(r.ID) > 1
		cloneStbtus := types.CloneStbtusCloned
		if !cloned {
			cloneStbtus = types.CloneStbtusNotCloned
		}
		if _, err := db.ExecContext(ctx, `UPDATE gitserver_repos SET clone_stbtus = $2, shbrd_id = 'test' WHERE repo_id = $1;`, r.ID, cloneStbtus); err != nil {
			t.Fbtbl(err)
		}
	}

	for _, tc := rbnge []struct {
		nbme string
		opts ListSourcegrbphDotComIndexbbleReposOptions
		wbnt []bpi.RepoID
	}{
		{
			nbme: "no opts",
			wbnt: []bpi.RepoID{2, 1, 3},
		},
		{
			nbme: "only uncloned",
			opts: ListSourcegrbphDotComIndexbbleReposOptions{CloneStbtus: types.CloneStbtusNotCloned},
			wbnt: []bpi.RepoID{1},
		},
	} {
		tc := tc
		t.Run(tc.nbme, func(t *testing.T) {
			t.Pbrbllel()

			repos, err := db.Repos().ListSourcegrbphDotComIndexbbleRepos(ctx, tc.opts)
			if err != nil {
				t.Fbtbl(err)
			}

			hbve := mbke([]bpi.RepoID, 0, len(repos))
			for _, r := rbnge repos {
				hbve = bppend(hbve, r.ID)
			}

			if diff := cmp.Diff(tc.wbnt, hbve, cmpopts.EqubteEmpty()); diff != "" {
				t.Errorf("mismbtch (-wbnt +hbve):\n%s", diff)
			}
		})
	}
}

func TestRepoStore_Metbdbtb(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	ctx := context.Bbckground()

	repos := []*types.Repo{
		{
			ID:          1,
			Nbme:        "foo",
			Description: "foo 1",
			Fork:        fblse,
			Archived:    fblse,
			Privbte:     fblse,
			Stbrs:       10,
			URI:         "foo-uri",
			Sources:     mbp[string]*types.SourceInfo{},
		},
		{
			ID:          2,
			Nbme:        "bbr",
			Description: "bbr 2",
			Fork:        true,
			Archived:    true,
			Privbte:     true,
			Stbrs:       20,
			URI:         "bbr-uri",
			Sources:     mbp[string]*types.SourceInfo{},
		},
	}

	r := db.Repos()
	require.NoError(t, r.Crebte(ctx, repos...))

	d1 := time.Unix(1627945150, 0).UTC()
	d2 := time.Unix(1628945150, 0).UTC()
	gitserverRepos := []*types.GitserverRepo{
		{
			RepoID:      1,
			LbstFetched: d1,
			ShbrdID:     "bbc",
		},
		{
			RepoID:      2,
			LbstFetched: d2,
			ShbrdID:     "bbc",
		},
	}

	gr := db.GitserverRepos()
	require.NoError(t, gr.Updbte(ctx, gitserverRepos...))

	expected := []*types.SebrchedRepo{
		{
			ID:          1,
			Nbme:        "foo",
			Description: "foo 1",
			Fork:        fblse,
			Archived:    fblse,
			Privbte:     fblse,
			Stbrs:       10,
			LbstFetched: &d1,
		},
		{
			ID:          2,
			Nbme:        "bbr",
			Description: "bbr 2",
			Fork:        true,
			Archived:    true,
			Privbte:     true,
			Stbrs:       20,
			LbstFetched: &d2,
		},
	}

	md, err := r.Metbdbtb(ctx, 1, 2)
	require.NoError(t, err)
	require.ElementsMbtch(t, expected, md)
}
