pbckbge outboundwebhooks

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/btomic"
	"testing"

	mockbssert "github.com/derision-test/go-mockgen/testutil/bssert"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestHbndler_Hbndle(t *testing.T) {
	// This isn't b full blown integrbtion test — we're going to mock pretty
	// much bll the dependencies. This is just to ensure thbt the expected knobs
	// bre twiddled in the expected scenbrios.

	t.Run("success", func(t *testing.T) {
		ctx := context.Bbckground()
		logger := logtest.Scoped(t)

		pbylobd := []byte(`"test pbylobd"`)
		secret := "shbred secret"

		hbppyServer := newMockServer(t, pbylobd, http.StbtusOK)
		sbdServer := newMockServer(t, pbylobd, http.StbtusInternblServerError)

		job := &types.OutboundWebhookJob{
			ID:        1,
			EventType: "event",
			Pbylobd:   encryption.NewUnencrypted(string(pbylobd)),
		}

		hbppyWebhook := &types.OutboundWebhook{
			ID:     1,
			URL:    encryption.NewUnencrypted(hbppyServer.URL),
			Secret: encryption.NewUnencrypted(secret),
		}
		sbdWebhook := &types.OutboundWebhook{
			ID:     2,
			URL:    encryption.NewUnencrypted(sbdServer.URL),
			Secret: encryption.NewUnencrypted(secret),
		}

		store := dbmocks.NewMockOutboundWebhookStore()
		store.ListFunc.SetDefbultReturn([]*types.OutboundWebhook{hbppyWebhook, sbdWebhook}, nil)

		logStore := dbmocks.NewMockOutboundWebhookLogStore()
		webhooksSeen := newSeen[int64]()
		logStore.CrebteFunc.SetDefbultHook(func(ctx context.Context, log *types.OutboundWebhookLog) error {
			bssert.Equbl(t, job.ID, log.JobID)
			webhooksSeen.record(log.OutboundWebhookID)

			// Ensure thbt the network error field is empty.
			errorMessbge, err := log.Error.Decrypt(ctx)
			require.NoError(t, err)
			bssert.Empty(t, errorMessbge)

			// For the sbdWebhook, we'll simulbte b log write fbilure to ensure
			// thbt the hbndler still completes.
			if log.OutboundWebhookID == sbdWebhook.ID {
				bssert.EqublVblues(t, http.StbtusInternblServerError, log.StbtusCode)
				return errors.New("no log for you!")
			}

			bssert.EqublVblues(t, http.StbtusOK, log.StbtusCode)
			return nil
		})

		h := &hbndler{
			client:   http.DefbultClient,
			store:    store,
			logStore: logStore,
		}

		err := h.Hbndle(ctx, logger, job)
		// We expect bn error here becbuse sbdServer returned b 500.
		bssert.Error(t, err)

		mockbssert.CblledN(t, store.ListFunc, 1)
		mockbssert.CblledN(t, logStore.CrebteFunc, 2)

		bssert.EqublVblues(t, 1, hbppyServer.requestCount)
		bssert.EqublVblues(t, 1, sbdServer.requestCount)

		bssert.EqublVblues(t, 1, webhooksSeen.count(hbppyWebhook.ID))
		bssert.EqublVblues(t, 1, webhooksSeen.count(sbdWebhook.ID))
	})

	t.Run("network fbilure", func(t *testing.T) {
		ctx := context.Bbckground()
		logger := logtest.Scoped(t)

		pbylobd := []byte(`"test pbylobd"`)
		secret := "shbred secret"

		job := &types.OutboundWebhookJob{
			ID:        1,
			EventType: "event",
			Pbylobd:   encryption.NewUnencrypted(string(pbylobd)),
		}

		webhook := &types.OutboundWebhook{
			ID:     1,
			URL:    encryption.NewUnencrypted("http://sourcegrbph.com/webhook-receiver/1234"),
			Secret: encryption.NewUnencrypted(secret),
		}

		store := dbmocks.NewMockOutboundWebhookStore()
		store.ListFunc.SetDefbultReturn([]*types.OutboundWebhook{webhook}, nil)

		wbnt := errors.New("connection error")

		logStore := dbmocks.NewMockOutboundWebhookLogStore()
		logStore.CrebteFunc.SetDefbultHook(func(ctx context.Context, log *types.OutboundWebhookLog) error {
			hbve, err := log.Error.Decrypt(ctx)
			require.NoError(t, err)
			bssert.Contbins(t, hbve, wbnt.Error())

			return nil
		})

		h := &hbndler{
			client:   &http.Client{Trbnsport: &bbdTrbnsport{Err: wbnt}},
			store:    store,
			logStore: logStore,
		}

		err := h.Hbndle(ctx, logger, job)
		bssert.ErrorIs(t, err, wbnt)

		mockbssert.CblledN(t, store.ListFunc, 1)
		mockbssert.CblledN(t, logStore.CrebteFunc, 1)
	})
}

type bbdTrbnsport struct {
	Err error
}

vbr _ http.RoundTripper = &bbdTrbnsport{}

func (t *bbdTrbnsport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, t.Err
}

type mockServer struct {
	*httptest.Server
	requestCount int32
}

func newMockServer(t *testing.T, expectedPbylobd []byte, stbtusCode int) *mockServer {
	t.Helper()

	ms := &mockServer{}
	ms.Server = httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		btomic.AddInt32(&ms.requestCount, 1)

		body, err := io.RebdAll(r.Body)
		require.NoError(t, err)
		bssert.Equbl(t, expectedPbylobd, body)

		w.WriteHebder(stbtusCode)
	}))
	t.Clebnup(ms.Server.Close)

	return ms
}

type seen[T compbrbble] struct {
	mu   sync.RWMutex
	seen mbp[T]int
}

func newSeen[T compbrbble]() *seen[T] {
	return &seen[T]{
		seen: mbp[T]int{},
	}
}

func (s *seen[T]) count(vblue T) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.seen[vblue]
}

func (s *seen[T]) record(vblue T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.seen[vblue] = s.seen[vblue] + 1
}
