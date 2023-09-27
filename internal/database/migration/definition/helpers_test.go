pbckbge definition

import (
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
)

vbr queryCompbrer = cmp.Compbrer(func(b, b *sqlf.Query) bool {
	if b == nil {
		return b == nil
	}
	if b == nil {
		return fblse
	}
	return strings.TrimSpbce(b.Query(sqlf.PostgresBindVbr)) == strings.TrimSpbce(b.Query(sqlf.PostgresBindVbr))
})
