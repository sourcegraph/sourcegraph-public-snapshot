package output

type Progress interface {
	Context

	// Complete stops the set of progress bars and marks them all as completed.
	Complete()

	// Destroy stops the set of progress bars and clears them from the
	// terminal.
	Destroy()

	// SetLabel updates the label for the given bar.
	SetLabel(i int, label string)

	// SetLabelAndRecalc updates the label for the given bar and recalculates
	// the maximum width of the labels.
	SetLabelAndRecalc(i int, label string)

	// SetValue updates the value for the given bar.
	SetValue(i int, v float64)
}

type ProgressBar struct {
	Label string
	Max   float64
	Value float64

	labelWidth int
}

type ProgressOpts struct {
	PendingStyle Style
	SuccessEmoji string
	SuccessStyle Style

	// NoSpinner turns of the automatic updating of the progress bar and
	// spinner in a background goroutine.
	// Used for testing only!
	NoSpinner bool
}

func (opt *ProgressOpts) WithNoSpinner(noSpinner bool) *ProgressOpts {
	c := *opt
	c.NoSpinner = noSpinner
	return &c
}

func newProgress(bars []ProgressBar, o *Output, opts *ProgressOpts) Progress {
	barPtrs := make([]*ProgressBar, len(bars))
	for i := range bars {
		barPtrs[i] = &bars[i]
	}

	if !o.caps.Isatty {
		return newProgressSimple(barPtrs, o, opts)
	}

	return newProgressTTY(barPtrs, o, opts)
}
