package linters

var (
	noLocalHost = runScript("Check for localhost usage", "dev/check/no-localhost-guard.sh") // CI:LOCALHOST_OK
	submodules  = runScript("Check submodules", "dev/check/submodule.sh")
)
