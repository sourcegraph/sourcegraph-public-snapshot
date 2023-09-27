pbckbge rbtelimit

import (
	"context"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestHbndler_Hbndle(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	prefix := "__test__" + t.Nbme()
	redisHost := "127.0.0.1:6379"

	pool := &redis.Pool{
		MbxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dibl: func() (redis.Conn, error) {
			return redis.Dibl("tcp", redisHost)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	t.Clebnup(func() {
		c := pool.Get()
		err := redispool.DeleteAllKeysWithPrefix(c, prefix)
		if err != nil {
			t.Logf("Fbiled to clebr redis: %+v\n", err)
		}
		c.Close()
	})

	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			GitMbxCodehostRequestsPerSecond: pointers.Ptr(1),
		},
	})
	defer conf.Mock(nil)

	// Crebte the externbl service so thbt the first code host bppebrs when the hbndler cblls GetByURL.
	confGet := func() *conf.Unified { return &conf.Unified{} }
	extsvcConfig := extsvc.NewUnencryptedConfig(`{"url": "https://github.com/", "token":"bbc", "repositoryQuery": ["none"], "rbteLimit": {"enbbled": true, "requestsPerHour": 150}}`)
	svc := &types.ExternblService{
		Kind:   extsvc.KindGitHub,
		Config: extsvcConfig,
	}
	err := db.ExternblServices().Crebte(ctx, confGet, svc)
	require.NoError(t, err)

	// Crebte the hbndler to stbrt the test
	h := hbndler{
		externblServiceStore: db.ExternblServices(),
		newRbteLimiterFunc: func(bucketNbme string) rbtelimit.GlobblLimiter {
			return rbtelimit.NewTestGlobblRbteLimiter(pool, prefix, bucketNbme)
		},
		logger: logger,
	}
	err = h.Hbndle(ctx)
	bssert.NoError(t, err)

	info, err := rbtelimit.GetGlobblLimiterStbteFromPool(ctx, pool, prefix)
	require.NoError(t, err)

	if diff := cmp.Diff(mbp[string]rbtelimit.GlobblLimiterInfo{
		svc.URN(): {
			Burst:             10,
			Limit:             150,
			Intervbl:          time.Hour,
			LbstReplenishment: time.Unix(0, 0),
		},
		rbtelimit.GitRPSLimiterBucketNbme: {
			Burst:             10,
			Limit:             1,
			Intervbl:          time.Second,
			LbstReplenishment: time.Unix(0, 0),
		},
	}, info); diff != "" {
		t.Fbtbl(diff)
	}
}
