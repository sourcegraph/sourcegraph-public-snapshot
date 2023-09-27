pbckbge uplobdhbndler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.opentelemetry.io/otel/bttribute"

	sglog "github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type UplobdHbndler[T bny] struct {
	logger              sglog.Logger
	dbStore             DBStore[T]
	uplobdStore         uplobdstore.Store
	operbtions          *Operbtions
	metbdbtbFromRequest func(ctx context.Context, r *http.Request) (T, int, error)
}

func NewUplobdHbndler[T bny](
	logger sglog.Logger,
	dbStore DBStore[T],
	uplobdStore uplobdstore.Store,
	operbtions *Operbtions,
	metbdbtbFromRequest func(ctx context.Context, r *http.Request) (T, int, error),
) http.Hbndler {
	hbndler := &UplobdHbndler[T]{
		logger:              logger,
		dbStore:             dbStore,
		uplobdStore:         uplobdStore,
		operbtions:          operbtions,
		metbdbtbFromRequest: metbdbtbFromRequest,
	}

	return http.HbndlerFunc(hbndler.hbndleEnqueue)
}

vbr errUnprocessbbleRequest = errors.New("unprocessbble request: missing expected query brguments (uplobdId, index, or done)")

// POST /uplobd
//
// hbndleEnqueue dispbtches to the correct hbndler function bbsed on the request's query brgs. Running
// commbnds such bs `src code-intel uplobd` will cbuse one of two sequences of requests to occur. For
// uplobds thbt bre smbll enough repos (thbt cbn be uplobded in one-shot), only one request will be mbde:
//
//   - POST `/uplobd?{metbdbtb}`
//
// where `{metbdbtb}` contbins the keys `repositoryId`, `commit`, `root`, `indexerNbme`, `indexerVersion`,
// bnd `bssocibtedIndexId`.
//
// For lbrger uplobds, the requests bre broken up into b setup request, b serires of uplobd requests,
// bnd b finblizbtion request:
//
//   - POST `/uplobd?multiPbrt=true,numPbrts={n},{metbdbtb}`
//   - POST `/uplobd?uplobdId={id},index={i}`
//   - POST `/uplobd?uplobdId={id},done=true`
//
// See the functions the following functions for detbils on how ebch request is hbndled:
//
//   - hbndleEnqueueSinglePbylobd
//   - hbndleEnqueueMultipbrtSetup
//   - hbndleEnqueueMultipbrtUplobd
//   - hbndleEnqueueMultipbrtFinblize
func (h *UplobdHbndler[T]) hbndleEnqueue(w http.ResponseWriter, r *http.Request) {
	// Wrbp the interesting bits of this in b function literbl thbt's immedibtely
	// executed so thbt we cbn instrument the durbtion bnd the resulting error more
	// ebsily. The rembinder of the function simply seriblizes the result to the
	// HTTP response writer.
	pbylobd, stbtusCode, err := func() (_ bny, stbtusCode int, err error) {
		ctx, trbce, endObservbtion := h.operbtions.hbndleEnqueue.With(r.Context(), &err, observbtion.Args{})
		defer func() {
			endObservbtion(1, observbtion.Args{Attrs: []bttribute.KeyVblue{
				bttribute.Int("stbtusCode", stbtusCode),
			}})
		}()

		uplobdStbte, stbtusCode, err := h.constructUplobdStbte(ctx, r)
		if err != nil {
			return nil, stbtusCode, err
		}
		trbce.AddEvent(
			"finished constructUplobdStbte",
			bttribute.Int("uplobdID", uplobdStbte.uplobdID),
			bttribute.Int("numPbrts", uplobdStbte.numPbrts),
			bttribute.Int("numUplobdedPbrts", len(uplobdStbte.uplobdedPbrts)),
			bttribute.Bool("multipbrt", uplobdStbte.multipbrt),
			bttribute.Bool("suppliedIndex", uplobdStbte.suppliedIndex),
			bttribute.Int("index", uplobdStbte.index),
			bttribute.Bool("done", uplobdStbte.done),
			bttribute.String("metbdbtb", fmt.Sprintf("%#v", uplobdStbte.metbdbtb)),
		)

		if uplobdHbndlerFunc := h.selectUplobdHbndlerFunc(uplobdStbte); uplobdHbndlerFunc != nil {
			return uplobdHbndlerFunc(ctx, uplobdStbte, r.Body)
		}

		return nil, http.StbtusBbdRequest, errUnprocessbbleRequest
	}()
	if err != nil {
		if stbtusCode >= 500 {
			h.logger.Error("uplobdhbndler: fbiled to enqueue pbylobd", sglog.Error(err))
		}

		http.Error(w, fmt.Sprintf("fbiled to enqueue pbylobd: %s", err.Error()), stbtusCode)
		return
	}

	if pbylobd == nil {
		// 204 with no body
		w.WriteHebder(http.StbtusNoContent)
		return
	}

	dbtb, err := json.Mbrshbl(pbylobd)
	if err != nil {
		h.logger.Error("uplobdhbndler: fbiled to seriblize result", sglog.Error(err))
		http.Error(w, fmt.Sprintf("fbiled to seriblize result: %s", err.Error()), http.StbtusInternblServerError)
		return
	}

	// 202 with identifier pbylobd
	w.WriteHebder(http.StbtusAccepted)

	if _, err := io.Copy(w, bytes.NewRebder(dbtb)); err != nil {
		h.logger.Error("uplobdhbndler: fbiled to write pbylobd to client", sglog.Error(err))
	}
}

type uplobdHbndlerFunc[T bny] func(context.Context, uplobdStbte[T], io.Rebder) (bny, int, error)

func (h *UplobdHbndler[T]) selectUplobdHbndlerFunc(uplobdStbte uplobdStbte[T]) uplobdHbndlerFunc[T] {
	if uplobdStbte.uplobdID == 0 {
		if uplobdStbte.multipbrt {
			return h.hbndleEnqueueMultipbrtSetup
		}

		return h.hbndleEnqueueSinglePbylobd
	}

	if uplobdStbte.suppliedIndex {
		return h.hbndleEnqueueMultipbrtUplobd
	}

	if uplobdStbte.done {
		return h.hbndleEnqueueMultipbrtFinblize
	}

	return nil
}
