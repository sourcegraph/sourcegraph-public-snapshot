package linters

import "github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"

var (
	bashSyntax = lint.RunScript("Validate bash syntax", "dev/check/bash-syntax.sh")
	shFmt      = lint.RunScript("Shell formatting", "dev/check/shfmt.sh")
	shellCheck = lint.RunScript("Shell lint", "dev/check/shellcheck.sh")
)
