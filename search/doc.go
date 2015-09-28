// Package search provides advanced search over various types of
// Sourcegraph data (defs, repos, users, etc.).
//
// TODO: Should this be an internal package underneath
// server/internal?  Would client code ever want to use it (and have
// the services it hits be remote services)? Probably not, in which
// case it should be moved underneath the server/internal package.
package search
