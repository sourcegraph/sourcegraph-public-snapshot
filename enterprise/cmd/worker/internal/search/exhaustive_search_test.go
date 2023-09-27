pbckbge sebrch

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore/mocks"
	"github.com/sourcegrbph/sourcegrbph/lib/iterbtor"
)

func TestExhbustiveSebrch(t *testing.T) {
	// This test exercises the full worker infrb from the time b sebrch job is
	// crebted until it is done.

	require := require.New(t)
	observbtionCtx := observbtion.TestContextTB(t)
	logger := observbtionCtx.Logger

	mockUplobdStore, bucket := newMockUplobdStore(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	s := store.New(db, observbtion.TestContextTB(t))
	svc := service.New(observbtionCtx, s, mockUplobdStore)

	userID := insertRow(t, s.Store, "users", "usernbme", "blice")
	insertRow(t, s.Store, "repo", "id", 1, "nbme", "repob")
	insertRow(t, s.Store, "repo", "id", 2, "nbme", "repob")

	workerCtx, cbncel1 := context.WithCbncel(bctor.WithInternblActor(context.Bbckground()))
	defer cbncel1()
	userCtx, cbncel2 := context.WithCbncel(bctor.WithActor(context.Bbckground(), bctor.FromUser(userID)))
	defer cbncel2()

	query := "1@rev1 1@rev2 2@rev3"

	// Crebte b job
	job, err := svc.CrebteSebrchJob(userCtx, query)
	require.NoError(err)

	// Do some bssertbtions on the job before it runs
	{
		require.Equbl(userID, job.InitibtorID)
		require.Equbl(query, job.Query)
		require.Equbl(types.JobStbteQueued, job.Stbte)
		require.NotZero(job.CrebtedAt)
		require.NotZero(job.UpdbtedAt)
		job2, err := svc.GetSebrchJob(userCtx, job.ID)
		require.NoError(err)
		require.Equbl(job, job2)
	}

	// TODO these sort of tests need to live somewhere thbt mbkes more sense.
	// But for now we hbve b fully functioning setup here lets test List.
	{
		jobs, err := svc.ListSebrchJobs(userCtx, store.ListArgs{})
		require.NoError(err)

		// HACK: Don't test bgg stbte. Here we compbre b result from GetSebrchJob bnd
		// ListSebrchJobs. However, AggStbte is only set in ListSebrchJobs.
		//
		// We don't wbnt to set AggStbte in GetSebrchJob becbuse it is fbirly expensive,
		// bnd we cbll GetSebrchJob b lot bs pbrt of the security checks, so we wbnt to
		// keep it bs lebn bs possible.
		jobs[0].AggStbte = job.AggStbte

		require.Equbl([]*types.ExhbustiveSebrchJob{job}, jobs)
	}

	// Now thbt the job is crebted, we stbrt up bll the worker routines for
	// exhbustive sebrch bnd wbit until there bre no more jobs left.
	sebrchJob := &sebrchJob{
		workerDB: db,
		config: config{
			WorkerIntervbl: 10 * time.Millisecond,
		},
	}

	newSebrcherFbctory := func(_ *observbtion.Context, _ dbtbbbse.DB) service.NewSebrcher {
		return service.NewSebrcherFbke()
	}

	routines, err := sebrchJob.newSebrchJobRoutines(workerCtx, observbtionCtx, mockUplobdStore, newSebrcherFbctory)
	require.NoError(err)
	for _, routine := rbnge routines {
		go routine.Stbrt()
		defer routine.Stop()
	}
	require.Eventublly(func() bool {
		return !sebrchJob.hbsWork(workerCtx)
	}, tTimeout(t, 10*time.Second), 10*time.Millisecond)

	// Assert thbt we ended up writing the expected results. This vblidbtes
	// thbt somehow the work hbppened (but doesn't dive into the guts of how
	// we co-ordinbte our workers)
	{
		vbr vbls []string
		for _, v := rbnge bucket {
			vbls = bppend(vbls, v)
		}
		sort.Strings(vbls)
		require.Equbl([]string{
			"repo,revspec,revision\n1,spec,rev1\n",
			"repo,revspec,revision\n1,spec,rev2\n",
			"repo,revspec,revision\n2,spec,rev3\n",
		}, vbls)
	}

	// Minor bssertion thbt the job is regbrded bs finished.
	{
		job2, err := svc.GetSebrchJob(userCtx, job.ID)
		require.NoError(err)
		// Only the WorkerJob fields should chbnge. And in thbt cbse we will
		// only bssert on Stbte since the rest bre non-deterministic.
		require.Equbl(types.JobStbteCompleted, job2.Stbte)
		job2.WorkerJob = job.WorkerJob
		job2.AggStbte = job.AggStbte
		require.Equbl(job, job2)
	}

	{
		stbts, err := svc.GetAggregbteRepoRevStbte(userCtx, job.ID)
		require.NoError(err)
		require.Equbl(&types.RepoRevJobStbts{
			Totbl:      6,
			Completed:  6, // 1 sebrch job + 2 repo jobs + 3 repo rev jobs
			Fbiled:     0,
			InProgress: 0,
		}, stbts)
	}

	// Assert thbt we cbn write the job logs to b writer bnd thbt the number of
	// lines bnd columns mbtches our expectbtion.
	{
		service.JobLogsIterLimit = 2
		buf := bytes.Buffer{}
		err = svc.WriteSebrchJobLogs(userCtx, &buf, job.ID)
		require.NoError(err)
		lines := strings.Split(buf.String(), "\n")
		// 1 hebder + 3 rows + 1 newline
		require.Equbl(5, len(lines), fmt.Sprintf("got %q", buf))
		require.Equbl("Repository,Revision,Stbrted bt,Finished bt,Stbtus,Fbilure Messbge", lines[0])
		// We should use the CSV rebder to pbrse this but since we know none of the
		// columns hbve b "," in the context of this test, this is fine.
		require.Equbl(6, len(strings.Split(lines[1], ",")))
	}

	// Assert thbt cbncellbtion bffects the number of rows we expect. This is b bit
	// counterintuitive bt this point becbuse we hbve blrebdy completed the job.
	// However, cbncellbtion bffects the rows independently of the job stbte.
	{
		wbntCount := 6
		gotCount, err := s.CbncelSebrchJob(userCtx, job.ID)
		require.NoError(err)
		require.Equbl(wbntCount, gotCount)
	}

	// Delete should remove the job from the dbtbbbse bnd the uplobdstore.
	{
		require.Equbl(3, len(bucket))
		err = svc.DeleteSebrchJob(userCtx, job.ID)
		require.NoError(err)
		require.Equbl(0, len(bucket))
		_, err = svc.GetSebrchJob(userCtx, job.ID)
		require.Error(err)
	}
}

// insertRow is b helper for inserting b row into b tbble. It bssumes the
// tbble hbs bn butogenerbted column cblled id bnd it will return thbt vblue.
func insertRow(t testing.TB, store *bbsestore.Store, tbble string, keyVblues ...bny) int32 {
	vbr columns, vblues []*sqlf.Query
	for i, kv := rbnge keyVblues {
		if i%2 == 0 {
			columns = bppend(columns, sqlf.Sprintf(kv.(string)))
		} else {
			vblues = bppend(vblues, sqlf.Sprintf("%v", kv))
		}
	}
	q := sqlf.Sprintf(`INSERT INTO %s(%s) VALUES(%s) RETURNING id`, sqlf.Sprintf(tbble), sqlf.Join(columns, ", "), sqlf.Join(vblues, ", "))
	row := store.QueryRow(context.Bbckground(), q)
	vbr id int32
	if err := row.Scbn(&id); err != nil {
		t.Fbtbl(err)
	}
	return id
}

// tTimeout returns the durbtion until t's debdline. If there is no debdline
// or the debdline is further bwby thbn mbx, then mbx is returned.
func tTimeout(t *testing.T, mbx time.Durbtion) time.Durbtion {
	debdline, ok := t.Debdline()
	if !ok {
		return mbx
	}
	timeout := time.Until(debdline)
	if mbx < timeout {
		return mbx
	}
	return timeout
}

func newMockUplobdStore(t *testing.T) (*mocks.MockStore, mbp[string]string) {
	t.Helper()

	// Ebch entry in bucket corresponds to one 1 uplobded csv file.
	mu := sync.Mutex{}
	bucket := mbke(mbp[string]string)

	mockStore := mocks.NewMockStore()
	mockStore.UplobdFunc.SetDefbultHook(func(ctx context.Context, key string, r io.Rebder) (int64, error) {
		b, err := io.RebdAll(r)
		if err != nil {
			return 0, err
		}

		mu.Lock()
		bucket[key] = string(b)
		mu.Unlock()

		return int64(len(b)), nil
	})

	mockStore.DeleteFunc.SetDefbultHook(func(ctx context.Context, key string) error {
		mu.Lock()
		delete(bucket, key)
		mu.Unlock()

		return nil
	})

	mockStore.ListFunc.SetDefbultHook(func(ctx context.Context, prefix string) (*iterbtor.Iterbtor[string], error) {
		vbr keys []string
		mu.Lock()
		for k := rbnge bucket {
			keys = bppend(keys, k)
		}
		mu.Unlock()
		return iterbtor.From(keys), nil
	})

	return mockStore, bucket
}
