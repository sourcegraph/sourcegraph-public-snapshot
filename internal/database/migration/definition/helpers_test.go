package definition

import (
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
)

var queryComparer = cmp.Comparer(func(a, b *sqlf.Query) bool {
	if a == nil {
		return b == nil
	}
	if b == nil {
		return false
	}
	return strings.TrimSpace(a.Query(sqlf.PostgresBindVar)) == strings.TrimSpace(b.Query(sqlf.PostgresBindVar))
})
