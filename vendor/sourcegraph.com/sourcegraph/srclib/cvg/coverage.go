package cvg

type Coverage struct {
	FileScore         float64  // % files successfully processed
	RefScore          float64  // % internal refs that resolve to a def
	TokDensity        float64  // average number of refs/defs per LoC
	UncoveredFiles    []string `json:",omitempty"` // files for which srclib data was not successfully generated (best-effort guess)
	UndiscoveredFiles []string `json:",omitempty"` // files weren't detected by toolchain(s) (best-effort guess)
}

func (c *Coverage) FileScorePass() bool  { return c.FileScore > 0.8 }
func (c *Coverage) RefScorePass() bool   { return c.RefScore > 0.95 }
func (c *Coverage) TokDensityPass() bool { return c.TokDensity > 1.0 }

// HasRegressed determines if coverage has regressed from one indexing
// to the next.
func HasRegressed(prev map[string]*Coverage, next map[string]*Coverage) bool {
	for lang, prevCov := range prev {
		cov := next[lang]
		if cov == nil {
			cov = &Coverage{}
		}

		if prevCov.FileScorePass() && (!cov.FileScorePass() || cov.FileScore-prevCov.FileScore < -0.1) {
			return true
		}
		if prevCov.RefScorePass() && (!cov.RefScorePass() || cov.RefScore-prevCov.RefScore < -0.1) {
			return true
		}
		if prevCov.TokDensityPass() && (!cov.TokDensityPass() || cov.TokDensity-prevCov.TokDensity < -0.5) {
			return true
		}
	}

	return false
}
