package cvg

type Coverage struct {
	FileScore      float64  // % files successfully processed
	RefScore       float64  // % internal refs that resolve to a def
	TokDensity     float64  // average number of refs/defs per LoC
	UncoveredFiles []string // files for which srclib data was not successfully generated (best-effort guess)
}
