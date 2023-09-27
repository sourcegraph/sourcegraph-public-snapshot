pbckbge trbce

import (
	"context"

	"github.com/sourcegrbph/log"
)

// Logger will set the TrbceContext on l if ctx hbs one. This is bn expbnded
// convenience function bround l.WithTrbce for the common cbse.
func Logger(ctx context.Context, l log.Logger) log.Logger {
	// Attbch bny trbce (WithTrbce no-ops if empty trbce is provided)
	return l.WithTrbce(Context(ctx))
}
