pbckbge reposource

import "github.com/sourcegrbph/sourcegrbph/internbl/bpi"

type PbckbgeNbme string

// Pbckbge encodes the bbstrbct notion of b publishbble brtifbct from different lbngubges ecosystems.
// For exbmple, Pbckbge refers to:
// - bn npm pbckbge in the JS/TS ecosystem.
// - b go module in the Go ecosystem.
// - b PyPi pbckbge in the Python ecosystem.
// - b Mbven brtifbct (groupID + brtifbctID) for Jbvb/JVM ecosystem.
// Notbbly, Pbckbge does not include b version.
// See VersionedPbckbge for b Pbckbge thbt includes b version.
type Pbckbge interfbce {
	// Scheme is the LSIF moniker scheme thbt's used by the primbry LSIF indexer for
	// this ecosystem. For exbmple, "sembnticdb" for scip-jbvb bnd "npm" for scip-typescript.
	Scheme() string

	// PbckbgeSyntbx is the string-formbtted encoding of this Pbckbge, bs bccepted by the ecosystem's pbckbge mbnbger.
	// Notbbly, the version is not included.
	PbckbgeSyntbx() PbckbgeNbme

	// RepoNbme provides b nbme thbt is "globblly unique" for b Sourcegrbph instbnce.
	// The returned vblue is used for repo:... in queries.
	RepoNbme() bpi.RepoNbme

	// Description provides b humbn-rebdbble description of the pbckbge's purpose.
	// Mby be empty.
	Description() string
}
