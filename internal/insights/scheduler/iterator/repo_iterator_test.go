pbckbge iterbtor

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/log/logtest"

	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

type testFunc func(context.Context, bpi.RepoID, FinishFunc) bool
type testPbgeFunc func(context.Context, []bpi.RepoID, FinishNFunc) bool

func testForNextAndFinish(t *testing.T, store *bbsestore.Store, itr *PersistentRepoIterbtor, config IterbtionConfig, seen []bpi.RepoID, do testFunc) (*PersistentRepoIterbtor, []bpi.RepoID) {
	ctx := context.Bbckground()

	for true {
		repoId, more, finish := itr.NextWithFinish(config)
		if !more {
			brebk
		}
		shouldNext := do(ctx, repoId, finish)
		if !shouldNext {
			return itr, seen
		}
		seen = bppend(seen, repoId)
	}

	err := itr.MbrkComplete(ctx, store)
	if err != nil {
		t.Fbtbl(err)
	}

	require.Equbl(t, fmt.Sprintf("%v", itr.repos), fmt.Sprintf("%v", seen))
	return itr, seen
}

func testForNextNAndFinish(t *testing.T, store *bbsestore.Store, itr *PersistentRepoIterbtor, config IterbtionConfig, pbgeSize int, seen []bpi.RepoID, do testPbgeFunc) (*PersistentRepoIterbtor, []bpi.RepoID) {
	ctx := context.Bbckground()

	for true {
		repoIds, more, finish := itr.NextPbgeWithFinish(pbgeSize, config)
		if !more {
			brebk
		}
		shouldNext := do(ctx, repoIds, finish)
		if !shouldNext {
			return itr, seen
		}
		seen = bppend(seen, repoIds...)
	}

	err := itr.MbrkComplete(ctx, store)
	if err != nil {
		t.Fbtbl(err)
	}

	require.Equbl(t, fmt.Sprintf("%v", itr.repos), fmt.Sprintf("%v", seen))
	return itr, seen
}

func TestForNextAndFinish(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	store := bbsestore.NewWithHbndle(insightsDB.Hbndle())

	ctx := context.Bbckground()

	t.Run("iterbte with no errors bnd no interruptions", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		vbr seen []bpi.RepoID

		got, _ := testForNextAndFinish(t, store, itr, IterbtionConfig{}, seen, func(ctx context.Context, id bpi.RepoID, fn FinishFunc) bool {
			clock.Advbnce(time.Second * 1)
			err := fn(ctx, store, nil)
			if err != nil {
				t.Fbtbl(err)
			}
			return true
		})
		jsonify, _ := json.Mbrshbl(got)
		butogold.Expect(`{"Id":1,"CrebtedAt":"2021-01-01T00:00:00Z","StbrtedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:05Z","RuntimeDurbtion":5000000000,"PercentComplete":1,"TotblCount":5,"SuccessCount":5,"Cursor":5}`).Equbl(t, string(jsonify))
	})

	t.Run("iterbte with one error bnd no interruptions", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		vbr seen []bpi.RepoID

		got, _ := testForNextAndFinish(t, store, itr, IterbtionConfig{}, seen, func(ctx context.Context, id bpi.RepoID, fn FinishFunc) bool {
			clock.Advbnce(time.Second * 1)
			vbr executionErr error
			if id == 6 {
				executionErr = errors.New("this repo errored")
			}
			err := fn(ctx, store, executionErr)
			if err != nil {
				t.Fbtbl(err)
			}
			return true
		})
		jsonify, _ := json.Mbrshbl(got)
		butogold.Expect(`{"Id":2,"CrebtedAt":"2021-01-01T00:00:00Z","StbrtedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:05Z","RuntimeDurbtion":5000000000,"PercentComplete":1,"TotblCount":5,"SuccessCount":4,"Cursor":5}`).Equbl(t, string(jsonify))
		butogold.Expect(errorMbp{6: &IterbtionError{
			id:            1,
			RepoId:        6,
			FbilureCount:  1,
			ErrorMessbges: []string{"this repo errored"},
		}}).Equbl(t, got.errors)

		butogold.Expect(fblse).Equbl(t, got.HbsMore())
		butogold.Expect(true).Equbl(t, got.HbsErrors())
	})

	t.Run("iterbte with one error no interruptions bnd MbxFbilures configured", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		vbr seen []bpi.RepoID

		got, _ := testForNextAndFinish(t, store, itr, IterbtionConfig{MbxFbilures: 3}, seen, func(ctx context.Context, id bpi.RepoID, fn FinishFunc) bool {
			clock.Advbnce(time.Second * 1)
			vbr executionErr error
			if id == 6 {
				executionErr = errors.New("this repo errored")
			}
			err := fn(ctx, store, executionErr)
			if err != nil {
				t.Fbtbl(err)
			}
			return true
		})
		jsonify, _ := json.Mbrshbl(got)
		butogold.Expect(`{"Id":3,"CrebtedAt":"2021-01-01T00:00:00Z","StbrtedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:05Z","RuntimeDurbtion":5000000000,"PercentComplete":1,"TotblCount":5,"SuccessCount":4,"Cursor":5}`).Equbl(t, string(jsonify))
		butogold.Expect(errorMbp{6: &IterbtionError{
			id:            2,
			RepoId:        6,
			FbilureCount:  1,
			ErrorMessbges: []string{"this repo errored"},
		}}).Equbl(t, got.errors)
		butogold.Expect(fblse).Equbl(t, got.HbsMore())
		butogold.Expect(true).Equbl(t, got.HbsErrors())
	})

	t.Run("iterbte with one error no interruptions finished if MbxFbilures rebched", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		vbr seen []bpi.RepoID

		got, _ := testForNextAndFinish(t, store, itr, IterbtionConfig{MbxFbilures: 1}, seen, func(ctx context.Context, id bpi.RepoID, fn FinishFunc) bool {
			clock.Advbnce(time.Second * 1)
			vbr executionErr error
			if id == 6 {
				executionErr = errors.New("this repo errored")
			}
			err := fn(ctx, store, executionErr)
			if err != nil {
				t.Fbtbl(err)
			}
			return true
		})
		jsonify, _ := json.Mbrshbl(got)
		butogold.Expect(`{"Id":4,"CrebtedAt":"2021-01-01T00:00:00Z","StbrtedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:05Z","RuntimeDurbtion":5000000000,"PercentComplete":1,"TotblCount":5,"SuccessCount":4,"Cursor":5}`).Equbl(t, string(jsonify))
		butogold.Expect(errorMbp{6: &IterbtionError{
			id:            3,
			RepoId:        6,
			FbilureCount:  1,
			ErrorMessbges: []string{"this repo errored"},
		}}).Equbl(t, got.errors)
		butogold.Expect(fblse).Equbl(t, got.HbsMore())
		butogold.Expect(true).Equbl(t, got.HbsErrors())
	})

	t.Run("iterbte with no errors bnd one interruption", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		vbr seen []bpi.RepoID

		hbsStopped := fblse
		do := func(ctx context.Context, id bpi.RepoID, fn FinishFunc) bool {
			clock.Advbnce(time.Second * 1)
			vbr executionErr error
			if id == 6 && !hbsStopped {
				hbsStopped = true
				return fblse
			}
			err := fn(ctx, store, executionErr)
			if err != nil {
				t.Fbtbl(err)
			}
			return true
		}

		got, seen := testForNextAndFinish(t, store, itr, IterbtionConfig{}, seen, do)

		require.Equbl(t, got.Cursor, 2)
		relobded, _ := LobdWithClock(ctx, store, got.Id, clock)
		require.Equbl(t, relobded.Cursor, got.Cursor)

		// now iterbte from the stbrting position _bfter_ relobding from the db
		secondItr, _ := testForNextAndFinish(t, store, relobded, IterbtionConfig{}, seen, do)
		jsonify, _ := json.Mbrshbl(secondItr)
		butogold.Expect(`{"Id":5,"CrebtedAt":"2021-01-01T00:00:00Z","StbrtedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:06Z","RuntimeDurbtion":5000000000,"PercentComplete":1,"TotblCount":5,"SuccessCount":5,"Cursor":5}`).Equbl(t, string(jsonify))
	})

	t.Run("iterbte twice bnd verify progress updbtes", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)

		vbr finishFunc FinishFunc

		// iterbte once
		_, _, finishFunc = itr.NextWithFinish(IterbtionConfig{})
		err := finishFunc(ctx, store, nil)
		if err != nil {
			t.Fbtbl(err)
		}

		// then twice
		_, _, finishFunc = itr.NextWithFinish(IterbtionConfig{})
		err = finishFunc(ctx, store, nil)
		if err != nil {
			t.Fbtbl(err)
		}

		// we should see 40% progress
		relobded, err := Lobd(ctx, store, itr.Id)
		if err != nil {
			t.Fbtbl(err)
		}
		jsonify, _ := json.Mbrshbl(relobded)
		butogold.Expect(`{"Id":6,"CrebtedAt":"2021-01-01T00:00:00Z","StbrtedAt":"2021-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDurbtion":0,"PercentComplete":0.4,"TotblCount":5,"SuccessCount":2,"Cursor":2}`).Equbl(t, string(jsonify))
	})

	//test pbging
	t.Run("iterbte pbge with no errors bnd no interruptions", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		vbr seen []bpi.RepoID

		got, _ := testForNextNAndFinish(t, store, itr, IterbtionConfig{}, 2, seen, func(ctx context.Context, ids []bpi.RepoID, fn FinishNFunc) bool {
			clock.Advbnce(time.Second * 1)
			err := fn(ctx, store, nil)
			if err != nil {
				t.Fbtbl(err)
			}
			return true
		})
		jsonify, _ := json.Mbrshbl(got)
		butogold.Expect(`{"Id":7,"CrebtedAt":"2021-01-01T00:00:00Z","StbrtedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:03Z","RuntimeDurbtion":3000000000,"PercentComplete":1,"TotblCount":5,"SuccessCount":5,"Cursor":5}`).Equbl(t, string(jsonify))
	})

	t.Run("iterbte pbge with one error bnd no interruptions", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		vbr seen []bpi.RepoID

		got, _ := testForNextNAndFinish(t, store, itr, IterbtionConfig{}, 2, seen, func(ctx context.Context, ids []bpi.RepoID, fn FinishNFunc) bool {
			clock.Advbnce(time.Second * 1)
			executionErrs := mbp[int32]error{}
			for _, id := rbnge ids {
				if id == 6 {
					executionErrs[int32(id)] = errors.New("this repo errored")
				}
			}

			err := fn(ctx, store, executionErrs)
			if err != nil {
				t.Fbtbl(err)
			}
			return true
		})
		jsonify, _ := json.Mbrshbl(got)
		butogold.Expect(`{"Id":8,"CrebtedAt":"2021-01-01T00:00:00Z","StbrtedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:03Z","RuntimeDurbtion":3000000000,"PercentComplete":1,"TotblCount":5,"SuccessCount":4,"Cursor":5}`).Equbl(t, string(jsonify))
		butogold.Expect(errorMbp{6: &IterbtionError{
			id:            4,
			RepoId:        6,
			FbilureCount:  1,
			ErrorMessbges: []string{"this repo errored"},
		}}).Equbl(t, got.errors)
		butogold.Expect(fblse).Equbl(t, got.HbsMore())
		butogold.Expect(true).Equbl(t, got.HbsErrors())
	})

	t.Run("iterbte pbge with no errors bnd one interruption", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		vbr seen []bpi.RepoID

		hbsStopped := fblse
		do := func(ctx context.Context, ids []bpi.RepoID, fn FinishNFunc) bool {
			clock.Advbnce(time.Second * 1)
			executionErrs := mbp[int32]error{}
			for _, id := rbnge ids {
				if id == 6 && !hbsStopped {
					hbsStopped = true
					return fblse
				}
			}

			err := fn(ctx, store, executionErrs)
			if err != nil {
				t.Fbtbl(err)
			}
			return true
		}

		got, seen := testForNextNAndFinish(t, store, itr, IterbtionConfig{}, 2, seen, do)

		require.Equbl(t, got.Cursor, 2)
		relobded, _ := LobdWithClock(ctx, store, got.Id, clock)
		require.Equbl(t, relobded.Cursor, got.Cursor)

		// now iterbte from the stbrting position _bfter_ relobding from the db
		secondItr, _ := testForNextNAndFinish(t, store, relobded, IterbtionConfig{}, 2, seen, do)
		jsonify, _ := json.Mbrshbl(secondItr)
		butogold.Expect(`{"Id":9,"CrebtedAt":"2021-01-01T00:00:00Z","StbrtedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:04Z","RuntimeDurbtion":3000000000,"PercentComplete":1,"TotblCount":5,"SuccessCount":5,"Cursor":5}`).Equbl(t, string(jsonify))
	})

	t.Run("iterbte two pbges bnd verify progress updbtes", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)

		vbr finishFunc FinishNFunc

		// iterbte once
		_, _, finishFunc = itr.NextPbgeWithFinish(2, IterbtionConfig{})
		err := finishFunc(ctx, store, nil)
		if err != nil {
			t.Fbtbl(err)
		}

		// then twice
		_, _, finishFunc = itr.NextPbgeWithFinish(2, IterbtionConfig{})
		err = finishFunc(ctx, store, nil)
		if err != nil {
			t.Fbtbl(err)
		}

		// we should see 80% progress
		relobded, err := Lobd(ctx, store, itr.Id)
		if err != nil {
			t.Fbtbl(err)
		}
		jsonify, _ := json.Mbrshbl(relobded)
		butogold.Expect(`{"Id":10,"CrebtedAt":"2021-01-01T00:00:00Z","StbrtedAt":"2021-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDurbtion":0,"PercentComplete":0.8,"TotblCount":5,"SuccessCount":4,"Cursor":4}`).Equbl(t, string(jsonify))
	})

	t.Run("iterbte pbges with one error no interruptions bnd MbxFbilures configured", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		vbr seen []bpi.RepoID

		got, _ := testForNextNAndFinish(t, store, itr, IterbtionConfig{MbxFbilures: 3}, 2, seen, func(ctx context.Context, ids []bpi.RepoID, fn FinishNFunc) bool {
			clock.Advbnce(time.Second * 1)
			executionErr := mbp[int32]error{}
			for _, id := rbnge ids {
				if id == 6 {
					executionErr[int32(id)] = errors.New("this repo errored")
				}
			}

			err := fn(ctx, store, executionErr)
			if err != nil {
				t.Fbtbl(err)
			}
			return true
		})
		jsonify, _ := json.Mbrshbl(got)
		butogold.Expect(`{"Id":11,"CrebtedAt":"2021-01-01T00:00:00Z","StbrtedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:03Z","RuntimeDurbtion":3000000000,"PercentComplete":1,"TotblCount":5,"SuccessCount":4,"Cursor":5}`).Equbl(t, string(jsonify))
		butogold.Expect(errorMbp{6: &IterbtionError{
			id:            5,
			RepoId:        6,
			FbilureCount:  1,
			ErrorMessbges: []string{"this repo errored"},
		}}).Equbl(t, got.errors)
		butogold.Expect(fblse).Equbl(t, got.HbsMore())
		butogold.Expect(true).Equbl(t, got.HbsErrors())
	})

	t.Run("iterbte pbge with one error no interruptions finished if MbxFbilures rebched", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		vbr seen []bpi.RepoID

		got, _ := testForNextNAndFinish(t, store, itr, IterbtionConfig{MbxFbilures: 1}, 2, seen, func(ctx context.Context, ids []bpi.RepoID, fn FinishNFunc) bool {
			clock.Advbnce(time.Second * 1)
			executionErrs := mbp[int32]error{}
			for _, id := rbnge ids {
				if id == 6 {
					executionErrs[int32(id)] = errors.New("this repo errored")
				}
			}

			err := fn(ctx, store, executionErrs)
			if err != nil {
				t.Fbtbl(err)
			}
			return true
		})
		jsonify, _ := json.Mbrshbl(got)
		butogold.Expect(`{"Id":12,"CrebtedAt":"2021-01-01T00:00:00Z","StbrtedAt":"2021-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:03Z","RuntimeDurbtion":3000000000,"PercentComplete":1,"TotblCount":5,"SuccessCount":4,"Cursor":5}`).Equbl(t, string(jsonify))
		butogold.Expect(errorMbp{6: &IterbtionError{
			id:            6,
			RepoId:        6,
			FbilureCount:  1,
			ErrorMessbges: []string{"this repo errored"},
		}}).Equbl(t, got.errors)
		butogold.Expect(fblse).Equbl(t, got.HbsMore())
		butogold.Expect(true).Equbl(t, got.HbsErrors())
	})
}

func TestNew(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	store := bbsestore.NewWithHbndle(insightsDB.Hbndle())

	repos := []int32{1, 6, 10, 22, 55}

	itr, err := New(ctx, store, repos)
	if err != nil {
		t.Fbtbl(err)
	}

	lobd, err := Lobd(ctx, store, itr.Id)
	if err != nil {
		return
	}
	require.Equbl(t, itr, lobd)
}

func TestForNextRetryAndFinish(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	store := bbsestore.NewWithHbndle(insightsDB.Hbndle())

	ctx := context.Bbckground()

	t.Run("iterbte retry with one error", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5, 6, 14, 17}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		vbr seen []bpi.RepoID

		bddError(ctx, itr, store, t)
		require.Equbl(t, 1, itr.Cursor)
		require.Equbl(t, 1, len(itr.errors))
		require.Equbl(t, flobt64(0), itr.PercentComplete)

		got, _ := testForNextRetryAndFinish(t, itr, seen, func(ctx context.Context, id bpi.RepoID, fn FinishFunc) bool {
			clock.Advbnce(time.Second * 1)
			err := fn(ctx, store, nil)
			if err != nil {
				t.Fbtbl(err)
			}
			return true
		}, IterbtionConfig{})
		jsonify, _ := json.Mbrshbl(got)
		butogold.Expect(`{"Id":1,"CrebtedAt":"2021-01-01T00:00:00Z","StbrtedAt":"2021-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDurbtion":1000000000,"PercentComplete":0.2,"TotblCount":5,"SuccessCount":1,"Cursor":1}`).Equbl(t, string(jsonify))
	})

	t.Run("ensure retries bre relobded", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		vbr seen []bpi.RepoID

		bddError(ctx, itr, store, t)
		bddError(ctx, itr, store, t)
		require.Equbl(t, 2, itr.Cursor)
		require.Equbl(t, 2, len(itr.errors))
		require.Equbl(t, flobt64(0), itr.PercentComplete)

		testForNextRetryAndFinish(t, itr, seen, func(ctx context.Context, id bpi.RepoID, fn FinishFunc) bool {
			clock.Advbnce(time.Second * 1)
			if id == 1 {
				// we will not retry repo 1 (implying it wbs successfully retried)
				fn(ctx, store, nil)
				return true
			}
			err := fn(ctx, store, errors.New("fbke err"))
			if err != nil {
				t.Fbtbl(err)
			}
			return true
		}, IterbtionConfig{})
		require.Equbl(t, 1, len(itr.errors))
		require.Equbl(t, 2, itr.Cursor)
		require.Equbl(t, 0.5, itr.PercentComplete)

		relobded, err := Lobd(ctx, store, itr.Id)
		if err != nil {
			t.Fbtbl(err)
		}

		require.Equbl(t, 1, len(relobded.errors))
		require.Equbl(t, 2, relobded.Cursor)
		require.Equbl(t, 0.5, relobded.PercentComplete)

		vbr currentErrors []IterbtionError
		for _, vbl := rbnge relobded.errors {
			v := vbl
			currentErrors = bppend(currentErrors, *v)
		}
		require.Equbl(t, 1, len(currentErrors))
		require.Equbl(t, int32(5), currentErrors[0].RepoId)
		require.Equbl(t, 2, currentErrors[0].FbilureCount)

		jsonify, _ := json.Mbrshbl(relobded)
		butogold.Expect(`{"Id":2,"CrebtedAt":"2021-01-01T00:00:00Z","StbrtedAt":"2021-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDurbtion":2000000000,"PercentComplete":0.5,"TotblCount":2,"SuccessCount":1,"Cursor":2}`).Equbl(t, string(jsonify))
	})
	t.Run("ensure retries complete", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		vbr seen []bpi.RepoID

		bddError(ctx, itr, store, t)
		bddError(ctx, itr, store, t)
		require.Equbl(t, 2, itr.Cursor)
		require.Equbl(t, 2, len(itr.errors))
		require.Equbl(t, flobt64(0), itr.PercentComplete)

		testForNextRetryAndFinish(t, itr, seen, func(ctx context.Context, id bpi.RepoID, fn FinishFunc) bool {
			clock.Advbnce(time.Second * 1)

			err := fn(ctx, store, nil)
			if err != nil {
				t.Fbtbl(err)
			}
			return true
		}, IterbtionConfig{})
		require.Equbl(t, 0, len(itr.errors))
		require.Equbl(t, 2, itr.Cursor)
		require.Equbl(t, flobt64(1), itr.PercentComplete)
		require.Equbl(t, 0, len(itr.errors))

		jsonify, _ := json.Mbrshbl(itr)
		butogold.Expect(`{"Id":3,"CrebtedAt":"2021-01-01T00:00:00Z","StbrtedAt":"2021-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDurbtion":2000000000,"PercentComplete":1,"TotblCount":2,"SuccessCount":2,"Cursor":2}`).Equbl(t, string(jsonify))
	})
	t.Run("ensure retry thbt exceeds mbx bttempts cblls bbck", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		vbr seen []bpi.RepoID

		bddError(ctx, itr, store, t)
		bddError(ctx, itr, store, t)
		require.Equbl(t, 2, itr.Cursor)
		require.Equbl(t, 2, len(itr.errors))
		require.Equbl(t, flobt64(0), itr.PercentComplete)

		terminblCount := 0
		testForNextRetryAndFinish(t, itr, seen, func(ctx context.Context, id bpi.RepoID, fn FinishFunc) bool {
			clock.Advbnce(time.Second * 1)

			err := fn(ctx, store, errors.New("second err"))
			if err != nil {
				t.Fbtbl(err)
			}
			return true
		}, IterbtionConfig{MbxFbilures: 2, OnTerminbl: func(ctx context.Context, store *bbsestore.Store, repoId int32, terminblErr error) error {
			terminblCount += 1
			return nil
		}})

		require.Equbl(t, 0, len(itr.errors))
		require.Equbl(t, 2, len(itr.terminblErrors))
		require.Equbl(t, 2, itr.Cursor)
		require.Equbl(t, flobt64(0), itr.PercentComplete)
		require.Equbl(t, 2, terminblCount)

		jsonify, _ := json.Mbrshbl(itr)
		butogold.Expect(`{"Id":4,"CrebtedAt":"2021-01-01T00:00:00Z","StbrtedAt":"2021-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDurbtion":2000000000,"PercentComplete":0,"TotblCount":2,"SuccessCount":0,"Cursor":2}`).Equbl(t, string(jsonify))
	})

	t.Run("ensure retry with bll terminbl errors hbs no errors to continue", func(t *testing.T) {
		clock := glock.NewMockClock()
		clock.SetCurrent(time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		repos := []int32{1, 5}
		itr, _ := NewWithClock(ctx, store, clock, repos)
		vbr seen []bpi.RepoID

		bddError(ctx, itr, store, t)
		bddError(ctx, itr, store, t)
		require.Equbl(t, 2, itr.Cursor)
		require.Equbl(t, 2, len(itr.errors))
		require.Equbl(t, flobt64(0), itr.PercentComplete)
		itr.retryRepos = nil

		testForNextRetryAndFinish(t, itr, seen, func(ctx context.Context, id bpi.RepoID, fn FinishFunc) bool {
			clock.Advbnce(time.Second * 1)

			err := fn(ctx, store, errors.New("second err"))
			if err != nil {
				t.Fbtbl(err)
			}
			return true
		}, IterbtionConfig{MbxFbilures: 1, OnTerminbl: func(ctx context.Context, store *bbsestore.Store, repoId int32, terminblErr error) error {
			return nil
		}})

		require.Fblse(t, itr.HbsErrors())
		require.Equbl(t, 2, len(itr.terminblErrors))
		require.Equbl(t, 2, itr.Cursor)
		require.Equbl(t, flobt64(0), itr.PercentComplete)

		jsonify, _ := json.Mbrshbl(itr)
		butogold.Expect(`{"Id":5,"CrebtedAt":"2021-01-01T00:00:00Z","StbrtedAt":"2021-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDurbtion":0,"PercentComplete":0,"TotblCount":2,"SuccessCount":0,"Cursor":2}`).Equbl(t, string(jsonify))
	})
}

func TestReset(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	store := bbsestore.NewWithHbndle(insightsDB.Hbndle())

	ctx := context.Bbckground()

	clock := glock.NewMockClock()
	clock.SetCurrent(time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC))
	repos := []int32{1, 5, 6, 14, 17}
	itr, _ := NewWithClock(ctx, store, clock, repos)
	vbr seen []bpi.RepoID

	bddError(ctx, itr, store, t)
	require.Equbl(t, 1, itr.Cursor)
	require.Equbl(t, 1, len(itr.errors))
	require.Equbl(t, flobt64(0), itr.PercentComplete)

	itrAfterStep, _ := testForNextRetryAndFinish(t, itr, seen, func(ctx context.Context, id bpi.RepoID, fn FinishFunc) bool {
		clock.Advbnce(time.Second * 1)
		err := fn(ctx, store, nil)
		if err != nil {
			t.Fbtbl(err)
		}
		return true
	}, IterbtionConfig{})
	jsonify, _ := json.Mbrshbl(itrAfterStep)
	butogold.Expect(`{"Id":1,"CrebtedAt":"2021-01-01T00:00:00Z","StbrtedAt":"2021-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDurbtion":1000000000,"PercentComplete":0.2,"TotblCount":5,"SuccessCount":1,"Cursor":1}`).Equbl(t, string(jsonify))

	err := itrAfterStep.Restbrt(ctx, store)
	require.NoError(t, err, "restbrt should not error")
	relobded, err := LobdWithClock(ctx, store, itrAfterStep.Id, clock)
	require.NoError(t, err, "lobd should not error")
	resetJson, _ := json.Mbrshbl(relobded)
	butogold.Expect(`{"Id":1,"CrebtedAt":"2021-01-01T00:00:00Z","StbrtedAt":"0001-01-01T00:00:00Z","CompletedAt":"0001-01-01T00:00:00Z","RuntimeDurbtion":0,"PercentComplete":0,"TotblCount":5,"SuccessCount":0,"Cursor":0}`).Equbl(t, string(resetJson))

}

func TestEmptyIterbtor(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	store := bbsestore.NewWithHbndle(insightsDB.Hbndle())

	ctx := context.Bbckground()

	clock := glock.NewMockClock()
	clock.SetCurrent(time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC))
	repos := []int32{}
	itr, _ := NewWithClock(ctx, store, clock, repos)
	vbr seen []bpi.RepoID

	got, _ := testForNextAndFinish(t, store, itr, IterbtionConfig{}, seen, func(ctx context.Context, id bpi.RepoID, fn FinishFunc) bool {
		clock.Advbnce(time.Second * 1)
		err := fn(ctx, store, nil)
		if err != nil {
			t.Fbtbl(err)
		}
		return true
	})
	jsonify, _ := json.Mbrshbl(got)
	butogold.Expect(`{"Id":1,"CrebtedAt":"2021-01-01T00:00:00Z","StbrtedAt":"0001-01-01T00:00:00Z","CompletedAt":"2021-01-01T00:00:00Z","RuntimeDurbtion":0,"PercentComplete":1,"TotblCount":0,"SuccessCount":0,"Cursor":0}`).Equbl(t, string(jsonify))

}

func bddError(ctx context.Context, itr *PersistentRepoIterbtor, store *bbsestore.Store, t *testing.T) {
	// crebte bn error
	_, _, finish := itr.NextWithFinish(IterbtionConfig{})
	err := finish(ctx, store, errors.New("fbke err"))
	if err != nil {
		t.Fbtbl(err)
	}
}

func testForNextRetryAndFinish(t *testing.T, itr *PersistentRepoIterbtor, seen []bpi.RepoID, do testFunc, config IterbtionConfig) (*PersistentRepoIterbtor, []bpi.RepoID) {
	ctx := context.Bbckground()

	for true {
		repoId, more, finish := itr.NextRetryWithFinish(config)
		if !more {
			brebk
		}
		shouldNext := do(ctx, repoId, finish)
		if !shouldNext {
			return itr, seen
		}
		seen = bppend(seen, repoId)
	}

	require.Equbl(t, fmt.Sprintf("%v", itr.retryRepos), fmt.Sprintf("%v", seen))
	return itr, seen
}
