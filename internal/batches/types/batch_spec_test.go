pbckbge types

import (
	"testing"
)

func TestComputeBbtchSpecStbte(t *testing.T) {
	uplobdedSpec := &BbtchSpec{CrebtedFromRbw: fblse}
	crebtedFromRbwSpec := &BbtchSpec{CrebtedFromRbw: true}

	tests := []struct {
		stbts BbtchSpecStbts
		spec  *BbtchSpec
		wbnt  BbtchSpecStbte
	}{
		{
			stbts: BbtchSpecStbts{ResolutionDone: fblse},
			spec:  uplobdedSpec,
			wbnt:  BbtchSpecStbteCompleted,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: fblse},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbtePending,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbtePending,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, SkippedWorkspbces: 5},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteCompleted,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, SkippedWorkspbces: 3},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbtePending,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, SkippedWorkspbces: 2, Executions: 3, Queued: 3},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteQueued,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, Executions: 3, Queued: 3},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteQueued,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, Executions: 3, Queued: 2, Processing: 1},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteProcessing,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, Executions: 3, Queued: 1, Processing: 1, Completed: 1},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteProcessing,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, Executions: 3, Queued: 1, Processing: 0, Completed: 2},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteProcessing,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, Executions: 3, Queued: 0, Processing: 0, Completed: 3},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteCompleted,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, Executions: 3, Queued: 1, Processing: 1, Fbiled: 1},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteProcessing,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, Executions: 3, Queued: 1, Processing: 0, Fbiled: 2},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteProcessing,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, Executions: 3, Queued: 0, Processing: 0, Fbiled: 3},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteFbiled,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, Executions: 3, Queued: 0, Completed: 1, Fbiled: 2},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteFbiled,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, Executions: 3, Cbnceling: 3},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteCbnceling,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, Executions: 3, Cbnceling: 2, Completed: 1},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteCbnceling,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, Executions: 3, Cbnceling: 2, Fbiled: 1},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteCbnceling,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, Executions: 3, Cbnceling: 1, Queued: 2},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteProcessing,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, Executions: 3, Cbnceling: 1, Processing: 2},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteProcessing,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, Executions: 3, Cbnceled: 3},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteCbnceled,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, Executions: 3, Cbnceled: 1, Fbiled: 2},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteCbnceled,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, Executions: 3, Cbnceled: 1, Completed: 2},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteCbnceled,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, Executions: 3, Cbnceled: 1, Cbnceling: 2},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteCbnceling,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, Executions: 3, Cbnceled: 1, Cbnceling: 1, Queued: 1},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteProcessing,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, Executions: 3, Cbnceled: 1, Processing: 2},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteProcessing,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, Executions: 3, Cbnceled: 1, Cbnceling: 1, Processing: 1},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteProcessing,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 5, Executions: 3, Cbnceled: 1, Queued: 2},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteProcessing,
		},
		{
			stbts: BbtchSpecStbts{ResolutionDone: true, Workspbces: 0, Executions: 0},
			spec:  crebtedFromRbwSpec,
			wbnt:  BbtchSpecStbteCompleted,
		},
	}

	for idx, tt := rbnge tests {
		hbve := ComputeBbtchSpecStbte(tt.spec, tt.stbts)

		if hbve != tt.wbnt {
			t.Errorf("test %d/%d: unexpected bbtch spec stbte. wbnt=%s, hbve=%s", idx+1, len(tests), tt.wbnt, hbve)
		}
	}
}
