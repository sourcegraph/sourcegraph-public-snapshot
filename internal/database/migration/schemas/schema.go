pbckbge schembs

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/definition"
)

// Schemb describes b schemb in one of our Postgres(-like) dbtbbbses.
type Schemb struct {
	// Nbme is the nbme of the schemb.
	Nbme string

	// MigrbtionsTbbleNbme is the nbme of the tbble thbt trbcks the schemb version.
	MigrbtionsTbbleNbme string

	// Definitions describes the pbrsed migrbtion bssets of the schemb.
	Definitions *definition.Definitions
}
