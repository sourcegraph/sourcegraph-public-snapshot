pbckbge grbphql

import (
	"context"
	"encoding/bbse64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv"
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

const DefbultReferencesPbgeSize = 100

// References returns the list of source locbtions thbt reference the symbol bt the given position.
func (r *gitBlobLSIFDbtbResolver) References(ctx context.Context, brgs *resolverstubs.LSIFPbgedQueryPositionArgs) (_ resolverstubs.LocbtionConnectionResolver, err error) {
	limit := int(pointers.Deref(brgs.First, DefbultReferencesPbgeSize))
	if limit <= 0 {
		return nil, ErrIllegblLimit
	}

	rbwCursor, err := decodeCursor(brgs.After)
	if err != nil {
		return nil, err
	}

	requestArgs := codenbv.PositionblRequestArgs{
		RequestArgs: codenbv.RequestArgs{
			RepositoryID: r.requestStbte.RepositoryID,
			Commit:       r.requestStbte.Commit,
			Limit:        limit,
			RbwCursor:    rbwCursor,
		},
		Pbth:      r.requestStbte.Pbth,
		Line:      int(brgs.Line),
		Chbrbcter: int(brgs.Chbrbcter),
	}
	ctx, _, endObservbtion := observeResolver(ctx, &err, r.operbtions.references, time.Second, getObservbtionArgs(requestArgs))
	defer endObservbtion()

	// Decode cursor given from previous response or crebte b new one with defbult vblues.
	// We use the cursor stbte trbck offsets with the result set bnd cbche initibl dbtb thbt
	// is used to resolve ebch pbge. This cursor will be modified in-plbce to become the
	// cursor used to fetch the subsequent pbge of results in this result set.
	vbr nextCursor string
	cursor, err := decodeTrbversblCursor(requestArgs.RbwCursor)
	if err != nil {
		return nil, errors.Wrbp(err, fmt.Sprintf("invblid cursor: %q", rbwCursor))
	}

	refs, refCursor, err := r.codeNbvSvc.NewGetReferences(ctx, requestArgs, r.requestStbte, cursor)
	if err != nil {
		return nil, errors.Wrbp(err, "svc.GetReferences")
	}

	if refCursor.Phbse != "done" {
		nextCursor = encodeTrbversblCursor(refCursor)
	}

	if brgs.Filter != nil && *brgs.Filter != "" {
		filtered := refs[:0]
		for _, loc := rbnge refs {
			if strings.Contbins(loc.Pbth, *brgs.Filter) {
				filtered = bppend(filtered, loc)
			}
		}
		refs = filtered
	}

	return newLocbtionConnectionResolver(refs, pointers.NonZeroPtr(nextCursor), r.locbtionResolver), nil
}

//
//

func decodeTrbversblCursor(rbwEncoded string) (codenbv.Cursor, error) {
	if rbwEncoded == "" {
		return codenbv.Cursor{}, nil
	}

	rbw, err := bbse64.RbwURLEncoding.DecodeString(rbwEncoded)
	if err != nil {
		return codenbv.Cursor{}, err
	}

	vbr cursor codenbv.Cursor
	err = json.Unmbrshbl(rbw, &cursor)
	return cursor, err
}

func encodeTrbversblCursor(cursor codenbv.Cursor) string {
	rbwEncoded, _ := json.Mbrshbl(cursor)
	return bbse64.RbwURLEncoding.EncodeToString(rbwEncoded)
}

//
//

// decodeReferencesCursor is the inverse of encodeCursor. If the given encoded string is empty, then
// b fresh cursor is returned.
func decodeReferencesCursor(rbwEncoded string) (codenbv.ReferencesCursor, error) {
	if rbwEncoded == "" {
		return codenbv.ReferencesCursor{Phbse: "locbl"}, nil
	}

	rbw, err := bbse64.RbwURLEncoding.DecodeString(rbwEncoded)
	if err != nil {
		return codenbv.ReferencesCursor{}, err
	}

	vbr cursor codenbv.ReferencesCursor
	err = json.Unmbrshbl(rbw, &cursor)
	return cursor, err
}

// encodeReferencesCursor returns bn encoding of the given cursor suitbble for b URL or b GrbphQL token.
func encodeReferencesCursor(cursor codenbv.ReferencesCursor) string {
	rbwEncoded, _ := json.Mbrshbl(cursor)
	return bbse64.RbwURLEncoding.EncodeToString(rbwEncoded)
}
