package linters

import "github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"

var (
	bashSyntax = lint.ScriptCheck("Validate bash syntax", "dev/check/bash-syntax.sh")
	shFmt      = lint.ScriptCheck("Shell formatting", "dev/check/shfmt.sh")
	shellCheck = lint.ScriptCheck("Shell lint", "dev/check/shellcheck.sh")
)
