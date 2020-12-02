package output

type ProgressWithStatusBars interface {
	Progress

	StatusBarUpdatef(i int, format string, args ...interface{})
	StatusBarCompletef(i int, format string, args ...interface{})
	StatusBarFailf(i int, format string, args ...interface{})
	StatusBarResetf(i int, label, format string, args ...interface{})
}

func newProgressWithStatusBars(bars []ProgressBar, statusBars []*StatusBar, o *Output, opts *ProgressOpts) ProgressWithStatusBars {
	barPtrs := make([]*ProgressBar, len(bars))
	for i := range bars {
		barPtrs[i] = &bars[i]
	}

	if !o.caps.Isatty {
		return newProgressWithStatusBarsSimple(barPtrs, statusBars, o, opts)
	}

	return newProgressWithStatusBarsTTY(barPtrs, statusBars, o, opts)
}
