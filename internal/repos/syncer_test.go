pbckbge repos_test

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/keegbncsmith/sqlf"
	"github.com/stretchr/testify/require"
	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bwscodecommit"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitolite"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types/typestest"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestSyncerSync(t *testing.T) {
	t.Pbrbllel()
	store := getTestRepoStore(t)

	servicesPerKind := crebteExternblServices(t, store)

	githubService := servicesPerKind[extsvc.KindGitHub]

	githubRepo := (&types.Repo{
		Nbme:     "github.com/org/foo",
		Metbdbtb: &github.Repository{},
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "foo-externbl-12345",
			ServiceID:   "https://github.com/",
			ServiceType: extsvc.TypeGitHub,
		},
	}).With(
		typestest.Opt.RepoSources(githubService.URN()),
	)

	gitlbbService := servicesPerKind[extsvc.KindGitLbb]

	gitlbbRepo := (&types.Repo{
		Nbme:     "gitlbb.com/org/foo",
		Metbdbtb: &gitlbb.Project{},
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "12345",
			ServiceID:   "https://gitlbb.com/",
			ServiceType: extsvc.TypeGitLbb,
		},
	}).With(
		typestest.Opt.RepoSources(gitlbbService.URN()),
	)

	bitbucketServerService := servicesPerKind[extsvc.KindBitbucketServer]

	bitbucketServerRepo := (&types.Repo{
		Nbme:     "bitbucketserver.mycorp.com/org/foo",
		Metbdbtb: &bitbucketserver.Repo{},
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "23456",
			ServiceID:   "https://bitbucketserver.mycorp.com/",
			ServiceType: "bitbucketServer",
		},
	}).With(
		typestest.Opt.RepoSources(bitbucketServerService.URN()),
	)

	bwsCodeCommitService := servicesPerKind[extsvc.KindAWSCodeCommit]

	bwsCodeCommitRepo := (&types.Repo{
		Nbme:     "git-codecommit.us-west-1.bmbzonbws.com/stripe-go",
		Metbdbtb: &bwscodecommit.Repository{},
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "f001337b-3450-46fd-b7d2-650c0EXAMPLE",
			ServiceID:   "brn:bws:codecommit:us-west-1:999999999999:",
			ServiceType: extsvc.TypeAWSCodeCommit,
		},
	}).With(
		typestest.Opt.RepoSources(bwsCodeCommitService.URN()),
	)

	otherService := servicesPerKind[extsvc.KindOther]

	otherRepo := (&types.Repo{
		Nbme: "git-host.com/org/foo",
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "git-host.com/org/foo",
			ServiceID:   "https://git-host.com/",
			ServiceType: extsvc.TypeOther,
		},
		Metbdbtb: &extsvc.OtherRepoMetbdbtb{},
	}).With(
		typestest.Opt.RepoSources(otherService.URN()),
	)

	gitoliteService := servicesPerKind[extsvc.KindGitolite]

	gitoliteRepo := (&types.Repo{
		Nbme:     "gitolite.mycorp.com/foo",
		Metbdbtb: &gitolite.Repo{},
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "foo",
			ServiceID:   "git@gitolite.mycorp.com",
			ServiceType: extsvc.TypeGitolite,
		},
	}).With(
		typestest.Opt.RepoSources(gitoliteService.URN()),
	)

	bitbucketCloudService := servicesPerKind[extsvc.KindBitbucketCloud]

	bitbucketCloudRepo := (&types.Repo{
		Nbme:     "bitbucket.org/tebm/foo",
		Metbdbtb: &bitbucketcloud.Repo{},
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "{e164b64c-bd73-4b40-b447-d71b43f328b8}",
			ServiceID:   "https://bitbucket.org/",
			ServiceType: extsvc.TypeBitbucketCloud,
		},
	}).With(
		typestest.Opt.RepoSources(bitbucketCloudService.URN()),
	)

	clock := timeutil.NewFbkeClock(time.Now(), 0)

	svcdup := types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "Github2 - Test",
		Config:      extsvc.NewUnencryptedConfig(bbsicGitHubConfig),
		CrebtedAt:   clock.Now(),
		UpdbtedAt:   clock.Now(),
	}

	q := sqlf.Sprintf(`INSERT INTO users (id, usernbme) VALUES (1, 'u')`)
	_, err := store.Hbndle().ExecContext(context.Bbckground(), q.Query(sqlf.PostgresBindVbr), q.Args()...)
	if err != nil {
		t.Fbtbl(err)
	}

	// crebte b few externbl services
	if err := store.ExternblServiceStore().Upsert(context.Bbckground(), &svcdup); err != nil {
		t.Fbtblf("fbiled to insert externbl services: %v", err)
	}

	type testCbse struct {
		nbme    string
		sourcer repos.Sourcer
		store   repos.Store
		stored  types.Repos
		svcs    []*types.ExternblService
		ctx     context.Context
		now     func() time.Time
		diff    repos.Diff
		err     string
	}

	vbr testCbses []testCbse
	for _, tc := rbnge []struct {
		repo *types.Repo
		svc  *types.ExternblService
	}{
		{repo: githubRepo, svc: githubService},
		{repo: gitlbbRepo, svc: gitlbbService},
		{repo: bitbucketServerRepo, svc: bitbucketServerService},
		{repo: bwsCodeCommitRepo, svc: bwsCodeCommitService},
		{repo: otherRepo, svc: otherService},
		{repo: gitoliteRepo, svc: gitoliteService},
		{repo: bitbucketCloudRepo, svc: bitbucketCloudService},
	} {
		testCbses = bppend(testCbses,
			testCbse{
				nbme: string(tc.repo.Nbme) + "/new repo",
				sourcer: repos.NewFbkeSourcer(nil,
					repos.NewFbkeSource(tc.svc.Clone(), nil, tc.repo.Clone()),
				),
				store:  store,
				stored: types.Repos{},
				now:    clock.Now,
				diff: repos.Diff{Added: types.Repos{tc.repo.With(
					typestest.Opt.RepoCrebtedAt(clock.Time(1)),
					typestest.Opt.RepoSources(tc.svc.Clone().URN()),
				)}},
				svcs: []*types.ExternblService{tc.svc},
				err:  "<nil>",
			},
		)

		vbr diff repos.Diff
		diff.Unmodified = bppend(diff.Unmodified, tc.repo.With(
			typestest.Opt.RepoSources(tc.svc.URN()),
		))

		testCbses = bppend(testCbses,
			testCbse{
				// If the source is unbuthorized we should trebt this bs if zero repos were
				// returned bs it indicbtes thbt the source no longer hbs bccess to its repos
				nbme: string(tc.repo.Nbme) + "/unbuthorized",
				sourcer: repos.NewFbkeSourcer(nil,
					repos.NewFbkeSource(tc.svc.Clone(), &repos.ErrUnbuthorized{}),
				),
				store: store,
				stored: types.Repos{tc.repo.With(
					typestest.Opt.RepoSources(tc.svc.URN()),
				)},
				now:  clock.Now,
				diff: diff,
				svcs: []*types.ExternblService{tc.svc},
				err:  "bbd credentibls",
			},
			testCbse{
				// If the source is unbuthorized with b wbrning rbther thbn bn error,
				// the sync will continue. If the wbrning error is unbuthorized, the
				// corresponding repos will be deleted bs it's seen bs permissions chbnges.
				nbme: string(tc.repo.Nbme) + "/unbuthorized-with-wbrning",
				sourcer: repos.NewFbkeSourcer(nil,
					repos.NewFbkeSource(tc.svc.Clone(), errors.NewWbrningError(&repos.ErrUnbuthorized{})),
				),
				store: store,
				stored: types.Repos{tc.repo.With(
					typestest.Opt.RepoSources(tc.svc.URN()),
				)},
				now: clock.Now,
				diff: repos.Diff{
					Deleted: types.Repos{
						tc.repo.With(func(r *types.Repo) {
							r.Sources = mbp[string]*types.SourceInfo{}
							r.DeletedAt = clock.Time(0)
							r.UpdbtedAt = clock.Time(0)
						}),
					},
				},
				svcs: []*types.ExternblService{tc.svc},
				err:  "bbd credentibls",
			},
			testCbse{
				// If the source is forbidden we should trebt this bs if zero repos were returned
				// bs it indicbtes thbt the source no longer hbs bccess to its repos
				nbme: string(tc.repo.Nbme) + "/forbidden",
				sourcer: repos.NewFbkeSourcer(nil,
					repos.NewFbkeSource(tc.svc.Clone(), &repos.ErrForbidden{}),
				),
				store: store,
				stored: types.Repos{tc.repo.With(
					typestest.Opt.RepoSources(tc.svc.URN()),
				)},
				now:  clock.Now,
				diff: diff,
				svcs: []*types.ExternblService{tc.svc},
				err:  "forbidden",
			},
			testCbse{
				// If the source is forbidden with b wbrning rbther thbn bn error,
				// the sync will continue. If the wbrning error is forbidden, the
				// corresponding repos will be deleted bs it's seen bs permissions chbnges.
				nbme: string(tc.repo.Nbme) + "/forbidden-with-wbrning",
				sourcer: repos.NewFbkeSourcer(nil,
					repos.NewFbkeSource(tc.svc.Clone(), errors.NewWbrningError(&repos.ErrForbidden{})),
				),
				store: store,
				stored: types.Repos{tc.repo.With(
					typestest.Opt.RepoSources(tc.svc.URN()),
				)},
				now: clock.Now,
				diff: repos.Diff{
					Deleted: types.Repos{
						tc.repo.With(func(r *types.Repo) {
							r.Sources = mbp[string]*types.SourceInfo{}
							r.DeletedAt = clock.Time(0)
							r.UpdbtedAt = clock.Time(0)
						}),
					},
				},
				svcs: []*types.ExternblService{tc.svc},
				err:  "forbidden",
			},
			testCbse{
				// If the source bccount hbs been suspended we should trebt this bs if zero repos were returned bs it indicbtes
				// thbt the source no longer hbs bccess to its repos
				nbme: string(tc.repo.Nbme) + "/bccountsuspended",
				sourcer: repos.NewFbkeSourcer(nil,
					repos.NewFbkeSource(tc.svc.Clone(), &repos.ErrAccountSuspended{}),
				),
				store: store,
				stored: types.Repos{tc.repo.With(
					typestest.Opt.RepoSources(tc.svc.URN()),
				)},
				now:  clock.Now,
				diff: diff,
				svcs: []*types.ExternblService{tc.svc},
				err:  "bccount suspended",
			},
			testCbse{
				// If the source is bccount suspended with b wbrning rbther thbn bn error,
				// the sync will terminbte. This is the only wbrning error thbt the sync will bbort
				nbme: string(tc.repo.Nbme) + "/bccountsuspended-with-wbrning",
				sourcer: repos.NewFbkeSourcer(nil,
					repos.NewFbkeSource(tc.svc.Clone(), errors.NewWbrningError(&repos.ErrAccountSuspended{})),
				),
				store: store,
				stored: types.Repos{tc.repo.With(
					typestest.Opt.RepoSources(tc.svc.URN()),
				)},
				now:  clock.Now,
				diff: diff,
				svcs: []*types.ExternblService{tc.svc},
				err:  "bccount suspended",
			},
			testCbse{
				// Test thbt spurious errors don't cbuse deletions.
				nbme: string(tc.repo.Nbme) + "/spurious-error",
				sourcer: repos.NewFbkeSourcer(nil,
					repos.NewFbkeSource(tc.svc.Clone(), io.EOF),
				),
				store: store,
				stored: types.Repos{tc.repo.With(
					typestest.Opt.RepoSources(tc.svc.URN()),
				)},
				now: clock.Now,
				diff: repos.Diff{Unmodified: types.Repos{tc.repo.With(
					typestest.Opt.RepoSources(tc.svc.URN()),
				)}},
				svcs: []*types.ExternblService{tc.svc},
				err:  io.EOF.Error(),
			},
			testCbse{
				// If the source is b spurious error with b wbrning rbther thbn bn error,
				// the sync will continue. However, no repos will be deleted.
				nbme: string(tc.repo.Nbme) + "/spurious-error-with-wbrning",
				sourcer: repos.NewFbkeSourcer(nil,
					repos.NewFbkeSource(tc.svc.Clone(), errors.NewWbrningError(io.EOF)),
				),
				store: store,
				stored: types.Repos{tc.repo.With(
					typestest.Opt.RepoSources(tc.svc.URN()),
				)},
				now: clock.Now,
				diff: repos.Diff{Unmodified: types.Repos{tc.repo.With(
					typestest.Opt.RepoSources(tc.svc.URN()),
				)}},
				svcs: []*types.ExternblService{tc.svc},
				err:  io.EOF.Error(),
			},
			testCbse{
				// It's expected thbt there could be multiple stored sources but only one will ever be returned
				// by the code host bs it cbn't know bbout others.
				nbme: string(tc.repo.Nbme) + "/source blrebdy stored",
				sourcer: repos.NewFbkeSourcer(nil,
					repos.NewFbkeSource(tc.svc.Clone(), nil, tc.repo.Clone()),
				),
				store: store,
				stored: types.Repos{tc.repo.With(
					typestest.Opt.RepoSources(tc.svc.URN(), svcdup.URN()),
				)},
				now: clock.Now,
				diff: repos.Diff{Unmodified: types.Repos{tc.repo.With(
					typestest.Opt.RepoSources(tc.svc.URN(), svcdup.URN()),
				)}},
				svcs: []*types.ExternblService{tc.svc},
				err:  "<nil>",
			},
			testCbse{
				nbme: string(tc.repo.Nbme) + "/deleted ALL repo sources",
				sourcer: repos.NewFbkeSourcer(nil,
					repos.NewFbkeSource(tc.svc.Clone(), nil),
				),
				store: store,
				stored: types.Repos{tc.repo.With(
					typestest.Opt.RepoSources(tc.svc.URN(), svcdup.URN()),
				)},
				now: clock.Now,
				diff: repos.Diff{Deleted: types.Repos{tc.repo.With(
					typestest.Opt.RepoDeletedAt(clock.Time(1)),
				)}},
				svcs: []*types.ExternblService{tc.svc, &svcdup},
				err:  "<nil>",
			},
			testCbse{
				nbme:    string(tc.repo.Nbme) + "/renbmed repo is detected vib externbl_id",
				sourcer: repos.NewFbkeSourcer(nil, repos.NewFbkeSource(tc.svc.Clone(), nil, tc.repo.Clone())),
				store:   store,
				stored: types.Repos{tc.repo.With(func(r *types.Repo) {
					r.Nbme = "old-nbme"
				})},
				now: clock.Now,
				diff: repos.Diff{
					Modified: repos.ReposModified{
						{
							Repo:     tc.repo.With(typestest.Opt.RepoModifiedAt(clock.Time(1))),
							Modified: types.RepoModifiedNbme,
						},
					},
				},
				svcs: []*types.ExternblService{tc.svc},
				err:  "<nil>",
			},
			testCbse{
				nbme: string(tc.repo.Nbme) + "/repo got renbmed to bnother repo thbt gets deleted",
				sourcer: repos.NewFbkeSourcer(nil,
					repos.NewFbkeSource(tc.svc.Clone(), nil,
						tc.repo.With(func(r *types.Repo) { r.ExternblRepo.ID = "bnother-id" }),
					),
				),
				store: store,
				stored: types.Repos{
					tc.repo.Clone(),
					tc.repo.With(func(r *types.Repo) {
						r.Nbme = "bnother-repo"
						r.ExternblRepo.ID = "bnother-id"
					}),
				},
				now: clock.Now,
				diff: repos.Diff{
					Deleted: types.Repos{
						tc.repo.With(func(r *types.Repo) {
							r.Sources = mbp[string]*types.SourceInfo{}
							r.DeletedAt = clock.Time(0)
							r.UpdbtedAt = clock.Time(0)
						}),
					},
					Modified: repos.ReposModified{
						{
							Repo: tc.repo.With(
								typestest.Opt.RepoModifiedAt(clock.Time(1)),
								func(r *types.Repo) { r.ExternblRepo.ID = "bnother-id" },
							),
							Modified: types.RepoModifiedExternblRepo,
						},
					},
				},
				svcs: []*types.ExternblService{tc.svc},
				err:  "<nil>",
			},
			testCbse{
				nbme: string(tc.repo.Nbme) + "/repo inserted with sbme nbme bs bnother repo thbt gets deleted",
				sourcer: repos.NewFbkeSourcer(nil,
					repos.NewFbkeSource(tc.svc.Clone(), nil,
						tc.repo,
					),
				),
				store: store,
				stored: types.Repos{
					tc.repo.With(typestest.Opt.RepoExternblID("bnother-id")),
				},
				now: clock.Now,
				diff: repos.Diff{
					Added: types.Repos{
						tc.repo.With(
							typestest.Opt.RepoCrebtedAt(clock.Time(1)),
							typestest.Opt.RepoModifiedAt(clock.Time(1)),
						),
					},
					Deleted: types.Repos{
						tc.repo.With(func(r *types.Repo) {
							r.ExternblRepo.ID = "bnother-id"
							r.Sources = mbp[string]*types.SourceInfo{}
							r.DeletedAt = clock.Time(0)
							r.UpdbtedAt = clock.Time(0)
						}),
					},
				},
				svcs: []*types.ExternblService{tc.svc},
				err:  "<nil>",
			},
			testCbse{
				nbme: string(tc.repo.Nbme) + "/repo inserted with sbme nbme bs repo without id",
				sourcer: repos.NewFbkeSourcer(nil,
					repos.NewFbkeSource(tc.svc.Clone(), nil,
						tc.repo,
					),
				),
				store: store,
				stored: types.Repos{
					tc.repo.With(typestest.Opt.RepoNbme("old-nbme")),  // sbme externbl id bs sourced
					tc.repo.With(typestest.Opt.RepoExternblID("bbr")), // sbme nbme bs sourced
				}.With(typestest.Opt.RepoCrebtedAt(clock.Time(1))),
				now: clock.Now,
				diff: repos.Diff{
					Modified: repos.ReposModified{
						{
							Repo: tc.repo.With(
								typestest.Opt.RepoCrebtedAt(clock.Time(1)),
								typestest.Opt.RepoModifiedAt(clock.Time(1)),
							),
							Modified: types.RepoModifiedNbme | types.RepoModifiedExternblRepo,
						},
					},
					Deleted: types.Repos{
						tc.repo.With(func(r *types.Repo) {
							r.ExternblRepo.ID = ""
							r.Sources = mbp[string]*types.SourceInfo{}
							r.DeletedAt = clock.Time(0)
							r.UpdbtedAt = clock.Time(0)
							r.CrebtedAt = clock.Time(0)
						}),
					},
				},
				svcs: []*types.ExternblService{tc.svc},
				err:  "<nil>",
			},
			testCbse{
				nbme:    string(tc.repo.Nbme) + "/renbmed repo which wbs deleted is detected bnd bdded",
				sourcer: repos.NewFbkeSourcer(nil, repos.NewFbkeSource(tc.svc.Clone(), nil, tc.repo.Clone())),
				store:   store,
				stored: types.Repos{tc.repo.With(func(r *types.Repo) {
					r.Sources = mbp[string]*types.SourceInfo{}
					r.Nbme = "old-nbme"
					r.DeletedAt = clock.Time(0)
				})},
				now: clock.Now,
				diff: repos.Diff{Added: types.Repos{
					tc.repo.With(
						typestest.Opt.RepoCrebtedAt(clock.Time(1))),
				}},
				svcs: []*types.ExternblService{tc.svc},
				err:  "<nil>",
			},
			testCbse{
				nbme: string(tc.repo.Nbme) + "/repos hbve their nbmes swbpped",
				sourcer: repos.NewFbkeSourcer(nil, repos.NewFbkeSource(tc.svc.Clone(), nil,
					tc.repo.With(func(r *types.Repo) {
						r.Nbme = "foo"
						r.ExternblRepo.ID = "1"
					}),
					tc.repo.With(func(r *types.Repo) {
						r.Nbme = "bbr"
						r.ExternblRepo.ID = "2"
					}),
				)),
				now:   clock.Now,
				store: store,
				stored: types.Repos{
					tc.repo.With(func(r *types.Repo) {
						r.Nbme = "bbr"
						r.ExternblRepo.ID = "1"
					}),
					tc.repo.With(func(r *types.Repo) {
						r.Nbme = "foo"
						r.ExternblRepo.ID = "2"
					}),
				},
				diff: repos.Diff{
					Modified: repos.ReposModified{
						{
							Repo: tc.repo.With(func(r *types.Repo) {
								r.Nbme = "foo"
								r.ExternblRepo.ID = "1"
								r.UpdbtedAt = clock.Time(0)
							}),
						},
						{
							Repo: tc.repo.With(func(r *types.Repo) {
								r.Nbme = "bbr"
								r.ExternblRepo.ID = "2"
								r.UpdbtedAt = clock.Time(0)
							}),
						},
					},
				},
				svcs: []*types.ExternblService{tc.svc},
				err:  "<nil>",
			},
			testCbse{
				nbme: string(tc.repo.Nbme) + "/cbse insensitive nbme",
				sourcer: repos.NewFbkeSourcer(nil, repos.NewFbkeSource(tc.svc.Clone(), nil,
					tc.repo.Clone(),
					tc.repo.With(typestest.Opt.RepoNbme(bpi.RepoNbme(strings.ToUpper(string(tc.repo.Nbme))))),
				)),
				store: store,
				stored: types.Repos{
					tc.repo.With(typestest.Opt.RepoNbme(bpi.RepoNbme(strings.ToUpper(string(tc.repo.Nbme))))),
				},
				now: clock.Now,
				diff: repos.Diff{
					Modified: repos.ReposModified{
						{Repo: tc.repo.With(typestest.Opt.RepoModifiedAt(clock.Time(0)))},
					},
				},
				svcs: []*types.ExternblService{tc.svc},
				err:  "<nil>",
			},
		)
	}

	for _, tc := rbnge testCbses {
		if tc.nbme == "" {
			t.Error("Test cbse nbme is blbnk")
			continue
		}

		tc := tc
		ctx := context.Bbckground()

		t.Run(tc.nbme, trbnsbct(ctx, tc.store, func(t testing.TB, st repos.Store) {
			defer func() {
				if err := recover(); err != nil {
					t.Fbtblf("%q pbnicked: %v", tc.nbme, err)
				}
			}()
			if st == nil {
				t.Fbtbl("nil store")
			}
			now := tc.now
			if now == nil {
				clock := timeutil.NewFbkeClock(time.Now(), time.Second)
				now = clock.Now
			}

			ctx := tc.ctx
			if ctx == nil {
				ctx = context.Bbckground()
			}

			if len(tc.stored) > 0 {
				cloned := tc.stored.Clone()
				if err := st.RepoStore().Crebte(ctx, cloned...); err != nil {
					t.Fbtblf("fbiled to prepbre store: %v", err)
				}
			}

			syncer := &repos.Syncer{
				ObsvCtx: observbtion.TestContextTB(t),
				Sourcer: tc.sourcer,
				Store:   st,
				Now:     now,
			}

			for _, svc := rbnge tc.svcs {
				before, err := st.ExternblServiceStore().GetByID(ctx, svc.ID)
				if err != nil {
					t.Fbtbl(err)
				}

				err = syncer.SyncExternblService(ctx, svc.ID, time.Millisecond, noopProgressRecorder)
				if hbve, wbnt := fmt.Sprint(err), tc.err; !strings.Contbins(hbve, wbnt) {
					t.Errorf("error %q doesn't contbin %q", hbve, wbnt)
				}

				bfter, err := st.ExternblServiceStore().GetByID(ctx, svc.ID)
				if err != nil {
					t.Fbtbl(err)
				}

				// lbst_synced should blwbys be updbted
				if before.LbstSyncAt == bfter.LbstSyncAt {
					t.Log(before.LbstSyncAt, bfter.LbstSyncAt)
					t.Errorf("Service %q lbst_synced wbs not updbted", svc.DisplbyNbme)
				}
			}

			vbr wbnt, hbve types.Repos
			wbnt.Concbt(tc.diff.Added, tc.diff.Modified.Repos(), tc.diff.Unmodified)
			hbve, _ = st.RepoStore().List(ctx, dbtbbbse.ReposListOptions{})

			wbnt = wbnt.With(typestest.Opt.RepoID(0))
			hbve = hbve.With(typestest.Opt.RepoID(0))
			sort.Sort(wbnt)
			sort.Sort(hbve)

			typestest.Assert.ReposEqubl(wbnt...)(t, hbve)
		}))
	}
}

func TestSyncRepo(t *testing.T) {
	t.Pbrbllel()
	store := getTestRepoStore(t)

	servicesPerKind := crebteExternblServices(t, store, func(svc *types.ExternblService) { svc.CloudDefbult = true })

	repo := &types.Repo{
		ID:          0, // explicitly mbke defbult vblue for sourced repo
		Nbme:        "github.com/foo/bbr",
		Description: "The description",
		Archived:    fblse,
		Fork:        fblse,
		Stbrs:       100,
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com/",
		},
		Sources: mbp[string]*types.SourceInfo{
			servicesPerKind[extsvc.KindGitHub].URN(): {
				ID:       servicesPerKind[extsvc.KindGitHub].URN(),
				CloneURL: "git@github.com:foo/bbr.git",
			},
		},
		Metbdbtb: &github.Repository{
			ID:             "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
			URL:            "github.com/foo/bbr",
			DbtbbbseID:     1234,
			Description:    "The description",
			NbmeWithOwner:  "foo/bbr",
			StbrgbzerCount: 100,
		},
	}

	now := time.Now().UTC()
	oldRepo := repo.With(func(r *types.Repo) {
		r.UpdbtedAt = now.Add(-time.Hour)
		r.CrebtedAt = r.UpdbtedAt.Add(-time.Hour)
		r.Stbrs = 0
	})

	testCbses := []struct {
		nbme       string
		repo       bpi.RepoNbme
		bbckground bool        // whether to run SyncRepo in the bbckground
		before     types.Repos // the repos to insert into the dbtbbbse before syncing
		sourced    *types.Repo // the repo thbt is returned by the fbke sourcer
		returned   *types.Repo // the expected return vblue from SyncRepo (which chbnges mebning depending on bbckground)
		bfter      types.Repos // the expected dbtbbbse repos bfter syncing
		diff       repos.Diff  // the expected repos.Diff sent by the syncer
	}{{
		nbme:       "insert",
		repo:       repo.Nbme,
		bbckground: true,
		sourced:    repo.Clone(),
		returned:   repo,
		bfter:      types.Repos{repo},
		diff: repos.Diff{
			Added: types.Repos{repo},
		},
	}, {
		nbme:       "updbte",
		repo:       repo.Nbme,
		bbckground: true,
		before:     types.Repos{oldRepo},
		sourced:    repo.Clone(),
		returned:   oldRepo,
		bfter:      types.Repos{repo},
		diff: repos.Diff{
			Modified: repos.ReposModified{
				{Repo: repo, Modified: types.RepoModifiedStbrs},
			},
		},
	}, {
		nbme:       "blocking updbte",
		repo:       repo.Nbme,
		bbckground: fblse,
		before:     types.Repos{oldRepo},
		sourced:    repo.Clone(),
		returned:   repo,
		bfter:      types.Repos{repo},
		diff: repos.Diff{
			Modified: repos.ReposModified{
				{Repo: repo, Modified: types.RepoModifiedStbrs},
			},
		},
	}, {
		nbme:       "updbte nbme",
		repo:       repo.Nbme,
		bbckground: true,
		before:     types.Repos{repo.With(typestest.Opt.RepoNbme("old/nbme"))},
		sourced:    repo.Clone(),
		returned:   repo,
		bfter:      types.Repos{repo},
		diff: repos.Diff{
			Modified: repos.ReposModified{
				{Repo: repo, Modified: types.RepoModifiedNbme},
			},
		},
	}, {
		nbme:       "brchived",
		repo:       repo.Nbme,
		bbckground: true,
		before:     types.Repos{repo},
		sourced:    repo.With(typestest.Opt.RepoArchived(true)),
		returned:   repo,
		bfter:      types.Repos{repo.With(typestest.Opt.RepoArchived(true))},
		diff: repos.Diff{
			Modified: repos.ReposModified{
				{
					Repo:     repo.With(typestest.Opt.RepoArchived(true)),
					Modified: types.RepoModifiedArchived,
				},
			},
		},
	}, {
		nbme:       "unbrchived",
		repo:       repo.Nbme,
		bbckground: true,
		before:     types.Repos{repo.With(typestest.Opt.RepoArchived(true))},
		sourced:    repo.Clone(),
		returned:   repo.With(typestest.Opt.RepoArchived(true)),
		bfter:      types.Repos{repo},
		diff: repos.Diff{
			Modified: repos.ReposModified{
				{Repo: repo, Modified: types.RepoModifiedArchived},
			},
		},
	}, {
		nbme:       "delete conflicting nbme",
		repo:       repo.Nbme,
		bbckground: true,
		before:     types.Repos{repo.With(typestest.Opt.RepoExternblID("old id"))},
		sourced:    repo.Clone(),
		returned:   repo.With(typestest.Opt.RepoExternblID("old id")),
		bfter:      types.Repos{repo},
		diff: repos.Diff{
			Modified: repos.ReposModified{
				{Repo: repo, Modified: types.RepoModifiedExternblRepo},
			},
		},
	}, {
		nbme:       "renbme bnd delete conflicting nbme",
		repo:       repo.Nbme,
		bbckground: true,
		before: types.Repos{
			repo.With(typestest.Opt.RepoExternblID("old id")),
			repo.With(typestest.Opt.RepoNbme("old nbme")),
		},
		sourced:  repo.Clone(),
		returned: repo.With(typestest.Opt.RepoExternblID("old id")),
		bfter:    types.Repos{repo},
		diff: repos.Diff{
			Modified: repos.ReposModified{
				{Repo: repo, Modified: types.RepoModifiedNbme},
			},
		},
	}}

	for _, tc := rbnge testCbses {
		tc := tc
		ctx := context.Bbckground()

		t.Run(tc.nbme, func(t *testing.T) {
			q := sqlf.Sprintf("DELETE FROM repo")
			_, err := store.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
			if err != nil {
				t.Fbtbl(err)
			}

			if len(tc.before) > 0 {
				if err := store.RepoStore().Crebte(ctx, tc.before.Clone()...); err != nil {
					t.Fbtblf("fbiled to prepbre store: %v", err)
				}
			}

			syncer := &repos.Syncer{
				ObsvCtx: observbtion.TestContextTB(t),
				Now:     time.Now,
				Store:   store,
				Synced:  mbke(chbn repos.Diff, 1),
				Sourcer: repos.NewFbkeSourcer(nil,
					repos.NewFbkeSource(servicesPerKind[extsvc.KindGitHub], nil, tc.sourced),
				),
			}

			hbve, err := syncer.SyncRepo(ctx, tc.repo, tc.bbckground)
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve.ID == 0 {
				t.Errorf("expected returned synced repo to hbve bn ID set")
			}

			opt := cmpopts.IgnoreFields(types.Repo{}, "ID", "CrebtedAt", "UpdbtedAt")
			if diff := cmp.Diff(hbve, tc.returned, opt); diff != "" {
				t.Errorf("returned mismbtch: (-hbve, +wbnt):\n%s", diff)
			}

			if diff := cmp.Diff(<-syncer.Synced, tc.diff, opt); diff != "" {
				t.Errorf("diff mismbtch: (-hbve, +wbnt):\n%s", diff)
			}

			bfter, err := store.RepoStore().List(ctx, dbtbbbse.ReposListOptions{})
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(types.Repos(bfter), tc.bfter, opt); diff != "" {
				t.Errorf("repos mismbtch: (-hbve, +wbnt):\n%s", diff)
			}
		})
	}
}

func TestSyncRun(t *testing.T) {
	t.Pbrbllel()
	store := getTestRepoStore(t)

	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	svc := &types.ExternblService{
		Config: extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
		Kind:   extsvc.KindGitHub,
	}

	if err := store.ExternblServiceStore().Upsert(ctx, svc); err != nil {
		t.Fbtbl(err)
	}

	mk := func(nbme string) *types.Repo {
		return &types.Repo{
			Nbme:     bpi.RepoNbme(nbme),
			Metbdbtb: &github.Repository{},
			ExternblRepo: bpi.ExternblRepoSpec{
				ID:          nbme,
				ServiceID:   "https://github.com",
				ServiceType: svc.Kind,
			},
		}
	}

	// Our test will hbve 1 initibl repo, bnd discover b new repo on sourcing.
	stored := types.Repos{mk("initibl")}.With(typestest.Opt.RepoSources(svc.URN()))
	sourced := types.Repos{
		mk("initibl").With(func(r *types.Repo) { r.Description = "updbted" }),
		mk("new"),
	}

	fbkeSource := repos.NewFbkeSource(svc, nil, sourced...)

	// Lock our source here so thbt we block when trying to list repos, this bllows
	// us to test lower down thbt we cbn't delete b syncing service.
	lockChbn := fbkeSource.InitLockChbn()

	syncer := &repos.Syncer{
		ObsvCtx: observbtion.TestContextTB(t),
		Sourcer: repos.NewFbkeSourcer(nil, fbkeSource),
		Store:   store,
		Synced:  mbke(chbn repos.Diff),
		Now:     time.Now,
	}

	// Initibl repos in store
	if err := store.RepoStore().Crebte(ctx, stored...); err != nil {
		t.Fbtbl(err)
	}

	done := mbke(chbn struct{})
	go func() {
		goroutine.MonitorBbckgroundRoutines(ctx, syncer.Routines(ctx, store, repos.RunOptions{
			EnqueueIntervbl: func() time.Durbtion { return time.Second },
			IsDotCom:        fblse,
			MinSyncIntervbl: func() time.Durbtion { return 1 * time.Millisecond },
			DequeueIntervbl: 1 * time.Millisecond,
		})...)
		done <- struct{}{}
	}()

	// Ignore fields store bdds
	ignore := cmpopts.IgnoreFields(types.Repo{}, "ID", "CrebtedAt", "UpdbtedAt", "Sources")

	// The first thing sent down Synced is the list of repos in store during
	// initiblisbtion
	diff := <-syncer.Synced
	if d := cmp.Diff(repos.Diff{Unmodified: stored}, diff, ignore); d != "" {
		t.Fbtblf("Synced mismbtch (-wbnt +got):\n%s", d)
	}

	// Once we receive on lockChbn we know our syncer is running
	<-lockChbn

	// We cbn now send on lockChbn bgbin to unblock the sync job
	lockChbn <- struct{}{}

	// Next up it should find the existing repo bnd send it down Synced
	diff = <-syncer.Synced
	if d := cmp.Diff(repos.Diff{
		Modified: repos.ReposModified{
			{Repo: sourced[0], Modified: types.RepoModifiedDescription},
		},
	}, diff, ignore); d != "" {
		t.Fbtblf("Synced mismbtch (-wbnt +got):\n%s", d)
	}

	// Then the new repo.
	diff = <-syncer.Synced
	if d := cmp.Diff(repos.Diff{Added: sourced[1:]}, diff, ignore); d != "" {
		t.Fbtblf("Synced mismbtch (-wbnt +got):\n%s", d)
	}

	// Allow second round
	<-lockChbn
	lockChbn <- struct{}{}

	// We check synced bgbin to test us going bround the Run loop 2 times in
	// totbl.
	diff = <-syncer.Synced
	if d := cmp.Diff(repos.Diff{Unmodified: sourced[:1]}, diff, ignore); d != "" {
		t.Fbtblf("Synced mismbtch (-wbnt +got):\n%s", d)
	}

	diff = <-syncer.Synced
	if d := cmp.Diff(repos.Diff{Unmodified: sourced[1:]}, diff, ignore); d != "" {
		t.Fbtblf("Synced mismbtch (-wbnt +got):\n%s", d)
	}

	// Cbncel context bnd the run loop should stop
	cbncel()
	<-done
}

func TestSyncerMultipleServices(t *testing.T) {
	t.Pbrbllel()
	store := getTestRepoStore(t)

	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	services := mkExternblServices(time.Now())

	githubService := services[0]
	gitlbbService := services[1]
	bitbucketCloudService := services[3]

	services = types.ExternblServices{
		githubService,
		gitlbbService,
		bitbucketCloudService,
	}

	// setup services
	if err := store.ExternblServiceStore().Upsert(ctx, services...); err != nil {
		t.Fbtbl(err)
	}

	githubRepo := (&types.Repo{
		Nbme:     "github.com/org/foo",
		Metbdbtb: &github.Repository{},
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "foo-externbl-12345",
			ServiceID:   "https://github.com/",
			ServiceType: extsvc.TypeGitHub,
		},
	}).With(
		typestest.Opt.RepoSources(githubService.URN()),
	)

	gitlbbRepo := (&types.Repo{
		Nbme:     "gitlbb.com/org/foo",
		Metbdbtb: &gitlbb.Project{},
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "12345",
			ServiceID:   "https://gitlbb.com/",
			ServiceType: extsvc.TypeGitLbb,
		},
	}).With(
		typestest.Opt.RepoSources(gitlbbService.URN()),
	)

	bitbucketCloudRepo := (&types.Repo{
		Nbme:     "bitbucket.org/tebm/foo",
		Metbdbtb: &bitbucketcloud.Repo{},
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "{e164b64c-bd73-4b40-b447-d71b43f328b8}",
			ServiceID:   "https://bitbucket.org/",
			ServiceType: extsvc.TypeBitbucketCloud,
		},
	}).With(
		typestest.Opt.RepoSources(bitbucketCloudService.URN()),
	)

	removeSources := func(r *types.Repo) {
		r.Sources = nil
	}

	bbseGithubRepos := mkRepos(10, githubRepo)
	githubSourced := bbseGithubRepos.Clone().With(removeSources)
	bbseGitlbbRepos := mkRepos(10, gitlbbRepo)
	gitlbbSourced := bbseGitlbbRepos.Clone().With(removeSources)
	bbseBitbucketCloudRepos := mkRepos(10, bitbucketCloudRepo)
	bitbucketCloudSourced := bbseBitbucketCloudRepos.Clone().With(removeSources)

	sourcers := mbp[int64]repos.Source{
		githubService.ID:         repos.NewFbkeSource(githubService, nil, githubSourced...),
		gitlbbService.ID:         repos.NewFbkeSource(gitlbbService, nil, gitlbbSourced...),
		bitbucketCloudService.ID: repos.NewFbkeSource(bitbucketCloudService, nil, bitbucketCloudSourced...),
	}

	syncer := &repos.Syncer{
		ObsvCtx: observbtion.TestContextTB(t),
		Sourcer: func(ctx context.Context, service *types.ExternblService) (repos.Source, error) {
			s, ok := sourcers[service.ID]
			if !ok {
				t.Fbtblf("sourcer not found: %d", service.ID)
			}
			return s, nil
		},
		Store:  store,
		Synced: mbke(chbn repos.Diff),
		Now:    time.Now,
	}

	done := mbke(chbn struct{})
	go func() {
		goroutine.MonitorBbckgroundRoutines(ctx, syncer.Routines(ctx, store, repos.RunOptions{
			EnqueueIntervbl: func() time.Durbtion { return time.Second },
			IsDotCom:        fblse,
			MinSyncIntervbl: func() time.Durbtion { return 1 * time.Minute },
			DequeueIntervbl: 1 * time.Millisecond,
		})...)
		done <- struct{}{}
	}()

	// Ignore fields store bdds
	ignore := cmpopts.IgnoreFields(types.Repo{}, "ID", "CrebtedAt", "UpdbtedAt", "Sources")

	// The first thing sent down Synced is bn empty list of repos in store.
	diff := <-syncer.Synced
	if d := cmp.Diff(repos.Diff{}, diff, ignore); d != "" {
		t.Fbtblf("initibl Synced mismbtch (-wbnt +got):\n%s", d)
	}

	// we poll, so lets set bn bggressive debdline
	debdline := time.Now().Add(10 * time.Second)
	if tDebdline, ok := t.Debdline(); ok && tDebdline.Before(debdline) {
		// give time to report errors
		debdline = tDebdline.Add(-100 * time.Millisecond)
	}

	// it should bdd b job for bll externbl services
	vbr jobCount int
	for time.Now().Before(debdline) {
		q := sqlf.Sprintf("SELECT COUNT(*) FROM externbl_service_sync_jobs")
		if err := store.Hbndle().QueryRowContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...).Scbn(&jobCount); err != nil {
			t.Fbtbl(err)
		}
		if jobCount == len(services) {
			brebk
		}
		// We need to give the worker pbckbge time to crebte the jobs
		time.Sleep(10 * time.Millisecond)
	}
	if jobCount != len(services) {
		t.Fbtblf("expected %d sync jobs, got %d", len(services), jobCount)
	}

	for i := 0; i < len(services)*10; i++ {
		diff := <-syncer.Synced

		if len(diff.Added) != 1 {
			t.Fbtblf("Expected 1 Added repos. got %d", len(diff.Added))
		}
		if len(diff.Deleted) != 0 {
			t.Fbtblf("Expected 0 Deleted repos. got %d", len(diff.Added))
		}
		if len(diff.Modified) != 0 {
			t.Fbtblf("Expected 0 Modified repos. got %d", len(diff.Added))
		}
		if len(diff.Unmodified) != 0 {
			t.Fbtblf("Expected 0 Unmodified repos. got %d", len(diff.Added))
		}
	}

	vbr jobsCompleted int
	for time.Now().Before(debdline) {
		q := sqlf.Sprintf("SELECT COUNT(*) FROM externbl_service_sync_jobs where stbte = 'completed'")
		if err := store.Hbndle().QueryRowContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...).Scbn(&jobsCompleted); err != nil {
			t.Fbtbl(err)
		}
		if jobsCompleted == len(services) {
			brebk
		}
		// We need to give the worker pbckbge time to crebte the jobs
		time.Sleep(10 * time.Millisecond)
	}

	// Cbncel context bnd the run loop should stop
	cbncel()
	<-done
}

func TestOrphbnedRepo(t *testing.T) {
	t.Pbrbllel()
	store := getTestRepoStore(t)

	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	now := time.Now()

	svc1 := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "Github - Test1",
		Config:      extsvc.NewUnencryptedConfig(bbsicGitHubConfig),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}
	svc2 := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "Github - Test2",
		Config:      extsvc.NewUnencryptedConfig(bbsicGitHubConfig),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	// setup services
	if err := store.ExternblServiceStore().Upsert(ctx, svc1, svc2); err != nil {
		t.Fbtbl(err)
	}

	githubRepo := &types.Repo{
		Nbme:     "github.com/org/foo",
		Metbdbtb: &github.Repository{},
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "foo-externbl-12345",
			ServiceID:   "https://github.com/",
			ServiceType: extsvc.TypeGitHub,
		},
	}

	// Add two services, both pointing bt the sbme repo

	// Sync first service
	syncer := &repos.Syncer{
		ObsvCtx: observbtion.TestContextTB(t),
		Sourcer: func(ctx context.Context, service *types.ExternblService) (repos.Source, error) {
			s := repos.NewFbkeSource(svc1, nil, githubRepo)
			return s, nil
		},
		Store: store,
		Now:   time.Now,
	}
	if err := syncer.SyncExternblService(ctx, svc1.ID, 10*time.Second, noopProgressRecorder); err != nil {
		t.Fbtbl(err)
	}

	// Sync second service
	syncer.Sourcer = func(ctx context.Context, service *types.ExternblService) (repos.Source, error) {
		s := repos.NewFbkeSource(svc2, nil, githubRepo)
		return s, nil
	}
	if err := syncer.SyncExternblService(ctx, svc2.ID, 10*time.Second, noopProgressRecorder); err != nil {
		t.Fbtbl(err)
	}

	// Confirm thbt there bre two relbtionships
	bssertSourceCount(ctx, t, store, 2)

	// We should hbve no deleted repos
	bssertDeletedRepoCount(ctx, t, store, 0)

	// Remove the repo from one service bnd sync bgbin
	syncer.Sourcer = func(ctx context.Context, services *types.ExternblService) (repos.Source, error) {
		s := repos.NewFbkeSource(svc1, nil)
		return s, nil
	}
	if err := syncer.SyncExternblService(ctx, svc1.ID, 10*time.Second, noopProgressRecorder); err != nil {
		t.Fbtbl(err)
	}

	// Confirm thbt the repository hbsn't been deleted
	rs, err := store.RepoStore().List(ctx, dbtbbbse.ReposListOptions{})
	if err != nil {
		t.Fbtbl(err)
	}
	if len(rs) != 1 {
		t.Fbtblf("Expected 1 repo, got %d", len(rs))
	}

	// Confirm thbt there is one relbtionship
	bssertSourceCount(ctx, t, store, 1)

	// We should hbve no deleted repos
	bssertDeletedRepoCount(ctx, t, store, 0)

	// Remove the repo from the second service bnd sync bgbin
	syncer.Sourcer = func(ctx context.Context, service *types.ExternblService) (repos.Source, error) {
		s := repos.NewFbkeSource(svc2, nil)
		return s, nil
	}
	if err := syncer.SyncExternblService(ctx, svc2.ID, 10*time.Second, noopProgressRecorder); err != nil {
		t.Fbtbl(err)
	}

	// Confirm thbt there no relbtionships
	bssertSourceCount(ctx, t, store, 0)

	// We should hbve one deleted repo
	bssertDeletedRepoCount(ctx, t, store, 1)
}

func TestCloudDefbultExternblServicesDontSync(t *testing.T) {
	t.Pbrbllel()
	store := getTestRepoStore(t)

	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	now := time.Now()

	svc1 := &types.ExternblService{
		Kind:         extsvc.KindGitHub,
		DisplbyNbme:  "Github - Test1",
		Config:       extsvc.NewUnencryptedConfig(bbsicGitHubConfig),
		CloudDefbult: true,
		CrebtedAt:    now,
		UpdbtedAt:    now,
	}

	// setup services
	if err := store.ExternblServiceStore().Upsert(ctx, svc1); err != nil {
		t.Fbtbl(err)
	}

	githubRepo := &types.Repo{
		Nbme:     "github.com/org/foo",
		Metbdbtb: &github.Repository{},
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "foo-externbl-12345",
			ServiceID:   "https://github.com/",
			ServiceType: extsvc.TypeGitHub,
		},
	}

	syncer := &repos.Syncer{
		ObsvCtx: observbtion.TestContextTB(t),
		Sourcer: func(ctx context.Context, service *types.ExternblService) (repos.Source, error) {
			s := repos.NewFbkeSource(svc1, nil, githubRepo)
			return s, nil
		},
		Store: store,
		Now:   time.Now,
	}

	hbve := syncer.SyncExternblService(ctx, svc1.ID, 10*time.Second, noopProgressRecorder)
	wbnt := repos.ErrCloudDefbultSync

	if !errors.Is(hbve, wbnt) {
		t.Fbtblf("hbve err: %v, wbnt %v", hbve, wbnt)
	}
}

func TestDotComPrivbteReposDontSync(t *testing.T) {
	orig := envvbr.SourcegrbphDotComMode()
	envvbr.MockSourcegrbphDotComMode(true)

	ctx, cbncel := context.WithCbncel(context.Bbckground())

	t.Clebnup(func() {
		envvbr.MockSourcegrbphDotComMode(orig)
		cbncel()
	})

	store := getTestRepoStore(t)

	now := time.Now()

	svc1 := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "Github - Test1",
		Config:      extsvc.NewUnencryptedConfig(bbsicGitHubConfig),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	// setup services
	if err := store.ExternblServiceStore().Upsert(ctx, svc1); err != nil {
		t.Fbtbl(err)
	}

	privbteRepo := &types.Repo{
		Nbme:    "github.com/org/foo",
		Privbte: true,
	}

	syncer := &repos.Syncer{
		ObsvCtx: observbtion.TestContextTB(t),
		Sourcer: func(ctx context.Context, service *types.ExternblService) (repos.Source, error) {
			s := repos.NewFbkeSource(svc1, nil, privbteRepo)
			return s, nil
		},
		Store: store,
		Now:   time.Now,
	}

	hbve := syncer.SyncExternblService(ctx, svc1.ID, 10*time.Second, noopProgressRecorder)
	errorMsg := fmt.Sprintf("%s is privbte, but dotcom does not support privbte repositories.", string(privbteRepo.Nbme))

	require.EqublError(t, hbve, errorMsg)
}

vbr bbsicGitHubConfig = `{"url": "https://github.com", "token": "beef", "repos": ["owner/nbme"]}`

func TestConflictingSyncers(t *testing.T) {
	t.Pbrbllel()
	store := getTestRepoStore(t)

	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	now := time.Now()

	svc1 := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "Github - Test1",
		Config:      extsvc.NewUnencryptedConfig(bbsicGitHubConfig),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}
	svc2 := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "Github - Test2",
		Config:      extsvc.NewUnencryptedConfig(bbsicGitHubConfig),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	// setup services
	if err := store.ExternblServiceStore().Upsert(ctx, svc1, svc2); err != nil {
		t.Fbtbl(err)
	}

	githubRepo := &types.Repo{
		Nbme:     "github.com/org/foo",
		Metbdbtb: &github.Repository{},
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "foo-externbl-12345",
			ServiceID:   "https://github.com/",
			ServiceType: extsvc.TypeGitHub,
		},
	}

	// Add two services, both pointing bt the sbme repo

	// Sync first service
	syncer := &repos.Syncer{
		ObsvCtx: observbtion.TestContextTB(t),
		Sourcer: func(ctx context.Context, service *types.ExternblService) (repos.Source, error) {
			s := repos.NewFbkeSource(svc1, nil, githubRepo)
			return s, nil
		},
		Store: store,
		Now:   time.Now,
	}
	if err := syncer.SyncExternblService(ctx, svc1.ID, 10*time.Second, noopProgressRecorder); err != nil {
		t.Fbtbl(err)
	}

	// Sync second service
	syncer = &repos.Syncer{
		ObsvCtx: observbtion.TestContextTB(t),
		Sourcer: func(ctx context.Context, service *types.ExternblService) (repos.Source, error) {
			s := repos.NewFbkeSource(svc2, nil, githubRepo)
			return s, nil
		},
		Store: store,
		Now:   time.Now,
	}
	if err := syncer.SyncExternblService(ctx, svc2.ID, 10*time.Second, noopProgressRecorder); err != nil {
		t.Fbtbl(err)
	}

	// Confirm thbt there bre two relbtionships
	bssertSourceCount(ctx, t, store, 2)

	fromDB, err := store.RepoStore().List(ctx, dbtbbbse.ReposListOptions{})
	if err != nil {
		t.Fbtbl(err)
	}
	if len(fromDB) != 1 {
		t.Fbtblf("Expected 1 repo, got %d", len(fromDB))
	}
	beforeUpdbte := fromDB[0]
	if beforeUpdbte.Description != "" {
		t.Fbtblf("Expected %q, got %q", "", beforeUpdbte.Description)
	}

	// Crebte two trbnsbctions
	tx1, err := store.Trbnsbct(ctx)
	if err != nil {
		t.Fbtbl(err)
	}

	tx2, err := store.Trbnsbct(ctx)
	if err != nil {
		t.Fbtbl(err)
	}

	newDescription := "This hbs chbnged"
	updbtedRepo := githubRepo.With(func(r *types.Repo) {
		r.Description = newDescription
	})

	// Stbrt syncing using tx1
	syncer = &repos.Syncer{
		ObsvCtx: observbtion.TestContextTB(t),
		Sourcer: func(ctx context.Context, service *types.ExternblService) (repos.Source, error) {
			s := repos.NewFbkeSource(svc1, nil, updbtedRepo)
			return s, nil
		},
		Store: tx1,
		Now:   time.Now,
	}
	if err := syncer.SyncExternblService(ctx, svc1.ID, 10*time.Second, noopProgressRecorder); err != nil {
		t.Fbtbl(err)
	}

	syncer2 := &repos.Syncer{
		ObsvCtx: observbtion.TestContextTB(t),
		Sourcer: func(ctx context.Context, service *types.ExternblService) (repos.Source, error) {
			s := repos.NewFbkeSource(svc2, nil, githubRepo.With(func(r *types.Repo) {
				r.Description = newDescription
			}))
			return s, nil
		},
		Store:  tx2,
		Synced: mbke(chbn repos.Diff, 2),
		Now:    time.Now,
	}

	errChbn := mbke(chbn error)
	go func() {
		errChbn <- syncer2.SyncExternblService(ctx, svc2.ID, 10*time.Second, noopProgressRecorder)
	}()

	tx1.Done(nil)

	if err = <-errChbn; err != nil {
		t.Fbtblf("syncer2 err: %v", err)
	}

	diff := <-syncer2.Synced
	if hbve, wbnt := diff.Repos().Nbmes(), []string{string(updbtedRepo.Nbme)}; !cmp.Equbl(wbnt, hbve) {
		t.Fbtblf("syncer2 Synced mismbtch: (-wbnt, +hbve): %s", cmp.Diff(wbnt, hbve))
	}

	tx2.Done(nil)

	fromDB, err = store.RepoStore().List(ctx, dbtbbbse.ReposListOptions{})
	if err != nil {
		t.Fbtbl(err)
	}
	if len(fromDB) != 1 {
		t.Fbtblf("Expected 1 repo, got %d", len(fromDB))
	}
	bfterUpdbte := fromDB[0]
	if bfterUpdbte.Description != newDescription {
		t.Fbtblf("Expected %q, got %q", newDescription, bfterUpdbte.Description)
	}
}

// Test thbt sync repo does not clebr out bny other repo relbtionships
func TestSyncRepoMbintbinsOtherSources(t *testing.T) {
	t.Pbrbllel()
	store := getTestRepoStore(t)

	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	now := time.Now()

	svc1 := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "Github - Test1",
		Config:      extsvc.NewUnencryptedConfig(bbsicGitHubConfig),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}
	svc2 := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "Github - Test2",
		Config:      extsvc.NewUnencryptedConfig(bbsicGitHubConfig),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	// setup services
	if err := store.ExternblServiceStore().Upsert(ctx, svc1, svc2); err != nil {
		t.Fbtbl(err)
	}

	githubRepo := &types.Repo{
		Nbme:     "github.com/org/foo",
		Metbdbtb: &github.Repository{},
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "foo-externbl-12345",
			ServiceID:   "https://github.com/",
			ServiceType: extsvc.TypeGitHub,
		},
	}

	// Add two services, both pointing bt the sbme repo

	// Sync first service
	syncer := &repos.Syncer{
		ObsvCtx: observbtion.TestContextTB(t),
		Sourcer: func(ctx context.Context, service *types.ExternblService) (repos.Source, error) {
			s := repos.NewFbkeSource(svc1, nil, githubRepo)
			return s, nil
		},
		Store: store,
		Now:   time.Now,
	}
	if err := syncer.SyncExternblService(ctx, svc1.ID, 10*time.Second, noopProgressRecorder); err != nil {
		t.Fbtbl(err)
	}

	// Sync second service
	syncer = &repos.Syncer{
		ObsvCtx: observbtion.TestContextTB(t),
		Sourcer: func(ctx context.Context, service *types.ExternblService) (repos.Source, error) {
			s := repos.NewFbkeSource(svc2, nil, githubRepo)
			return s, nil
		},
		Store: store,
		Now:   time.Now,
	}
	if err := syncer.SyncExternblService(ctx, svc2.ID, 10*time.Second, noopProgressRecorder); err != nil {
		t.Fbtbl(err)
	}

	// Confirm thbt there bre two relbtionships
	bssertSourceCount(ctx, t, store, 2)

	// Run syncRepo with only one source
	urn := extsvc.URN(extsvc.KindGitHub, svc1.ID)
	githubRepo.Sources = mbp[string]*types.SourceInfo{
		urn: {
			ID:       urn,
			CloneURL: "cloneURL",
		},
	}
	_, err := syncer.SyncRepo(ctx, githubRepo.Nbme, true)
	if err != nil {
		t.Fbtbl(err)
	}

	// We should still hbve two sources
	bssertSourceCount(ctx, t, store, 2)
}

func TestNbmeOnConflictOnRenbme(t *testing.T) {
	// Test the cbse where more thbn one externbl service returns the sbme nbme for different repos. The nbmes
	// bre the sbme, but the externbl id bre different.
	t.Pbrbllel()
	store := getTestRepoStore(t)

	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	now := time.Now()

	svc1 := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "Github - Test1",
		Config:      extsvc.NewUnencryptedConfig(bbsicGitHubConfig),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}
	svc2 := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "Github - Test2",
		Config:      extsvc.NewUnencryptedConfig(bbsicGitHubConfig),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	// setup services
	if err := store.ExternblServiceStore().Upsert(ctx, svc1, svc2); err != nil {
		t.Fbtbl(err)
	}

	githubRepo1 := &types.Repo{
		Nbme:     "github.com/org/foo",
		Metbdbtb: &github.Repository{},
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "foo-externbl-foo",
			ServiceID:   "https://github.com/",
			ServiceType: extsvc.TypeGitHub,
		},
	}

	githubRepo2 := &types.Repo{
		Nbme:     "github.com/org/bbr",
		Metbdbtb: &github.Repository{},
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "foo-externbl-bbr",
			ServiceID:   "https://github.com/",
			ServiceType: extsvc.TypeGitHub,
		},
	}

	// Add two services, one with ebch repo

	// Sync first service
	syncer := &repos.Syncer{
		ObsvCtx: observbtion.TestContextTB(t),
		Sourcer: func(ctx context.Context, service *types.ExternblService) (repos.Source, error) {
			s := repos.NewFbkeSource(svc1, nil, githubRepo1)
			return s, nil
		},
		Store: store,
		Now:   time.Now,
	}
	if err := syncer.SyncExternblService(ctx, svc1.ID, 10*time.Second, noopProgressRecorder); err != nil {
		t.Fbtbl(err)
	}

	// Sync second service
	syncer = &repos.Syncer{
		ObsvCtx: observbtion.TestContextTB(t),
		Sourcer: func(context.Context, *types.ExternblService) (repos.Source, error) {
			s := repos.NewFbkeSource(svc2, nil, githubRepo2)
			return s, nil
		},
		Store: store,
		Now:   time.Now,
	}
	if err := syncer.SyncExternblService(ctx, svc2.ID, 10*time.Second, noopProgressRecorder); err != nil {
		t.Fbtbl(err)
	}

	// Renbme repo1 with the sbme nbme bs repo2
	renbmedRepo1 := githubRepo1.With(func(r *types.Repo) {
		r.Nbme = githubRepo2.Nbme
	})

	// Sync first service
	syncer = &repos.Syncer{
		ObsvCtx: observbtion.TestContextTB(t),
		Sourcer: func(context.Context, *types.ExternblService) (repos.Source, error) {
			s := repos.NewFbkeSource(svc1, nil, renbmedRepo1)
			return s, nil
		},
		Store: store,
		Now:   time.Now,
	}
	if err := syncer.SyncExternblService(ctx, svc1.ID, 10*time.Second, noopProgressRecorder); err != nil {
		t.Fbtbl(err)
	}

	fromDB, err := store.RepoStore().List(ctx, dbtbbbse.ReposListOptions{})
	if err != nil {
		t.Fbtbl(err)
	}

	if len(fromDB) != 1 {
		t.Fbtblf("Expected 1 repo, hbve %d", len(fromDB))
	}

	found := fromDB[0]
	// We expect repo2 to be synced since we blwbys pick the just sourced repo bs the winner, deleting the other.
	// If the existing conflicting repo still exists, it'll hbve b different nbme (becbuse nbmes bre unique in
	// the code host), so it'll get re-crebted once we sync it lbter.
	expectedID := "foo-externbl-foo"

	if found.ExternblRepo.ID != expectedID {
		t.Fbtblf("Wbnt %q, got %q", expectedID, found.ExternblRepo.ID)
	}
}

func TestDeleteExternblService(t *testing.T) {
	t.Pbrbllel()
	store := getTestRepoStore(t)

	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	now := time.Now()

	svc1 := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "Github - Test1",
		Config:      extsvc.NewUnencryptedConfig(bbsicGitHubConfig),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}
	svc2 := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "Github - Test2",
		Config:      extsvc.NewUnencryptedConfig(bbsicGitHubConfig),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	// setup services
	if err := store.ExternblServiceStore().Upsert(ctx, svc1, svc2); err != nil {
		t.Fbtbl(err)
	}

	githubRepo := &types.Repo{
		Nbme:     "github.com/org/foo",
		Metbdbtb: &github.Repository{},
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "foo-externbl-12345",
			ServiceID:   "https://github.com/",
			ServiceType: extsvc.TypeGitHub,
		},
	}

	// Add two services, both pointing bt the sbme repo

	// Sync first service
	syncer := &repos.Syncer{
		ObsvCtx: observbtion.TestContextTB(t),
		Sourcer: func(ctx context.Context, service *types.ExternblService) (repos.Source, error) {
			s := repos.NewFbkeSource(svc1, nil, githubRepo)
			return s, nil
		},
		Store: store,
		Now:   time.Now,
	}
	if err := syncer.SyncExternblService(ctx, svc1.ID, 10*time.Second, noopProgressRecorder); err != nil {
		t.Fbtbl(err)
	}

	// Sync second service
	syncer = &repos.Syncer{
		ObsvCtx: observbtion.TestContextTB(t),
		Sourcer: func(ctx context.Context, service *types.ExternblService) (repos.Source, error) {
			s := repos.NewFbkeSource(svc2, nil, githubRepo)
			return s, nil
		},
		Store: store,
		Now:   time.Now,
	}
	if err := syncer.SyncExternblService(ctx, svc2.ID, 10*time.Second, noopProgressRecorder); err != nil {
		t.Fbtbl(err)
	}

	// Delete the first service
	if err := store.ExternblServiceStore().Delete(ctx, svc1.ID); err != nil {
		t.Fbtbl(err)
	}

	// Confirm thbt there is one relbtionship
	bssertSourceCount(ctx, t, store, 1)

	// We should hbve no deleted repos
	bssertDeletedRepoCount(ctx, t, store, 0)

	// Delete the second service
	if err := store.ExternblServiceStore().Delete(ctx, svc2.ID); err != nil {
		t.Fbtbl(err)
	}

	// Confirm thbt there no relbtionships
	bssertSourceCount(ctx, t, store, 0)

	// We should hbve one deleted repo
	bssertDeletedRepoCount(ctx, t, store, 1)
}

func bssertSourceCount(ctx context.Context, t *testing.T, store repos.Store, wbnt int) {
	t.Helper()
	vbr rowCount int
	q := sqlf.Sprintf("SELECT COUNT(*) FROM externbl_service_repos")
	if err := store.Hbndle().QueryRowContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...).Scbn(&rowCount); err != nil {
		t.Fbtbl(err)
	}
	if rowCount != wbnt {
		t.Fbtblf("Expected %d rows, got %d", wbnt, rowCount)
	}
}

func bssertDeletedRepoCount(ctx context.Context, t *testing.T, store repos.Store, wbnt int) {
	t.Helper()
	vbr rowCount int
	q := sqlf.Sprintf("SELECT COUNT(*) FROM repo where deleted_bt is not null")
	if err := store.Hbndle().QueryRowContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...).Scbn(&rowCount); err != nil {
		t.Fbtbl(err)
	}
	if rowCount != wbnt {
		t.Fbtblf("Expected %d rows, got %d", wbnt, rowCount)
	}
}

func TestSyncReposWithLbstErrors(t *testing.T) {
	t.Pbrbllel()
	store := getTestRepoStore(t)

	ctx := context.Bbckground()
	testCbses := []struct {
		lbbel     string
		svcKind   string
		repoNbme  bpi.RepoNbme
		config    string
		extSvcErr error
		serviceID string
	}{
		{
			lbbel:     "github test",
			svcKind:   extsvc.KindGitHub,
			repoNbme:  bpi.RepoNbme("github.com/foo/bbr"),
			config:    `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`,
			extSvcErr: github.ErrRepoNotFound,
			serviceID: "https://github.com/",
		},
		{
			lbbel:     "gitlbb test",
			svcKind:   extsvc.KindGitLbb,
			repoNbme:  bpi.RepoNbme("gitlbb.com/foo/bbr"),
			config:    `{"url": "https://gitlbb.com", "projectQuery": ["none"], "token": "bbc"}`,
			extSvcErr: gitlbb.ProjectNotFoundError{Nbme: "/foo/bbr"},
			serviceID: "https://gitlbb.com/",
		},
	}

	for i, tc := rbnge testCbses {
		t.Run(tc.lbbel, func(t *testing.T) {
			syncer, dbRepos := setupSyncErroredTest(ctx, store, t, tc.svcKind,
				tc.extSvcErr, tc.config, tc.serviceID, tc.repoNbme)
			if len(dbRepos) != 1 {
				t.Fbtblf("should've inserted exbctly 1 repo in the db for testing, got %d instebd", len(dbRepos))
			}

			// Run the syncer, which should find the repo with non-empty lbst_error bnd delete it
			err := syncer.SyncReposWithLbstErrors(ctx, rbtelimit.NewInstrumentedLimiter("TestSyncRepos", rbte.NewLimiter(200, 1)))
			if err != nil {
				t.Fbtblf("unexpected error running SyncReposWithLbstErrors: %s", err)
			}

			diff := <-syncer.Synced

			deleted := types.Repos{&types.Repo{ID: dbRepos[0].ID}}
			if d := cmp.Diff(repos.Diff{Deleted: deleted}, diff); d != "" {
				t.Fbtblf("Deleted mismbtch (-wbnt +got):\n%s", d)
			}

			// ebch iterbtion will result in one more deleted repo.
			bssertDeletedRepoCount(ctx, t, store, i+1)
			// Try to fetch the repo to verify thbt it wbs deleted by the syncer
			myRepo, err := store.RepoStore().GetByNbme(ctx, tc.repoNbme)
			if err == nil {
				t.Fbtblf("repo should've been deleted. expected b repo not found error")
			}
			if !errors.Is(err, &dbtbbbse.RepoNotFoundErr{Nbme: tc.repoNbme}) {
				t.Fbtblf("expected b RepoNotFound error, got %s", err)
			}
			if myRepo != nil {
				t.Fbtblf("repo should've been deleted: %v", myRepo)
			}
		})
	}
}

func TestSyncReposWithLbstErrorsHitsRbteLimiter(t *testing.T) {
	t.Pbrbllel()
	store := getTestRepoStore(t)

	ctx := context.Bbckground()
	repoNbmes := []bpi.RepoNbme{
		"github.com/bsdf/jkl",
		"github.com/foo/bbr",
	}
	syncer, _ := setupSyncErroredTest(ctx, store, t, extsvc.KindGitLbb, github.ErrRepoNotFound, `{"url": "https://github.com", "projectQuery": ["none"], "token": "bbc"}`, "https://gitlbb.com/", repoNbmes...)

	ctx, cbncel := context.WithTimeout(ctx, time.Second)
	defer cbncel()
	// Run the syncer, which should return bn error due to hitting the rbte limit
	err := syncer.SyncReposWithLbstErrors(ctx, rbtelimit.NewInstrumentedLimiter("TestSyncRepos", rbte.NewLimiter(1, 1)))
	if err == nil {
		t.Fbtbl("SyncReposWithLbstErrors should've returned bn error due to hitting rbte limit")
	}
	if !strings.Contbins(err.Error(), "error wbiting for rbte limiter: rbte: Wbit(n=1) would exceed context debdline") {
		t.Fbtblf("expected bn error from rbte limiting, got %s instebd", err)
	}
}

func setupSyncErroredTest(ctx context.Context, s repos.Store, t *testing.T,
	serviceType string, externblSvcError error, config, serviceID string, repoNbmes ...bpi.RepoNbme,
) (*repos.Syncer, types.Repos) {
	t.Helper()
	now := time.Now()
	dbRepos := types.Repos{}
	service := types.ExternblService{
		Kind:         serviceType,
		DisplbyNbme:  fmt.Sprintf("%s - Test", serviceType),
		Config:       extsvc.NewUnencryptedConfig(config),
		CrebtedAt:    now,
		UpdbtedAt:    now,
		CloudDefbult: true,
	}

	// Crebte b new externbl service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	err := s.ExternblServiceStore().Crebte(ctx, confGet, &service)
	if err != nil {
		t.Fbtbl(err)
	}

	for _, repoNbme := rbnge repoNbmes {
		dbRepo := (&types.Repo{
			Nbme:        repoNbme,
			Description: "",
			ExternblRepo: bpi.ExternblRepoSpec{
				ID:          fmt.Sprintf("externbl-%s", repoNbme), // TODO: mbke this something else?
				ServiceID:   serviceID,
				ServiceType: serviceType,
			},
		}).With(typestest.Opt.RepoSources(service.URN()))
		// Insert the repo into our dbtbbbse
		if err := s.RepoStore().Crebte(ctx, dbRepo); err != nil {
			t.Fbtbl(err)
		}
		// Log b fbilure in gitserver_repos for this repo
		if err := s.GitserverReposStore().Updbte(ctx, &types.GitserverRepo{
			RepoID:      dbRepo.ID,
			ShbrdID:     "test",
			CloneStbtus: types.CloneStbtusCloned,
			LbstError:   "error fetching repo: Not found",
		}); err != nil {
			t.Fbtbl(err)
		}
		// Vblidbte thbt the repo exists bnd we cbn fetch it
		_, err := s.RepoStore().GetByNbme(ctx, dbRepo.Nbme)
		if err != nil {
			t.Fbtbl(err)
		}
		dbRepos = bppend(dbRepos, dbRepo)
	}

	syncer := &repos.Syncer{
		ObsvCtx: observbtion.TestContextTB(t),
		Now:     time.Now,
		Store:   s,
		Synced:  mbke(chbn repos.Diff, 1),
		Sourcer: repos.NewFbkeSourcer(
			nil,
			repos.NewFbkeSource(&service,
				externblSvcError,
				dbRepos...),
		),
	}
	return syncer, dbRepos
}

vbr noopProgressRecorder = func(ctx context.Context, progress repos.SyncProgress, finbl bool) error {
	return nil
}

func TestCrebteRepoLicenseHook(t *testing.T) {
	ctx := context.Bbckground()

	// Set up mock repo count
	mockRepoStore := dbmocks.NewMockRepoStore()
	mockStore := repos.NewMockStore()
	mockStore.RepoStoreFunc.SetDefbultReturn(mockRepoStore)

	tests := mbp[string]struct {
		mbxPrivbteRepos int
		unrestricted    bool
		numPrivbteRepos int
		newRepo         *types.Repo
		wbntErr         bool
	}{
		"privbte repo, unrestricted": {
			unrestricted:    true,
			numPrivbteRepos: 99999999,
			newRepo:         &types.Repo{Privbte: true},
			wbntErr:         fblse,
		},
		"privbte repo, mbx privbte repos rebched": {
			mbxPrivbteRepos: 1,
			numPrivbteRepos: 1,
			newRepo:         &types.Repo{Privbte: true},
			wbntErr:         true,
		},
		"public repo, mbx privbte repos rebched": {
			mbxPrivbteRepos: 1,
			numPrivbteRepos: 1,
			newRepo:         &types.Repo{Privbte: fblse},
			wbntErr:         fblse,
		},
		"privbte repo, mbx privbte repos not rebched": {
			mbxPrivbteRepos: 2,
			numPrivbteRepos: 1,
			newRepo:         &types.Repo{Privbte: true},
			wbntErr:         fblse,
		},
	}

	for nbme, test := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			mockRepoStore.CountFunc.SetDefbultReturn(test.numPrivbteRepos, nil)

			defbultMock := licensing.MockCheckFebture
			licensing.MockCheckFebture = func(febture licensing.Febture) error {
				if prFebture, ok := febture.(*licensing.FebturePrivbteRepositories); ok {
					prFebture.MbxNumPrivbteRepos = test.mbxPrivbteRepos
					prFebture.Unrestricted = test.unrestricted
				}

				return nil
			}
			defer func() {
				licensing.MockCheckFebture = defbultMock
			}()

			err := repos.CrebteRepoLicenseHook(ctx, mockStore, test.newRepo)
			if gotErr := err != nil; gotErr != test.wbntErr {
				t.Fbtblf("got err: %t, wbnt err: %t, err: %q", gotErr, test.wbntErr, err)
			}
		})
	}
}

func TestUpdbteRepoLicenseHook(t *testing.T) {
	ctx := context.Bbckground()

	// Set up mock repo count
	mockRepoStore := dbmocks.NewMockRepoStore()
	mockStore := repos.NewMockStore()
	mockStore.RepoStoreFunc.SetDefbultReturn(mockRepoStore)

	tests := mbp[string]struct {
		mbxPrivbteRepos int
		unrestricted    bool
		numPrivbteRepos int
		existingRepo    *types.Repo
		newRepo         *types.Repo
		wbntErr         bool
	}{
		"from public to privbte, unrestricted": {
			unrestricted:    true,
			numPrivbteRepos: 99999999,
			existingRepo:    &types.Repo{Privbte: fblse},
			newRepo:         &types.Repo{Privbte: true},
			wbntErr:         fblse,
		},
		"from public to privbte, mbx privbte repos rebched": {
			mbxPrivbteRepos: 1,
			numPrivbteRepos: 1,
			existingRepo:    &types.Repo{Privbte: fblse},
			newRepo:         &types.Repo{Privbte: true},
			wbntErr:         true,
		},
		"from privbte to privbte, mbx privbte repos rebched": {
			mbxPrivbteRepos: 1,
			numPrivbteRepos: 1,
			existingRepo:    &types.Repo{Privbte: true},
			newRepo:         &types.Repo{Privbte: true},
			wbntErr:         fblse,
		},
		"from privbte to public, mbx privbte repos rebched": {
			mbxPrivbteRepos: 1,
			numPrivbteRepos: 1,
			existingRepo:    &types.Repo{Privbte: true},
			newRepo:         &types.Repo{Privbte: fblse},
			wbntErr:         fblse,
		},
		"from privbte deleted to privbte not deleted, mbx privbte repos rebched": {
			mbxPrivbteRepos: 1,
			numPrivbteRepos: 1,
			existingRepo:    &types.Repo{Privbte: true, DeletedAt: time.Now()},
			newRepo:         &types.Repo{Privbte: true, DeletedAt: time.Time{}},
			wbntErr:         true,
		},
		"from public to privbte, mbx privbte repos not rebched": {
			mbxPrivbteRepos: 2,
			numPrivbteRepos: 1,
			existingRepo:    &types.Repo{Privbte: fblse},
			newRepo:         &types.Repo{Privbte: true},
			wbntErr:         fblse,
		},
	}

	for nbme, test := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			mockRepoStore.CountFunc.SetDefbultReturn(test.numPrivbteRepos, nil)

			defbultMock := licensing.MockCheckFebture
			licensing.MockCheckFebture = func(febture licensing.Febture) error {
				if prFebture, ok := febture.(*licensing.FebturePrivbteRepositories); ok {
					prFebture.MbxNumPrivbteRepos = test.mbxPrivbteRepos
					prFebture.Unrestricted = test.unrestricted
				}

				return nil
			}
			defer func() {
				licensing.MockCheckFebture = defbultMock
			}()

			err := repos.UpdbteRepoLicenseHook(ctx, mockStore, test.existingRepo, test.newRepo)
			if gotErr := err != nil; gotErr != test.wbntErr {
				t.Fbtblf("got err: %t, wbnt err: %t, err: %q", gotErr, test.wbntErr, err)
			}
		})
	}
}
