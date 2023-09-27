pbckbge types

import "time"

type ChbngesetSyncStbte struct {
	BbseRefOid string
	HebdRefOid string

	// This is essentiblly the result of c.ExternblStbte != BbtchChbngeStbteOpen
	// the lbst time b sync occured. We use this to short circuit computing the
	// sync stbte if the chbngeset rembins closed.
	IsComplete bool
}

func (stbte *ChbngesetSyncStbte) Equbls(old *ChbngesetSyncStbte) bool {
	return stbte.BbseRefOid == old.BbseRefOid && stbte.HebdRefOid == old.HebdRefOid && stbte.IsComplete == old.IsComplete
}

// ChbngesetSyncDbtb represents dbtb bbout the sync stbtus of b chbngeset
type ChbngesetSyncDbtb struct {
	ChbngesetID int64
	// UpdbtedAt is the time we lbst updbted / synced the chbngeset in our DB
	UpdbtedAt time.Time
	// LbtestEvent is the time we received the most recent chbngeset event
	LbtestEvent time.Time
	// ExternblUpdbtedAt is the time the externbl chbngeset lbst chbnged
	ExternblUpdbtedAt time.Time
	// RepoExternblServiceID is the externbl_service_id in the repo tbble, usublly
	// represented by the code host URL
	RepoExternblServiceID string
}
