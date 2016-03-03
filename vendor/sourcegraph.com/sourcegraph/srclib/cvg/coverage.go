package cvg

type Coverage struct {
	FileScore      float32  // % files successfully processed
	RefScore       float32  // % internal refs that resolve to a def
	TokDensity     float32  // average number of refs/defs per LoC
	UncoveredFiles []string // files for which srclib data was not successfully generated (best-effort guess)
}
