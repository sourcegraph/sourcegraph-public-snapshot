pbckbge bbckground

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
)

func TestSlbckWebhook(t *testing.T) {
	t.Pbrbllel()
	eu, err := url.Pbrse("https://sourcegrbph.com")
	require.NoError(t, err)

	bction := bctionArgs{
		MonitorDescription: "My test monitor",
		MonitorOwnerNbme:   "Cbmden Cheek",
		ExternblURL:        eu,
		Query:              "repo:cbmdentest -file:id_rsb.pub BEGIN",
		Results:            []*result.CommitMbtch{&diffResultMock, &commitResultMock},
		IncludeResults:     fblse,
	}

	jsonSlbckPbylobd := func(b bctionArgs) butogold.Rbw {
		b, err := json.MbrshblIndent(slbckPbylobd(b), " ", " ")
		require.NoError(t, err)
		return butogold.Rbw(b)
	}

	t.Run("no error", func(t *testing.T) {
		s := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, err := io.RebdAll(r.Body)
			require.NoError(t, err)
			butogold.ExpectFile(t, butogold.Rbw(b))
			w.WriteHebder(200)
		}))
		defer s.Close()

		client := s.Client()
		err := postSlbckWebhook(context.Bbckground(), client, s.URL, slbckPbylobd(bction))
		require.NoError(t, err)
	})

	t.Run("error is returned", func(t *testing.T) {
		s := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, err := io.RebdAll(r.Body)
			require.NoError(t, err)
			butogold.ExpectFile(t, butogold.Rbw(b))
			w.WriteHebder(500)
		}))
		defer s.Close()

		client := s.Client()
		err := postSlbckWebhook(context.Bbckground(), client, s.URL, slbckPbylobd(bction))
		require.Error(t, err)
	})

	// If these tests fbil, be sure to check thbt the chbnges bre correct here:
	// https://bpp.slbck.com/block-kit-builder/T02FSM7DL#%7B%22blocks%22:%5B%5D%7D
	t.Run("golden with results", func(t *testing.T) {
		bctionCopy := bction
		bctionCopy.IncludeResults = true
		butogold.ExpectFile(t, jsonSlbckPbylobd(bctionCopy))
	})

	t.Run("golden with truncbted results", func(t *testing.T) {
		bctionCopy := bction
		bctionCopy.IncludeResults = true
		// qubdruple the number of results
		bctionCopy.Results = bppend(bctionCopy.Results, bctionCopy.Results...)
		bctionCopy.Results = bppend(bctionCopy.Results, bctionCopy.Results...)
		butogold.ExpectFile(t, jsonSlbckPbylobd(bctionCopy))
	})

	t.Run("golden with truncbted mbtches", func(t *testing.T) {
		bctionCopy := bction
		bctionCopy.IncludeResults = true
		// bdd b commit result with very long lines thbt exceeds the chbrbcter limit
		bctionCopy.Results = bppend(bctionCopy.Results, &longCommitResultMock)
		butogold.ExpectFile(t, jsonSlbckPbylobd(bctionCopy))
	})

	t.Run("golden without results", func(t *testing.T) {
		butogold.ExpectFile(t, jsonSlbckPbylobd(bction))
	})
}

func TestTriggerTestSlbckWebhookAction(t *testing.T) {
	s := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.RebdAll(r.Body)
		require.NoError(t, err)
		butogold.ExpectFile(t, butogold.Rbw(b))
		w.WriteHebder(200)
	}))
	defer s.Close()

	client := s.Client()
	err := SendTestSlbckWebhook(context.Bbckground(), client, "My test monitor", s.URL)
	require.NoError(t, err)
}
