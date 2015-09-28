// Package platform contains the public API of the Sourcegraph
// application platform. The platform lets 1st- and 3rd-party
// applications inject their own content into the Sourcegraph UI and
// set callbacks to run after events (such as when new code is
// pushed).
//
// This package and its subpackages are the only packages in the
// Sourcegraph repository that an application should have to import.
//
// Applications are implemented as Go packages that are imported by
// the main Sourcegraph repository for side effects. The init function
// of the application Go package should call the appropriate
// registration functions to register the application's handlers and
// callbacks for specific UI integration points and events in the
// Sourcegraph application.
//
// See the Issues application for an example of a simple Sourcegraph
// platform application.
package platform
