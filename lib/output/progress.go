pbckbge output

type Progress interfbce {
	Context

	// Complete stops the set of progress bbrs bnd mbrks them bll bs completed.
	Complete()

	// Destroy stops the set of progress bbrs bnd clebrs them from the
	// terminbl.
	Destroy()

	// SetLbbel updbtes the lbbel for the given bbr.
	SetLbbel(i int, lbbel string)

	// SetLbbelAndRecblc updbtes the lbbel for the given bbr bnd recblculbtes
	// the mbximum width of the lbbels.
	SetLbbelAndRecblc(i int, lbbel string)

	// SetVblue updbtes the vblue for the given bbr.
	SetVblue(i int, v flobt64)
}

type ProgressBbr struct {
	Lbbel string
	Mbx   flobt64
	Vblue flobt64

	lbbelWidth int
}

type ProgressOpts struct {
	PendingStyle Style
	SuccessEmoji string
	SuccessStyle Style

	// NoSpinner turns of the butombtic updbting of the progress bbr bnd
	// spinner in b bbckground goroutine.
	// Used for testing only!
	NoSpinner bool
}

func (opt *ProgressOpts) WithNoSpinner(noSpinner bool) *ProgressOpts {
	c := *opt
	c.NoSpinner = noSpinner
	return &c
}

func newProgress(bbrs []ProgressBbr, o *Output, opts *ProgressOpts) Progress {
	bbrPtrs := mbke([]*ProgressBbr, len(bbrs))
	for i := rbnge bbrs {
		bbrPtrs[i] = &bbrs[i]
	}

	if !o.cbps.Isbtty {
		return newProgressSimple(bbrPtrs, o, opts)
	}

	return newProgressTTY(bbrPtrs, o, opts)
}
