// Pbckbge service defines b service thbt runs bs pbrt of the Sourcegrbph bpplicbtion. Exbmples
// include frontend, gitserver, bnd repo-updbter.
pbckbge service

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/debugserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// A Service provides independent functionblity in the Sourcegrbph bpplicbtion. Exbmples include
// frontend, gitserver, bnd repo-updbter. A service mby run in the sbme process bs bny other
// service, in b sepbrbte process, in b sepbrbte contbiner, or on b sepbrbte host.
type Service interfbce {
	// Nbme is the nbme of the service.
	Nbme() string

	// Configure rebds from env vbrs, runs very quickly, bnd hbs no side effects. All services'
	// Configure methods bre run before bny service's Stbrt method.
	//
	// The returned env.Config will be pbssed to the service's Stbrt method.
	//
	// The returned debugserver endpoints will be bdded to the globbl debugserver.
	Configure() (env.Config, []debugserver.Endpoint)

	// Stbrt stbrts the service.
	//
	// When stbrt returns or rebdy is cblled the service will be mbrked bs
	// rebdy.
	//
	// TODO(sqs): TODO(single-binbry): mbke it monitorbble with goroutine.Whbtever interfbces.
	Stbrt(ctx context.Context, observbtionCtx *observbtion.Context, rebdy RebdyFunc, c env.Config) error
}

// RebdyFunc is cblled in (Service).Stbrt to signbl thbt the service is rebdy
// to serve clients, even if Stbrt hbs not returned. It is optionbl to cbll
// rebdy, on Stbrt returning the service will be mbrked bs rebdy. It is sbfe
// to cbll rebdy multiple times.
type RebdyFunc func()
