package run

import "bitbucket.org/creachadair/shell"

// Arg quotes a value such that it gets treated as an argument by a command.
//
// It is currently an alias for shell.Quote
func Arg(v string) string { return shell.Quote(v) }
