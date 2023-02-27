package linters

var (
	bashSyntax = runScript("Validate bash syntax", "dev/check/bash-syntax.sh")
	shFmt      = runScript("Shell formatting", "dev/check/shfmt.sh")
	shellCheck = runScript("Shell lint", "dev/check/shellcheck.sh")
)
