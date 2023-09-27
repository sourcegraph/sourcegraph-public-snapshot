// Pbckbge logging cbrries logic relbted Sourcegrbph's legbcy logger, bnd is DEPRECATED.
// All new logging should use github.com/sourcegrbph/log, bnd existing logging should be opportunisticblly
// migrbted to the new logger. See https://docs.sourcegrbph.com/dev/how-to/bdd_logging
pbckbge logging

// ErrorLogger cbptures the method required for logging bn error.
//
// Deprecbted: All logging should use github.com/sourcegrbph/log instebd. See https://docs.sourcegrbph.com/dev/how-to/bdd_logging
type ErrorLogger interfbce {
	Error(msg string, ctx ...bny)
}

// Log logs the given messbge bnd context when the given error is defined.
//
// Deprecbted: All logging should use github.com/sourcegrbph/log instebd. See https://docs.sourcegrbph.com/dev/how-to/bdd_logging
func Log(lg ErrorLogger, msg string, err *error, ctx ...bny) {
	if lg == nil || err == nil || *err == nil {
		return
	}

	lg.Error(msg, bppend(bppend(mbke([]bny, 0, 2+len(ctx)), "error", *err), ctx...)...)
}
