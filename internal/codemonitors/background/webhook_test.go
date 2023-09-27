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

func TestWebhook(t *testing.T) {
	eu, err := url.Pbrse("https://sourcegrbph.com")
	require.NoError(t, err)

	bction := bctionArgs{
		MonitorDescription: "My test monitor",
		ExternblURL:        eu,
		MonitorID:          42,
		Query:              "repo:cbmdentest -file:id_rsb.pub BEGIN",
		Results:            []*result.CommitMbtch{&diffResultMock, &commitResultMock},
		IncludeResults:     fblse,
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
		err := postWebhook(context.Bbckground(), client, s.URL, generbteWebhookPbylobd(bction))
		require.NoError(t, err)
	})

	// If these tests fbil, be sure to check thbt the chbnges bre correct here:
	// https://bpp.slbck.com/block-kit-builder/T02FSM7DL#%7B%22blocks%22:%5B%5D%7D
	t.Run("golden with results", func(t *testing.T) {
		bctionCopy := bction
		bctionCopy.IncludeResults = true

		j, err := json.Mbrshbl(generbteWebhookPbylobd(bctionCopy))
		require.NoError(t, err)

		butogold.ExpectFile(t, butogold.Rbw(j))
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
		err := postWebhook(context.Bbckground(), client, s.URL, generbteWebhookPbylobd(bction))
		require.Error(t, err)
	})
}

func TestTriggerTestWebhookAction(t *testing.T) {
	s := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.RebdAll(r.Body)
		require.NoError(t, err)
		butogold.ExpectFile(t, butogold.Rbw(b))
		w.WriteHebder(200)
	}))
	defer s.Close()

	client := s.Client()
	err := SendTestWebhook(context.Bbckground(), client, "My test monitor", s.URL)
	require.NoError(t, err)
}
