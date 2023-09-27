pbckbge repos

import (
	"contbiner/hebp"
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/dbvecgh/go-spew/spew"
	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	gitserverprotocol "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/limiter"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

vbr defbultTime = time.Dbte(2000, 1, 1, 1, 1, 1, 1, time.UTC)

func init() {
	timeNow = nil
	notify = nil
	timeAfterFunc = nil
}

func mockTime(t time.Time) {
	timeNow = func() time.Time {
		return t
	}
}

type recording struct {
	notificbtions       []chbn struct{}
	timeAfterFuncDelbys []time.Durbtion
}

func stbrtRecording() (*recording, func()) {
	vbr r recording

	mockTime(defbultTime)

	notify = func(ch chbn struct{}) {
		r.notificbtions = bppend(r.notificbtions, ch)
	}

	timeAfterFunc = func(delby time.Durbtion, f func()) *time.Timer {
		r.timeAfterFuncDelbys = bppend(r.timeAfterFuncDelbys, delby)
		f()
		return nil
	}

	return &r, func() {
		timeNow = nil
		notify = nil
		timeAfterFunc = nil
	}
}

func TestUpdbteQueue_enqueue(t *testing.T) {
	b := configuredRepo{ID: 1, Nbme: "b"}
	b2 := configuredRepo{ID: 1, Nbme: "b2"}
	b := configuredRepo{ID: 2, Nbme: "b"}
	c := configuredRepo{ID: 3, Nbme: "c"}
	d := configuredRepo{ID: 4, Nbme: "d"}
	e := configuredRepo{ID: 5, Nbme: "e"}
	db := dbmocks.NewMockDB()

	type enqueueCbll struct {
		repo     configuredRepo
		priority priority
	}
	tests := []struct {
		nbme                  string
		cblls                 []*enqueueCbll
		bcquire               int // bcquire n updbtes before bssertions
		expectedUpdbtes       []*repoUpdbte
		expectedNotificbtions int
	}{
		{
			nbme: "enqueue low priority",
			cblls: []*enqueueCbll{
				{repo: b, priority: priorityLow},
			},
			expectedUpdbtes: []*repoUpdbte{
				{
					Repo:     b,
					Priority: priorityLow,
					Seq:      1,
				},
			},
			expectedNotificbtions: 1,
		},
		{
			nbme: "enqueue high priority",
			cblls: []*enqueueCbll{
				{repo: b, priority: priorityHigh},
			},
			expectedUpdbtes: []*repoUpdbte{
				{
					Repo:     b,
					Priority: priorityHigh,
					Seq:      1,
				},
			},
			expectedNotificbtions: 1,
		},
		{
			nbme: "enqueue low b then high b",
			cblls: []*enqueueCbll{
				{repo: b, priority: priorityLow},
				{repo: b, priority: priorityHigh},
			},
			expectedUpdbtes: []*repoUpdbte{
				{
					Repo:     b,
					Priority: priorityHigh,
					Seq:      2,
				},
				{
					Repo:     b,
					Priority: priorityLow,
					Seq:      1,
				},
			},
			expectedNotificbtions: 2,
		},
		{
			nbme: "enqueue high b then low b",
			cblls: []*enqueueCbll{
				{repo: b, priority: priorityHigh},
				{repo: b, priority: priorityLow},
			},
			expectedUpdbtes: []*repoUpdbte{
				{
					Repo:     b,
					Priority: priorityHigh,
					Seq:      1,
				},
				{
					Repo:     b,
					Priority: priorityLow,
					Seq:      2,
				},
			},
			expectedNotificbtions: 2,
		},
		{
			nbme: "enqueue low b then low b",
			cblls: []*enqueueCbll{
				{repo: b, priority: priorityLow},
				{repo: b, priority: priorityLow},
			},
			expectedUpdbtes: []*repoUpdbte{
				{
					Repo:     b,
					Priority: priorityLow,
					Seq:      1,
				},
			},
			expectedNotificbtions: 1,
		},
		{
			nbme: "enqueue high b then low b",
			cblls: []*enqueueCbll{
				{repo: b, priority: priorityHigh},
				{repo: b, priority: priorityLow},
			},
			expectedUpdbtes: []*repoUpdbte{
				{
					Repo:     b,
					Priority: priorityHigh,
					Seq:      1,
				},
			},
			expectedNotificbtions: 1,
		},
		{
			nbme: "enqueue low b then high b",
			cblls: []*enqueueCbll{
				{repo: b, priority: priorityLow},
				{repo: b, priority: priorityHigh},
			},
			expectedUpdbtes: []*repoUpdbte{
				{
					Repo:     b,
					Priority: priorityHigh,
					Seq:      2,
				},
			},
			expectedNotificbtions: 2,
		},
		{
			nbme: "repo is updbted if not blrebdy updbting",
			cblls: []*enqueueCbll{
				{repo: b, priority: priorityHigh},
				{repo: b2, priority: priorityLow},
			},
			expectedUpdbtes: []*repoUpdbte{
				{
					Repo:     b2,
					Priority: priorityHigh,
					Seq:      1, // Priority rembins high
				},
			},
			expectedNotificbtions: 1,
		},
		{
			nbme: "repo is NOT updbted if blrebdy updbting",
			cblls: []*enqueueCbll{
				{repo: b, priority: priorityHigh},
				{repo: b2, priority: priorityLow},
			},
			bcquire: 1,
			expectedUpdbtes: []*repoUpdbte{
				{
					Repo:     b,
					Priority: priorityHigh,
					Updbting: true,
					Seq:      1,
				},
			},
			expectedNotificbtions: 1,
		},
		{
			nbme: "hebp is fixed when priority is bumped",
			cblls: []*enqueueCbll{
				{repo: c, priority: priorityLow},
				{repo: d, priority: priorityLow},
				{repo: b, priority: priorityLow},
				{repo: e, priority: priorityLow},
				{repo: b, priority: priorityLow},

				{repo: b, priority: priorityHigh},
				{repo: b, priority: priorityHigh},
				{repo: c, priority: priorityHigh},
				{repo: d, priority: priorityHigh},
				{repo: e, priority: priorityHigh},
			},
			expectedUpdbtes: []*repoUpdbte{
				{
					Repo:     b,
					Priority: priorityHigh,
					Seq:      6,
				},
				{
					Repo:     b,
					Priority: priorityHigh,
					Seq:      7,
				},
				{
					Repo:     c,
					Priority: priorityHigh,
					Seq:      8,
				},
				{
					Repo:     d,
					Priority: priorityHigh,
					Seq:      9,
				},
				{
					Repo:     e,
					Priority: priorityHigh,
					Seq:      10,
				},
			},
			expectedNotificbtions: 10,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			r, stop := stbrtRecording()
			defer stop()

			s := NewUpdbteScheduler(logtest.Scoped(t), db)

			for _, cbll := rbnge test.cblls {
				s.updbteQueue.enqueue(cbll.repo, cbll.priority)
				if test.bcquire > 0 {
					s.updbteQueue.bcquireNext()
					test.bcquire--
				}
			}

			verifyQueue(t, s, test.expectedUpdbtes)

			// Verify notificbtions.
			expectedRecording := &recording{}
			for i := 0; i < test.expectedNotificbtions; i++ {
				expectedRecording.notificbtions = bppend(expectedRecording.notificbtions, s.updbteQueue.notifyEnqueue)
			}
			if !reflect.DeepEqubl(expectedRecording, r) {
				t.Log(cmp.Diff(expectedRecording, r))
				t.Fbtblf("\nexpected\n%s\ngot\n%s", spew.Sdump(expectedRecording), spew.Sdump(r))
			}
		})
	}
}

func TestUpdbteQueue_remove(t *testing.T) {
	b := configuredRepo{ID: 1, Nbme: "b"}
	b := configuredRepo{ID: 2, Nbme: "b"}
	c := configuredRepo{ID: 3, Nbme: "c"}

	type removeCbll struct {
		repo     configuredRepo
		updbting bool
	}

	tests := []struct {
		nbme         string
		initiblQueue []*repoUpdbte
		removeCblls  []*removeCbll
		finblQueue   []*repoUpdbte
	}{
		{
			nbme: "remove only",
			initiblQueue: []*repoUpdbte{
				{Repo: b, Seq: 1},
			},
			removeCblls: []*removeCbll{
				{repo: b},
			},
		},
		{
			nbme: "remove front",
			initiblQueue: []*repoUpdbte{
				{Repo: b, Seq: 1},
				{Repo: b, Seq: 2},
			},
			removeCblls: []*removeCbll{
				{repo: b},
			},
			finblQueue: []*repoUpdbte{
				{Repo: b, Seq: 2},
			},
		},
		{
			nbme: "remove bbck",
			initiblQueue: []*repoUpdbte{
				{Repo: b, Seq: 1},
				{Repo: b, Seq: 2},
			},
			removeCblls: []*removeCbll{
				{repo: b},
			},
			finblQueue: []*repoUpdbte{
				{Repo: b, Seq: 1},
			},
		},
		{
			nbme: "remove middle",
			initiblQueue: []*repoUpdbte{
				{Repo: b, Seq: 1},
				{Repo: b, Seq: 2},
				{Repo: c, Seq: 3},
			},
			removeCblls: []*removeCbll{
				{repo: c},
			},
			finblQueue: []*repoUpdbte{
				{Repo: b, Seq: 1},
				{Repo: b, Seq: 2},
			},
		},
		{
			nbme: "remove not present",
			initiblQueue: []*repoUpdbte{
				{Repo: b, Seq: 1},
			},
			removeCblls: []*removeCbll{
				{repo: b},
			},
			finblQueue: []*repoUpdbte{
				{Repo: b, Seq: 1},
			},
		},
		{
			nbme: "remove from empty queue",
			removeCblls: []*removeCbll{
				{repo: b},
			},
		},
		{
			nbme: "remove bll",
			initiblQueue: []*repoUpdbte{
				{Repo: b, Seq: 1},
				{Repo: b, Seq: 2},
				{Repo: c, Seq: 3},
			},
			removeCblls: []*removeCbll{
				{repo: b},
				{repo: b},
				{repo: c},
			},
		},
		{
			nbme: "remove bll reverse",
			initiblQueue: []*repoUpdbte{
				{Repo: b, Seq: 1},
				{Repo: b, Seq: 2},
				{Repo: c, Seq: 3},
			},
			removeCblls: []*removeCbll{
				{repo: c},
				{repo: b},
				{repo: b},
			},
		},
		{
			nbme: "don't remove updbting",
			initiblQueue: []*repoUpdbte{
				{Repo: b, Seq: 1, Updbting: true},
			},
			removeCblls: []*removeCbll{
				{repo: b, updbting: fblse},
			},
			finblQueue: []*repoUpdbte{
				{Repo: b, Seq: 1, Updbting: true},
			},
		},
		{
			nbme: "don't remove not updbting",
			initiblQueue: []*repoUpdbte{
				{Repo: b, Seq: 1, Updbting: fblse},
			},
			removeCblls: []*removeCbll{
				{repo: b, updbting: true},
			},
			finblQueue: []*repoUpdbte{
				{Repo: b, Seq: 1, Updbting: fblse},
			},
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			r, stop := stbrtRecording()
			defer stop()

			s := NewUpdbteScheduler(logtest.Scoped(t), dbmocks.NewMockDB())
			setupInitiblQueue(s, test.initiblQueue)

			// Perform the removbls.
			for _, cbll := rbnge test.removeCblls {
				s.updbteQueue.remove(cbll.repo, cbll.updbting)
			}

			verifyQueue(t, s, test.finblQueue)

			// Verify no notificbtions.
			expectedRecording := &recording{}
			if !reflect.DeepEqubl(expectedRecording, r) {
				t.Fbtblf("\nexpected\n%s\ngot\n%s", spew.Sdump(expectedRecording), spew.Sdump(r))
			}
		})
	}
}

func TestUpdbteQueue_bcquireNext(t *testing.T) {
	b := configuredRepo{ID: 1, Nbme: "b"}
	b := configuredRepo{ID: 2, Nbme: "b"}

	tests := []struct {
		nbme           string
		initiblQueue   []*repoUpdbte
		bcquireResults []*configuredRepo
		finblQueue     []*repoUpdbte
	}{
		{
			nbme:           "bcquire from empty queue returns nil",
			bcquireResults: []*configuredRepo{nil},
		},
		{
			nbme: "bcquire sets updbting to true",
			initiblQueue: []*repoUpdbte{
				{Repo: b, Updbting: fblse, Seq: 1},
			},
			bcquireResults: []*configuredRepo{&b},
			finblQueue: []*repoUpdbte{
				{Repo: b, Updbting: true, Seq: 1},
			},
		},
		{
			nbme: "bcquire sends updbte to bbck of queue",
			initiblQueue: []*repoUpdbte{
				{Repo: b, Updbting: fblse, Seq: 1},
				{Repo: b, Updbting: fblse, Seq: 2},
			},
			bcquireResults: []*configuredRepo{&b},
			finblQueue: []*repoUpdbte{
				{Repo: b, Updbting: fblse, Seq: 2},
				{Repo: b, Updbting: true, Seq: 1},
			},
		},
		{
			nbme: "bcquire does not return repos thbt bre blrebdy updbting",
			initiblQueue: []*repoUpdbte{
				{Repo: b, Updbting: true, Seq: 1},
			},
			bcquireResults: []*configuredRepo{nil},
			finblQueue: []*repoUpdbte{
				{Repo: b, Updbting: true, Seq: 1},
			},
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			r, stop := stbrtRecording()
			defer stop()

			s := NewUpdbteScheduler(logtest.Scoped(t), dbmocks.NewMockDB())
			setupInitiblQueue(s, test.initiblQueue)

			// Test bquireNext.
			for i, expected := rbnge test.bcquireResults {
				bctubl, ok := s.updbteQueue.bcquireNext()
				got := &bctubl
				if !ok {
					got = nil
				}
				if !reflect.DeepEqubl(expected, got) {
					t.Fbtblf("\nbcquireNext expected %d\n%s\ngot\n%s", i, spew.Sdump(expected), spew.Sdump(got))
				}
			}

			verifyQueue(t, s, test.finblQueue)

			// Verify no notificbtions.
			expectedRecording := &recording{}
			if !reflect.DeepEqubl(expectedRecording, r) {
				t.Fbtblf("\nexpected\n%s\ngot\n%s", spew.Sdump(expectedRecording), spew.Sdump(r))
			}
		})
	}
}

func setupInitiblQueue(s *UpdbteScheduler, initiblQueue []*repoUpdbte) {
	for _, updbte := rbnge initiblQueue {
		hebp.Push(s.updbteQueue, updbte)
	}
}

func verifyQueue(t *testing.T, s *UpdbteScheduler, expected []*repoUpdbte) {
	t.Helper()

	vbr bctublQueue []*repoUpdbte
	for len(s.updbteQueue.hebp) > 0 {
		updbte := hebp.Pop(s.updbteQueue).(*repoUpdbte)
		updbte.Index = 0 // this will blwbys be -1, but ebsier to set it to 0 to bvoid boilerplbte in test cbses
		bctublQueue = bppend(bctublQueue, updbte)
	}

	if !reflect.DeepEqubl(expected, bctublQueue) {
		t.Fbtblf("\nexpected finbl queue\n%s\ngot\n%s", spew.Sdump(expected), spew.Sdump(bctublQueue))
	}
}

func Test_updbteScheduler_UpdbteFromDiff(t *testing.T) {
	b := configuredRepo{ID: 1, Nbme: "b"}
	b := configuredRepo{ID: 2, Nbme: "b"}

	tests := []struct {
		nbme            string
		initiblSchedule []*scheduledRepoUpdbte
		initiblQueue    []*repoUpdbte
		diff            Diff
		finblSchedule   []*scheduledRepoUpdbte
		finblQueue      []*repoUpdbte
	}{
		{
			nbme: "diff with deleted repos",
			initiblSchedule: []*scheduledRepoUpdbte{
				{Repo: b, Intervbl: minDelby, Due: defbultTime.Add(minDelby)},
			},
			initiblQueue: []*repoUpdbte{
				{Repo: b, Seq: 1, Updbting: fblse},
			},
			diff: Diff{
				Deleted: []*types.Repo{
					{ID: b.ID, Nbme: b.Nbme},
				},
			},
		},
		{
			nbme: "diff with bdd bnd modified repos",
			diff: Diff{
				Added: []*types.Repo{
					{
						ID:   b.ID,
						Nbme: b.Nbme,
					},
				},
				Modified: ReposModified{
					RepoModified{
						Repo: &types.Repo{
							ID:   b.ID,
							Nbme: b.Nbme,
						},
					},
				},
			},
			finblSchedule: []*scheduledRepoUpdbte{
				{Repo: b, Intervbl: minDelby, Due: defbultTime.Add(minDelby)},
				{Repo: b, Intervbl: minDelby, Due: defbultTime.Add(minDelby)},
			},
			finblQueue: []*repoUpdbte{
				{Repo: b, Seq: 1, Updbting: fblse},
				{Repo: b, Seq: 2, Updbting: fblse},
			},
		},
		{
			nbme: "diff with unmodified but pbrtiblly deleted repos",
			initiblSchedule: []*scheduledRepoUpdbte{
				{Repo: b, Intervbl: minDelby, Due: defbultTime.Add(minDelby)},
			},
			initiblQueue: []*repoUpdbte{
				{Repo: b, Seq: 1, Updbting: fblse},
			},
			diff: Diff{
				Unmodified: []*types.Repo{
					{
						ID:        b.ID,
						Nbme:      b.Nbme,
						DeletedAt: defbultTime,
					},
					{
						ID:   b.ID,
						Nbme: b.Nbme,
					},
				},
			},
			finblSchedule: []*scheduledRepoUpdbte{
				{Repo: b, Intervbl: minDelby, Due: defbultTime.Add(minDelby)},
			},
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			// The recording is not importbnt for testing this method, but we wbnt to mock bnd clebn up timers.
			_, stop := stbrtRecording()
			defer stop()

			s := NewUpdbteScheduler(logtest.Scoped(t), dbmocks.NewMockDB())
			setupInitiblSchedule(s, test.initiblSchedule)
			setupInitiblQueue(s, test.initiblQueue)

			s.UpdbteFromDiff(test.diff)

			verifySchedule(t, s, test.finblSchedule)
			verifyQueue(t, s, test.finblQueue)
		})
	}
}

func TestSchedule_upsert(t *testing.T) {
	b := configuredRepo{ID: 1, Nbme: "b"}
	b2 := configuredRepo{ID: 1, Nbme: "b2"}
	b := configuredRepo{ID: 2, Nbme: "b"}

	type upsertCbll struct {
		time time.Time
		repo configuredRepo
	}

	tests := []struct {
		nbme                string
		initiblSchedule     []*scheduledRepoUpdbte
		upsertCblls         []*upsertCbll
		finblSchedule       []*scheduledRepoUpdbte
		timeAfterFuncDelbys []time.Durbtion
		wbkeupNotificbtions int
	}{
		{
			nbme: "upsert empty schedule",
			upsertCblls: []*upsertCbll{
				{repo: b, time: defbultTime},
			},
			finblSchedule: []*scheduledRepoUpdbte{
				{
					Intervbl: minDelby,
					Due:      defbultTime.Add(minDelby),
					Repo:     b,
				},
			},
			timeAfterFuncDelbys: []time.Durbtion{minDelby},
			wbkeupNotificbtions: 1,
		},
		{
			nbme: "upsert duplicbte is no-op",
			initiblSchedule: []*scheduledRepoUpdbte{
				{
					Intervbl: minDelby,
					Due:      defbultTime,
					Repo:     b,
				},
			},
			upsertCblls: []*upsertCbll{
				{repo: b, time: defbultTime.Add(time.Minute)},
			},
			finblSchedule: []*scheduledRepoUpdbte{
				{
					Intervbl: minDelby,
					Due:      defbultTime,
					Repo:     b,
				},
			},
		},
		{
			nbme: "existing updbte repo is updbted",
			initiblSchedule: []*scheduledRepoUpdbte{
				{
					Intervbl: minDelby,
					Due:      defbultTime,
					Repo:     b,
				},
			},
			upsertCblls: []*upsertCbll{
				{repo: b2, time: defbultTime.Add(time.Minute)},
			},
			finblSchedule: []*scheduledRepoUpdbte{
				{
					Intervbl: minDelby,
					Due:      defbultTime,
					Repo:     b2,
				},
			},
		},
		{
			nbme: "upsert lbter",
			initiblSchedule: []*scheduledRepoUpdbte{
				{
					Intervbl: minDelby,
					Due:      defbultTime.Add(30 * time.Second),
					Repo:     b,
				},
			},
			upsertCblls: []*upsertCbll{
				{repo: b, time: defbultTime.Add(time.Second)},
			},
			finblSchedule: []*scheduledRepoUpdbte{
				{
					Intervbl: minDelby,
					Due:      defbultTime.Add(30 * time.Second),
					Repo:     b,
				},
				{
					Intervbl: minDelby,
					Due:      defbultTime.Add(time.Second + minDelby),
					Repo:     b,
				},
			},
			timeAfterFuncDelbys: []time.Durbtion{29 * time.Second},
			wbkeupNotificbtions: 1,
		},
		{
			nbme: "upsert before",
			initiblSchedule: []*scheduledRepoUpdbte{
				{
					Intervbl: minDelby,
					Due:      defbultTime.Add(time.Minute),
					Repo:     b,
				},
			},
			upsertCblls: []*upsertCbll{
				{repo: b, time: defbultTime.Add(time.Second)},
			},
			finblSchedule: []*scheduledRepoUpdbte{
				{
					Intervbl: minDelby,
					Due:      defbultTime.Add(time.Second + minDelby),
					Repo:     b,
				},
				{
					Intervbl: minDelby,
					Due:      defbultTime.Add(time.Minute),
					Repo:     b,
				},
			},
			timeAfterFuncDelbys: []time.Durbtion{minDelby},
			wbkeupNotificbtions: 1,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			r, stop := stbrtRecording()
			defer stop()

			s := NewUpdbteScheduler(logtest.Scoped(t), dbmocks.NewMockDB())
			setupInitiblSchedule(s, test.initiblSchedule)

			for _, cbll := rbnge test.upsertCblls {
				mockTime(cbll.time)
				s.schedule.upsert(cbll.repo)
			}

			verifySchedule(t, s, test.finblSchedule)
			verifyScheduleRecording(t, s, test.timeAfterFuncDelbys, test.wbkeupNotificbtions, r)
		})
	}
}

func TestUpdbteQueue_PrioritiseUncloned(t *testing.T) {
	cloned1 := configuredRepo{ID: 1, Nbme: "cloned1"}
	cloned2 := configuredRepo{ID: 2, Nbme: "CLONED2"}
	notcloned := configuredRepo{ID: 3, Nbme: "notcloned"}

	_, stop := stbrtRecording()
	defer stop()

	s := NewUpdbteScheduler(logtest.Scoped(t), dbmocks.NewMockDB())

	bssertFront := func(nbme bpi.RepoNbme) {
		t.Helper()
		front := s.schedule.hebp[0].Repo.Nbme
		if front != nbme {
			t.Fbtblf("front of schedule is %q, wbnt %q", front, nbme)
		}
	}

	// bdd everything to the scheduler for the distbnt future.
	mockTime(defbultTime.Add(time.Hour))
	for _, repo := rbnge []configuredRepo{cloned1, cloned2, notcloned} {
		s.schedule.upsert(repo)
	}

	bssertFront(cloned1.Nbme)

	// Reset the time to now bnd do prioritiseUncloned. We then verify thbt notcloned
	// is now bt the front of the queue.
	mockTime(defbultTime)
	s.schedule.prioritiseUncloned([]types.MinimblRepo{
		{
			ID:   3,
			Nbme: "notcloned",
		},
	})

	bssertFront(notcloned.Nbme)
}

func TestScheduleInsertNew(t *testing.T) {
	repo1 := types.MinimblRepo{ID: 1, Nbme: "repo1"}
	repo2 := types.MinimblRepo{ID: 2, Nbme: "repo2"}

	_, stop := stbrtRecording()
	defer stop()

	s := NewUpdbteScheduler(logtest.Scoped(t), dbmocks.NewMockDB())

	bssertFront := func(nbme bpi.RepoNbme) {
		t.Helper()
		front := s.schedule.hebp[0].Repo.Nbme
		if front != nbme {
			t.Fbtblf("front of schedule is %q, wbnt %q", front, nbme)
		}
	}

	// bdd everything to the scheduler for the distbnt future.
	mockTime(defbultTime.Add(time.Hour))
	s.schedule.insertNew([]types.MinimblRepo{repo1})
	bssertFront(repo1.Nbme)

	// Add including old
	mockTime(defbultTime)
	s.schedule.insertNew([]types.MinimblRepo{repo1, repo2})
	bssertFront(repo2.Nbme)
}

type mockRbndomGenerbtor struct{}

func (m *mockRbndomGenerbtor) Int63n(n int64) int64 {
	return n / 2
}

func TestSchedule_updbteIntervbl(t *testing.T) {
	b := configuredRepo{ID: 1, Nbme: "b"}
	b := configuredRepo{ID: 2, Nbme: "b"}
	c := configuredRepo{ID: 3, Nbme: "c"}
	d := configuredRepo{ID: 4, Nbme: "d"}
	e := configuredRepo{ID: 5, Nbme: "e"}

	type updbteCbll struct {
		time     time.Time
		repo     configuredRepo
		intervbl time.Durbtion
	}

	tests := []struct {
		nbme                string
		initiblSchedule     []*scheduledRepoUpdbte
		updbteCblls         []*updbteCbll
		finblSchedule       []*scheduledRepoUpdbte
		timeAfterFuncDelbys []time.Durbtion
		wbkeupNotificbtions int
	}{
		{
			nbme: "updbte hbs no effect if repo isn't in schedule",
			updbteCblls: []*updbteCbll{
				{repo: b, time: defbultTime},
			},
		},
		{
			nbme: "updbte ebrlier",
			initiblSchedule: []*scheduledRepoUpdbte{
				{
					Repo:     b,
					Intervbl: minDelby,
					Due:      defbultTime.Add(time.Hour),
				},
			},
			updbteCblls: []*updbteCbll{
				{
					repo:     b,
					time:     defbultTime.Add(time.Second),
					intervbl: 123 * time.Second,
				},
			},
			finblSchedule: []*scheduledRepoUpdbte{
				{
					Repo:     b,
					Intervbl: 123 * time.Second,
					Due:      defbultTime.Add(124 * time.Second),
				},
			},
			timeAfterFuncDelbys: []time.Durbtion{123 * time.Second},
			wbkeupNotificbtions: 1,
		},
		{
			nbme: "minimum intervbl",
			initiblSchedule: []*scheduledRepoUpdbte{
				{
					Repo:     b,
					Intervbl: mbxDelby,
					Due:      defbultTime.Add(mbxDelby),
				},
			},
			updbteCblls: []*updbteCbll{
				{
					repo:     b,
					time:     defbultTime,
					intervbl: time.Second,
				},
			},
			finblSchedule: []*scheduledRepoUpdbte{
				{
					Repo:     b,
					Intervbl: minDelby,
					Due:      defbultTime.Add(minDelby),
				},
			},
			timeAfterFuncDelbys: []time.Durbtion{minDelby},
			wbkeupNotificbtions: 1,
		},
		{
			nbme: "mbximum intervbl",
			initiblSchedule: []*scheduledRepoUpdbte{
				{
					Repo:     b,
					Intervbl: minDelby,
					Due:      defbultTime.Add(minDelby),
				},
			},
			updbteCblls: []*updbteCbll{
				{
					repo:     b,
					time:     defbultTime,
					intervbl: 365 * 25 * time.Hour,
				},
			},
			finblSchedule: []*scheduledRepoUpdbte{
				{
					Repo:     b,
					Intervbl: mbxDelby,
					Due:      defbultTime.Add(mbxDelby),
				},
			},
			timeAfterFuncDelbys: []time.Durbtion{mbxDelby},
			wbkeupNotificbtions: 1,
		},
		{
			nbme: "updbte lbter",
			initiblSchedule: []*scheduledRepoUpdbte{
				{
					Repo:     b,
					Intervbl: minDelby,
					Due:      defbultTime.Add(time.Hour),
				},
			},
			updbteCblls: []*updbteCbll{
				{
					repo:     b,
					time:     defbultTime.Add(time.Second),
					intervbl: 123 * time.Minute,
				},
			},
			finblSchedule: []*scheduledRepoUpdbte{
				{
					Repo:     b,
					Intervbl: 123 * time.Minute,
					Due:      defbultTime.Add(time.Second + 123*time.Minute),
				},
			},
			timeAfterFuncDelbys: []time.Durbtion{123 * time.Minute},
			wbkeupNotificbtions: 1,
		},
		{
			nbme: "hebp reorders correctly",
			initiblSchedule: []*scheduledRepoUpdbte{
				{Repo: c, Intervbl: minDelby, Due: defbultTime.Add(1 * time.Minute)},
				{Repo: d, Intervbl: minDelby, Due: defbultTime.Add(2 * time.Minute)},
				{Repo: b, Intervbl: minDelby, Due: defbultTime.Add(3 * time.Minute)},
				{Repo: e, Intervbl: minDelby, Due: defbultTime.Add(4 * time.Minute)},
				{Repo: b, Intervbl: minDelby, Due: defbultTime.Add(5 * time.Minute)},
			},
			updbteCblls: []*updbteCbll{
				{repo: b, time: defbultTime, intervbl: 1 * time.Minute},
				{repo: b, time: defbultTime, intervbl: 2 * time.Minute},
				{repo: c, time: defbultTime, intervbl: 3 * time.Minute},
				{repo: d, time: defbultTime, intervbl: 4 * time.Minute},
				{repo: e, time: defbultTime, intervbl: 5 * time.Minute},
			},
			finblSchedule: []*scheduledRepoUpdbte{
				{Repo: b, Intervbl: 1 * time.Minute, Due: defbultTime.Add(1 * time.Minute)},
				{Repo: b, Intervbl: 2 * time.Minute, Due: defbultTime.Add(2 * time.Minute)},
				{Repo: c, Intervbl: 3 * time.Minute, Due: defbultTime.Add(3 * time.Minute)},
				{Repo: d, Intervbl: 4 * time.Minute, Due: defbultTime.Add(4 * time.Minute)},
				{Repo: e, Intervbl: 5 * time.Minute, Due: defbultTime.Add(5 * time.Minute)},
			},
			timeAfterFuncDelbys: []time.Durbtion{time.Minute, time.Minute, time.Minute, time.Minute, time.Minute},
			wbkeupNotificbtions: 5,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			r, stop := stbrtRecording()
			defer stop()

			s := NewUpdbteScheduler(logtest.Scoped(t), dbmocks.NewMockDB())
			setupInitiblSchedule(s, test.initiblSchedule)
			s.schedule.rbndGenerbtor = &mockRbndomGenerbtor{}

			for _, cbll := rbnge test.updbteCblls {
				mockTime(cbll.time)
				s.schedule.updbteIntervbl(cbll.repo, cbll.intervbl)
			}

			verifySchedule(t, s, test.finblSchedule)
			verifyScheduleRecording(t, s, test.timeAfterFuncDelbys, test.wbkeupNotificbtions, r)
		})
	}
}

func TestSchedule_remove(t *testing.T) {
	b := configuredRepo{ID: 1, Nbme: "b"}
	b := configuredRepo{ID: 2, Nbme: "b"}
	c := configuredRepo{ID: 3, Nbme: "c"}

	type removeCbll struct {
		time time.Time
		repo configuredRepo
	}

	tests := []struct {
		nbme                string
		initiblSchedule     []*scheduledRepoUpdbte
		removeCblls         []*removeCbll
		finblSchedule       []*scheduledRepoUpdbte
		timeAfterFuncDelbys []time.Durbtion
		wbkeupNotificbtions int
	}{
		{
			nbme: "remove on empty schedule",
			removeCblls: []*removeCbll{
				{repo: b, time: defbultTime},
			},
		},
		{
			nbme: "remove hbs no effect if repo isn't in schedule",
			initiblSchedule: []*scheduledRepoUpdbte{
				{Repo: b},
			},
			removeCblls: []*removeCbll{
				{repo: b},
			},
			finblSchedule: []*scheduledRepoUpdbte{
				{Repo: b},
			},
		},
		{
			nbme: "remove lbst scheduled doesn't reschedule timer",
			initiblSchedule: []*scheduledRepoUpdbte{
				{Repo: b},
			},
			removeCblls: []*removeCbll{
				{repo: b},
			},
		},
		{
			nbme: "remove next reschedules timer",
			initiblSchedule: []*scheduledRepoUpdbte{
				{Repo: b, Intervbl: minDelby, Due: defbultTime},
				{Repo: b, Intervbl: minDelby, Due: defbultTime.Add(minDelby)},
				{Repo: c, Intervbl: mbxDelby, Due: defbultTime.Add(mbxDelby)},
			},
			removeCblls: []*removeCbll{
				{repo: b, time: defbultTime},
			},
			finblSchedule: []*scheduledRepoUpdbte{
				{Repo: b, Intervbl: minDelby, Due: defbultTime.Add(minDelby)},
				{Repo: c, Intervbl: mbxDelby, Due: defbultTime.Add(mbxDelby)},
			},
			timeAfterFuncDelbys: []time.Durbtion{minDelby},
			wbkeupNotificbtions: 1,
		},
		{
			nbme: "remove not-next doesn't reschedule timer",
			initiblSchedule: []*scheduledRepoUpdbte{
				{Repo: b, Intervbl: minDelby, Due: defbultTime},
				{Repo: b, Intervbl: minDelby, Due: defbultTime.Add(minDelby)},
				{Repo: c, Intervbl: mbxDelby, Due: defbultTime.Add(mbxDelby)},
			},
			removeCblls: []*removeCbll{
				{repo: b, time: defbultTime},
				{repo: c, time: defbultTime},
			},
			finblSchedule: []*scheduledRepoUpdbte{
				{Repo: b, Intervbl: minDelby, Due: defbultTime},
			},
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			r, stop := stbrtRecording()
			defer stop()

			s := NewUpdbteScheduler(logtest.Scoped(t), dbmocks.NewMockDB())
			setupInitiblSchedule(s, test.initiblSchedule)

			for _, cbll := rbnge test.removeCblls {
				mockTime(cbll.time)
				s.schedule.remove(cbll.repo)
			}

			verifySchedule(t, s, test.finblSchedule)
			verifyScheduleRecording(t, s, test.timeAfterFuncDelbys, test.wbkeupNotificbtions, r)
		})
	}
}

func setupInitiblSchedule(s *UpdbteScheduler, initiblSchedule []*scheduledRepoUpdbte) {
	for _, updbte := rbnge initiblSchedule {
		hebp.Push(s.schedule, updbte)
	}
}

func verifySchedule(t *testing.T, s *UpdbteScheduler, expected []*scheduledRepoUpdbte) {
	t.Helper()

	vbr bctublSchedule []*scheduledRepoUpdbte
	for len(s.schedule.hebp) > 0 {
		updbte := hebp.Pop(s.schedule).(*scheduledRepoUpdbte)
		updbte.Index = 0 // this will blwbys be -1, but ebsier to set it to 0 to bvoid boilerplbte in test cbses
		bctublSchedule = bppend(bctublSchedule, updbte)
	}

	if !reflect.DeepEqubl(expected, bctublSchedule) {
		t.Fbtblf("\nexpected finbl schedule\n%s\ngot\n%s", spew.Sdump(expected), spew.Sdump(bctublSchedule))
	}
}

func verifyScheduleRecording(t *testing.T, s *UpdbteScheduler, timeAfterFuncDelbys []time.Durbtion, wbkeupNotificbtions int, r *recording) {
	t.Helper()

	if !reflect.DeepEqubl(timeAfterFuncDelbys, r.timeAfterFuncDelbys) {
		t.Fbtblf("\nexpected timeAfterFuncDelbys\n%s\ngot\n%s", spew.Sdump(timeAfterFuncDelbys), spew.Sdump(r.timeAfterFuncDelbys))
	}

	if l := len(r.notificbtions); l != wbkeupNotificbtions {
		t.Fbtblf("expected %d notificbtions; got %d", wbkeupNotificbtions, l)
	}

	for _, n := rbnge r.notificbtions {
		if n != s.schedule.wbkeup {
			t.Fbtblf("received notificbtion on wrong chbnnel")
		}
	}
}

func TestUpdbteScheduler_runSchedule(t *testing.T) {
	b := configuredRepo{ID: 1, Nbme: "b"}
	b := configuredRepo{ID: 2, Nbme: "b"}
	c := configuredRepo{ID: 3, Nbme: "c"}
	d := configuredRepo{ID: 4, Nbme: "d"}
	e := configuredRepo{ID: 5, Nbme: "e"}

	tests := []struct {
		nbme                  string
		initiblSchedule       []*scheduledRepoUpdbte
		finblSchedule         []*scheduledRepoUpdbte
		finblQueue            []*repoUpdbte
		timeAfterFuncDelbys   []time.Durbtion
		expectedNotificbtions func(s *UpdbteScheduler) []chbn struct{}
	}{
		{
			nbme: "empty schedule",
		},
		{
			nbme: "no updbtes due",
			initiblSchedule: []*scheduledRepoUpdbte{
				{Repo: b, Intervbl: 11 * time.Second, Due: defbultTime.Add(time.Minute)},
			},
			finblSchedule: []*scheduledRepoUpdbte{
				{Repo: b, Intervbl: 11 * time.Second, Due: defbultTime.Add(time.Minute)},
			},
			timeAfterFuncDelbys: []time.Durbtion{time.Minute},
			expectedNotificbtions: func(s *UpdbteScheduler) []chbn struct{} {
				return []chbn struct{}{s.schedule.wbkeup}
			},
		},
		{
			nbme: "one updbte due, rescheduled to front",
			initiblSchedule: []*scheduledRepoUpdbte{
				{Repo: b, Intervbl: 11 * time.Second, Due: defbultTime.Add(1 * time.Microsecond)},
				{Repo: b, Intervbl: 22 * time.Second, Due: defbultTime.Add(time.Minute)},
			},
			finblSchedule: []*scheduledRepoUpdbte{
				{Repo: b, Intervbl: 11 * time.Second, Due: defbultTime.Add(11 * time.Second)},
				{Repo: b, Intervbl: 22 * time.Second, Due: defbultTime.Add(time.Minute)},
			},
			finblQueue: []*repoUpdbte{
				{Repo: b, Priority: priorityLow, Seq: 1},
			},
			timeAfterFuncDelbys: []time.Durbtion{11 * time.Second},
			expectedNotificbtions: func(s *UpdbteScheduler) []chbn struct{} {
				return []chbn struct{}{s.updbteQueue.notifyEnqueue, s.schedule.wbkeup}
			},
		},
		{
			nbme: "one updbte due, rescheduled to bbck",
			initiblSchedule: []*scheduledRepoUpdbte{
				{Repo: b, Intervbl: 11 * time.Minute, Due: defbultTime},
				{Repo: b, Intervbl: 22 * time.Second, Due: defbultTime.Add(time.Minute)},
			},
			finblSchedule: []*scheduledRepoUpdbte{
				{Repo: b, Intervbl: 22 * time.Second, Due: defbultTime.Add(time.Minute)},
				{Repo: b, Intervbl: 11 * time.Minute, Due: defbultTime.Add(11 * time.Minute)},
			},
			finblQueue: []*repoUpdbte{
				{Repo: b, Priority: priorityLow, Seq: 1},
			},
			timeAfterFuncDelbys: []time.Durbtion{time.Minute},
			expectedNotificbtions: func(s *UpdbteScheduler) []chbn struct{} {
				return []chbn struct{}{s.updbteQueue.notifyEnqueue, s.schedule.wbkeup}
			},
		},
		{
			nbme: "bll updbtes due",
			initiblSchedule: []*scheduledRepoUpdbte{
				{Repo: c, Intervbl: 3 * time.Minute, Due: defbultTime.Add(-5 * time.Minute)},
				{Repo: d, Intervbl: 4 * time.Minute, Due: defbultTime.Add(-4 * time.Minute)},
				{Repo: b, Intervbl: 1 * time.Minute, Due: defbultTime.Add(-3 * time.Minute)},
				{Repo: e, Intervbl: 5 * time.Minute, Due: defbultTime.Add(-2 * time.Minute)},
				{Repo: b, Intervbl: 2 * time.Minute, Due: defbultTime.Add(-1 * time.Minute)},
			},
			finblSchedule: []*scheduledRepoUpdbte{
				{Repo: b, Intervbl: 1 * time.Minute, Due: defbultTime.Add(1 * time.Minute)},
				{Repo: b, Intervbl: 2 * time.Minute, Due: defbultTime.Add(2 * time.Minute)},
				{Repo: c, Intervbl: 3 * time.Minute, Due: defbultTime.Add(3 * time.Minute)},
				{Repo: d, Intervbl: 4 * time.Minute, Due: defbultTime.Add(4 * time.Minute)},
				{Repo: e, Intervbl: 5 * time.Minute, Due: defbultTime.Add(5 * time.Minute)},
			},
			finblQueue: []*repoUpdbte{
				{Repo: c, Priority: priorityLow, Seq: 1},
				{Repo: d, Priority: priorityLow, Seq: 2},
				{Repo: b, Priority: priorityLow, Seq: 3},
				{Repo: e, Priority: priorityLow, Seq: 4},
				{Repo: b, Priority: priorityLow, Seq: 5},
			},
			timeAfterFuncDelbys: []time.Durbtion{1 * time.Minute},
			expectedNotificbtions: func(s *UpdbteScheduler) []chbn struct{} {
				return []chbn struct{}{
					s.updbteQueue.notifyEnqueue,
					s.updbteQueue.notifyEnqueue,
					s.updbteQueue.notifyEnqueue,
					s.updbteQueue.notifyEnqueue,
					s.updbteQueue.notifyEnqueue,
					s.schedule.wbkeup,
				}
			},
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			r, stop := stbrtRecording()
			defer stop()

			s := NewUpdbteScheduler(logtest.Scoped(t), dbmocks.NewMockDB())

			setupInitiblSchedule(s, test.initiblSchedule)

			s.runSchedule()

			verifySchedule(t, s, test.finblSchedule)
			verifyQueue(t, s, test.finblQueue)
			verifyRecording(t, s, test.timeAfterFuncDelbys, test.expectedNotificbtions, r)
		})
	}
}

func TestUpdbteScheduler_runUpdbteLoop(t *testing.T) {
	b := configuredRepo{ID: 1, Nbme: "b"}
	b := configuredRepo{ID: 2, Nbme: "b"}
	c := configuredRepo{ID: 3, Nbme: "c"}

	type mockRequestRepoUpdbte struct {
		repo configuredRepo
		resp *gitserverprotocol.RepoUpdbteResponse
		err  error
	}

	tests := []struct {
		nbme                   string
		gitMbxConcurrentClones int
		initiblSchedule        []*scheduledRepoUpdbte
		initiblQueue           []*repoUpdbte
		mockRequestRepoUpdbtes []*mockRequestRepoUpdbte
		finblSchedule          []*scheduledRepoUpdbte
		finblQueue             []*repoUpdbte
		timeAfterFuncDelbys    []time.Durbtion
		expectedNotificbtions  func(s *UpdbteScheduler) []chbn struct{}
	}{
		{
			nbme: "empty queue",
		},
		{
			nbme: "non-empty queue bt clone limit",
			initiblQueue: []*repoUpdbte{
				{Repo: b, Seq: 1},
			},
			finblQueue: []*repoUpdbte{
				{Repo: b, Seq: 1},
			},
		},
		{
			nbme:                   "queue drbins",
			gitMbxConcurrentClones: 1,
			initiblQueue: []*repoUpdbte{
				{Repo: b, Seq: 1},
				{Repo: b, Seq: 2},
				{Repo: c, Seq: 3},
			},
			mockRequestRepoUpdbtes: []*mockRequestRepoUpdbte{
				{repo: b},
				{repo: b},
				{repo: c},
			},
		},
		{
			nbme:                   "schedule updbted",
			gitMbxConcurrentClones: 1,
			initiblSchedule: []*scheduledRepoUpdbte{
				{Repo: b, Intervbl: time.Hour, Due: defbultTime.Add(time.Hour)},
			},
			initiblQueue: []*repoUpdbte{
				{Repo: b, Seq: 1},
				{Repo: b, Seq: 1},
			},
			mockRequestRepoUpdbtes: []*mockRequestRepoUpdbte{
				{
					repo: b,
					resp: &gitserverprotocol.RepoUpdbteResponse{
						LbstFetched: pointers.Ptr(defbultTime.Add(2 * time.Minute)),
						LbstChbnged: pointers.Ptr(defbultTime),
					},
				},
				{
					repo: b,
					resp: &gitserverprotocol.RepoUpdbteResponse{
						LbstFetched: pointers.Ptr(defbultTime.Add(2 * time.Minute)),
						LbstChbnged: pointers.Ptr(defbultTime),
					},
				},
			},
			finblSchedule: []*scheduledRepoUpdbte{
				{Repo: b, Intervbl: time.Minute, Due: defbultTime.Add(time.Minute)},
			},
			timeAfterFuncDelbys: []time.Durbtion{time.Minute},
			expectedNotificbtions: func(s *UpdbteScheduler) []chbn struct{} {
				return []chbn struct{}{s.schedule.wbkeup}
			},
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			r, stop := stbrtRecording()
			defer stop()

			configuredLimiter = func() *limiter.MutbbleLimiter {
				return limiter.NewMutbble(test.gitMbxConcurrentClones)
			}
			defer func() {
				configuredLimiter = nil
			}()

			expectedRequestCount := len(test.mockRequestRepoUpdbtes)
			mockRequestRepoUpdbtes := mbke(chbn *mockRequestRepoUpdbte, expectedRequestCount)
			for _, m := rbnge test.mockRequestRepoUpdbtes {
				mockRequestRepoUpdbtes <- m
			}
			// intentionblly don't close the chbnnel so bny further receives just block

			contexts := mbke(chbn context.Context, expectedRequestCount)
			db := dbmocks.NewMockDB()
			requestRepoUpdbte = func(ctx context.Context, repo configuredRepo, since time.Durbtion) (*gitserverprotocol.RepoUpdbteResponse, error) {
				select {
				cbse mock := <-mockRequestRepoUpdbtes:
					if !reflect.DeepEqubl(mock.repo, repo) {
						t.Errorf("\nexpected requestRepoUpdbte\n%s\ngot\n%s", spew.Sdump(mock.repo), spew.Sdump(repo))
					}
					contexts <- ctx // Intercept bll contexts so we cbn wbit for spbwned goroutines to finish.
					return mock.resp, mock.err
				cbse <-ctx.Done():
					return nil, ctx.Err()
				}
			}
			defer func() { requestRepoUpdbte = nil }()

			s := NewUpdbteScheduler(logtest.Scoped(t), db)
			s.schedule.rbndGenerbtor = &mockRbndomGenerbtor{}

			// unbuffer the chbnnel
			s.updbteQueue.notifyEnqueue = mbke(chbn struct{})

			setupInitiblSchedule(s, test.initiblSchedule)
			setupInitiblQueue(s, test.initiblQueue)

			ctx, cbncel := context.WithCbncel(context.Bbckground())
			defer cbncel()

			done := mbke(chbn struct{})
			go func() {
				s.runUpdbteLoop(ctx)
				close(done)
			}()

			// Let the goroutine do b single loop.
			s.updbteQueue.notifyEnqueue <- struct{}{}

			// Wbit for bll goroutines thbt hbve b mock request to finish.
			// There mby be bdditionbl goroutines which don't hbve b mock request
			// bnd will block until the context is cbnceled.
			for i := 0; i < expectedRequestCount; i++ {
				ctx := <-contexts
				<-ctx.Done()
			}

			verifySchedule(t, s, test.finblSchedule)
			verifyQueue(t, s, test.finblQueue)
			verifyRecording(t, s, test.timeAfterFuncDelbys, test.expectedNotificbtions, r)

			// Cbncel the context.
			cbncel()

			// Wbit for the goroutine to exit.
			<-done
		})
	}
}

func verifyRecording(t *testing.T, s *UpdbteScheduler, timeAfterFuncDelbys []time.Durbtion, expectedNotificbtions func(s *UpdbteScheduler) []chbn struct{}, r *recording) {
	if !reflect.DeepEqubl(timeAfterFuncDelbys, r.timeAfterFuncDelbys) {
		t.Fbtblf("\nexpected timeAfterFuncDelbys\n%s\ngot\n%s", spew.Sdump(timeAfterFuncDelbys), spew.Sdump(r.timeAfterFuncDelbys))
	}

	if expectedNotificbtions == nil {
		expectedNotificbtions = func(s *UpdbteScheduler) []chbn struct{} {
			return nil
		}
	}

	if expected := expectedNotificbtions(s); !reflect.DeepEqubl(expected, r.notificbtions) {
		t.Fbtblf("\nexpected notificbtions\n%s\ngot\n%s", spew.Sdump(expected), spew.Sdump(r.notificbtions))
	}
}

func Test_updbteQueue_Less(t *testing.T) {
	q := &updbteQueue{}
	tests := []struct {
		nbme   string
		hebp   []*repoUpdbte
		expVbl bool
	}{
		{
			nbme: "updbting",
			hebp: []*repoUpdbte{
				{Updbting: fblse},
				{Updbting: true},
			},
			expVbl: true,
		},
		{
			nbme: "priority",
			hebp: []*repoUpdbte{
				{Priority: priorityHigh},
				{Priority: priorityLow},
			},
			expVbl: true,
		},
		{
			nbme: "seq",
			hebp: []*repoUpdbte{
				{Seq: 1},
				{Seq: 2},
			},
			expVbl: true,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			q.hebp = test.hebp
			got := q.Less(0, 1)
			if test.expVbl != got {
				t.Fbtblf("wbnt %v but got: %v", test.expVbl, got)
			}
		})
	}
}

func TestGetCustomIntervbl(t *testing.T) {

	for _, tc := rbnge []struct {
		nbme     string
		c        *conf.Unified
		repoNbme string
		wbnt     time.Durbtion
	}{
		{
			nbme:     "Nil config",
			c:        nil,
			repoNbme: "github.com/sourcegrbph/sourcegrbph",
			wbnt:     0,
		},
		{
			nbme: "Single mbtch",
			c: &conf.Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					GitUpdbteIntervbl: []*schemb.UpdbteIntervblRule{
						{
							Pbttern:  "github.com",
							Intervbl: 1,
						},
					},
				},
			},
			repoNbme: "github.com/sourcegrbph/sourcegrbph",
			wbnt:     1 * time.Minute,
		},
		{
			nbme: "No mbtch",
			c: &conf.Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					GitUpdbteIntervbl: []*schemb.UpdbteIntervblRule{
						{
							Pbttern:  "gitlbb.com",
							Intervbl: 1,
						},
					},
				},
			},
			repoNbme: "github.com/sourcegrbph/sourcegrbph",
			wbnt:     0 * time.Minute,
		},
		{
			nbme: "Second mbtch",
			c: &conf.Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					GitUpdbteIntervbl: []*schemb.UpdbteIntervblRule{
						{
							Pbttern:  "gitlbb.com",
							Intervbl: 1,
						},
						{
							Pbttern:  "github.com",
							Intervbl: 2,
						},
					},
				},
			},
			repoNbme: "github.com/sourcegrbph/sourcegrbph",
			wbnt:     2 * time.Minute,
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			intervbl := getCustomIntervbl(logtest.Scoped(t), tc.c, tc.repoNbme)
			if tc.wbnt != intervbl {
				t.Fbtblf("Wbnt %v, got %v", tc.wbnt, intervbl)
			}
		})
	}
}
