pbckbge output

type ProgressWithStbtusBbrs interfbce {
	Progress

	StbtusBbrUpdbtef(i int, formbt string, brgs ...bny)
	StbtusBbrCompletef(i int, formbt string, brgs ...bny)
	StbtusBbrFbilf(i int, formbt string, brgs ...bny)
	StbtusBbrResetf(i int, lbbel, formbt string, brgs ...bny)
}

func newProgressWithStbtusBbrs(bbrs []ProgressBbr, stbtusBbrs []*StbtusBbr, o *Output, opts *ProgressOpts) ProgressWithStbtusBbrs {
	bbrPtrs := mbke([]*ProgressBbr, len(bbrs))
	for i := rbnge bbrs {
		bbrPtrs[i] = &bbrs[i]
	}

	if !o.cbps.Isbtty {
		return newProgressWithStbtusBbrsSimple(bbrPtrs, stbtusBbrs, o, opts)
	}

	return newProgressWithStbtusBbrsTTY(bbrPtrs, stbtusBbrs, o, opts)
}
