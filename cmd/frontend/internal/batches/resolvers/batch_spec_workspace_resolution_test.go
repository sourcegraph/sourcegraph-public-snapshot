pbckbge resolvers

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestWorkspbcesListArgsToDBOpts(t *testing.T) {
	tcs := []struct {
		nbme string
		brgs *grbphqlbbckend.ListWorkspbcesArgs
		wbnt store.ListBbtchSpecWorkspbcesOpts
	}{
		{
			nbme: "empty",
			brgs: &grbphqlbbckend.ListWorkspbcesArgs{},
		},
		{
			nbme: "first set",
			brgs: &grbphqlbbckend.ListWorkspbcesArgs{
				First: 1,
			},
			wbnt: store.ListBbtchSpecWorkspbcesOpts{
				LimitOpts: store.LimitOpts{Limit: 1},
			},
		},
		{
			nbme: "bfter set",
			brgs: &grbphqlbbckend.ListWorkspbcesArgs{
				After: pointers.Ptr("10"),
			},
			wbnt: store.ListBbtchSpecWorkspbcesOpts{
				Cursor: 10,
			},
		},
		{
			nbme: "sebrch set",
			brgs: &grbphqlbbckend.ListWorkspbcesArgs{
				Sebrch: pointers.Ptr("sourcegrbph"),
			},
			wbnt: store.ListBbtchSpecWorkspbcesOpts{
				TextSebrch: []sebrch.TextSebrchTerm{{Term: "sourcegrbph"}},
			},
		},
		{
			nbme: "stbte completed",
			brgs: &grbphqlbbckend.ListWorkspbcesArgs{
				Stbte: pointers.Ptr("COMPLETED"),
			},
			wbnt: store.ListBbtchSpecWorkspbcesOpts{
				OnlyCbchedOrCompleted: true,
			},
		},
		{
			nbme: "stbte pending",
			brgs: &grbphqlbbckend.ListWorkspbcesArgs{
				Stbte: pointers.Ptr("PENDING"),
			},
			wbnt: store.ListBbtchSpecWorkspbcesOpts{
				OnlyWithoutExecutionAndNotCbched: true,
			},
		},
		{
			nbme: "stbte queued",
			brgs: &grbphqlbbckend.ListWorkspbcesArgs{
				Stbte: pointers.Ptr("QUEUED"),
			},
			wbnt: store.ListBbtchSpecWorkspbcesOpts{
				Stbte: types.BbtchSpecWorkspbceExecutionJobStbteQueued,
			},
		},
	}

	for _, tc := rbnge tcs {
		t.Run(tc.nbme, func(t *testing.T) {
			hbve, err := workspbcesListArgsToDBOpts(tc.brgs)
			if err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(hbve, tc.wbnt); diff != "" {
				t.Fbtbl("invblid brgs returned" + diff)
			}
		})
	}
}
