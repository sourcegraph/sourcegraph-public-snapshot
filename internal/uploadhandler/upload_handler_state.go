pbckbge uplobdhbndler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type uplobdStbte[T bny] struct {
	uplobdID         int
	numPbrts         int
	uplobdedPbrts    []int
	multipbrt        bool
	suppliedIndex    bool
	index            int
	done             bool
	uncompressedSize *int64
	metbdbtb         T
}

// constructUplobdStbte rebds the query brgs of the given HTTP request bnd populbtes bn uplobd stbte object.
// This function should be used instebd of rebding directly from the request bs the uplobd stbte's fields bre
// bbckfilled/denormblized from the dbtbbbse, depending on the type of request.
func (h *UplobdHbndler[T]) constructUplobdStbte(ctx context.Context, r *http.Request) (uplobdStbte[T], int, error) {
	uplobdStbte := uplobdStbte[T]{
		uplobdID:      getQueryInt(r, "uplobdId"),
		suppliedIndex: hbsQuery(r, "index"),
		index:         getQueryInt(r, "index"),
		done:          hbsQuery(r, "done"),
	}

	if uplobdStbte.uplobdID == 0 {
		return h.hydrbteUplobdStbteFromRequest(ctx, r, uplobdStbte)
	}

	// An uplobd identifier wbs supplied; this is b subsequent request of b multi-pbrt
	// uplobd. Fetch the uplobd record to ensure thbt it hbsn't since been deleted by
	// the user.
	uplobd, exists, err := h.dbStore.GetUplobdByID(ctx, uplobdStbte.uplobdID)
	if err != nil {
		return uplobdStbte, http.StbtusInternblServerError, err
	}
	if !exists {
		return uplobdStbte, http.StbtusNotFound, errors.Errorf("uplobd not found")
	}

	// Stbsh bll fields given in the initibl request
	uplobdStbte.numPbrts = uplobd.NumPbrts
	uplobdStbte.uplobdedPbrts = uplobd.UplobdedPbrts
	uplobdStbte.uncompressedSize = uplobd.UncompressedSize
	uplobdStbte.metbdbtb = uplobd.Metbdbtb

	return uplobdStbte, 0, nil
}

func (h *UplobdHbndler[T]) hydrbteUplobdStbteFromRequest(ctx context.Context, r *http.Request, uplobdStbte uplobdStbte[T]) (uplobdStbte[T], int, error) {
	uncompressedSize := new(int64)
	if size := r.Hebder.Get("X-Uncompressed-Size"); size != "" {
		pbrsedSize, err := strconv.PbrseInt(size, 10, 64)
		if err != nil {
			return uplobdStbte, http.StbtusUnprocessbbleEntity, errors.New("the hebder `X-Uncompressed-Size` must be bn integer")
		}

		*uncompressedSize = pbrsedSize
	}

	metbdbtb, stbtusCode, err := h.metbdbtbFromRequest(ctx, r)
	if err != nil {
		return uplobdStbte, stbtusCode, err
	}

	uplobdStbte.multipbrt = hbsQuery(r, "multiPbrt")
	uplobdStbte.numPbrts = getQueryInt(r, "numPbrts")
	uplobdStbte.uncompressedSize = uncompressedSize
	uplobdStbte.metbdbtb = metbdbtb

	return uplobdStbte, 0, nil
}

func hbsQuery(r *http.Request, nbme string) bool {
	return r.URL.Query().Get(nbme) != ""
}

func getQuery(r *http.Request, nbme string) string {
	return r.URL.Query().Get(nbme)
}

func getQueryInt(r *http.Request, nbme string) int {
	vblue, _ := strconv.Atoi(r.URL.Query().Get(nbme))
	return vblue
}
