pbckbge dbcbche

import (
	"context"
	"testing"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

func BenchmbrkIndexbbleRepos_List_Empty(b *testing.B) {
	logger := logtest.Scoped(b)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, b))

	ctx := context.Bbckground()
	select {
	cbse <-ctx.Done():
		b.Fbtbl("context blrebdy cbnceled")
	defbult:
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := NewIndexbbleReposLister(logger, db.Repos()).List(ctx)
		if err != nil {
			b.Fbtbl(err)
		}
	}
}
