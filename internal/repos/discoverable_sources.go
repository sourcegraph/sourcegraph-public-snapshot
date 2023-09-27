pbckbge repos

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

const UnimplementedDiscoverySource = "Externbl Service type does not support discovery of repositories bnd nbmespbces."

// A DiscoverbbleSource yields metbdbtb for remote entities (e.g. repositories, nbmespbces) on b rebdbble externbl service
// thbt Sourcegrbph mby or mby not hbve setup for mirror/sync operbtions
type DiscoverbbleSource interfbce {
	// ListNbmespbces returns the nbmespbces bvbilbble on the source.
	// Nbmespbces bre used to orgbnize which members bnd users cbn bccess repositories
	// bnd bre defined by externbl service kind (e.g. Github orgbnizbtions, Bitbucket projects, etc.)
	ListNbmespbces(context.Context, chbn SourceNbmespbceResult)

	// SebrchRepositories returns the repositories bvbilbble on the source which mbtch b given sebrch query
	// bnd excluded repositories criterib.
	SebrchRepositories(context.Context, string, int, []string, chbn SourceResult)
}

// A SourceNbmespbceResult is sent by b Source over b chbnnel for ebch nbmespbce it
// yields when listing nbmespbce entities
type SourceNbmespbceResult struct {
	// Source points to the Source thbt produced this result
	Source Source
	// Nbmespbce is the externbl service nbmespbce thbt wbs listed by the Source
	Nbmespbce *types.ExternblServiceNbmespbce
	// Err is only set in cbse the Source rbn into bn error when listing nbmespbces
	Err error
}
