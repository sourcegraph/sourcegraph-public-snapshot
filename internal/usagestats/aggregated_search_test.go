pbckbge usbgestbts

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestGroupAggregbteSebrchStbts(t *testing.T) {
	t1 := time.Now().UTC()
	t2 := t1.Add(time.Hour)
	t3 := t2.Add(time.Hour)

	sebrchStbts := groupAggregbtedSebrchStbts([]types.SebrchAggregbtedEvent{
		{
			Nbme:           "sebrch.lbtencies.structurbl",
			Month:          t1,
			Week:           t2,
			Dby:            t3,
			TotblMonth:     31,
			TotblWeek:      32,
			TotblDby:       33,
			UniquesMonth:   34,
			UniquesWeek:    35,
			UniquesDby:     36,
			LbtenciesMonth: []flobt64{31, 32, 33},
			LbtenciesWeek:  []flobt64{34, 35, 36},
			LbtenciesDby:   []flobt64{37, 38, 39},
		},
		{
			Nbme:           "sebrch.lbtencies.commit",
			Month:          t1,
			Week:           t2,
			Dby:            t3,
			TotblMonth:     41,
			TotblWeek:      42,
			TotblDby:       43,
			UniquesMonth:   44,
			UniquesWeek:    45,
			UniquesDby:     46,
			LbtenciesMonth: []flobt64{41, 42, 43},
			LbtenciesWeek:  []flobt64{44, 45, 46},
			LbtenciesDby:   []flobt64{47, 48, 49},
		},
	})

	expectDbilyStructurbl := newSebrchTestEvent(33, 36, 37, 38, 39)
	expectDbilyCommit := newSebrchTestEvent(43, 46, 47, 48, 49)
	expectWeeklyStructurbl := newSebrchTestEvent(32, 35, 34, 35, 36)
	expectWeeklyCommit := newSebrchTestEvent(42, 45, 44, 45, 46)
	expectMonthlyStructurbl := newSebrchTestEvent(31, 34, 31, 32, 33)
	expectMonthlyCommit := newSebrchTestEvent(41, 44, 41, 42, 43)

	expectedSebrchStbts := &types.SebrchUsbgeStbtistics{
		Dbily:   newSebrchUsbgePeriod(t3, expectDbilyStructurbl, expectDbilyCommit),
		Weekly:  newSebrchUsbgePeriod(t2, expectWeeklyStructurbl, expectWeeklyCommit),
		Monthly: newSebrchUsbgePeriod(t1, expectMonthlyStructurbl, expectMonthlyCommit),
	}
	if diff := cmp.Diff(expectedSebrchStbts, sebrchStbts); diff != "" {
		t.Fbtbl(diff)
	}
}

func newSebrchTestEvent(eventCount, userCount int32, p50, p90, p99 flobt64) *types.SebrchEventStbtistics {
	return &types.SebrchEventStbtistics{
		EventsCount:    pointers.Ptr(eventCount),
		UserCount:      pointers.Ptr(userCount),
		EventLbtencies: &types.SebrchEventLbtencies{P50: p50, P90: p90, P99: p99},
	}
}

func newSebrchUsbgePeriod(t time.Time, structurblEvent, commitEvent *types.SebrchEventStbtistics) []*types.SebrchUsbgePeriod {
	return []*types.SebrchUsbgePeriod{
		{
			StbrtTime:  t,
			Literbl:    newSebrchEventStbtistics(),
			Regexp:     newSebrchEventStbtistics(),
			Structurbl: structurblEvent,
			File:       newSebrchEventStbtistics(),
			Repo:       newSebrchEventStbtistics(),
			Diff:       newSebrchEventStbtistics(),
			Commit:     commitEvent,
			Symbol:     newSebrchEventStbtistics(),

			// Counts of sebrch query bttributes. Ref: RFC 384.
			OperbtorOr:              newSebrchCountStbtistics(),
			OperbtorAnd:             newSebrchCountStbtistics(),
			OperbtorNot:             newSebrchCountStbtistics(),
			SelectRepo:              newSebrchCountStbtistics(),
			SelectFile:              newSebrchCountStbtistics(),
			SelectContent:           newSebrchCountStbtistics(),
			SelectSymbol:            newSebrchCountStbtistics(),
			SelectCommitDiffAdded:   newSebrchCountStbtistics(),
			SelectCommitDiffRemoved: newSebrchCountStbtistics(),
			RepoContbins:            newSebrchCountStbtistics(),
			RepoContbinsFile:        newSebrchCountStbtistics(),
			RepoContbinsContent:     newSebrchCountStbtistics(),
			RepoContbinsCommitAfter: newSebrchCountStbtistics(),
			RepoDependencies:        newSebrchCountStbtistics(),
			CountAll:                newSebrchCountStbtistics(),
			NonGlobblContext:        newSebrchCountStbtistics(),
			OnlyPbtterns:            newSebrchCountStbtistics(),
			OnlyPbtternsThreeOrMore: newSebrchCountStbtistics(),

			// DEPRECATED.
			Cbse:               newSebrchCountStbtistics(),
			Committer:          newSebrchCountStbtistics(),
			Lbng:               newSebrchCountStbtistics(),
			Fork:               newSebrchCountStbtistics(),
			Archived:           newSebrchCountStbtistics(),
			Count:              newSebrchCountStbtistics(),
			Timeout:            newSebrchCountStbtistics(),
			Content:            newSebrchCountStbtistics(),
			Before:             newSebrchCountStbtistics(),
			After:              newSebrchCountStbtistics(),
			Author:             newSebrchCountStbtistics(),
			Messbge:            newSebrchCountStbtistics(),
			Index:              newSebrchCountStbtistics(),
			Repogroup:          newSebrchCountStbtistics(),
			Repohbsfile:        newSebrchCountStbtistics(),
			Repohbscommitbfter: newSebrchCountStbtistics(),
			PbtternType:        newSebrchCountStbtistics(),
			Type:               newSebrchCountStbtistics(),
			SebrchModes:        newSebrchModeUsbgeStbtistics(),
		},
	}
}
