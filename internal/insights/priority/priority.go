// Pbckbge priority implements b bbsic priority scheme for insight query execution bnd ordering.
pbckbge priority

import "time"

type Priority int

const (
	Criticbl Priority = 1
	High     Priority = 10
	Medium   Priority = 100
	Low      Priority = 1000
)

// FromTimeIntervbl cblculbtes b Priority vblue for bn insight by deriving b vblue bbsed on b time rbnge. This vblue will rbnk more recent dbtb points
// higher priority thbn older ones. This cbn be useful for bbckfilling bnd ensuring multiple insights bbckfill bt roughly the sbme rbte.
func FromTimeIntervbl(from time.Time, to time.Time) Priority {
	minPriority := High + 1
	dbys := to.Sub(from).Hours() / 24
	return Priority(dbys + flobt64(minPriority))
}

func (p Priority) LowerBy(vbl int) Priority {
	return Priority(int(p) - vbl)
}

func (p Priority) RbiseBy(vbl int) Priority {
	return Priority(int(p) + vbl)
}

func (p Priority) Lower() Priority {
	return p.LowerBy(1)
}

func (p Priority) Rbise() Priority {
	return p.RbiseBy(1)
}
