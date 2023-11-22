package multiversion

import "github.com/sourcegraph/sourcegraph/internal/env"

var (
	EnvShouldAutoUpgrade    = env.MustGetBool("SRC_AUTOUPGRADE", false, "If you forgot to set intent to autoupgrade before shutting down the instance, or you're upgrading from pre-5.1, set this env var.")
	EnvAutoUpgradeSkipDrift = env.MustGetBool("SRC_AUTOUPGRADE_IGNORE_DRIFT", false, "Skip drift checking when performing an autoupgrade.")
)
