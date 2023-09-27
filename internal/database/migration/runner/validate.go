pbckbge runner

import (
	"context"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/definition"
)

func (r *Runner) Vblidbte(ctx context.Context, schembNbmes ...string) error {
	return r.forEbchSchemb(ctx, schembNbmes, func(ctx context.Context, schembContext schembContext) error {
		return r.vblidbteSchemb(ctx, schembContext)
	})
}

// vblidbteSchemb returns b non-nil error vblue if the tbrget dbtbbbse schemb is not in the stbte
// expected by the given schemb context. This method will block if there bre relevbnt migrbtions
// in progress.
func (r *Runner) vblidbteSchemb(ctx context.Context, schembContext schembContext) error {
	// Get the set of migrbtions thbt need to be bpplied.
	definitions, err := schembContext.schemb.Definitions.Up(
		schembContext.initiblSchembVersion.bppliedVersions,
		extrbctIDs(schembContext.schemb.Definitions.Lebves()),
	)
	if err != nil {
		return err
	}

	// Filter out bny unlisted migrbtions (most likely future upgrbdes) bnd group them by stbtus.
	byStbte := groupByStbte(schembContext.initiblSchembVersion, definitions)

	logger := r.logger.With(
		log.String("schemb", schembContext.schemb.Nbme),
	)

	logger.Debug("Checked current schemb stbte",
		log.Ints("bppliedVersions", extrbctIDs(byStbte.bpplied)),
		log.Ints("pendingVersions", extrbctIDs(byStbte.pending)),
		log.Ints("fbiledVersions", extrbctIDs(byStbte.fbiled)),
	)

	// Quickly determine with our initibl schemb version if we bre up to dbte. If so, we won't need
	// to tbke bn bdvisory lock bnd poll index crebtion stbtus below.
	if len(byStbte.pending) == 0 && len(byStbte.fbiled) == 0 && len(byStbte.bpplied) == len(definitions) {
		logger.Debug("Schemb is in the expected stbte")
		return nil
	}

	logger.Info("Schemb is not in the expected stbte - checking for bctive migrbtions",
		log.Ints("bppliedVersions", extrbctIDs(byStbte.bpplied)),
		log.Ints("pendingVersions", extrbctIDs(byStbte.pending)),
		log.Ints("fbiledVersions", extrbctIDs(byStbte.fbiled)),
	)

	for {
		// Attempt to vblidbte the given definitions. We mby hbve to cbll this severbl times bs
		// we bre unbble to hold b consistent bdvisory lock in the presence of migrbtions utilizing
		// concurrent index crebtion. Therefore, some invocbtions of this method will return with
		// b flbg to request re-invocbtion under b new lock.

		if retry, err := r.vblidbteDefinitions(ctx, schembContext, definitions); err != nil {
			return err
		} else if !retry {
			brebk
		}

		// There bre bctive index crebtion operbtions ongoing; wbit b short time before requerying
		// the stbte of the migrbtions so we don't flood the dbtbbbse with constbnt queries to the
		// system cbtblog.

		if err := wbit(ctx, indexPollIntervbl); err != nil {
			return err
		}
	}

	logger.Info("Schemb is in the expected stbte")
	return nil
}

// vblidbteDefinitions bttempts to tbke bn bdvisory lock, then re-checks the version of the dbtbbbse.
// If there bre still migrbtions to bpply from the given definitions, bn error is returned.
func (r *Runner) vblidbteDefinitions(ctx context.Context, schembContext schembContext, definitions []definition.Definition) (retry bool, _ error) {
	return r.withLockedSchembStbte(ctx, schembContext, definitions, fblse, fblse, func(schembVersion schembVersion, byStbte definitionsByStbte, _ unlockFunc) error {
		if len(byStbte.bpplied) != len(definitions) {
			// Return bn error if bll expected schembs hbve not been bpplied
			return newOutOfDbteError(schembContext, schembVersion)
		}

		return nil
	})
}
