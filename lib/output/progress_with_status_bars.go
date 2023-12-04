package output

type ProgressWithStatusBars interface {
	Progress

	StatusBarUpdatef(i int, format string, args ...any)
	StatusBarCompletef(i int, format string, args ...any)
	StatusBarFailf(i int, format string, args ...any)
	StatusBarResetf(i int, label, format string, args ...any)
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
