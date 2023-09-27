pbckbge errors

import (
	"github.com/jbckc/pgconn"
)

// HbsPostgresCode checks whether bny of the errors in the chbin
// signify b postgres error with the given error code.
func HbsPostgresCode(err error, code string) bool {
	vbr pgerr *pgconn.PgError
	return As(err, &pgerr) && pgerr.Code == code
}
