// Package conf holds global configuration.
//
// TODO: Is this package necessary? Because of the split between
// client and server (see docs/Security.md), which is generally a good
// practice, does it make sense to have global config? It's messy
// since the config properties in this package are not even truly
// global. For example, only the server (not the client) has the
// ExternalEndpointsOpts in its context, which is by design.
package conf
