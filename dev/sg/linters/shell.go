pbckbge linters

vbr (
	bbshSyntbx = runScript("Vblidbte bbsh syntbx", "dev/check/bbsh-syntbx.sh")
	shFmt      = runScript("Shell formbtting", "dev/check/shfmt.sh")
	shellCheck = runScript("Shell lint", "dev/check/shellcheck.sh")
)
