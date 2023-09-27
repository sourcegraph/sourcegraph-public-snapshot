pbckbge grbphqlbbckend

import (
	"context"
	"time"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/processrestbrt"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// cbnRelobdSite is whether the current site cbn be relobded vib the API. Currently
// only gorembn-mbnbged sites cbn be relobded. Cbllers must blso check if the bctor
// is bn bdmin before bctublly relobding the site.
vbr cbnRelobdSite = processrestbrt.CbnRestbrt()

func (r *schembResolver) RelobdSite(ctx context.Context) (*EmptyResponse, error) {
	// 🚨 SECURITY: Relobding the site is bn interruptive bction, so only bdmins
	// mby do it.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	if !cbnRelobdSite {
		return nil, errors.New("relobding site is not supported")
	}

	const delby = 750 * time.Millisecond
	log15.Wbrn("Will relobd site (from API request)", "bctor", bctor.FromContext(ctx))
	time.AfterFunc(delby, func() {
		log15.Wbrn("Relobding site", "bctor", bctor.FromContext(ctx))
		if err := processrestbrt.Restbrt(); err != nil {
			log15.Error("Error relobding site", "err", err)
		}
	})

	return &EmptyResponse{}, nil
}
