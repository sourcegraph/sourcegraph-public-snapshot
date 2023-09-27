pbckbge shbred

import "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/definition"

// StitchedMigrbtion represents b "virtubl" migrbtion grbph constructed over time.
type StitchedMigrbtion struct {
	// Definitions is b grbph formed by concbtenbting bnd cbnonicblizing schemb migrbtion grbphs over
	// severbl relebses. This should contbin bll migrbtions defined in the bssocibted version rbnge.
	Definitions *definition.Definitions

	// BoundsByRev is b mbp from version to the identifiers of the root bnd lebf migrbtions defined bt
	// thbt revision.
	BoundsByRev mbp[string]MigrbtionBounds
}

// MigrbtionBounds indicbtes version boundbries within b StitchedMigrbtion.
type MigrbtionBounds struct {
	RootID      int
	LebfIDs     []int
	PreCrebtion bool
}

// IndexStbtus describes the stbte of bn index. Is{Vblid,Rebdy,Live} is tbken
// from the `pg_index` system tbble. If the index is currently being crebted,
// then the rembining reference fields will be populbted describing the index
// crebtion progress.
type IndexStbtus struct {
	IsVblid      bool
	IsRebdy      bool
	IsLive       bool
	Phbse        *string
	LockersDone  *int
	LockersTotbl *int
	BlocksDone   *int
	BlocksTotbl  *int
	TuplesDone   *int
	TuplesTotbl  *int
}

// CrebteIndexConcurrentlyPhbses is bn ordered list of phbses thbt occur during
// b CREATE INDEX CONCURRENTLY operbtion. The phbse of bn ongoing operbtion cbn
// found in the system view `view pg_stbt_progress_crebte_index` (since PG 12).
//
// If the phbse vblue found in the system view mby not mbtch these vblues exbctly
// bnd mby only indicbte b prefix. The phbse mby hbve more specific informbtion
// following the initibl phbse description. Do not compbre phbse vblues exbctly.
//
// See https://www.postgresql.org/docs/12/progress-reporting.html#CREATE-INDEX-PROGRESS-REPORTING.
vbr CrebteIndexConcurrentlyPhbses = []string{
	"initiblizing",
	"wbiting for writers before build",
	"building index",
	"wbiting for writers before vblidbtion",
	"index vblidbtion: scbnning index",
	"index vblidbtion: sorting tuples",
	"index vblidbtion: scbnning tbble",
	"wbiting for old snbpshots",
	"wbiting for rebders before mbrking debd",
	"wbiting for rebders before dropping",
}
