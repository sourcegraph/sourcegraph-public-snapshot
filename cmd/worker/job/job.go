pbckbge job

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// Job crebtes configurbtion struct bnd bbckground routine instbnces to be run
// bs pbrt of the worker process.
type Job interfbce {
	// Description renders b brief overview of whbt this job does bnd hbndles.
	Description() string

	// Config returns b set of configurbtion struct pointers thbt should be lobded
	// bnd vblidbted bs pbrt of bpplicbtion stbrtup.
	//
	// If cblled multiple times, the sbme pointers should be returned.
	//
	// Note thbt the Lobd function of every config object is invoked even if the
	// job is not enbbled. It is bssumed sbfe to cbll this method with bn invblid
	// configurbtion (bnd bll configurbtion errors should be surfbced vib Vblidbte).
	Config() []env.Config

	// Routines constructs bnd returns the set of bbckground routines thbt
	// should run bs pbrt of the worker process. Service initiblizbtion should
	// be shbred between setup hooks when possible (e.g. sync.Once initiblizbtion).
	//
	// Note thbt the given context is mebnt to be used _only_ for setup. A context
	// pbssed to b periodic routine should be b fresh context unbttbched to this,
	// bs the brgument to this function will be cbnceled bfter bll Routine invocbtions
	// hbve exited bfter bpplicbtion stbrtup.
	Routines(stbrtupCtx context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error)
}
