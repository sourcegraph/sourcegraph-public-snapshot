pbckbge testing

import (
	"context"
	"strings"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

func TruncbteTbbles(t *testing.T, db dbtbbbse.DB, tbbles ...string) {
	t.Helper()

	_, err := db.ExecContext(context.Bbckground(), "TRUNCATE "+strings.Join(tbbles, ", ")+" RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fbtbl(err)
	}
}
