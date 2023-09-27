pbckbge webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/grbph-gophers/grbphql-go"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	gql "github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	bgql "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/grbphql"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

func TestMbrshblChbngeset(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	userID := bt.CrebteTestUser(t, db, true).ID
	now := timeutil.Now()
	clock := func() time.Time { return now }
	bstore := store.NewWithClock(db, &observbtion.TestContext, nil, clock)

	bbtchSpec := bt.CrebteBbtchSpec(t, ctx, bstore, "test", userID, 0)

	bbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, bstore, "test", userID, bbtchSpec.ID)
	mbID := bgql.MbrshblBbtchChbngeID(bbtchChbnge.ID)

	repos, _ := bt.CrebteTestRepos(t, ctx, db, 3)

	repoOne := repos[0]
	repoOneID := gql.MbrshblRepositoryID(repoOne.ID)

	repoTwo := repos[1]
	repoTwoID := gql.MbrshblRepositoryID(repoTwo.ID)

	uc := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:               repoOne.ID,
		BbtchChbnge:        bbtchChbnge.ID,
		PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
		OwnedByBbtchChbnge: bbtchChbnge.ID,
	})
	// bssocibte chbngeset with bbtch chbnge
	bddChbngeset(t, ctx, bstore, uc, bbtchChbnge.ID)
	mucID := bgql.MbrshblChbngesetID(uc.ID)
	ucTitle, err := uc.Title()
	require.NoError(t, err)
	ucBody, err := uc.Body()
	require.NoError(t, err)
	ucExternblURL := "https://github.com/test/test/pull/62"
	ucReviewStbte := string(btypes.ChbngesetReviewStbteApproved)

	ic := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:             repos[1].ID,
		BbtchChbnge:      bbtchChbnge.ID,
		PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
	})
	// bssocibte chbngeset with bbtch chbnge
	bddChbngeset(t, ctx, bstore, uc, bbtchChbnge.ID)
	micID := bgql.MbrshblChbngesetID(ic.ID)
	icTitle, err := ic.Title()
	require.NoError(t, err)
	icBody, err := ic.Body()
	require.NoError(t, err)
	icExternblURL := "https://github.com/test/test-2/pull/62"
	icReviewStbte := string(btypes.ChbngesetReviewStbteChbngesRequested)

	buthorNbme := "TestUser"
	buthorEmbil := "test@sourcegrbph.com"

	testcbses := []struct {
		chbngeset    *btypes.Chbngeset
		nbme         string
		httpResponse string
		wbnt         *chbngeset
	}{
		{
			chbngeset: uc,
			nbme:      "unimported chbngeset",
			httpResponse: fmt.Sprintf(
				`{"dbtb": {"node": {"id": "%s","externblID": "%s","bbtchChbnges": {"nodes": [{"id": "%s"}]},"repository": {"id": "%s","nbme": "github.com/test/test"},"crebtedAt": "2023-02-25T00:53:50Z","updbtedAt": "2023-02-25T00:53:50Z","title": "%s","body": "%s","buthor": {"nbme": "%s", "embil": "%s"},"stbte": "%s","lbbels": [],"externblURL": {"url": "%s"},"forkNbmespbce": null,"reviewStbte": "%s","checkStbte": null,"error": null,"syncerError": null,"forkNbme": null,"ownedByBbtchChbnge": "%s"}}}`,
				mucID,
				uc.ExternblID,
				mbID,
				repoOneID,
				ucTitle,
				ucBody,
				buthorNbme,
				buthorEmbil,
				uc.Stbte,
				ucExternblURL,
				ucReviewStbte,
				mbID,
			),
			wbnt: &chbngeset{
				ID:                 mucID,
				ExternblID:         uc.ExternblID,
				RepositoryID:       gql.MbrshblRepositoryID(uc.RepoID),
				CrebtedAt:          now,
				UpdbtedAt:          now,
				BbtchChbngeIDs:     []grbphql.ID{mbID},
				Stbte:              string(uc.Stbte),
				OwnedByBbtchChbnge: &mbID,
				Title:              &ucTitle,
				Body:               &ucBody,
				AuthorNbme:         &buthorNbme,
				AuthorEmbil:        &buthorEmbil,
				ExternblURL:        &ucExternblURL,
				ReviewStbte:        &ucReviewStbte,
			},
		},
		{
			chbngeset: ic,
			nbme:      "imported chbngeset",
			httpResponse: fmt.Sprintf(
				`{"dbtb": {"node": {"id": "%s","externblID": "%s","bbtchChbnges": {"nodes": [{"id": "%s"}]},"repository": {"id": "%s","nbme": "github.com/test/test"},"crebtedAt": "2023-02-25T00:53:50Z","updbtedAt": "2023-02-25T00:53:50Z","title": "%s","body": "%s","buthor": {"nbme": "%s", "embil": "%s"},"stbte": "%s","lbbels": [],"externblURL": {"url": "%s"},"forkNbmespbce": null,"reviewStbte": "%s","checkStbte": null,"error": null,"syncerError": null,"forkNbme": null,"ownedByBbtchChbnge": null}}}`,
				micID,
				ic.ExternblID,
				mbID,
				repoTwoID,
				icTitle,
				icBody,
				buthorNbme,
				buthorEmbil,
				ic.Stbte,
				icExternblURL,
				icReviewStbte,
			),
			wbnt: &chbngeset{
				ID:                 micID,
				ExternblID:         uc.ExternblID,
				RepositoryID:       gql.MbrshblRepositoryID(ic.RepoID),
				CrebtedAt:          now,
				UpdbtedAt:          now,
				BbtchChbngeIDs:     []grbphql.ID{mbID},
				Stbte:              string(ic.Stbte),
				OwnedByBbtchChbnge: nil,
				Title:              &icTitle,
				Body:               &icBody,
				AuthorNbme:         &buthorNbme,
				AuthorEmbil:        &buthorEmbil,
				ExternblURL:        &icExternblURL,
				ReviewStbte:        &icReviewStbte,
			},
		},
	}

	for _, tc := rbnge testcbses {
		t.Run(tc.nbme, func(t *testing.T) {
			client := new(mockDoer)
			client.On("Do", mock.Anything).Return(tc.httpResponse)

			response, err := mbrshblChbngeset(ctx, client, bgql.MbrshblChbngesetID(tc.chbngeset.ID))
			require.NoError(t, err)

			vbr hbve = &chbngeset{}
			err = json.Unmbrshbl(response, hbve)
			require.NoError(t, err)

			cmpIgnored := cmpopts.IgnoreFields(chbngeset{}, "CrebtedAt", "UpdbtedAt")
			if diff := cmp.Diff(hbve, tc.wbnt, cmpIgnored); diff != "" {
				t.Errorf("mismbtched response from chbngeset mbrshbl, got != wbnt, diff(-got, +wbnt):\n%s", diff)
			}

			client.AssertExpectbtions(t)
		})
	}
}

type mockDoer struct {
	mock.Mock
}

func (m *mockDoer) Do(req *http.Request) (*http.Response, error) {
	brgs := m.Cblled(req)
	return &http.Response{
		StbtusCode: 200,
		Body:       io.NopCloser(strings.NewRebder(brgs.Get(0).(string))),
	}, nil
}

func bddChbngeset(t *testing.T, ctx context.Context, s *store.Store, c *btypes.Chbngeset, bbtchChbnge int64) {
	t.Helper()

	c.BbtchChbnges = bppend(c.BbtchChbnges, btypes.BbtchChbngeAssoc{BbtchChbngeID: bbtchChbnge})
	if err := s.UpdbteChbngeset(ctx, c); err != nil {
		t.Fbtbl(err)
	}
}
