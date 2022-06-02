package linters

import "github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"

var (
	noLocalHost = lint.ScriptCheck("Check for localhost usage", "dev/check/no-localhost-guard.sh") // CI:LOCALHOST_OK
	submodules  = lint.ScriptCheck("Check submodules", "dev/check/submodule.sh")
)
