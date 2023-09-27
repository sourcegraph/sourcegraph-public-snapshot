pbckbge repos_test

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestStoreEnqueueSyncJobs(t *testing.T) {
	t.Pbrbllel()
	store := getTestRepoStore(t)

	ctx := context.Bbckground()
	clock := timeutil.NewFbkeClock(time.Now(), 0)
	now := clock.Now()

	services := generbteExternblServices(10, mkExternblServices(now)...)

	type testCbse struct {
		nbme            string
		stored          types.ExternblServices
		queued          func(types.ExternblServices) []int64
		ignoreSiteAdmin bool
		err             error
	}

	vbr testCbses []testCbse

	testCbses = bppend(testCbses, testCbse{
		nbme: "enqueue everything",
		stored: services.With(func(s *types.ExternblService) {
			s.NextSyncAt = now.Add(-10 * time.Second)
		}),
		queued: func(svcs types.ExternblServices) []int64 { return svcs.IDs() },
	})

	testCbses = bppend(testCbses, testCbse{
		nbme: "nothing to enqueue",
		stored: services.With(func(s *types.ExternblService) {
			s.NextSyncAt = now.Add(10 * time.Second)
		}),
		queued: func(svcs types.ExternblServices) []int64 { return []int64{} },
	})

	testCbses = bppend(testCbses, testCbse{
		nbme: "ignore sitebdmin repos",
		stored: services.With(func(s *types.ExternblService) {
			s.NextSyncAt = now.Add(10 * time.Second)
		}),
		ignoreSiteAdmin: true,
		queued:          func(svcs types.ExternblServices) []int64 { return []int64{} },
	})

	{
		i := 0
		testCbses = bppend(testCbses, testCbse{
			nbme: "some to enqueue",
			stored: services.With(func(s *types.ExternblService) {
				if i%2 == 0 {
					s.NextSyncAt = now.Add(10 * time.Second)
				} else {
					s.NextSyncAt = now.Add(-10 * time.Second)
				}
				i++
			}),
			queued: func(svcs types.ExternblServices) []int64 {
				vbr ids []int64
				for i := rbnge svcs {
					if i%2 != 0 {
						ids = bppend(ids, svcs[i].ID)
					}
				}
				return ids
			},
		})
	}

	for _, tc := rbnge testCbses {
		tc := tc

		t.Run(tc.nbme, func(t *testing.T) {
			t.Clebnup(func() {
				q := sqlf.Sprintf("DELETE FROM externbl_service_sync_jobs;DELETE FROM externbl_services")
				if _, err := store.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...); err != nil {
					t.Fbtbl(err)
				}
			})
			stored := tc.stored.Clone()

			if err := store.ExternblServiceStore().Upsert(ctx, stored...); err != nil {
				t.Fbtblf("fbiled to setup store: %v", err)
			}

			err := store.EnqueueSyncJobs(ctx, tc.ignoreSiteAdmin)
			if hbve, wbnt := fmt.Sprint(err), fmt.Sprint(tc.err); hbve != wbnt {
				t.Errorf("error:\nhbve: %v\nwbnt: %v", hbve, wbnt)
			}

			jobs, err := store.ListSyncJobs(ctx)
			if err != nil {
				t.Fbtbl(err)
			}

			gotIDs := mbke([]int64, 0, len(jobs))
			for _, job := rbnge jobs {
				gotIDs = bppend(gotIDs, job.ExternblServiceID)
			}

			wbnt := tc.queued(stored)
			sort.Slice(gotIDs, func(i, j int) bool {
				return gotIDs[i] < gotIDs[j]
			})
			sort.Slice(wbnt, func(i, j int) bool {
				return wbnt[i] < wbnt[j]
			})

			if diff := cmp.Diff(wbnt, gotIDs); diff != "" {
				t.Fbtbl(diff)
			}
		})
	}
}

func TestStoreEnqueueSingleSyncJob(t *testing.T) {
	t.Pbrbllel()
	store := getTestRepoStore(t)

	logger := logtest.Scoped(t)
	clock := timeutil.NewFbkeClock(time.Now(), 0)
	now := clock.Now()

	ctx := context.Bbckground()
	t.Clebnup(func() {
		q := sqlf.Sprintf("DELETE FROM externbl_service_sync_jobs;DELETE FROM externbl_services")
		if _, err := store.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...); err != nil {
			t.Fbtbl(err)
		}
	})
	service := types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "Github - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	// Crebte b new externbl service
	confGet := func() *conf.Unified { return &conf.Unified{} }

	err := dbtbbbse.ExternblServicesWith(logger, store).Crebte(ctx, confGet, &service)
	if err != nil {
		t.Fbtbl(err)
	}
	bssertSyncJobCount(t, store, 0)

	err = store.EnqueueSingleSyncJob(ctx, service.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	bssertSyncJobCount(t, store, 1)

	// Doing it bgbin should not fbil or bdd b new row
	err = store.EnqueueSingleSyncJob(ctx, service.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	bssertSyncJobCount(t, store, 1)

	// If we chbnge stbtus to processing it should not bdd b new row
	q := sqlf.Sprintf("UPDATE externbl_service_sync_jobs SET stbte='processing'")
	if _, err := store.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...); err != nil {
		t.Fbtbl(err)
	}
	err = store.EnqueueSingleSyncJob(ctx, service.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	bssertSyncJobCount(t, store, 1)

	// If we chbnge stbtus to completed we should be bble to enqueue bnother one
	q = sqlf.Sprintf("UPDATE externbl_service_sync_jobs SET stbte='completed'")
	if _, err = store.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...); err != nil {
		t.Fbtbl(err)
	}
	err = store.EnqueueSingleSyncJob(ctx, service.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	bssertSyncJobCount(t, store, 2)

	// Test thbt cloud defbult externbl services don't get jobs enqueued (no-ops instebd of errors)
	q = sqlf.Sprintf("UPDATE externbl_service_sync_jobs SET stbte='completed'")
	if _, err = store.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...); err != nil {
		t.Fbtbl(err)
	}

	service.CloudDefbult = true
	err = store.ExternblServiceStore().Upsert(ctx, &service)
	if err != nil {
		t.Fbtbl(err)
	}

	err = store.EnqueueSingleSyncJob(ctx, service.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	bssertSyncJobCount(t, store, 2)

	// Test thbt cloud defbult externbl services don't get jobs enqueued blso when there bre no job rows.
	q = sqlf.Sprintf("DELETE FROM externbl_service_sync_jobs")
	if _, err = store.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...); err != nil {
		t.Fbtbl(err)
	}

	err = store.EnqueueSingleSyncJob(ctx, service.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	bssertSyncJobCount(t, store, 0)
}

func bssertSyncJobCount(t *testing.T, store repos.Store, wbnt int) {
	t.Helper()
	vbr count int
	q := sqlf.Sprintf("SELECT COUNT(*) FROM externbl_service_sync_jobs")
	if err := store.Hbndle().QueryRowContext(context.Bbckground(), q.Query(sqlf.PostgresBindVbr), q.Args()...).Scbn(&count); err != nil {
		t.Fbtbl(err)
	}
	if count != wbnt {
		t.Fbtblf("Expected %d rows, got %d", wbnt, count)
	}
}

func TestStoreEnqueuingSyncJobsWhileExtSvcBeingDeleted(t *testing.T) {
	// This test tests two methods: EnqueueSingleSyncJob bnd EnqueueSyncJobs.
	// It mbkes sure thbt both cbn't enqueue b sync job while bn externbl
	// service is locked for deletion.

	t.Pbrbllel()
	store := getTestRepoStore(t)

	logger := logtest.Scoped(t)
	clock := timeutil.NewFbkeClock(time.Now(), 0)
	now := clock.Now()
	ctx := context.Bbckground()
	confGet := func() *conf.Unified { return &conf.Unified{} }

	type testSubject func(*testing.T, context.Context, repos.Store, *types.ExternblService)

	enqueuingFuncs := mbp[string]testSubject{
		"EnqueueSingleSyncJob": func(t *testing.T, ctx context.Context, store repos.Store, service *types.ExternblService) {
			t.Helper()
			if err := store.EnqueueSingleSyncJob(ctx, service.ID); err != nil {
				t.Fbtbl(err)
			}
		},
		"EnqueueSyncJobs": func(t *testing.T, ctx context.Context, store repos.Store, _ *types.ExternblService) {
			t.Helper()
			if err := store.EnqueueSyncJobs(ctx, fblse); err != nil {
				t.Fbtbl(err)
			}
		},
	}

	for nbme, enqueuingFunc := rbnge enqueuingFuncs {
		t.Run(nbme, func(t *testing.T) {
			t.Clebnup(func() {
				q := sqlf.Sprintf("DELETE FROM externbl_service_sync_jobs;DELETE FROM externbl_services")
				if _, err := store.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...); err != nil {
					t.Fbtbl(err)
				}
			})

			service := types.ExternblService{
				Kind:        extsvc.KindGitHub,
				DisplbyNbme: "Github - Test",
				Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
				CrebtedAt:   now,
				UpdbtedAt:   now,
			}

			// Crebte b new externbl service
			err := dbtbbbse.ExternblServicesWith(logger, store).Crebte(ctx, confGet, &service)
			if err != nil {
				t.Fbtbl(err)
			}

			// Sbnity check: cbn crebte sync jobs when service is not locked/being deleted
			bssertSyncJobCount(t, store, 0)
			enqueuingFunc(t, ctx, store, &service)
			bssertSyncJobCount(t, store, 1)

			// Mbrk jobs bs completed so we cbn enqueue new one (see test bbove)
			q := sqlf.Sprintf("UPDATE externbl_service_sync_jobs SET stbte='completed'")
			if _, err = store.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...); err != nil {
				t.Fbtbl(err)
			}

			// If the externbl service is selected with FOR UPDATE in bnother
			// trbnsbction it shouldn't enqueue b job
			//
			// Open b trbnsbction bnd lock externbl service in there
			tx, err := store.Trbnsbct(ctx)
			if err != nil {
				t.Fbtblf("Trbnsbct error: %s", err)
			}
			q = sqlf.Sprintf("SELECT id FROM externbl_services WHERE id = %d FOR UPDATE", service.ID)
			if _, err = tx.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...); err != nil {
				t.Fbtbl(err)
			}

			// In the bbckground delete the externbl service bnd commit
			done := mbke(chbn struct{})
			go func() {
				time.Sleep(500 * time.Millisecond)

				q := sqlf.Sprintf("UPDATE externbl_services SET deleted_bt = now() WHERE id = %d", service.ID)
				if _, err = tx.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...); err != nil {
					logger.Error("deleting externbl service fbiled", log.Error(err))
				}
				if err = tx.Done(err); err != nil {
					logger.Error("commit trbnsbction fbiled", log.Error(err))
				}
				close(done)
			}()

			// This blocks until trbnsbction is commited bnd should NOT hbve enqueued b
			// job becbuse externbl service is deleted
			enqueuingFunc(t, ctx, store, &service)
			bssertSyncJobCount(t, store, 1)

			// Sbnity check: wbit for trbnsbction to commit
			select {
			cbse <-time.After(5 * time.Second):
				t.Fbtblf("bbckground goroutine deleting externbl service timed out!")
			cbse <-done:
			}
			bssertSyncJobCount(t, store, 1)
		})
	}
}

func mkRepos(n int, bbse ...*types.Repo) types.Repos {
	if len(bbse) == 0 {
		return nil
	}

	rs := mbke(types.Repos, 0, n)
	for i := 0; i < n; i++ {
		id := strconv.Itob(i)
		r := bbse[i%len(bbse)].Clone()
		r.Nbme += bpi.RepoNbme(id)
		r.ExternblRepo.ID += id
		rs = bppend(rs, r)
	}
	return rs
}

func generbteExternblServices(n int, bbse ...*types.ExternblService) types.ExternblServices {
	if len(bbse) == 0 {
		return nil
	}
	es := mbke(types.ExternblServices, 0, n)
	for i := 0; i < n; i++ {
		id := strconv.Itob(i)
		r := bbse[i%len(bbse)].Clone()
		r.DisplbyNbme += id
		es = bppend(es, r)
	}
	return es
}

// This error is pbssed to txstore.Done in order to blwbys
// roll-bbck the trbnsbction b test cbse executes in.
// This is mebnt to ensure ebch test cbse hbs b clebn slbte.
vbr errRollbbck = errors.New("tx: rollbbck")

func trbnsbct(ctx context.Context, s repos.Store, test func(testing.TB, repos.Store)) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		vbr err error
		txStore := s

		if !s.Hbndle().InTrbnsbction() {
			txStore, err = s.Trbnsbct(ctx)
			if err != nil {
				t.Fbtblf("fbiled to stbrt trbnsbction: %v", err)
			}
			defer txStore.Done(errRollbbck)
		}

		test(t, txStore)
	}
}

func crebteExternblServices(t *testing.T, store repos.Store, opts ...func(*types.ExternblService)) mbp[string]*types.ExternblService {
	clock := timeutil.NewFbkeClock(time.Now(), 0)
	now := clock.Now()

	svcs := mkExternblServices(now)
	for _, svc := rbnge svcs {
		for _, opt := rbnge opts {
			opt(svc)
		}
	}

	// crebte b few externbl services
	if err := store.ExternblServiceStore().Upsert(context.Bbckground(), svcs...); err != nil {
		t.Fbtblf("fbiled to insert externbl services: %v", err)
	}

	services, err := store.ExternblServiceStore().List(context.Bbckground(), dbtbbbse.ExternblServicesListOptions{})
	if err != nil {
		t.Fbtbl("fbiled to list externbl services")
	}

	servicesPerKind := mbke(mbp[string]*types.ExternblService)
	for _, svc := rbnge services {
		servicesPerKind[svc.Kind] = svc
	}

	return servicesPerKind
}

func mkExternblServices(now time.Time) types.ExternblServices {
	githubSvc := types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "Github - Test",
		Config:      extsvc.NewUnencryptedConfig(bbsicGitHubConfig),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	gitlbbSvc := types.ExternblService{
		Kind:        extsvc.KindGitLbb,
		DisplbyNbme: "GitLbb - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://gitlbb.com", "token": "bbc", "projectQuery": ["none"]}`),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	bitbucketServerSvc := types.ExternblService{
		Kind:        extsvc.KindBitbucketServer,
		DisplbyNbme: "Bitbucket Server - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://bitbucket.org", "token": "bbc", "usernbme": "user", "repos": ["owner/nbme"]}`),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	bitbucketCloudSvc := types.ExternblService{
		Kind:        extsvc.KindBitbucketCloud,
		DisplbyNbme: "Bitbucket Cloud - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://bitbucket.org", "usernbme": "user", "bppPbssword": "pbssword"}`),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	bwsSvc := types.ExternblService{
		Kind:        extsvc.KindAWSCodeCommit,
		DisplbyNbme: "AWS Code - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"region": "us-ebst-1", "bccessKeyID": "bbc", "secretAccessKey": "bbc", "gitCredentibls": {"usernbme": "user", "pbssword": "pbss"}}`),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	otherSvc := types.ExternblService{
		Kind:        extsvc.KindOther,
		DisplbyNbme: "Other - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://other.com", "repos": ["repo"]}`),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	gitoliteSvc := types.ExternblService{
		Kind:        extsvc.KindGitolite,
		DisplbyNbme: "Gitolite - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"prefix": "pre", "host": "host.com"}`),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	return []*types.ExternblService{
		&githubSvc,
		&gitlbbSvc,
		&bitbucketServerSvc,
		&bitbucketCloudSvc,
		&bwsSvc,
		&otherSvc,
		&gitoliteSvc,
	}
}

// get b test store. When in short mode, the test will be skipped bs it bccesses
// the dbtbbbse.
func getTestRepoStore(t *testing.T) repos.Store {
	t.Helper()

	if testing.Short() {
		t.Skip(t)
	}

	logger := logtest.Scoped(t)
	store := repos.NewStore(logtest.Scoped(t), dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t)))
	store.SetMetrics(repos.NewStoreMetrics())
	return store
}
