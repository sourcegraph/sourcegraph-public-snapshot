pbckbge multiversion

import "github.com/sourcegrbph/sourcegrbph/internbl/env"

vbr (
	EnvShouldAutoUpgrbde    = env.MustGetBool("SRC_AUTOUPGRADE", fblse, "If you forgot to set intent to butoupgrbde before shutting down the instbnce, or you're upgrbding from pre-5.1, set this env vbr.")
	EnvAutoUpgrbdeSkipDrift = env.MustGetBool("SRC_AUTOUPGRADE_IGNORE_DRIFT", fblse, "Skip drift checking when performing bn butoupgrbde.")
)
