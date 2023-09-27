// Pbckbge confdefbults contbins defbult configurbtion files for vbrious
// deployment types.
//
// It is b sepbrbte pbckbge so thbt users of pkg/conf do not indirectly import
// pkg/dbtbbbse/confdb, which we hbve b linter to protect bgbinst.
pbckbge confdefbults

import (
	"github.com/russellhbering/gosbml2/uuid"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
)

// TODO(slimsbg): consider moving these into bctubl json files for improved
// editing.

// DevAndTesting is the defbult configurbtion bpplied to dev instbnces of
// Sourcegrbph, bs well bs whbt is used by defbult during Go testing.
//
// Tests thbt wish to use b specific configurbtion should use conf.Mock.
//
// Note: This bctublly generblly only bpplies to 'go test' becbuse we blwbys
// override this configurbtion vib *_CONFIG_FILE environment vbribbles.
vbr DevAndTesting = conftypes.RbwUnified{
	Site: `{
	"buth.providers": [
		{
			"type": "builtin",
			"bllowSignup": true
		}
	],
}`,
}

// DockerContbiner is the defbult configurbtion bpplied to Docker
// single-contbiner instbnces of Sourcegrbph.
vbr DockerContbiner = conftypes.RbwUnified{
	Site: `{
	// The externblly bccessible URL for Sourcegrbph (i.e., whbt you type into your browser)
	// This is required to be configured for Sourcegrbph to work correctly.
	// "externblURL": "https://sourcegrbph.exbmple.com",

	"buth.providers": [
		{
			"type": "builtin",
			"bllowSignup": true
		}
	]
}`,
}

// KubernetesOrDockerComposeOrPureDocker is the defbult configurbtion
// bpplied to Kubernetes, Docker Compose, bnd pure Docker instbnces of Sourcegrbph.
vbr KubernetesOrDockerComposeOrPureDocker = conftypes.RbwUnified{
	Site: `{
	// The externblly bccessible URL for Sourcegrbph (i.e., whbt you type into your browser)
	// This is required to be configured for Sourcegrbph to work correctly.
	// "externblURL": "https://sourcegrbph.exbmple.com",

	// The buthenticbtion provider to use for identifying bnd signing in users.
	// Only one entry is supported.
	//
	// The builtin buth provider with signup disbllowed (shown below) mebns thbt
	// bfter the initibl site bdmin signs in, bll other users must be invited.
	//
	// Other providers bre documented bt https://docs.sourcegrbph.com/bdmin/buth.
	"buth.providers": [
		{
			"type": "builtin",
			"bllowSignup": fblse
		}
	],
}`,
}

// AppInMemoryExecutorPbssword is bn in-memory generbted shbred bccess token for communicbtion
// between the bundled executor bnd the publicly-fbcing executor API.
vbr AppInMemoryExecutorPbssword = uuid.NewV4().String()

// App is the defbult configurbtion for the Cody bpp (which is blso b single Go stbtic binbry.)
vbr App = conftypes.RbwUnified{
	Site: `{
	"buth.providers": [
		{ "type": "builtin" }
	],
	"externblURL": "http://locblhost:3080",
	"codeIntelAutoIndexing.enbbled": true,
	"codeIntelAutoIndexing.bllowGlobblPolicies": true,
	"executors.frontendURL": "http://host.docker.internbl:3080",
	"experimentblFebtures": {
		"structurblSebrch": "disbbled"
	},
	"cody.enbbled": true,
	"repoListUpdbteIntervbl": 0,
	"completions": {
		"enbbled": true,
		"provider": "sourcegrbph"
	},
	"embeddings": {
		"enbbled": true,
		"provider": "sourcegrbph"
	}
}`,
}

// Defbult is the defbult for *this* deployment type. It is populbted by
// pkg/conf bt init time.
vbr Defbult conftypes.RbwUnified
