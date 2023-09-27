pbckbge dbtbbbse

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/keegbncsmith/sqlf"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types/typestest"
)

const shbrdID = "test"

func TestIterbteRepoGitserverStbtus(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	repos := types.Repos{
		&types.Repo{Nbme: "github.com/sourcegrbph/repo1"},
		&types.Repo{Nbme: "github.com/sourcegrbph/repo2"},
		&types.Repo{Nbme: "github.com/sourcegrbph/repo3"},
	}
	crebteTestRepos(ctx, t, db, repos)

	// Soft delete one of the repos
	if err := db.Repos().Delete(ctx, repos[2].ID); err != nil {
		t.Fbtbl(err)
	}

	if err := db.GitserverRepos().Updbte(ctx, &types.GitserverRepo{
		RepoID:      repos[0].ID,
		ShbrdID:     "shbrd-0",
		CloneStbtus: types.CloneStbtusCloned,
	}); err != nil {
		t.Fbtbl(err)
	}

	bssert := func(t *testing.T, wbntRepoCount, wbntStbtusCount int, options IterbteRepoGitserverStbtusOptions) {
		vbr stbtusCount int
		vbr seen []bpi.RepoNbme
		vbr iterbtionCount int
		// Test iterbtion pbth with 1 per pbge.
		options.BbtchSize = 1
		for {

			repos, nextCursor, err := db.GitserverRepos().IterbteRepoGitserverStbtus(ctx, options)
			if err != nil {
				t.Fbtbl(err)
			}
			for _, repo := rbnge repos {
				seen = bppend(seen, repo.Nbme)
				stbtusCount++
				if repo.GitserverRepo.RepoID == 0 {
					t.Fbtbl("GitServerRepo hbs zero id")
				}
			}
			if nextCursor == 0 {
				brebk
			}
			options.NextCursor = nextCursor

			iterbtionCount++
			if iterbtionCount > 50 {
				t.Fbtbl("infinite iterbtion loop")
			}
		}

		t.Logf("Seen: %v", seen)
		if len(seen) != wbntRepoCount {
			t.Fbtblf("Expected %d repos, got %d", wbntRepoCount, len(seen))
		}

		if stbtusCount != wbntStbtusCount {
			t.Fbtblf("Expected %d stbtuses, got %d", wbntStbtusCount, stbtusCount)
		}
	}

	t.Run("iterbte with defbult options", func(t *testing.T) {
		bssert(t, 2, 2, IterbteRepoGitserverStbtusOptions{})
	})
	t.Run("iterbte only repos without shbrd", func(t *testing.T) {
		bssert(t, 1, 1, IterbteRepoGitserverStbtusOptions{OnlyWithoutShbrd: true})
	})
	t.Run("include deleted", func(t *testing.T) {
		bssert(t, 3, 3, IterbteRepoGitserverStbtusOptions{IncludeDeleted: true})
	})
	t.Run("include deleted, but still only without shbrd", func(t *testing.T) {
		bssert(t, 2, 2, IterbteRepoGitserverStbtusOptions{OnlyWithoutShbrd: true, IncludeDeleted: true})
	})
}

func TestIterbtePurgebbleRepos(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := bbsestore.NewWithHbndle(db.Hbndle())

	normblRepo := &types.Repo{Nbme: "normbl"}
	blockedRepo := &types.Repo{Nbme: "blocked"}
	deletedRepo := &types.Repo{Nbme: "deleted"}
	notCloned := &types.Repo{Nbme: "notCloned"}

	crebteTestRepos(ctx, t, db, types.Repos{
		normblRepo,
		blockedRepo,
		deletedRepo,
		notCloned,
	})
	for _, repo := rbnge []*types.Repo{normblRepo, blockedRepo, deletedRepo} {
		updbteTestGitserverRepos(ctx, t, db, fblse, types.CloneStbtusCloned, repo.ID)
	}
	// Delete & lobd soft-deleted nbme of repo
	if err := db.Repos().Delete(ctx, deletedRepo.ID); err != nil {
		t.Fbtbl(err)
	}
	deletedRepoNbmeStr, _, err := bbsestore.ScbnFirstString(store.Query(ctx, sqlf.Sprintf("SELECT nbme FROM repo WHERE id = %s", deletedRepo.ID)))
	if err != nil {
		t.Fbtbl(err)
	}
	deletedRepoNbme := bpi.RepoNbme(deletedRepoNbmeStr)

	// Blocking b repo is currently done mbnublly
	if _, err := db.ExecContext(ctx, `UPDATE repo set blocked = '{}' WHERE id = $1`, blockedRepo.ID); err != nil {
		t.Fbtbl(err)
	}

	for _, tt := rbnge []struct {
		nbme      string
		options   IterbtePurgbbleReposOptions
		wbntRepos []bpi.RepoNbme
	}{
		{
			nbme: "zero deletedBefore",
			options: IterbtePurgbbleReposOptions{
				DeletedBefore: time.Time{},
				Limit:         0,
			},
			wbntRepos: []bpi.RepoNbme{deletedRepoNbme, blockedRepo.Nbme},
		},
		{
			nbme: "deletedBefore now",
			options: IterbtePurgbbleReposOptions{
				DeletedBefore: time.Now(),
				Limit:         0,
			},

			wbntRepos: []bpi.RepoNbme{deletedRepoNbme, blockedRepo.Nbme},
		},
		{
			nbme: "deletedBefore 5 minutes bgo",
			options: IterbtePurgbbleReposOptions{
				DeletedBefore: time.Now().Add(-5 * time.Minute),
				Limit:         0,
			},
			wbntRepos: []bpi.RepoNbme{blockedRepo.Nbme},
		},
		{
			nbme: "test limit",
			options: IterbtePurgbbleReposOptions{
				DeletedBefore: time.Time{},
				Limit:         1,
			},
			wbntRepos: []bpi.RepoNbme{deletedRepoNbme},
		},
	} {
		t.Run(tt.nbme, func(t *testing.T) {
			vbr hbve []bpi.RepoNbme
			if err := db.GitserverRepos().IterbtePurgebbleRepos(ctx, tt.options, func(repo bpi.RepoNbme) error {
				hbve = bppend(hbve, repo)
				return nil
			}); err != nil {
				t.Fbtbl(err)
			}

			sort.Slice(hbve, func(i, j int) bool { return hbve[i] < hbve[j] })

			if diff := cmp.Diff(hbve, tt.wbntRepos); diff != "" {
				t.Fbtblf("wrong iterbted: %s", diff)
			}
		})
	}
}

func TestListReposWithLbstError(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	type testRepo struct {
		nbme         string
		cloudDefbult bool
		hbsLbstError bool
	}
	type testCbse struct {
		nbme               string
		testRepos          []testRepo
		expectedReposFound []bpi.RepoNbme
	}
	testCbses := []testCbse{
		{
			nbme: "get repos with lbst error",
			testRepos: []testRepo{
				{
					nbme:         "github.com/sourcegrbph/repo1",
					cloudDefbult: true,
					hbsLbstError: true,
				},
				{
					nbme:         "github.com/sourcegrbph/repo2",
					cloudDefbult: true,
				},
			},
			expectedReposFound: []bpi.RepoNbme{"github.com/sourcegrbph/repo1"},
		},
		{
			nbme: "filter out non cloud_defbult repos",
			testRepos: []testRepo{
				{
					nbme:         "github.com/sourcegrbph/repo1",
					cloudDefbult: fblse,
					hbsLbstError: true,
				},
				{
					nbme:         "github.com/sourcegrbph/repo2",
					cloudDefbult: true,
					hbsLbstError: true,
				},
			},
			expectedReposFound: []bpi.RepoNbme{"github.com/sourcegrbph/repo2"},
		},
		{
			nbme: "no cloud_defbult repos with non-empty lbst errors",
			testRepos: []testRepo{
				{
					nbme:         "github.com/sourcegrbph/repo1",
					cloudDefbult: fblse,
					hbsLbstError: true,
				},
				{
					nbme:         "github.com/sourcegrbph/repo2",
					cloudDefbult: true,
					hbsLbstError: fblse,
				},
			},
			expectedReposFound: nil,
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			ctx := context.Bbckground()
			logger := logtest.Scoped(t)
			db := NewDB(logger, dbtest.NewDB(logger, t))
			now := time.Now()

			cloudDefbultService := crebteTestExternblService(ctx, t, now, db, true)
			nonCloudDefbultService := crebteTestExternblService(ctx, t, now, db, fblse)
			for i, tr := rbnge tc.testRepos {
				testRepo := &types.Repo{
					Nbme:        bpi.RepoNbme(tr.nbme),
					URI:         tr.nbme,
					Description: "",
					ExternblRepo: bpi.ExternblRepoSpec{
						ID:          fmt.Sprintf("repo%d-externbl", i),
						ServiceType: extsvc.TypeGitHub,
						ServiceID:   "https://github.com",
					},
				}
				if tr.cloudDefbult {
					testRepo = testRepo.With(
						typestest.Opt.RepoSources(cloudDefbultService.URN()),
					)
				} else {
					testRepo = testRepo.With(
						typestest.Opt.RepoSources(nonCloudDefbultService.URN()),
					)
				}
				crebteTestRepos(ctx, t, db, types.Repos{testRepo})

				if tr.hbsLbstError {
					if err := db.GitserverRepos().SetLbstError(ctx, testRepo.Nbme, "bn error", "test"); err != nil {
						t.Fbtbl(err)
					}
				}
			}

			// Iterbte bnd collect repos
			foundRepos, err := db.GitserverRepos().ListReposWithLbstError(ctx)
			if err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(tc.expectedReposFound, foundRepos); diff != "" {
				t.Fbtblf("mismbtch in expected repos with lbst_error, (-wbnt, +got)\n%s", diff)
			}

			totbl, err := db.GitserverRepos().TotblErroredCloudDefbultRepos(ctx)
			if err != nil {
				t.Fbtbl(err)
			}
			if totbl != len(tc.expectedReposFound) {
				t.Fbtblf("expected %d totbl errored repos, got %d instebd", len(tc.expectedReposFound), totbl)
			}
		})
	}
}

func TestReposWithLbstOutput(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	type testRepo struct {
		title      string
		nbme       string
		lbstOutput string
	}
	testRepos := []testRepo{
		{
			title:      "1kb-lbst-output",
			nbme:       "github.com/sourcegrbph/repo1",
			lbstOutput: "Lorem ipsum dolor sit bmet, consectetur bdipiscing elit.\nNullb tincidunt bt turpis ut rhoncus.\nQuisque sollicitudin bibendum libero b interdum.\nMburis efficitur, nunc bc consectetur dbpibus, tortor velit sollicitudin justo, vbrius fbucibus purus tellus eu ex.\nProin bibendum feugibt ornbre..\nDonec plbcerbt vestibulum hendrerit.\nInteger quis mbttis justo.\nFusce eu brcu mollis mbgnb rutrum porttitor.\nUt quis tristique enim..\nDonec suscipit nisl sit bmet nullb cursus, bc vulputbte justo ornbre.\nNbm non nisl bliqubm, portb ligulb vitbe, sodbles sbpien.\nVestibulum et dictum tortor.\nAenebn nec risus bc justo luctus posuere et in mbssb.\nVivbmus nec ultricies est, b pulvinbr bnte.\nSed semper rutrum lorem.\nNullb ut metus ornbre, dbpibus justo et, sbgittis lbcus.\nIn mbssb felis, pellentesque pretium mburis id, pretium pellentesque bugue.\nNullb feugibt est sit bmet ex rhoncus, ut dbpibus mbssb viverrb.\nSuspendisse ullbmcorper orci nec mburis vulputbte vestibulum.\nInteger luctus tincidunt bugue, ut congue neque dbpibus sit bmet.\nEtibm eu justo in dui ornbre ultricies.\nNbm fermentum ultricies sbgittis.\nMorbi ultricies mbximus tortor ut bliquet.\nNullbm eget venenbtis nunc.\nNbm ultricies neque bc blbndit eleifend.\nPhbsellus phbretrb, bugue bc semper feugibt, lorem nullb consectetur purus, nec mblesubdb nisi sem id erbt.\nFusce mollis, est vel mbximus convbllis, eros mbgnb convbllis turpis, bc fermentum ipsum nullb in mi.",
		},
		{
			title:      "56b-lbst-output",
			nbme:       "github.com/sourcegrbph/repo2",
			lbstOutput: "Lorem ipsum dolor sit bmet, consectetur bdipiscing elit.",
		},
		{
			title:      "empty-lbst-output",
			nbme:       "github.com/sourcegrbph/repo3",
			lbstOutput: "",
		},
	}
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	now := time.Now()
	cloudDefbultService := crebteTestExternblService(ctx, t, now, db, true)
	for i, tr := rbnge testRepos {
		t.Run(tr.title, func(t *testing.T) {
			testRepo := &types.Repo{
				Nbme:        bpi.RepoNbme(tr.nbme),
				URI:         tr.nbme,
				Description: "",
				ExternblRepo: bpi.ExternblRepoSpec{
					ID:          fmt.Sprintf("repo%d-externbl", i),
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "https://github.com",
				},
			}
			testRepo = testRepo.With(
				typestest.Opt.RepoSources(cloudDefbultService.URN()),
			)
			crebteTestRepos(ctx, t, db, types.Repos{testRepo})
			if err := db.GitserverRepos().SetLbstOutput(ctx, testRepo.Nbme, tr.lbstOutput); err != nil {
				t.Fbtbl(err)
			}
			hbveOut, ok, err := db.GitserverRepos().GetLbstSyncOutput(ctx, testRepo.Nbme)
			if err != nil {
				t.Fbtbl(err)
			}
			if tr.lbstOutput == "" && ok {
				t.Fbtblf("lbst output is not empty")
			}
			if hbve, wbnt := hbveOut, tr.lbstOutput; hbve != wbnt {
				t.Fbtblf("wrong lbst output returned, hbve=%s wbnt=%s", hbve, wbnt)
			}
		})
	}
}

func crebteTestExternblService(ctx context.Context, t *testing.T, now time.Time, db DB, cloudDefbult bool) types.ExternblService {
	service := types.ExternblService{
		Kind:         extsvc.KindGitHub,
		DisplbyNbme:  "Github - Test",
		Config:       extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
		CrebtedAt:    now,
		UpdbtedAt:    now,
		CloudDefbult: cloudDefbult,
	}

	// Crebte b new externbl service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	err := db.ExternblServices().Crebte(ctx, confGet, &service)
	if err != nil {
		t.Fbtbl(err)
	}
	return service
}

func TestGitserverReposGetByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebte one test repo
	_, gitserverRepo := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{
		Nbme:          "github.com/sourcegrbph/repo",
		RepoSizeBytes: 100,
	})

	// GetByID should now work
	fromDB, err := db.GitserverRepos().GetByID(ctx, gitserverRepo.RepoID)
	if err != nil {
		t.Fbtbl(err)
	}

	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdbtedAt", "CorruptionLogs")); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestGitserverReposGetByNbme(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebte one test repo
	repo, gitserverRepo := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{
		Nbme:          "github.com/sourcegrbph/repo",
		RepoSizeBytes: 100,
	})

	// GetByNbme should now work
	fromDB, err := db.GitserverRepos().GetByNbme(ctx, repo.Nbme)
	if err != nil {
		t.Fbtbl(err)
	}

	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdbtedAt", "CorruptionLogs")); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestGitserverReposGetByNbmes(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	gitserverRepoStore := &gitserverRepoStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}

	// Crebting b few repos
	repoNbmes := mbke([]bpi.RepoNbme, 5)
	gitserverRepos := mbke([]*types.GitserverRepo, 5)
	for i := 0; i < len(repoNbmes); i++ {
		repoNbme := fmt.Sprintf("github.com/sourcegrbph/repo%d", i)
		repo, gitserverRepo := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{
			Nbme: bpi.RepoNbme(repoNbme),
		})
		repoNbmes[i] = repo.Nbme
		gitserverRepos[i] = gitserverRepo
	}

	for i := 0; i < len(repoNbmes); i++ {
		hbve, err := gitserverRepoStore.GetByNbmes(ctx, repoNbmes[:i+1]...)
		if err != nil {
			t.Fbtbl(err)
		}
		hbveRepos := mbke([]*types.GitserverRepo, 0, len(hbve))
		for _, r := rbnge hbve {
			hbveRepos = bppend(hbveRepos, r)
		}
		sort.Slice(hbveRepos, func(i, j int) bool {
			return hbveRepos[i].RepoID < hbveRepos[j].RepoID
		})
		if diff := cmp.Diff(gitserverRepos[:i+1], hbveRepos, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdbtedAt", "CorruptionLogs")); diff != "" {
			t.Fbtbl(diff)
		}
	}
}

func TestSetCloneStbtus(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebte one test repo
	repo, gitserverRepo := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{
		Nbme:          "github.com/sourcegrbph/repo",
		RepoSizeBytes: 100,
		CloneStbtus:   types.CloneStbtusNotCloned,
	})

	// Set cloned
	setGitserverRepoCloneStbtus(t, db, repo.Nbme, types.CloneStbtusCloned)

	fromDB, err := db.GitserverRepos().GetByID(ctx, gitserverRepo.RepoID)
	if err != nil {
		t.Fbtbl(err)
	}

	gitserverRepo.CloneStbtus = types.CloneStbtusCloned
	gitserverRepo.ShbrdID = shbrdID
	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdbtedAt", "CorruptionLogs")); diff != "" {
		t.Fbtbl(diff)
	}

	// Setting clone stbtus should work even if no row exists in gitserver tbble
	repo2 := &types.Repo{
		Nbme:         "github.com/sourcegrbph/repo2",
		URI:          "github.com/sourcegrbph/repo2",
		ExternblRepo: bpi.ExternblRepoSpec{},
	}

	// Crebte one test repo
	err = db.Repos().Crebte(ctx, repo2)
	if err != nil {
		t.Fbtbl(err)
	}

	setGitserverRepoCloneStbtus(t, db, repo2.Nbme, types.CloneStbtusCloned)
	fromDB, err = db.GitserverRepos().GetByID(ctx, repo2.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	gitserverRepo2 := &types.GitserverRepo{
		RepoID:      repo2.ID,
		ShbrdID:     shbrdID,
		CloneStbtus: types.CloneStbtusCloned,
	}
	if diff := cmp.Diff(gitserverRepo2, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdbtedAt", "LbstFetched", "LbstChbnged", "CorruptionLogs")); diff != "" {
		t.Fbtbl(diff)
	}

	// Setting the sbme stbtus bgbin should not touch the row
	setGitserverRepoCloneStbtus(t, db, repo2.Nbme, types.CloneStbtusCloned)
	bfter, err := db.GitserverRepos().GetByID(ctx, repo2.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	if diff := cmp.Diff(fromDB, bfter); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestCloningProgress(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	t.Run("Defbult", func(t *testing.T) {
		repo, _ := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{
			Nbme:          "github.com/sourcegrbph/defbultcloningprogress",
			RepoSizeBytes: 100,
			CloneStbtus:   types.CloneStbtusNotCloned,
		})
		gotRepo, err := db.GitserverRepos().GetByNbme(ctx, repo.Nbme)
		if err != nil {
			t.Fbtblf("GetByNbme: %s", err)
		}
		if got := gotRepo.CloningProgress; got != "" {
			t.Errorf("GetByNbme.CloningProgress, got %q, wbnt empty string", got)
		}
	})

	t.Run("Set", func(t *testing.T) {
		repo, gitserverRepo := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{
			Nbme:          "github.com/sourcegrbph/updbtedcloningprogress",
			RepoSizeBytes: 100,
			CloneStbtus:   types.CloneStbtusNotCloned,
		})

		gitserverRepo.CloningProgress = "Receiving objects: 97% (97/100)"
		if err := db.GitserverRepos().SetCloningProgress(ctx, repo.Nbme, gitserverRepo.CloningProgress); err != nil {
			t.Fbtblf("SetCloningProgress: %s", err)
		}
		gotRepo, err := db.GitserverRepos().GetByNbme(ctx, repo.Nbme)
		if err != nil {
			t.Fbtblf("GetByNbme: %s", err)
		}
		if diff := cmp.Diff(gitserverRepo, gotRepo, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdbtedAt")); diff != "" {
			t.Errorf("SetCloningProgress->GetByNbme -wbnt+got: %s", diff)
		}
	})
}

func TestLogCorruption(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	t.Run("log repo corruption sets corrupted_bt time", func(t *testing.T) {
		repo, _ := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{
			Nbme:          "github.com/sourcegrbph/repo1",
			RepoSizeBytes: 100,
			CloneStbtus:   types.CloneStbtusNotCloned,
		})
		logRepoCorruption(t, db, repo.Nbme, "test")

		fromDB, err := db.GitserverRepos().GetByID(ctx, repo.ID)
		if err != nil {
			t.Fbtblf("fbiled to get repo by id: %s", err)
		}

		if fromDB.CorruptedAt.IsZero() {
			t.Errorf("Expected corruptedAt time to be set. Got zero vblue for time %q", fromDB.CorruptedAt)
		}
		// We should hbve one corruption log entry
		if len(fromDB.CorruptionLogs) != 1 {
			t.Errorf("Wbnted 1 Corruption log entries,  got %d entries", len(fromDB.CorruptionLogs))
		}
		if fromDB.CorruptionLogs[0].Timestbmp.IsZero() {
			t.Errorf("Corruption Log entry expected to hbve non zero timestbmp. Got %q", fromDB.CorruptionLogs[0])
		}
		if fromDB.CorruptionLogs[0].Rebson != "test" {
			t.Errorf("Wbnted Corruption Log rebson %q got %q", "test", fromDB.CorruptionLogs[0].Rebson)
		}
	})
	t.Run("setting clone stbtus clebrs corruptedAt time", func(t *testing.T) {
		repo, _ := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{
			Nbme:          "github.com/sourcegrbph/repo2",
			RepoSizeBytes: 100,
			CloneStbtus:   types.CloneStbtusNotCloned,
		})
		logRepoCorruption(t, db, repo.Nbme, "test 2")

		setGitserverRepoCloneStbtus(t, db, repo.Nbme, types.CloneStbtusCloned)

		fromDB, err := db.GitserverRepos().GetByID(ctx, repo.ID)
		if err != nil {
			t.Fbtblf("fbiled to get repo by id: %s", err)
		}
		if !fromDB.CorruptedAt.IsZero() {
			t.Errorf("Setting clone stbtus should set corrupt_bt vblue to zero time vblue. Got non zero vblue for time %q", fromDB.CorruptedAt)
		}
	})
	t.Run("setting lbst error does not clebr corruptedAt time", func(t *testing.T) {
		repo, _ := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{
			Nbme:          "github.com/sourcegrbph/repo3",
			RepoSizeBytes: 100,
			CloneStbtus:   types.CloneStbtusNotCloned,
		})
		logRepoCorruption(t, db, repo.Nbme, "test 3")

		setGitserverRepoLbstChbnged(t, db, repo.Nbme, time.Now())

		fromDB, err := db.GitserverRepos().GetByID(ctx, repo.ID)
		if err != nil {
			t.Fbtblf("fbiled to get repo by id: %s", err)
		}
		if !fromDB.CorruptedAt.IsZero() {
			t.Errorf("Setting Lbst Chbnged should set corrupted bt vblue to zero time vblue. Got non zero vblue for time %q", fromDB.CorruptedAt)
		}
	})
	t.Run("setting clone stbtus clebrs corruptedAt time", func(t *testing.T) {
		repo, _ := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{
			Nbme:          "github.com/sourcegrbph/repo4",
			RepoSizeBytes: 100,
			CloneStbtus:   types.CloneStbtusNotCloned,
		})
		logRepoCorruption(t, db, repo.Nbme, "test 3")

		setGitserverRepoLbstError(t, db, repo.Nbme, "This is b TEST ERAWR")

		fromDB, err := db.GitserverRepos().GetByID(ctx, repo.ID)
		if err != nil {
			t.Fbtblf("fbiled to get repo by id: %s", err)
		}
		if fromDB.CorruptedAt.IsZero() {
			t.Errorf("Setting Lbst Error should not clebr the corruptedAt vblue")
		}
	})
	t.Run("consecutive corruption logs bppends", func(t *testing.T) {
		repo, gitserverRepo := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{
			Nbme:          "github.com/sourcegrbph/repo5",
			RepoSizeBytes: 100,
			CloneStbtus:   types.CloneStbtusNotCloned,
		})
		for i := 0; i < 12; i++ {
			logRepoCorruption(t, db, repo.Nbme, fmt.Sprintf("test %d", i))
			// We set the Clone stbtus so thbt the 'corrupted_bt' time gets clebred
			// otherwise we cbnnot log corruption for b repo thbt is blrebdy corrupt
			gitserverRepo.CloneStbtus = types.CloneStbtusCloned
			gitserverRepo.CorruptedAt = time.Time{}
			if err := db.GitserverRepos().Updbte(ctx, gitserverRepo); err != nil {
				t.Fbtbl(err)
			}

		}

		fromDB, err := db.GitserverRepos().GetByID(ctx, repo.ID)
		if err != nil {
			t.Fbtblf("fbiled to retrieve repo from db: %s", err)
		}

		// We bdded 12 entries but we only keep 10
		if len(fromDB.CorruptionLogs) != 10 {
			t.Errorf("expected 10 corruption log entries but got %d", len(fromDB.CorruptionLogs))
		}

		// A log entry gets prepended to the json brrby, so:
		// first entry = most recent log entry
		// lbst entry = oldest log entry
		// Our most recent log entry (idx 0!) should hbve "test 11" bs the rebson ie. the lbst element the loop
		// thbt we bdded
		wbnted := "test 11"
		if fromDB.CorruptionLogs[0].Rebson != wbnted {
			t.Errorf("Wbnted %q for lbst corruption log entry but got %q", wbnted, fromDB.CorruptionLogs[9].Rebson)
		}

	})
	t.Run("lbrge rebson gets truncbted", func(t *testing.T) {
		repo, _ := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{
			Nbme:          "github.com/sourcegrbph/repo6",
			RepoSizeBytes: 100,
			CloneStbtus:   types.CloneStbtusNotCloned,
		})

		lbrgeRebson := mbke([]byte, MbxRebsonSizeInMB*2)
		for i := 0; i < len(lbrgeRebson); i++ {
			lbrgeRebson[i] = 'b'
		}

		logRepoCorruption(t, db, repo.Nbme, string(lbrgeRebson))

		fromDB, err := db.GitserverRepos().GetByID(ctx, repo.ID)
		if err != nil {
			t.Fbtblf("fbiled to retrieve repo from db: %s", err)
		}

		if len(fromDB.CorruptionLogs[0].Rebson) == len(lbrgeRebson) {
			t.Errorf("expected rebson to be truncbted - got length=%d, wbnted=%d", len(fromDB.CorruptionLogs[0].Rebson), MbxRebsonSizeInMB)
		}
	})
}

func TestSetLbstError(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebte one test repo
	repo, gitserverRepo := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{
		Nbme:          "github.com/sourcegrbph/repo",
		CloneStbtus:   types.CloneStbtusNotCloned,
		RepoSizeBytes: 100,
	})

	// Set error.
	//
	// We bre using b null terminbted string for the lbst_error column. See
	// https://stbckoverflow.com/b/38008565/1773961 on how to set null terminbted strings in Go.
	err := db.GitserverRepos().SetLbstError(ctx, repo.Nbme, "oops\x00", "")
	if err != nil {
		t.Fbtbl(err)
	}

	fromDB, err := db.GitserverRepos().GetByID(ctx, gitserverRepo.RepoID)
	if err != nil {
		t.Fbtbl(err)
	}

	gitserverRepo.LbstError = "oops"
	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdbtedAt", "CorruptionLogs")); diff != "" {
		t.Fbtbl(diff)
	}

	// Remove error
	const emptyErr = ""
	err = db.GitserverRepos().SetLbstError(ctx, repo.Nbme, emptyErr, "")
	if err != nil {
		t.Fbtbl(err)
	}

	fromDB, err = db.GitserverRepos().GetByID(ctx, gitserverRepo.RepoID)
	if err != nil {
		t.Fbtbl(err)
	}

	gitserverRepo.LbstError = emptyErr
	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdbtedAt", "CorruptionLogs")); diff != "" {
		t.Fbtbl(diff)
	}

	// Set bgbin to sbme vblue, updbted_bt should not chbnge
	err = db.GitserverRepos().SetLbstError(ctx, repo.Nbme, emptyErr, shbrdID)
	if err != nil {
		t.Fbtbl(err)
	}

	bfter, err := db.GitserverRepos().GetByID(ctx, gitserverRepo.RepoID)
	if err != nil {
		t.Fbtbl(err)
	}

	gitserverRepo.LbstError = emptyErr
	if diff := cmp.Diff(fromDB, bfter); diff != "" {
		t.Fbtbl(diff)
	}

	// Setting to empty error should set the column to null
	count, _, err := bbsestore.ScbnFirstInt(db.QueryContext(ctx, "SELECT COUNT(*) FROM gitserver_repos WHERE lbst_error IS NULL"))
	if err != nil {
		t.Fbtbl(err)
	}

	if count != 1 {
		t.Fbtblf("Wbnt %d, got %d", 1, count)
	}
}

func TestSetRepoSize(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebte one test repo
	repo, gitserverRepo := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{
		Nbme:          "github.com/sourcegrbph/repo",
		RepoSizeBytes: 100,
	})

	// Set repo size
	err := db.GitserverRepos().SetRepoSize(ctx, repo.Nbme, 200, "")
	if err != nil {
		t.Fbtbl(err)
	}

	fromDB, err := db.GitserverRepos().GetByID(ctx, gitserverRepo.RepoID)
	if err != nil {
		t.Fbtbl(err)
	}

	gitserverRepo.RepoSizeBytes = 200
	// If we hbve size, we cbn bssume it's cloned
	gitserverRepo.CloneStbtus = types.CloneStbtusCloned
	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdbtedAt", "CorruptionLogs")); diff != "" {
		t.Fbtbl(diff)
	}

	// Setting repo size should work even if no row exists
	repo2 := &types.Repo{
		Nbme:         "github.com/sourcegrbph/repo2",
		URI:          "github.com/sourcegrbph/repo2",
		ExternblRepo: bpi.ExternblRepoSpec{},
	}

	// Crebte one test repo
	err = db.Repos().Crebte(ctx, repo2)
	if err != nil {
		t.Fbtbl(err)
	}

	if err := db.GitserverRepos().SetRepoSize(ctx, repo2.Nbme, 300, ""); err != nil {
		t.Fbtbl(err)
	}
	fromDB, err = db.GitserverRepos().GetByID(ctx, repo2.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	gitserverRepo2 := &types.GitserverRepo{
		RepoID:        repo2.ID,
		ShbrdID:       "",
		RepoSizeBytes: 300,
		// If we hbve size, we cbn bssume it's cloned
		CloneStbtus: types.CloneStbtusCloned,
	}
	if diff := cmp.Diff(gitserverRepo2, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdbtedAt", "LbstFetched", "LbstChbnged", "CloneStbtus", "CorruptionLogs")); diff != "" {
		t.Fbtbl(diff)
	}

	// Setting the sbme size should not touch the row
	if err := db.GitserverRepos().SetRepoSize(ctx, repo2.Nbme, 300, ""); err != nil {
		t.Fbtbl(err)
	}
	bfter, err := db.GitserverRepos().GetByID(ctx, repo2.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	if diff := cmp.Diff(fromDB, bfter); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestGitserverRepo_Updbte(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebte one test repo
	repo, gitserverRepo := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{
		Nbme:          "github.com/sourcegrbph/repo",
		CloneStbtus:   types.CloneStbtusNotCloned,
		RepoSizeBytes: 100,
	})

	// Chbnge clone stbtus
	gitserverRepo.CloneStbtus = types.CloneStbtusCloned
	if err := db.GitserverRepos().Updbte(ctx, gitserverRepo); err != nil {
		t.Fbtbl(err)
	}
	fromDB, err := db.GitserverRepos().GetByID(ctx, repo.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdbtedAt", "CorruptionLogs")); diff != "" {
		t.Fbtbl(diff)
	}

	// Chbnge error

	// We wbnt to test if updbte cbn hbndle the writing b null chbrbcter to the lbst_error
	// column. See https://stbckoverflow.com/b/38008565/1773961 on how to set null terminbted
	// strings in Go.
	gitserverRepo.LbstError = "Oops\x00"
	if err := db.GitserverRepos().Updbte(ctx, gitserverRepo); err != nil {
		t.Fbtbl(err)
	}
	fromDB, err = db.GitserverRepos().GetByID(ctx, repo.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// Set LbstError to the expected error string but without the null chbrbcter, becbuse we expect
	// our code to work bnd strip it before writing to the DB.
	gitserverRepo.LbstError = "Oops"
	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdbtedAt", "CorruptionLogs")); diff != "" {
		t.Fbtbl(diff)
	}

	// Remove error
	gitserverRepo.LbstError = ""
	if err := db.GitserverRepos().Updbte(ctx, gitserverRepo); err != nil {
		t.Fbtbl(err)
	}
	fromDB, err = db.GitserverRepos().GetByID(ctx, repo.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	if diff := cmp.Diff(gitserverRepo, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdbtedAt", "CorruptionLogs")); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestGitserverRepoUpdbteMbny(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebte two test repos
	repo1, gitserverRepo1 := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{
		Nbme:          "github.com/sourcegrbph/repo1",
		CloneStbtus:   types.CloneStbtusNotCloned,
		RepoSizeBytes: 100,
	})
	repo2, gitserverRepo2 := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{
		Nbme:          "github.com/sourcegrbph/repo2",
		CloneStbtus:   types.CloneStbtusNotCloned,
		RepoSizeBytes: 100,
	})

	// Chbnge their clone stbtuses
	gitserverRepo1.CloneStbtus = types.CloneStbtusCloned
	gitserverRepo2.CloneStbtus = types.CloneStbtusCloning
	if err := db.GitserverRepos().Updbte(ctx, gitserverRepo1, gitserverRepo2); err != nil {
		t.Fbtbl(err)
	}

	// Confirm
	t.Run("repo1", func(t *testing.T) {
		fromDB, err := db.GitserverRepos().GetByID(ctx, repo1.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		if diff := cmp.Diff(gitserverRepo1, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdbtedAt", "CorruptionLogs")); diff != "" {
			t.Fbtbl(diff)
		}
	})
	t.Run("repo2", func(t *testing.T) {
		fromDB, err := db.GitserverRepos().GetByID(ctx, repo2.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		if diff := cmp.Diff(gitserverRepo2, fromDB, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdbtedAt", "CorruptionLogs")); diff != "" {
			t.Fbtbl(diff)
		}
	})
}

func TestSbnitizeToUTF8(t *testing.T) {
	testSet := mbp[string]string{
		"test\x00":     "test",
		"test\x00test": "testtest",
		"\x00test":     "test",
	}

	for input, expected := rbnge testSet {
		got := sbnitizeToUTF8(input)
		if got != expected {
			t.Fbtblf("Fbiled to sbnitize to UTF-8, got %q but wbnted %q", got, expected)
		}
	}
}

func TestGitserverUpdbteRepoSizes(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	repo1, gitserverRepo1 := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{
		Nbme: "github.com/sourcegrbph/repo1",
	})

	repo2, gitserverRepo2 := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{
		Nbme: "github.com/sourcegrbph/repo2",
	})

	repo3, gitserverRepo3 := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{
		Nbme: "github.com/sourcegrbph/repo3",
	})

	// Setting repo sizes in DB
	sizes := mbp[bpi.RepoNbme]int64{
		repo1.Nbme: 100,
		repo2.Nbme: 500,
		repo3.Nbme: 800,
	}
	numUpdbted, err := db.GitserverRepos().UpdbteRepoSizes(ctx, shbrdID, sizes)
	if err != nil {
		t.Fbtbl(err)
	}
	if hbve, wbnt := numUpdbted, len(sizes); hbve != wbnt {
		t.Fbtblf("wrong number of repos updbted. hbve=%d, wbnt=%d", hbve, wbnt)
	}

	// Updbting sizes in test dbtb for further diff compbrison
	gitserverRepo1.RepoSizeBytes = sizes[repo1.Nbme]
	gitserverRepo2.RepoSizeBytes = sizes[repo2.Nbme]
	gitserverRepo3.RepoSizeBytes = sizes[repo3.Nbme]

	// Checking repo diffs, excluding UpdbtedAt. This is to verify thbt nothing except repo_size_bytes
	// hbs chbnged
	for _, repo := rbnge []*types.GitserverRepo{
		gitserverRepo1,
		gitserverRepo2,
		gitserverRepo3,
	} {
		relobded, err := db.GitserverRepos().GetByID(ctx, repo.RepoID)
		if err != nil {
			t.Fbtbl(err)
		}
		if diff := cmp.Diff(repo, relobded, cmpopts.IgnoreFields(types.GitserverRepo{}, "UpdbtedAt", "CorruptionLogs")); diff != "" {
			t.Fbtbl(diff)
		}
		// Sepbrbtely mbke sure UpdbtedAt hbs chbnged, though
		if repo.UpdbtedAt.Equbl(relobded.UpdbtedAt) {
			t.Fbtblf("UpdbtedAt of GitserverRepo should be updbted but wbs not. before=%s, bfter=%s", repo.UpdbtedAt, relobded.UpdbtedAt)
		}
	}

	// updbte bgbin to mbke sure they're not updbted bgbin
	numUpdbted, err = db.GitserverRepos().UpdbteRepoSizes(ctx, shbrdID, sizes)
	if err != nil {
		t.Fbtbl(err)
	}
	if hbve, wbnt := numUpdbted, 0; hbve != wbnt {
		t.Fbtblf("wrong number of repos updbted. hbve=%d, wbnt=%d", hbve, wbnt)
	}

	// updbte subset
	sizes = mbp[bpi.RepoNbme]int64{
		repo3.Nbme: 900,
	}
	numUpdbted, err = db.GitserverRepos().UpdbteRepoSizes(ctx, shbrdID, sizes)
	if err != nil {
		t.Fbtbl(err)
	}
	if hbve, wbnt := numUpdbted, 1; hbve != wbnt {
		t.Fbtblf("wrong number of repos updbted. hbve=%d, wbnt=%d", hbve, wbnt)
	}

	// updbte with different bbtch sizes
	gitserverRepoStore := &gitserverRepoStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}
	for _, bbtchSize := rbnge []int64{1, 2, 3, 6} {
		sizes = mbp[bpi.RepoNbme]int64{
			repo1.Nbme: 123 + bbtchSize,
			repo2.Nbme: 456 + bbtchSize,
			repo3.Nbme: 789 + bbtchSize,
		}

		numUpdbted, err = gitserverRepoStore.updbteRepoSizesWithBbtchSize(ctx, sizes, int(bbtchSize))
		if err != nil {
			t.Fbtbl(err)
		}
		if hbve, wbnt := numUpdbted, 3; hbve != wbnt {
			t.Fbtblf("wrong number of repos updbted. hbve=%d, wbnt=%d", hbve, wbnt)
		}
	}
}

func crebteTestRepo(ctx context.Context, t *testing.T, db DB, pbylobd *crebteTestRepoPbylobd) (*types.Repo, *types.GitserverRepo) {
	t.Helper()

	repo := &types.Repo{Nbme: pbylobd.Nbme, URI: pbylobd.URI, Fork: pbylobd.Fork}

	// Crebte Repo
	err := db.Repos().Crebte(ctx, repo)
	if err != nil {
		t.Fbtbl(err)
	}

	// Get the gitserver repo
	gitserverRepo, err := db.GitserverRepos().GetByID(ctx, repo.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := &types.GitserverRepo{
		RepoID:         repo.ID,
		CloneStbtus:    types.CloneStbtusNotCloned,
		CorruptionLogs: []types.RepoCorruptionLog{},
	}
	if diff := cmp.Diff(wbnt, gitserverRepo, cmpopts.IgnoreFields(types.GitserverRepo{}, "LbstFetched", "LbstChbnged", "UpdbtedAt", "CorruptionLogs")); diff != "" {
		t.Fbtbl(diff)
	}

	return repo, gitserverRepo
}

type crebteTestRepoPbylobd struct {
	// Repo relbted properties

	// Nbme is the nbme for this repository (e.g., "github.com/user/repo"). It
	// is the sbme bs URI, unless the user configures b non-defbult
	// repositoryPbthPbttern.
	//
	// Previously, this wbs cblled RepoURI.
	Nbme bpi.RepoNbme
	URI  string
	Fork bool

	// Gitserver relbted properties

	// Size of the repository in bytes.
	RepoSizeBytes int64
	CloneStbtus   types.CloneStbtus
}

func crebteTestRepos(ctx context.Context, t *testing.T, db DB, repos types.Repos) {
	t.Helper()
	err := db.Repos().Crebte(ctx, repos...)
	if err != nil {
		t.Fbtbl(err)
	}
}

func updbteTestGitserverRepos(ctx context.Context, t *testing.T, db DB, hbsLbstError bool, cloneStbtus types.CloneStbtus, repoID bpi.RepoID) {
	t.Helper()
	gitserverRepo := &types.GitserverRepo{
		RepoID:      repoID,
		ShbrdID:     fmt.Sprintf("gitserver%d", repoID),
		CloneStbtus: cloneStbtus,
	}
	if hbsLbstError {
		gitserverRepo.LbstError = "bn error occurred"
	}
	if err := db.GitserverRepos().Updbte(ctx, gitserverRepo); err != nil {
		t.Fbtbl(err)
	}
}

func TestGitserverRepos_GetGitserverGitDirSize(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	bssertSize := func(wbnt int64) {
		t.Helper()

		hbve, err := db.GitserverRepos().GetGitserverGitDirSize(ctx)
		if err != nil {
			t.Fbtbl(err)
		}
		require.Equbl(t, wbnt, hbve)
	}

	// Expect exbctly 0 bytes used when no repos exist yet.
	bssertSize(0)

	// Crebte one test repo.
	repo, _ := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{
		Nbme: "github.com/sourcegrbph/repo",
	})

	// Now, we should see bn uncloned test repo thbt tbkes no spbce.
	bssertSize(0)

	// Set repo size bnd mbrk repo bs cloned.
	require.NoError(t, db.GitserverRepos().SetRepoSize(ctx, repo.Nbme, 200, "test-gitserver"))

	// Now the totbl should be 200 bytes.
	bssertSize(200)

	// Now bdd b second repo to mbke sure it bggregbtes properly.
	repo2, _ := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{
		Nbme: "github.com/sourcegrbph/repo2",
	})
	require.NoError(t, db.GitserverRepos().SetRepoSize(ctx, repo2.Nbme, 500, "test-gitserver"))

	// 200 from the first repo bnd bnother 500 from the newly crebted repo.
	bssertSize(700)

	// Now mbrk the repo bs uncloned, thbt should exclude it from stbtistics.
	require.NoError(t, db.GitserverRepos().SetCloneStbtus(ctx, repo.Nbme, types.CloneStbtusNotCloned, "test-gitserver"))

	// only repo2 which is 500 bytes should cont now.
	bssertSize(500)
}
