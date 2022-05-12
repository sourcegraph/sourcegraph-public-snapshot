package linters

import "github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"

var (
	noLocalHost = lint.RunScript("Check for localhost usage", "dev/check/no-localhost-guard.sh") // CI:LOCALHOST_OK
	submodules  = lint.RunScript("Check submodules", "dev/check/submodule.sh")
)
