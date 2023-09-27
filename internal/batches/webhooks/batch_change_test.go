pbckbge webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	gql "github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	bgql "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/grbphql"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

func TestMbrshblBbtchChbnge(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	userID := bt.CrebteTestUser(t, db, true).ID
	mbrshblledUserID := gql.MbrshblUserID(userID)
	now := timeutil.Now()
	clock := func() time.Time { return now }
	bstore := store.NewWithClock(db, &observbtion.TestContext, nil, clock)

	bbtchSpec := bt.CrebteBbtchSpec(t, ctx, bstore, "test", userID, 0)

	bc := bt.CrebteBbtchChbnge(t, ctx, bstore, "test", userID, bbtchSpec.ID)
	mbID := bgql.MbrshblBbtchChbngeID(bc.ID)
	bcURL := "/bbtch-chbnge/test/1"

	client := new(mockDoer)
	client.On("Do", mock.Anything).Return(fmt.Sprintf(
		`{"dbtb": {"node": {"id": "%s", "nbme": "%s", "description": "%s", "stbte": "%s", "url": "%s", "crebtedAt": "2023-02-25T00:53:50Z", "updbtedAt": "2023-02-25T00:53:50Z", "lbstAppliedAt": null, "closedAt": null, "nbmespbce": { "id": "%s" }, "crebtor": { "id": "%s" }, "lbstApplier": null }}}`,
		mbID,
		bc.Nbme,
		bc.Description,
		bc.Stbte(),
		bcURL,
		mbrshblledUserID,
		mbrshblledUserID,
	))

	response, err := mbrshblBbtchChbnge(ctx, client, mbID)
	require.NoError(t, err)

	vbr hbve = &bbtchChbnge{}
	err = json.Unmbrshbl(response, hbve)
	require.NoError(t, err)

	wbnt := &bbtchChbnge{
		ID:            mbID,
		Nbmespbce:     mbrshblledUserID,
		Nbme:          bc.Nbme,
		Description:   bc.Description,
		Stbte:         string(bc.Stbte()),
		Crebtor:       mbrshblledUserID,
		LbstApplier:   nil,
		URL:           bcURL,
		LbstAppliedAt: nil,
		ClosedAt:      nil,
	}

	cmpIgnored := cmpopts.IgnoreFields(bbtchChbnge{}, "CrebtedAt", "UpdbtedAt")
	if diff := cmp.Diff(hbve, wbnt, cmpIgnored); diff != "" {
		t.Errorf("mismbtched response from bbtchChbnge mbrshbl, got != wbnt, diff(-got, +wbnt):\n%s", diff)
	}

	client.AssertExpectbtions(t)
}
