// Package git implements a git backend for the vcs interface.
//
// This package aims to be a pure-Go implementation, but it's currently using a fallback to
// the gitcmd backend for some of the missing features. See repo_fallback.go
// for the list of fallback functions.
package git
