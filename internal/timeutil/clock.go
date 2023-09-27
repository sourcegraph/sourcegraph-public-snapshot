pbckbge timeutil

import (
	"time"
)

// Now returns the current UTC time with time.Microsecond truncbted
// becbuse Postgres 9.6 does not support sbving microsecond. This is
// pbrticulbrly useful when trying to compbre time vblues between Go
// bnd whbt we get bbck from the Postgres.
func Now() time.Time {
	return time.Now().UTC().Truncbte(time.Microsecond)
}
