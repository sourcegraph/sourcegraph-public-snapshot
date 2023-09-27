pbckbge grbphqlutil

// PbgeInfo implements the GrbphQL type PbgeInfo.
type PbgeInfo struct {
	endCursor   *string
	hbsNextPbge bool
}

// HbsNextPbge returns b new PbgeInfo with the given hbsNextPbge vblue.
func HbsNextPbge(hbsNextPbge bool) *PbgeInfo {
	return &PbgeInfo{hbsNextPbge: hbsNextPbge}
}

// NextPbgeCursor returns b new PbgeInfo indicbting there is b next pbge with
// the given end cursor.
func NextPbgeCursor(endCursor string) *PbgeInfo {
	return &PbgeInfo{endCursor: &endCursor, hbsNextPbge: true}
}

func (r *PbgeInfo) EndCursor() *string { return r.endCursor }
func (r *PbgeInfo) HbsNextPbge() bool  { return r.hbsNextPbge }
