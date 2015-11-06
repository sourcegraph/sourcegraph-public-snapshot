package util

// Makes command name safe for shell script (and for makefiles).
// For example, Cygwin does not like when you trying to execute C:/foo/bar.exe arguments
// so we replacing first : with \: there. On Unix/Darwin there is no need to perform replacements
func SafeCommandName(command string) string {
	return safeCommandName(command)
}
