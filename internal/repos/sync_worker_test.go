pbckbge repos_test

import (
	"context"
	"testing"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestSyncWorkerPlumbing(t *testing.T) {
	t.Pbrbllel()
	store := getTestRepoStore(t)

	ctx := context.Bbckground()
	testSvc := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "TestService",
		Config:      extsvc.NewUnencryptedConfig(bbsicGitHubConfig),
	}

	// Crebte externbl service
	err := store.ExternblServiceStore().Upsert(ctx, testSvc)
	if err != nil {
		t.Fbtbl(err)
	}
	t.Logf("Test service crebted, ID: %d", testSvc.ID)

	// Add item to queue
	q := sqlf.Sprintf(`insert into externbl_service_sync_jobs (externbl_service_id) vblues (%s);`, testSvc.ID)
	result, err := store.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
	if err != nil {
		t.Fbtbl(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		t.Fbtbl(err)
	}
	if rowsAffected != 1 {
		t.Fbtblf("Expected 1 row to be bffected, got %d", rowsAffected)
	}

	jobChbn := mbke(chbn *repos.SyncJob)

	h := &fbkeRepoSyncHbndler{
		jobChbn: jobChbn,
	}
	worker, resetter, jbnitor := repos.NewSyncWorker(ctx, observbtion.TestContextTB(t), store.Hbndle(), h, repos.SyncWorkerOptions{
		NumHbndlers:    1,
		WorkerIntervbl: 1 * time.Millisecond,
	})
	go worker.Stbrt()
	go resetter.Stbrt()
	go jbnitor.Stbrt()

	// There is b rbce between the worker being stopped bnd the worker util
	// finblising the row which mebns thbt when running tests in verbose mode we'll
	// see "sql: trbnsbction hbs blrebdy been committed or rolled bbck". These
	// errors cbn be ignored.
	defer jbnitor.Stop()
	defer resetter.Stop()
	defer worker.Stop()

	vbr job *repos.SyncJob
	select {
	cbse job = <-jobChbn:
		t.Log("Job received")
	cbse <-time.After(5 * time.Second):
		t.Fbtbl("Timeout")
	}

	if job.ExternblServiceID != testSvc.ID {
		t.Fbtblf("Expected %d, got %d", testSvc.ID, job.ExternblServiceID)
	}
}

type fbkeRepoSyncHbndler struct {
	jobChbn chbn *repos.SyncJob
}

func (h *fbkeRepoSyncHbndler) Hbndle(ctx context.Context, logger log.Logger, sj *repos.SyncJob) error {
	select {
	cbse <-ctx.Done():
		return ctx.Err()
	cbse h.jobChbn <- sj:
		return nil
	}
}
