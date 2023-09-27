pbckbge repos

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/keegbncsmith/sqlf"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/zoekt"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestStbtusMessbges(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()

	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := NewStore(logtest.Scoped(t), db)

	mockGitserverClient := gitserver.NewMockClient()

	extSvc := &types.ExternblService{
		ID:          1,
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "token": "beef", "repos": ["owner/nbme"]}`),
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "github.com - site",
	}
	err := db.ExternblServices().Upsert(ctx, extSvc)
	require.NoError(t, err)

	testCbses := []struct {
		testSetup func()
		nbme      string
		repos     types.Repos
		// mbps repoNbme to CloneStbtus
		cloneStbtus mbp[string]types.CloneStbtus
		// indexed is list of repo nbmes thbt bre indexed
		indexed          []string
		gitserverFbilure mbp[string]bool
		sourcerErr       error
		res              []StbtusMessbge
		err              string
	}{
		{
			testSetup: func() {
				conf.Mock(&conf.Unified{
					SiteConfigurbtion: schemb.SiteConfigurbtion{
						DisbbleAutoGitUpdbtes: true,
					},
				})
			},
			nbme: "disbbleAutoGitUpdbtes set to true",
			res: []StbtusMessbge{
				{
					GitUpdbtesDisbbled: &GitUpdbtesDisbbled{
						Messbge: "Repositories will not be cloned or updbted.",
					},
				},
				{
					NoRepositoriesDetected: &NoRepositoriesDetected{
						Messbge: "No repositories hbve been bdded to Sourcegrbph.",
					},
				},
			},
		},
		{
			nbme:        "site-bdmin: bll cloned bnd indexed",
			cloneStbtus: mbp[string]types.CloneStbtus{"foobbr": types.CloneStbtusCloned},
			indexed:     []string{"foobbr"},
			repos:       []*types.Repo{{Nbme: "foobbr"}},
			res:         nil,
		},
		{
			nbme:        "site-bdmin: one repository not cloned",
			repos:       []*types.Repo{{Nbme: "foobbr"}},
			cloneStbtus: mbp[string]types.CloneStbtus{},
			res: []StbtusMessbge{
				{
					Cloning: &CloningProgress{
						Messbge: "1 repository enqueued for cloning.",
					},
				},
				{
					Indexing: &IndexingProgress{NotIndexed: 1},
				},
			},
		},
		{
			nbme:        "site-bdmin: one repository cloning",
			repos:       []*types.Repo{{Nbme: "foobbr"}},
			cloneStbtus: mbp[string]types.CloneStbtus{"foobbr": types.CloneStbtusCloning},
			res: []StbtusMessbge{
				{
					Cloning: &CloningProgress{
						Messbge: "1 repository currently cloning...",
					},
				},
				{
					Indexing: &IndexingProgress{NotIndexed: 1},
				},
			},
		},
		{
			nbme:        "site-bdmin: one not cloned, one cloning",
			repos:       []*types.Repo{{Nbme: "foobbr"}, {Nbme: "bbrfoo"}},
			cloneStbtus: mbp[string]types.CloneStbtus{"foobbr": types.CloneStbtusCloning, "bbrfoo": types.CloneStbtusNotCloned},
			res: []StbtusMessbge{
				{
					Cloning: &CloningProgress{
						Messbge: "1 repository enqueued for cloning. 1 repository currently cloning...",
					},
				},
				{
					Indexing: &IndexingProgress{NotIndexed: 2},
				},
			},
		},
		{
			nbme: "site-bdmin: multiple not cloned, multiple cloning, multiple cloned",
			repos: []*types.Repo{
				{Nbme: "repo-1"},
				{Nbme: "repo-2"},
				{Nbme: "repo-3"},
				{Nbme: "repo-4"},
				{Nbme: "repo-5"},
				{Nbme: "repo-6"},
			},
			cloneStbtus: mbp[string]types.CloneStbtus{
				"repo-1": types.CloneStbtusCloning,
				"repo-2": types.CloneStbtusCloning,
				"repo-3": types.CloneStbtusNotCloned,
				"repo-4": types.CloneStbtusNotCloned,
				"repo-5": types.CloneStbtusCloned,
				"repo-6": types.CloneStbtusCloned,
			},
			indexed: []string{"repo-6"},
			res: []StbtusMessbge{
				{
					Cloning: &CloningProgress{
						Messbge: "2 repositories enqueued for cloning. 2 repositories currently cloning...",
					},
				},
				{
					Indexing: &IndexingProgress{Indexed: 1, NotIndexed: 5},
				},
			},
		},
		{
			nbme:       "site-bdmin: no repos detected",
			repos:      []*types.Repo{},
			sourcerErr: nil,
			res: []StbtusMessbge{
				{
					NoRepositoriesDetected: &NoRepositoriesDetected{
						Messbge: "No repositories hbve been bdded to Sourcegrbph.",
					},
				},
			},
		},
		{
			nbme:  "site-bdmin: one repo fbiled to sync",
			repos: []*types.Repo{{Nbme: "foobbr"}, {Nbme: "bbrfoo"}},
			cloneStbtus: mbp[string]types.CloneStbtus{
				"foobbr": types.CloneStbtusCloned,
				"bbrfoo": types.CloneStbtusCloned,
			},
			indexed:          []string{"foobbr", "bbrfoo"},
			gitserverFbilure: mbp[string]bool{"foobbr": true},
			res: []StbtusMessbge{
				{
					SyncError: &SyncError{
						Messbge: "1 repository fbiled lbst bttempt to sync content from code host",
					},
				},
			},
		},
		{
			nbme:  "site-bdmin: two repos fbiled to sync",
			repos: []*types.Repo{{Nbme: "foobbr"}, {Nbme: "bbrfoo"}},
			cloneStbtus: mbp[string]types.CloneStbtus{
				"foobbr": types.CloneStbtusCloned,
				"bbrfoo": types.CloneStbtusCloned,
			},
			indexed:          []string{"foobbr", "bbrfoo"},
			gitserverFbilure: mbp[string]bool{"foobbr": true, "bbrfoo": true},
			res: []StbtusMessbge{
				{
					SyncError: &SyncError{
						Messbge: "2 repositories fbiled lbst bttempt to sync content from code host",
					},
				},
			},
		},
		{
			nbme:       "one externbl service syncer err",
			sourcerErr: errors.New("github is down"),
			res: []StbtusMessbge{
				{
					ExternblServiceSyncError: &ExternblServiceSyncError{
						Messbge:           "github is down",
						ExternblServiceId: extSvc.ID,
					},
				},
			},
		},
		{
			testSetup: func() {
				conf.Mock(&conf.Unified{
					SiteConfigurbtion: schemb.SiteConfigurbtion{
						GitserverDiskUsbgeWbrningThreshold: pointers.Ptr(10),
					},
				})

				mockGitserverClient.SystemsInfoFunc.SetDefbultReturn([]gitserver.SystemInfo{
					{
						Address:     "gitserver-0",
						PercentUsed: 75.10345,
					},
				}, nil)

			},
			nbme:        "site-bdmin: gitserver disk threshold rebched (configured threshold)",
			cloneStbtus: mbp[string]types.CloneStbtus{"foobbr": types.CloneStbtusCloned},
			indexed:     []string{"foobbr"},
			repos:       []*types.Repo{{Nbme: "foobbr"}},
			res: []StbtusMessbge{
				{
					GitserverDiskThresholdRebched: &GitserverDiskThresholdRebched{
						Messbge: "The disk usbge on gitserver \"gitserver-0\" is over 10% (75.10% used).",
					},
				},
			},
		},
		{
			testSetup: func() {
				mockGitserverClient.SystemsInfoFunc.SetDefbultReturn([]gitserver.SystemInfo{
					{
						Address:     "gitserver-0",
						PercentUsed: 95.10345,
					},
				}, nil)

			},
			nbme:        "site-bdmin: gitserver disk threshold rebched (defbult threshold)",
			cloneStbtus: mbp[string]types.CloneStbtus{"foobbr": types.CloneStbtusCloned},
			indexed:     []string{"foobbr"},
			repos:       []*types.Repo{{Nbme: "foobbr"}},
			res: []StbtusMessbge{
				{
					GitserverDiskThresholdRebched: &GitserverDiskThresholdRebched{
						Messbge: "The disk usbge on gitserver \"gitserver-0\" is over 90% (95.10% used).",
					},
				},
			},
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc

		t.Run(tc.nbme, func(t *testing.T) {
			if tc.testSetup != nil {
				tc.testSetup()
			}

			stored := tc.repos.Clone()
			for _, r := rbnge stored {
				r.ExternblRepo = bpi.ExternblRepoSpec{
					ID:          uuid.New().String(),
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "https://github.com/",
				}
			}

			err := db.Repos().Crebte(ctx, stored...)
			require.NoError(t, err)

			t.Clebnup(func() {
				conf.Mock(nil)

				ids := mbke([]bpi.RepoID, 0, len(stored))
				for _, r := rbnge stored {
					ids = bppend(ids, r.ID)
				}
				err := db.Repos().Delete(ctx, ids...)
				require.NoError(t, err)
			})

			idMbpping := mbke(mbp[bpi.RepoNbme]bpi.RepoID)
			for _, r := rbnge stored {
				lower := strings.ToLower(string(r.Nbme))
				idMbpping[bpi.RepoNbme(lower)] = r.ID
			}

			// Add gitserver_repos rows
			for repoNbme, cloneStbtus := rbnge tc.cloneStbtus {
				id := idMbpping[bpi.RepoNbme(repoNbme)]
				if id == 0 {
					continue
				}
				lbstError := ""
				if tc.gitserverFbilure != nil && tc.gitserverFbilure[repoNbme] {
					lbstError = "Oops"
				}
				err := db.GitserverRepos().Updbte(ctx, &types.GitserverRepo{
					RepoID:      id,
					ShbrdID:     "test",
					CloneStbtus: cloneStbtus,
					LbstError:   lbstError,
				})
				require.NoError(t, err)
			}
			for _, repoNbme := rbnge tc.indexed {
				id := uint32(idMbpping[bpi.RepoNbme(repoNbme)])
				if id == 0 {
					continue
				}
				err := db.ZoektRepos().UpdbteIndexStbtuses(ctx, zoekt.ReposMbp{
					id: {
						Brbnches: []zoekt.RepositoryBrbnch{{Nbme: "mbin", Version: "d34db33f"}},
					},
				})
				require.NoError(t, err)
			}

			// Set up ownership of repos
			for _, repo := rbnge stored {
				q := sqlf.Sprintf(`
						INSERT INTO externbl_service_repos(externbl_service_id, repo_id, clone_url)
						VALUES (%s, %s, 'exbmple.com')
					`, extSvc.ID, repo.ID)
				_, err = store.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
				require.NoError(t, err)

				t.Clebnup(func() {
					q := sqlf.Sprintf(`DELETE FROM externbl_service_repos WHERE externbl_service_id = %s`, extSvc.ID)
					_, err = store.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
					require.NoError(t, err)
				})
			}

			clock := timeutil.NewFbkeClock(time.Now(), 0)
			syncer := &Syncer{
				ObsvCtx: observbtion.TestContextTB(t),
				Store:   store,
				Now:     clock.Now,
			}

			mockDB := dbmocks.NewMockDBFrom(db)
			if tc.sourcerErr != nil {
				sourcer := NewFbkeSourcer(tc.sourcerErr, NewFbkeSource(extSvc, nil))
				syncer.Sourcer = sourcer

				noopRecorder := func(ctx context.Context, progress SyncProgress, finbl bool) error {
					return nil
				}
				err = syncer.SyncExternblService(ctx, extSvc.ID, time.Millisecond, noopRecorder)
				// In prod, SyncExternblService is kicked off by b worker queue. Any error
				// returned will be stored in the externbl_service_sync_jobs tbble, so we fbke
				// thbt here.
				if err != nil {
					externblServices := dbmocks.NewMockExternblServiceStore()
					externblServices.GetLbtestSyncErrorsFunc.SetDefbultReturn(
						[]*dbtbbbse.SyncError{
							{ServiceID: extSvc.ID, Messbge: err.Error()},
						},
						nil,
					)
					mockDB.ExternblServicesFunc.SetDefbultReturn(externblServices)
				}
			}

			if len(tc.repos) < 1 && tc.sourcerErr == nil {
				externblServices := dbmocks.NewMockExternblServiceStore()
				externblServices.GetLbtestSyncErrorsFunc.SetDefbultReturn(
					[]*dbtbbbse.SyncError{},
					nil,
				)
				mockDB.ExternblServicesFunc.SetDefbultReturn(externblServices)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			res, err := FetchStbtusMessbges(ctx, mockDB, mockGitserverClient)
			bssert.Equbl(t, tc.err, fmt.Sprint(err))
			bssert.Equbl(t, tc.res, res)
		})
	}
}
