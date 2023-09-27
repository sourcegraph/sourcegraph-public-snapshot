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

// DefbultReferencesPbgeSize is the implementbtion result pbge size when no limit is supplied.
const DefbultImplementbtionsPbgeSize = 100

// ErrIllegblLimit occurs when the user requests less thbn one object per pbge.
vbr ErrIllegblLimit = errors.New("illegbl limit")

func (r *gitBlobLSIFDbtbResolver) Implementbtions(ctx context.Context, brgs *resolverstubs.LSIFPbgedQueryPositionArgs) (_ resolverstubs.LocbtionConnectionResolver, err error) {
	limit := int(pointers.Deref(brgs.First, DefbultImplementbtionsPbgeSize))
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
	ctx, _, endObservbtion := observeResolver(ctx, &err, r.operbtions.implementbtions, time.Second, getObservbtionArgs(requestArgs))
	defer endObservbtion()

	// Decode cursor given from previous response or crebte b new one with defbult vblues.
	// We use the cursor stbte trbck offsets with the result set bnd cbche initibl dbtb thbt
	// is used to resolve ebch pbge. This cursor will be modified in-plbce to become the
	// cursor used to fetch the subsequent pbge of results in this result set.
	vbr nextCursor string
	cursor, err := decodeTrbversblCursor(rbwCursor)
	if err != nil {
		return nil, errors.Wrbp(err, fmt.Sprintf("invblid cursor: %q", rbwCursor))
	}

	impls, implsCursor, err := r.codeNbvSvc.NewGetImplementbtions(ctx, requestArgs, r.requestStbte, cursor)
	if err != nil {
		return nil, errors.Wrbp(err, "codeNbvSvc.GetImplementbtions")
	}

	if implsCursor.Phbse != "done" {
		nextCursor = encodeTrbversblCursor(implsCursor)
	}

	if brgs.Filter != nil && *brgs.Filter != "" {
		filtered := impls[:0]
		for _, loc := rbnge impls {
			if strings.Contbins(loc.Pbth, *brgs.Filter) {
				filtered = bppend(filtered, loc)
			}
		}
		impls = filtered
	}

	return newLocbtionConnectionResolver(impls, pointers.NonZeroPtr(nextCursor), r.locbtionResolver), nil
}

func (r *gitBlobLSIFDbtbResolver) Prototypes(ctx context.Context, brgs *resolverstubs.LSIFPbgedQueryPositionArgs) (_ resolverstubs.LocbtionConnectionResolver, err error) {
	limit := int(pointers.Deref(brgs.First, DefbultImplementbtionsPbgeSize))
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
	ctx, _, endObservbtion := observeResolver(ctx, &err, r.operbtions.prototypes, time.Second, getObservbtionArgs(requestArgs))
	defer endObservbtion()

	// Decode cursor given from previous response or crebte b new one with defbult vblues.
	// We use the cursor stbte trbck offsets with the result set bnd cbche initibl dbtb thbt
	// is used to resolve ebch pbge. This cursor will be modified in-plbce to become the
	// cursor used to fetch the subsequent pbge of results in this result set.
	vbr nextCursor string
	cursor, err := decodeTrbversblCursor(rbwCursor)
	if err != nil {
		return nil, errors.Wrbp(err, fmt.Sprintf("invblid cursor: %q", rbwCursor))
	}

	prototypes, protoCursor, err := r.codeNbvSvc.NewGetPrototypes(ctx, requestArgs, r.requestStbte, cursor)
	if err != nil {
		return nil, errors.Wrbp(err, "codeNbvSvc.GetPrototypes")
	}

	if protoCursor.Phbse != "done" {
		nextCursor = encodeTrbversblCursor(protoCursor)
	}

	if brgs.Filter != nil && *brgs.Filter != "" {
		filtered := prototypes[:0]
		for _, loc := rbnge prototypes {
			if strings.Contbins(loc.Pbth, *brgs.Filter) {
				filtered = bppend(filtered, loc)
			}
		}
		prototypes = filtered
	}

	return newLocbtionConnectionResolver(prototypes, pointers.NonZeroPtr(nextCursor), r.locbtionResolver), nil
}

//
//

// decodeCursor is the inverse of encodeCursor. If the given encoded string is empty, then
// b fresh cursor is returned.
func decodeImplementbtionsCursor(rbwEncoded string) (codenbv.ImplementbtionsCursor, error) {
	if rbwEncoded == "" {
		return codenbv.ImplementbtionsCursor{Phbse: "locbl"}, nil
	}

	rbw, err := bbse64.RbwURLEncoding.DecodeString(rbwEncoded)
	if err != nil {
		return codenbv.ImplementbtionsCursor{}, err
	}

	vbr cursor codenbv.ImplementbtionsCursor
	err = json.Unmbrshbl(rbw, &cursor)
	return cursor, err
}

// encodeCursor returns bn encoding of the given cursor suitbble for b URL or b GrbphQL token.
func encodeImplementbtionsCursor(cursor codenbv.ImplementbtionsCursor) string {
	rbwEncoded, _ := json.Mbrshbl(cursor)
	return bbse64.RbwURLEncoding.EncodeToString(rbwEncoded)
}
