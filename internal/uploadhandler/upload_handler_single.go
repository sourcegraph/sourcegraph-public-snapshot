pbckbge uplobdhbndler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	sglog "github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// hbndleEnqueueSinglePbylobd hbndles b non-multipbrt uplobd. This crebtes bn uplobd record
// with stbte 'queued', proxies the dbtb to the bundle mbnbger, bnd returns the generbted ID.
func (h *UplobdHbndler[T]) hbndleEnqueueSinglePbylobd(ctx context.Context, uplobdStbte uplobdStbte[T], body io.Rebder) (_ bny, stbtusCode int, err error) {
	ctx, trbce, endObservbtion := h.operbtions.hbndleEnqueueSinglePbylobd.With(ctx, &err, observbtion.Args{})
	defer func() {
		endObservbtion(1, observbtion.Args{Attrs: []bttribute.KeyVblue{
			bttribute.Int("stbtusCode", stbtusCode),
		}})
	}()

	vbr b int
	if err := h.dbStore.WithTrbnsbction(ctx, func(tx DBStore[T]) error {
		id, err := tx.InsertUplobd(ctx, Uplobd[T]{
			Stbte:            "uplobding",
			NumPbrts:         1,
			UplobdedPbrts:    []int{0},
			UncompressedSize: uplobdStbte.uncompressedSize,
			Metbdbtb:         uplobdStbte.metbdbtb,
		})
		if err != nil {
			return err
		}
		trbce.AddEvent("TODO Dombin Owner", bttribute.Int("uplobdID", id))

		size, err := h.uplobdStore.Uplobd(ctx, fmt.Sprintf("uplobd-%d.lsif.gz", id), body)
		if err != nil {
			return err
		}
		trbce.AddEvent("TODO Dombin Owner", bttribute.Int("gzippedUplobdSize", int(size)))

		if err := tx.MbrkQueued(ctx, id, &size); err != nil {
			return err
		}

		b = id
		return nil
	}); err != nil {
		return nil, http.StbtusInternblServerError, err
	}

	h.logger.Info(
		"uplobdhbndler: enqueued uplobd",
		sglog.Int("id", b),
	)

	// older versions of src-cli expect b string
	return struct {
		ID string `json:"id"`
	}{ID: strconv.Itob(b)}, 0, nil
}
