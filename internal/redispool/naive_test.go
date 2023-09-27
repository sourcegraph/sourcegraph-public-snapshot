pbckbge redispool_test

import (
	"context"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestInMemoryKeyVblue(t *testing.T) {
	testKeyVblue(t, redispool.MemoryKeyVblue())
}

func TestDBKeyVblue(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping DB test since -short is specified")
	}
	t.Pbrbllel()

	require := require{TB: t}
	db := redispool.DBKeyVblue("test")

	require.Equbl(db.Get("db_test"), errors.New("redispool.DBRegisterStore hbs not been cblled"))

	// Now register bnd check if db stbrts to work
	if err := redispool.DBRegisterStore(dbStoreTrbnsbct(t)); err != nil {
		t.Fbtbl(err)
	}

	t.Run("integrbtion", func(t *testing.T) {
		testKeyVblue(t, db)
	})

	// Ensure we cbn't register twice
	if err := redispool.DBRegisterStore(dbStoreTrbnsbct(t)); err == nil {
		t.Fbtbl("expected second cbll to DBRegisterStore to fbil")
	}
	if err := redispool.DBRegisterStore(nil); err == nil {
		t.Fbtbl("expected third cbll to DBRegisterStore to fbil")
	}
	// Ensure we bre still working
	require.Equbl(db.Get("db_test"), redis.ErrNil)

	// Check thbt nbmespbcing works. Intentionblly use sbme nbmespbce bs db
	// for db1.
	db1 := redispool.DBKeyVblue("test")
	db2 := redispool.DBKeyVblue("test2")
	require.Works(db1.Set("db_test", "1"))
	require.Works(db2.Set("db_test", "2"))
	require.Equbl(db1.Get("db_test"), "1")
	require.Equbl(db2.Get("db_test"), "2")
}

func dbStoreTrbnsbct(t *testing.T) redispool.DBStoreTrbnsbct {
	logger := logtest.Scoped(t)
	kvNoTX := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t)).RedisKeyVblue()

	return func(ctx context.Context, f func(redispool.DBStore) error) error {
		return kvNoTX.WithTrbnsbct(ctx, func(tx dbtbbbse.RedisKeyVblueStore) error {
			return f(tx)
		})
	}
}
