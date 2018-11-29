// Package confdefaults contains default configuration files for various
// deployment types.
//
// It is a separate package so that users of pkg/conf do not indirectly import
// pkg/db/confdb, which we have a linter to protect against.
package confdefaults

import (
	"github.com/sourcegraph/sourcegraph/pkg/conf/conftypes"
)

// DevAndTesting is the default configuration applied to dev instances of
// Sourcegraph, as well as what is used by default during Go testing.
//
// Tests that wish to use a specific configuration should use conf.Mock.
//
// TODO(slimsag): Unified
var DevAndTesting = conftypes.RawUnified{
	Critical: `{
	"auth.providers": [
		{
			"type": "builtin",
			"allowSignup": true
		}
	],
}`,
	Site: `{}`,
}

// DockerContainer is the default configuration applied to Docker
// single-container instances of Sourcegraph.
//
// TODO(slimsag): Unified
/*
	// The default site configuration.
	defaultSiteConfig := schema.SiteConfiguration{
		// TODO(slimsag): Unified
		//AuthProviders: []schema.AuthProviders{
		//	{Builtin: &schema.BuiltinAuthProvider{Type: "builtin"}},
		//},
		MaxReposToSearch: 50,

		DisablePublicRepoRedirects: true,
	}
*/
var DockerContainer = conftypes.RawUnified{
	Critical: `{
	"auth.providers": [
		{
			"type": "builtin",
			"allowSignup": true
		}
	]
}`,
	Site: `{
	"disablePublicRepoRedirects": true,
	"maxReposToSearch": 50,
}`,
}

// Cluster is the default configuration applied to Cluster instances of
// Sourcegraph.
//
// TODO(slimsag): Unified
var Cluster = conftypes.RawUnified{
	Critical: `{
	"auth.providers": [
		{
			"type": "builtin",
			"allowSignup": true
		}
	],
}`,
	Site: `{}`,
}

// Default is the default for *this* deployment type. It is populated by
// pkg/conf at init time.
//
// In the case of a migration from an old Sourcegraph version to 3.0, this is
// not strictly one of the declared defaults in this package but rather may be
// defaults from a user's old configuration.
//
// TODO(slimsag): Remove legacy warning above after 3.0 (not 3.0-preview).
var Default conftypes.RawUnified
