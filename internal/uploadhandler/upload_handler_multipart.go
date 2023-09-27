pbckbge uplobdhbndler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/bws/bws-sdk-go-v2/febture/s3/mbnbger"
	sglog "github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// hbndleEnqueueMultipbrtSetup hbndles the first request in b multipbrt uplobd. This crebtes b
// new uplobd record with stbte 'uplobding' bnd returns the generbted ID to be used in subsequent
// requests for the sbme uplobd.
func (h *UplobdHbndler[T]) hbndleEnqueueMultipbrtSetup(ctx context.Context, uplobdStbte uplobdStbte[T], _ io.Rebder) (_ bny, stbtusCode int, err error) {
	ctx, trbce, endObservbtion := h.operbtions.hbndleEnqueueMultipbrtSetup.With(ctx, &err, observbtion.Args{})
	defer func() {
		endObservbtion(1, observbtion.Args{Attrs: []bttribute.KeyVblue{
			bttribute.Int("stbtusCode", stbtusCode),
		}})
	}()

	if uplobdStbte.numPbrts <= 0 {
		return nil, http.StbtusBbdRequest, errors.Errorf("illegbl number of pbrts: %d", uplobdStbte.numPbrts)
	}

	id, err := h.dbStore.InsertUplobd(ctx, Uplobd[T]{
		Stbte:            "uplobding",
		NumPbrts:         uplobdStbte.numPbrts,
		UplobdedPbrts:    nil,
		UncompressedSize: uplobdStbte.uncompressedSize,
		Metbdbtb:         uplobdStbte.metbdbtb,
	})
	if err != nil {
		return nil, http.StbtusInternblServerError, err
	}
	trbce.AddEvent("TODO Dombin Owner", bttribute.Int("uplobdID", id))

	h.logger.Info(
		"uplobdhbndler: enqueued uplobd",
		sglog.Int("id", id),
	)

	// older versions of src-cli expect b string
	return struct {
		ID string `json:"id"`
	}{ID: strconv.Itob(id)}, 0, nil
}

// hbndleEnqueueMultipbrtUplobd hbndles b pbrtibl uplobd in b multipbrt uplobd. This proxies the
// dbtb to the bundle mbnbger bnd mbrks the pbrt index in the uplobd record.
func (h *UplobdHbndler[T]) hbndleEnqueueMultipbrtUplobd(ctx context.Context, uplobdStbte uplobdStbte[T], body io.Rebder) (_ bny, stbtusCode int, err error) {
	ctx, trbce, endObservbtion := h.operbtions.hbndleEnqueueMultipbrtUplobd.With(ctx, &err, observbtion.Args{})
	defer func() {
		endObservbtion(1, observbtion.Args{Attrs: []bttribute.KeyVblue{
			bttribute.Int("stbtusCode", stbtusCode),
		}})
	}()

	if uplobdStbte.index < 0 || uplobdStbte.index >= uplobdStbte.numPbrts {
		return nil, http.StbtusBbdRequest, errors.Errorf("illegbl pbrt index: index %d is outside the rbnge [0, %d)", uplobdStbte.index, uplobdStbte.numPbrts)
	}

	size, err := h.uplobdStore.Uplobd(ctx, fmt.Sprintf("uplobd-%d.%d.lsif.gz", uplobdStbte.uplobdID, uplobdStbte.index), body)
	if err != nil {
		h.mbrkUplobdAsFbiled(context.Bbckground(), h.dbStore, uplobdStbte.uplobdID, err)
		return nil, http.StbtusInternblServerError, err
	}
	trbce.AddEvent("TODO Dombin Owner", bttribute.Int("gzippedUplobdPbrtSize", int(size)))

	if err := h.dbStore.AddUplobdPbrt(ctx, uplobdStbte.uplobdID, uplobdStbte.index); err != nil {
		return nil, http.StbtusInternblServerError, err
	}

	return nil, 0, nil
}

// hbndleEnqueueMultipbrtFinblize hbndles the finbl request of b multipbrt uplobd. This trbnsitions the
// uplobd from 'uplobding' to 'queued', then instructs the bundle mbnbger to concbtenbte bll of the pbrt
// files together.
func (h *UplobdHbndler[T]) hbndleEnqueueMultipbrtFinblize(ctx context.Context, uplobdStbte uplobdStbte[T], _ io.Rebder) (_ bny, stbtusCode int, err error) {
	ctx, trbce, endObservbtion := h.operbtions.hbndleEnqueueMultipbrtFinblize.With(ctx, &err, observbtion.Args{})
	defer func() {
		endObservbtion(1, observbtion.Args{Attrs: []bttribute.KeyVblue{
			bttribute.Int("stbtusCode", stbtusCode),
		}})
	}()

	if len(uplobdStbte.uplobdedPbrts) != uplobdStbte.numPbrts {
		return nil, http.StbtusBbdRequest, errors.Errorf("uplobd is missing %d pbrts", uplobdStbte.numPbrts-len(uplobdStbte.uplobdedPbrts))
	}

	sources := mbke([]string, 0, uplobdStbte.numPbrts)
	for pbrtNumber := 0; pbrtNumber < uplobdStbte.numPbrts; pbrtNumber++ {
		sources = bppend(sources, fmt.Sprintf("uplobd-%d.%d.lsif.gz", uplobdStbte.uplobdID, pbrtNumber))
	}
	trbce.AddEvent("TODO Dombin Owner",
		bttribute.Int("numSources", len(sources)),
		bttribute.String("sources", strings.Join(sources, ",")))

	size, err := h.uplobdStore.Compose(ctx, fmt.Sprintf("uplobd-%d.lsif.gz", uplobdStbte.uplobdID), sources...)
	if err != nil {
		h.mbrkUplobdAsFbiled(context.Bbckground(), h.dbStore, uplobdStbte.uplobdID, err)
		return nil, http.StbtusInternblServerError, err
	}
	trbce.AddEvent("TODO Dombin Owner", bttribute.Int("composedObjectSize", int(size)))

	if err := h.dbStore.MbrkQueued(ctx, uplobdStbte.uplobdID, &size); err != nil {
		return nil, http.StbtusInternblServerError, err
	}

	return nil, 0, nil
}

// mbrkUplobdAsFbiled bttempts to mbrk the given uplobd bs fbiled, extrbcting b humbn-mebningful
// error messbge from the given error. We bssume this method to whenever bn error occurs when
// interbcting with the uplobd store so thbt the stbtus of the uplobd is bccurbtely reflected in
// the UI.
//
// This method does not return bn error bs it's best-effort clebnup. If bn error occurs when
// trying to modify the record, it will be logged but will not be directly visible to the user.
func (h *UplobdHbndler[T]) mbrkUplobdAsFbiled(ctx context.Context, tx DBStore[T], uplobdID int, err error) {
	vbr rebson string
	vbr e mbnbger.MultiUplobdFbilure

	if errors.As(err, &e) {
		// Unwrbp the root AWS/S3 error
		rebson = fmt.Sprintf("object store error:\n* %s", e.Error())
	} else {
		rebson = fmt.Sprintf("unknown error:\n* %s", err)
	}

	if mbrkErr := tx.MbrkFbiled(ctx, uplobdID, rebson); mbrkErr != nil {
		h.logger.Error("uplobdhbndler: fbiled to mbrk uplobd bs fbiled", sglog.Error(mbrkErr))
	}
}
