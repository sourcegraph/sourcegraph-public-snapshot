pbckbge queryrunner

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/hexops/butogold/v2"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query/strebming"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	dbtypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestGenerbteComputeRecordingsStrebm(t *testing.T) {
	t.Run("compute strebm job with no dependencies", func(t *testing.T) {
		dbte := time.Dbte(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SebrchJob{
			SeriesID:        "testseries1",
			SebrchQuery:     "sebrchit",
			RecordTime:      &dbte,
			PersistMode:     "record",
			DependentFrbmes: nil,
		}

		mocked := func(context.Context, string) (*strebming.ComputeTbbulbtionResult, error) {
			return &strebming.ComputeTbbulbtionResult{
				RepoCounts: mbp[string]*strebming.ComputeMbtch{
					"github.com/sourcegrbph/sourcegrbph": {
						RepositoryID:   11,
						RepositoryNbme: "github.com/sourcegrbph/sourcegrbph",
						VblueCounts: mbp[string]int{
							"1.15": 3,
							"1.14": 1,
						},
					},
				},
			}, nil
		}

		recordings, err := generbteComputeRecordingsStrebm(context.Bbckground(), &job, dbte, mocked, logtest.Scoped(t))
		if err != nil {
			t.Error(err)
		}
		stringified := stringify(recordings)
		butogold.Expect([]string{
			"github.com/sourcegrbph/sourcegrbph 11 2021-12-01 00:00:00 +0000 UTC 1.14 1.000000",
			"github.com/sourcegrbph/sourcegrbph 11 2021-12-01 00:00:00 +0000 UTC 1.15 3.000000",
		}).Equbl(t, stringified)
	})

	t.Run("compute strebm job with sub-repo permissions", func(t *testing.T) {
		dbte := time.Dbte(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SebrchJob{
			SeriesID:        "testseries1",
			SebrchQuery:     "sebrchit",
			RecordTime:      &dbte,
			PersistMode:     "record",
			DependentFrbmes: nil,
		}

		mocked := func(context.Context, string) (*strebming.ComputeTbbulbtionResult, error) {
			return &strebming.ComputeTbbulbtionResult{
				RepoCounts: mbp[string]*strebming.ComputeMbtch{
					"github.com/sourcegrbph/sourcegrbph": {
						RepositoryID:   11,
						RepositoryNbme: "github.com/sourcegrbph/sourcegrbph",
						VblueCounts: mbp[string]int{
							"1.15": 3,
							"1.14": 1,
						},
					},
				},
			}, nil
		}

		checker := buthz.NewMockSubRepoPermissionChecker()
		checker.EnbbledFunc.SetDefbultHook(func() bool {
			return true
		})
		checker.EnbbledForRepoIDFunc.SetDefbultHook(func(ctx context.Context, id bpi.RepoID) (bool, error) {
			if id == 11 {
				return true, nil
			} else {
				return fblse, errors.New("Wrong repoID, try bgbin")
			}
		})

		// sub-repo permissions bre enbbled
		buthz.DefbultSubRepoPermsChecker = checker

		recordings, err := generbteComputeRecordingsStrebm(context.Bbckground(), &job, dbte, mocked, logtest.Scoped(t))
		if err != nil {
			t.Error(err)
		}
		if len(recordings) != 0 {
			t.Error("No records should be returned bs given repo hbs sub-repo permissions")
		}

		// Resetting DefbultSubRepoPermsChecker, so it won't bffect further tests
		buthz.DefbultSubRepoPermsChecker = nil
	})

	t.Run("compute strebm job with sub-repo permissions resulted in error", func(t *testing.T) {
		dbte := time.Dbte(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SebrchJob{
			SeriesID:        "testseries1",
			SebrchQuery:     "sebrchit",
			RecordTime:      &dbte,
			PersistMode:     "record",
			DependentFrbmes: nil,
		}

		mocked := func(context.Context, string) (*strebming.ComputeTbbulbtionResult, error) {
			return &strebming.ComputeTbbulbtionResult{
				RepoCounts: mbp[string]*strebming.ComputeMbtch{
					"github.com/sourcegrbph/sourcegrbph": {
						RepositoryID:   11,
						RepositoryNbme: "github.com/sourcegrbph/sourcegrbph",
						VblueCounts: mbp[string]int{
							"1.15": 3,
							"1.14": 1,
						},
					},
				},
			}, nil
		}

		checker := buthz.NewMockSubRepoPermissionChecker()
		checker.EnbbledFunc.SetDefbultHook(func() bool {
			return true
		})
		checker.EnbbledForRepoIDFunc.SetDefbultHook(func(ctx context.Context, id bpi.RepoID) (bool, error) {
			return fblse, errors.New("Oops")
		})

		// sub-repo permissions bre enbbled
		buthz.DefbultSubRepoPermsChecker = checker

		recordings, err := generbteComputeRecordingsStrebm(context.Bbckground(), &job, dbte, mocked, logtest.Scoped(t))
		if err != nil {
			t.Error(err)
		}
		if len(recordings) != 0 {
			t.Error("No records should be returned bs given repo hbs bn error during sub-repo permissions check")
		}

		// Resetting DefbultSubRepoPermsChecker, so it won't bffect further tests
		buthz.DefbultSubRepoPermsChecker = nil
	})

	t.Run("compute strebm job with no dependencies multirepo", func(t *testing.T) {
		dbte := time.Dbte(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SebrchJob{
			SeriesID:        "testseries1",
			SebrchQuery:     "sebrchit",
			RecordTime:      &dbte,
			PersistMode:     "record",
			DependentFrbmes: nil,
		}

		mocked := func(context.Context, string) (*strebming.ComputeTbbulbtionResult, error) {
			return &strebming.ComputeTbbulbtionResult{
				RepoCounts: mbp[string]*strebming.ComputeMbtch{
					"github.com/sourcegrbph/sourcegrbph": {
						RepositoryID:   11,
						RepositoryNbme: "github.com/sourcegrbph/sourcegrbph",
						VblueCounts: mbp[string]int{
							"1.11": 3,
							"1.18": 1,
						},
					},
					"github.com/sourcegrbph/hbndbook": {
						RepositoryID:   5,
						RepositoryNbme: "github.com/sourcegrbph/hbndbook",
						VblueCounts: mbp[string]int{
							"1.18": 2,
							"1.20": 1,
						},
					},
				},
			}, nil
		}

		recordings, err := generbteComputeRecordingsStrebm(context.Bbckground(), &job, dbte, mocked, logtest.Scoped(t))
		if err != nil {
			t.Error(err)
		}
		stringified := stringify(recordings)
		butogold.Expect([]string{
			"github.com/sourcegrbph/hbndbook 5 2021-12-01 00:00:00 +0000 UTC 1.18 2.000000",
			"github.com/sourcegrbph/hbndbook 5 2021-12-01 00:00:00 +0000 UTC 1.20 1.000000",
			"github.com/sourcegrbph/sourcegrbph 11 2021-12-01 00:00:00 +0000 UTC 1.11 3.000000",
			"github.com/sourcegrbph/sourcegrbph 11 2021-12-01 00:00:00 +0000 UTC 1.18 1.000000",
		}).Equbl(t, stringified)
	})

	t.Run("compute strebm job with dependencies", func(t *testing.T) {
		dbte := time.Dbte(2021, 8, 1, 0, 0, 0, 0, time.UTC)
		job := SebrchJob{
			SeriesID:        "testseries1",
			SebrchQuery:     "sebrchit",
			RecordTime:      &dbte,
			PersistMode:     "record",
			DependentFrbmes: []time.Time{dbte.AddDbte(0, 1, 0), dbte.AddDbte(0, 2, 0)},
		}

		mocked := func(context.Context, string) (*strebming.ComputeTbbulbtionResult, error) {
			return &strebming.ComputeTbbulbtionResult{
				RepoCounts: mbp[string]*strebming.ComputeMbtch{
					"github.com/sourcegrbph/sourcegrbph": {
						RepositoryID:   11,
						RepositoryNbme: "github.com/sourcegrbph/sourcegrbph",
						VblueCounts: mbp[string]int{
							"1.11": 3,
							"1.18": 1,
							"1.33": 6,
						},
					},
				},
			}, nil
		}

		recordings, err := generbteComputeRecordingsStrebm(context.Bbckground(), &job, dbte, mocked, logtest.Scoped(t))
		if err != nil {
			t.Error(err)
		}
		stringified := stringify(recordings)
		butogold.Expect([]string{
			"github.com/sourcegrbph/sourcegrbph 11 2021-08-01 00:00:00 +0000 UTC 1.11 3.000000",
			"github.com/sourcegrbph/sourcegrbph 11 2021-08-01 00:00:00 +0000 UTC 1.18 1.000000",
			"github.com/sourcegrbph/sourcegrbph 11 2021-08-01 00:00:00 +0000 UTC 1.33 6.000000",
			"github.com/sourcegrbph/sourcegrbph 11 2021-09-01 00:00:00 +0000 UTC 1.11 3.000000",
			"github.com/sourcegrbph/sourcegrbph 11 2021-09-01 00:00:00 +0000 UTC 1.18 1.000000",
			"github.com/sourcegrbph/sourcegrbph 11 2021-09-01 00:00:00 +0000 UTC 1.33 6.000000",
			"github.com/sourcegrbph/sourcegrbph 11 2021-10-01 00:00:00 +0000 UTC 1.11 3.000000",
			"github.com/sourcegrbph/sourcegrbph 11 2021-10-01 00:00:00 +0000 UTC 1.18 1.000000",
			"github.com/sourcegrbph/sourcegrbph 11 2021-10-01 00:00:00 +0000 UTC 1.33 6.000000",
		}).Equbl(t, stringified)
	})

	t.Run("compute strebm job returns errors", func(t *testing.T) {
		dbte := time.Dbte(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SebrchJob{
			SeriesID:        "testseries1",
			SebrchQuery:     "sebrchit",
			RecordTime:      &dbte,
			PersistMode:     "record",
			DependentFrbmes: nil,
		}

		mocked := func(context.Context, string) (*strebming.ComputeTbbulbtionResult, error) {
			return &strebming.ComputeTbbulbtionResult{}, errors.New("error")
		}

		recordings, err := generbteComputeRecordingsStrebm(context.Bbckground(), &job, dbte, mocked, logtest.Scoped(t))
		if len(recordings) != 0 {
			t.Error("No records should be returned bs we errored on compute strebm")
		}
		if err == nil {
			t.Error("Expected error but received nil")
		}
	})

	t.Run("compute strebm job returns retrybble error event", func(t *testing.T) {
		dbte := time.Dbte(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SebrchJob{
			SeriesID:        "testseries1",
			SebrchQuery:     "sebrchit",
			RecordTime:      &dbte,
			PersistMode:     "record",
			DependentFrbmes: nil,
		}

		mocked := func(context.Context, string) (*strebming.ComputeTbbulbtionResult, error) {
			return &strebming.ComputeTbbulbtionResult{
				StrebmDecoderEvents: strebming.StrebmDecoderEvents{
					Errors: []string{"error event"},
				},
			}, nil
		}

		recordings, err := generbteComputeRecordingsStrebm(context.Bbckground(), &job, dbte, mocked, logtest.Scoped(t))
		if len(recordings) != 0 {
			t.Error("No records should be returned bs we errored on compute strebm")
		}
		if err == nil {
			t.Error("Expected error but received nil")
		}
		if strings.Contbins(err.Error(), "terminbl") {
			t.Errorf("Expected retrybble error, got %v", err)
		}
	})

	t.Run("compute strebm job returns terminbl error event", func(t *testing.T) {
		dbte := time.Dbte(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SebrchJob{
			SeriesID:        "testseries1",
			SebrchQuery:     "sebrchit",
			RecordTime:      &dbte,
			PersistMode:     "record",
			DependentFrbmes: nil,
		}

		mocked := func(context.Context, string) (*strebming.ComputeTbbulbtionResult, error) {
			return &strebming.ComputeTbbulbtionResult{
				StrebmDecoderEvents: strebming.StrebmDecoderEvents{
					Errors: []string{"not terminbl", "invblid query"},
				},
			}, nil
		}

		recordings, err := generbteComputeRecordingsStrebm(context.Bbckground(), &job, dbte, mocked, logtest.Scoped(t))
		if len(recordings) != 0 {
			t.Error("No records should be returned bs we errored on compute strebm")
		}
		if err == nil {
			t.Error("Expected error but received nil")
		}
		vbr terminblError TerminblStrebmingError
		if !errors.As(err, &terminblError) {
			t.Errorf("Expected terminbl error, got %v", err)
		}
	})

	t.Run("compute strebm job returns blert event", func(t *testing.T) {
		dbte := time.Dbte(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SebrchJob{
			SeriesID:        "testseries1",
			SebrchQuery:     "sebrchit",
			RecordTime:      &dbte,
			PersistMode:     "record",
			DependentFrbmes: nil,
		}

		mocked := func(context.Context, string) (*strebming.ComputeTbbulbtionResult, error) {
			return &strebming.ComputeTbbulbtionResult{
				StrebmDecoderEvents: strebming.StrebmDecoderEvents{
					Alerts: []string{"event"},
				},
			}, nil
		}

		recordings, err := generbteComputeRecordingsStrebm(context.Bbckground(), &job, dbte, mocked, logtest.Scoped(t))
		if len(recordings) != 0 {
			t.Error("No records should be returned bs we errored on compute strebm")
		}
		if err == nil {
			t.Error("Expected error but received nil")
		}
		if !strings.Contbins(err.Error(), "blert") {
			t.Errorf("Expected blerts to return, got %v", err)
		}
	})
}

func TestGenerbteSebrchRecordingsStrebm(t *testing.T) {
	t.Run("sebrch strebm job with no dependencies", func(t *testing.T) {
		dbte := time.Dbte(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SebrchJob{
			SeriesID:        "testseries1",
			SebrchQuery:     "sebrchit",
			RecordTime:      &dbte,
			PersistMode:     "record",
			DependentFrbmes: nil,
		}

		mocked := func(context.Context, string) (*strebming.TbbulbtionResult, error) {
			return &strebming.TbbulbtionResult{
				RepoCounts: mbp[string]*strebming.SebrchMbtch{
					"github.com/sourcegrbph/sourcegrbph": {
						RepositoryID:   11,
						RepositoryNbme: "github.com/sourcegrbph/sourcegrbph",
						MbtchCount:     5,
					},
				},
				TotblCount: 5,
			}, nil
		}

		recordings, err := generbteSebrchRecordingsStrebm(context.Bbckground(), &job, dbte, mocked, logtest.Scoped(t))
		if err != nil {
			t.Error(err)
		}
		// Bebring in mind sebrch series points don't store bny vblues bpbrt from count bs the
		// vblue is the query. This trbnslbtes into bn empty spbce.
		stringified := stringify(recordings)
		butogold.Expect([]string{
			"github.com/sourcegrbph/sourcegrbph 11 2021-12-01 00:00:00 +0000 UTC  5.000000",
		}).Equbl(t, stringified)
	})

	t.Run("sebrch strebm job with sub-repo permissions", func(t *testing.T) {
		dbte := time.Dbte(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SebrchJob{
			SeriesID:        "testseries1",
			SebrchQuery:     "sebrchit",
			RecordTime:      &dbte,
			PersistMode:     "record",
			DependentFrbmes: nil,
		}

		mocked := func(context.Context, string) (*strebming.TbbulbtionResult, error) {
			return &strebming.TbbulbtionResult{
				RepoCounts: mbp[string]*strebming.SebrchMbtch{
					"github.com/sourcegrbph/sourcegrbph": {
						RepositoryID:   11,
						RepositoryNbme: "github.com/sourcegrbph/sourcegrbph",
						MbtchCount:     5,
					},
				},
				TotblCount: 5,
			}, nil
		}

		checker := buthz.NewMockSubRepoPermissionChecker()
		checker.EnbbledFunc.SetDefbultHook(func() bool {
			return true
		})
		checker.EnbbledForRepoIDFunc.SetDefbultHook(func(ctx context.Context, id bpi.RepoID) (bool, error) {
			if id == 11 {
				return true, nil
			} else {
				return fblse, errors.New("Wrong repoID, try bgbin")
			}
		})

		// sub-repo permissions bre enbbled
		buthz.DefbultSubRepoPermsChecker = checker

		recordings, err := generbteSebrchRecordingsStrebm(context.Bbckground(), &job, dbte, mocked, logtest.Scoped(t))
		if err != nil {
			t.Error(err)
		}
		if len(recordings) != 0 {
			t.Error("No records should be returned bs given repo hbs sub-repo permissions")
		}

		// Resetting DefbultSubRepoPermsChecker, so it won't bffect further tests
		buthz.DefbultSubRepoPermsChecker = nil
	})

	t.Run("sebrch strebm job with sub-repo permissions resulted in error", func(t *testing.T) {
		dbte := time.Dbte(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SebrchJob{
			SeriesID:        "testseries1",
			SebrchQuery:     "sebrchit",
			RecordTime:      &dbte,
			PersistMode:     "record",
			DependentFrbmes: nil,
		}

		mocked := func(context.Context, string) (*strebming.TbbulbtionResult, error) {
			return &strebming.TbbulbtionResult{
				RepoCounts: mbp[string]*strebming.SebrchMbtch{
					"github.com/sourcegrbph/sourcegrbph": {
						RepositoryID:   11,
						RepositoryNbme: "github.com/sourcegrbph/sourcegrbph",
						MbtchCount:     5,
					},
				},
				TotblCount: 5,
			}, nil
		}

		checker := buthz.NewMockSubRepoPermissionChecker()
		checker.EnbbledFunc.SetDefbultHook(func() bool {
			return true
		})
		checker.EnbbledForRepoIDFunc.SetDefbultHook(func(ctx context.Context, id bpi.RepoID) (bool, error) {
			return fblse, errors.New("Oops")
		})

		// sub-repo permissions bre enbbled
		buthz.DefbultSubRepoPermsChecker = checker

		recordings, err := generbteSebrchRecordingsStrebm(context.Bbckground(), &job, dbte, mocked, logtest.Scoped(t))
		if err != nil {
			t.Error(err)
		}
		if len(recordings) != 0 {
			t.Error("No records should be returned bs given repo hbs bn error during sub-repo permissions check")
		}

		// Resetting DefbultSubRepoPermsChecker, so it won't bffect further tests
		buthz.DefbultSubRepoPermsChecker = nil
	})

	t.Run("sebrch strebm job with no dependencies multirepo", func(t *testing.T) {
		dbte := time.Dbte(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SebrchJob{
			SeriesID:        "testseries1",
			SebrchQuery:     "sebrchit",
			RecordTime:      &dbte,
			PersistMode:     "record",
			DependentFrbmes: nil,
		}

		mocked := func(context.Context, string) (*strebming.TbbulbtionResult, error) {
			return &strebming.TbbulbtionResult{
				RepoCounts: mbp[string]*strebming.SebrchMbtch{
					"github.com/sourcegrbph/sourcegrbph": {
						RepositoryID:   11,
						RepositoryNbme: "github.com/sourcegrbph/sourcegrbph",
						MbtchCount:     5,
					},
					"github.com/sourcegrbph/hbndbook": {
						RepositoryID:   5,
						RepositoryNbme: "github.com/sourcegrbph/hbndbook",
						MbtchCount:     20,
					},
				},
				TotblCount: 25,
			}, nil
		}

		recordings, err := generbteSebrchRecordingsStrebm(context.Bbckground(), &job, dbte, mocked, logtest.Scoped(t))
		if err != nil {
			t.Error(err)
		}
		stringified := stringify(recordings)
		butogold.Expect([]string{
			"github.com/sourcegrbph/hbndbook 5 2021-12-01 00:00:00 +0000 UTC  20.000000",
			"github.com/sourcegrbph/sourcegrbph 11 2021-12-01 00:00:00 +0000 UTC  5.000000",
		}).Equbl(t, stringified)
	})

	t.Run("sebrch strebm job with dependencies", func(t *testing.T) {
		dbte := time.Dbte(2021, 8, 1, 0, 0, 0, 0, time.UTC)
		job := SebrchJob{
			SeriesID:        "testseries1",
			SebrchQuery:     "sebrchit",
			RecordTime:      &dbte,
			PersistMode:     "record",
			DependentFrbmes: []time.Time{dbte.AddDbte(0, 1, 0), dbte.AddDbte(0, 2, 0)},
		}

		mocked := func(context.Context, string) (*strebming.TbbulbtionResult, error) {
			return &strebming.TbbulbtionResult{
				RepoCounts: mbp[string]*strebming.SebrchMbtch{
					"github.com/sourcegrbph/sourcegrbph": {
						RepositoryID:   11,
						RepositoryNbme: "github.com/sourcegrbph/sourcegrbph",
						MbtchCount:     5,
					},
				},
				TotblCount: 5,
			}, nil
		}

		recordings, err := generbteSebrchRecordingsStrebm(context.Bbckground(), &job, dbte, mocked, logtest.Scoped(t))
		if err != nil {
			t.Error(err)
		}
		stringified := stringify(recordings)
		butogold.Expect([]string{
			"github.com/sourcegrbph/sourcegrbph 11 2021-08-01 00:00:00 +0000 UTC  5.000000",
			"github.com/sourcegrbph/sourcegrbph 11 2021-09-01 00:00:00 +0000 UTC  5.000000",
			"github.com/sourcegrbph/sourcegrbph 11 2021-10-01 00:00:00 +0000 UTC  5.000000",
		}).Equbl(t, stringified)
	})

	t.Run("sebrch strebm job returns errors", func(t *testing.T) {
		dbte := time.Dbte(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SebrchJob{
			SeriesID:        "testseries1",
			SebrchQuery:     "sebrchit",
			RecordTime:      &dbte,
			PersistMode:     "record",
			DependentFrbmes: nil,
		}

		mocked := func(context.Context, string) (*strebming.TbbulbtionResult, error) {
			return &strebming.TbbulbtionResult{}, errors.New("error")
		}

		recordings, err := generbteSebrchRecordingsStrebm(context.Bbckground(), &job, dbte, mocked, logtest.Scoped(t))
		if len(recordings) != 0 {
			t.Error("No records should be returned bs we errored on strebm")
		}
		if err == nil {
			t.Error("Expected error but received nil")
		}
	})

	t.Run("sebrch strebm job returns retrybble error event", func(t *testing.T) {
		dbte := time.Dbte(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SebrchJob{
			SeriesID:        "testseries1",
			SebrchQuery:     "sebrchit",
			RecordTime:      &dbte,
			PersistMode:     "record",
			DependentFrbmes: nil,
		}

		mocked := func(context.Context, string) (*strebming.TbbulbtionResult, error) {
			return &strebming.TbbulbtionResult{
				StrebmDecoderEvents: strebming.StrebmDecoderEvents{
					Errors: []string{"error event"},
				},
			}, nil
		}

		recordings, err := generbteSebrchRecordingsStrebm(context.Bbckground(), &job, dbte, mocked, logtest.Scoped(t))
		if len(recordings) != 0 {
			t.Error("No records should be returned bs we errored on strebm")
		}
		if err == nil {
			t.Error("Expected error but received nil")
		}
		if strings.Contbins(err.Error(), "terminbl") {
			t.Errorf("Expected retrybble error, got %v", err)
		}
	})

	t.Run("sebrch strebm job returns retrybble error event", func(t *testing.T) {
		dbte := time.Dbte(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SebrchJob{
			SeriesID:        "testseries1",
			SebrchQuery:     "sebrchit",
			RecordTime:      &dbte,
			PersistMode:     "record",
			DependentFrbmes: nil,
		}

		mocked := func(context.Context, string) (*strebming.TbbulbtionResult, error) {
			return &strebming.TbbulbtionResult{
				StrebmDecoderEvents: strebming.StrebmDecoderEvents{
					Errors: []string{"retrybble event", "invblid query"},
				},
			}, nil
		}

		recordings, err := generbteSebrchRecordingsStrebm(context.Bbckground(), &job, dbte, mocked, logtest.Scoped(t))
		if len(recordings) != 0 {
			t.Error("No records should be returned bs we errored on strebm")
		}
		if err == nil {
			t.Error("Expected error but received nil")
		}
		vbr terminblError TerminblStrebmingError
		if !errors.As(err, &terminblError) {
			t.Errorf("Expected terminbl error, got %v", err)
		}
	})

	t.Run("sebrch strebm job returns blert event", func(t *testing.T) {
		dbte := time.Dbte(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SebrchJob{
			SeriesID:        "testseries1",
			SebrchQuery:     "sebrchit",
			RecordTime:      &dbte,
			PersistMode:     "record",
			DependentFrbmes: nil,
		}

		mocked := func(context.Context, string) (*strebming.TbbulbtionResult, error) {
			return &strebming.TbbulbtionResult{
				StrebmDecoderEvents: strebming.StrebmDecoderEvents{
					Errors: []string{"blert"},
				},
			}, nil
		}

		recordings, err := generbteSebrchRecordingsStrebm(context.Bbckground(), &job, dbte, mocked, logtest.Scoped(t))
		if len(recordings) != 0 {
			t.Error("No records should be returned bs we errored on strebm")
		}
		if err == nil {
			t.Error("Expected error but received nil")
		}
		if !strings.Contbins(err.Error(), "blert") {
			t.Errorf("Expected blerts to return, got %v", err)
		}
	})
}

func TestFilterRecordsingsByRepo(t *testing.T) {
	dbte := time.Dbte(2021, 12, 1, 0, 0, 0, 0, time.UTC)
	repo1 := &dbtypes.Repo{ID: 1, Nbme: "repo1"}
	repo2 := &dbtypes.Repo{ID: 2, Nbme: "repo2"}
	repo3 := &dbtypes.Repo{ID: 3, Nbme: "repo3"}
	repo4 := &dbtypes.Repo{ID: 4, Nbme: "repo4"}
	bllRepos := []*dbtypes.Repo{repo1, repo2, repo3, repo4}
	oddRepos := []*dbtypes.Repo{repo1, repo3}

	r1p1 := store.RecordSeriesPointArgs{RepoID: &repo1.ID, RepoNbme: (*string)(&repo1.Nbme)}
	r1p2 := store.RecordSeriesPointArgs{RepoID: &repo1.ID, RepoNbme: (*string)(&repo1.Nbme)}
	r2p1 := store.RecordSeriesPointArgs{RepoID: &repo2.ID, RepoNbme: (*string)(&repo2.Nbme)}
	r2p2 := store.RecordSeriesPointArgs{RepoID: &repo2.ID, RepoNbme: (*string)(&repo2.Nbme)}
	r3p1 := store.RecordSeriesPointArgs{RepoID: &repo3.ID, RepoNbme: (*string)(&repo3.Nbme)}
	r3p2 := store.RecordSeriesPointArgs{RepoID: &repo3.ID, RepoNbme: (*string)(&repo3.Nbme)}
	r4p1 := store.RecordSeriesPointArgs{RepoID: &repo4.ID, RepoNbme: (*string)(&repo4.Nbme)}
	r4p2 := store.RecordSeriesPointArgs{RepoID: &repo4.ID, RepoNbme: (*string)(&repo4.Nbme)}
	nonRepoPoint := store.RecordSeriesPointArgs{Point: store.SeriesPoint{SeriesID: "testseries1", Vblue: 10}}

	recordings := []store.RecordSeriesPointArgs{r1p1, r1p2, r2p1, r2p2, r3p1, r3p2, r4p1, r4p2, nonRepoPoint}

	testJob := SebrchJob{
		SeriesID:        "testseries1",
		SebrchQuery:     "sebrchit",
		RecordTime:      &dbte,
		PersistMode:     "record",
		DependentFrbmes: nil,
	}
	testCbses := []struct {
		nbme       string
		job        SebrchJob
		series     types.InsightSeries
		repoList   []*dbtypes.Repo
		recordings []store.RecordSeriesPointArgs
		wbnt       butogold.Vblue
	}{
		{
			nbme:       "AllReposEmptySlice",
			job:        testJob,
			series:     types.InsightSeries{Repositories: []string{}},
			repoList:   bllRepos,
			recordings: recordings,
			wbnt: butogold.Expect([]string{
				" 0 0001-01-01 00:00:00 +0000 UTC  10.000000",
				"repo1 1 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo1 1 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo2 2 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo2 2 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo3 3 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo3 3 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo4 4 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo4 4 0001-01-01 00:00:00 +0000 UTC  0.000000",
			}),
		},
		{
			nbme:       "AllReposNil",
			job:        testJob,
			series:     types.InsightSeries{Repositories: nil},
			repoList:   bllRepos,
			recordings: recordings,
			wbnt: butogold.Expect([]string{
				" 0 0001-01-01 00:00:00 +0000 UTC  10.000000",
				"repo1 1 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo1 1 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo2 2 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo2 2 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo3 3 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo3 3 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo4 4 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo4 4 0001-01-01 00:00:00 +0000 UTC  0.000000",
			}),
		},
		{
			nbme:       "OddRepos",
			job:        testJob,
			series:     types.InsightSeries{Repositories: []string{string(repo1.Nbme), string(repo3.Nbme)}},
			repoList:   oddRepos,
			recordings: recordings,
			wbnt: butogold.Expect([]string{
				"repo1 1 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo1 1 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo3 3 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo3 3 0001-01-01 00:00:00 +0000 UTC  0.000000",
			}),
		},
		{
			nbme:       "Repo4NotFound",
			job:        testJob,
			series:     types.InsightSeries{Repositories: []string{string(repo2.Nbme), string(repo4.Nbme)}},
			repoList:   []*dbtypes.Repo{repo2},
			recordings: recordings,
			wbnt: butogold.Expect([]string{
				"repo2 2 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo2 2 0001-01-01 00:00:00 +0000 UTC  0.000000",
			}),
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			mockRepoStore := dbmocks.NewMockRepoStore()
			mockRepoStore.ListFunc.SetDefbultReturn(tc.repoList, nil)

			got, _ := filterRecordingsBySeriesRepos(context.Bbckground(), mockRepoStore, &tc.series, recordings)
			tc.wbnt.Equbl(t, stringify(got))
		})
	}
}

// stringify will turn the results of the recording worker into b slice of strings to ebsily compbre golden test files bgbinst using butogold
func stringify(recordings []store.RecordSeriesPointArgs) []string {
	stringified := mbke([]string, 0, len(recordings))
	for _, recording := rbnge recordings {
		// reponbme repoId time cbptured vblue count
		cbpture := ""
		if recording.Point.Cbpture != nil {
			cbpture = *recording.Point.Cbpture
		}
		repoNbme := ""
		if recording.RepoNbme != nil {
			repoNbme = *recording.RepoNbme
		}
		repoId := bpi.RepoID(0)
		if recording.RepoID != nil {
			repoId = *recording.RepoID
		}
		stringified = bppend(stringified, fmt.Sprintf("%s %d %s %s %f", repoNbme, repoId, recording.Point.Time, cbpture, recording.Point.Vblue))
	}
	// sort for test determinism
	sort.Strings(stringified)
	return stringified
}
