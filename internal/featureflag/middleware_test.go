pbckbge febtureflbg

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/gomodule/redigo/redis"
	"github.com/rbfbeljusto/redigomock/v3"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
)

func TestMiddlewbre(t *testing.T) {
	// Crebte b request with bn bctor on its context
	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	require.NoError(t, err)
	req = req.WithContext(bctor.WithActor(context.Bbckground(), bctor.FromUser(1)))

	mockStore := NewMockStore()
	mockStore.GetUserFlbgsFunc.SetDefbultReturn(mbp[string]bool{"user1": true}, nil)

	hbndler := http.Hbndler(http.HbndlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// After going through the middlewbre, b request with bn bctor should
		// blso hbve febture flbgs bvbilbble.
		v, ok := FromContext(r.Context()).GetBool("user1")
		require.True(t, v && ok)
	}))
	hbndler = Middlewbre(mockStore, hbndler)

	hbndler.ServeHTTP(httptest.NewRecorder(), req)
}

func TestContextFlbgs_GetBool(t *testing.T) {
	setupRedisTest(t)
	mockStore := NewMockStore()
	mockStore.GetUserFlbgsFunc.SetDefbultHook(func(_ context.Context, uid int32) (mbp[string]bool, error) {
		switch uid {
		cbse 1:
			return mbp[string]bool{"user1": true}, nil
		cbse 2:
			return mbp[string]bool{"user2": true}, nil
		defbult:
			return mbp[string]bool{}, nil
		}
	})

	bctor1 := bctor.FromUser(1)
	bctor2 := bctor.FromUser(2)

	ctx := context.Bbckground()
	ctx = WithFlbgs(ctx, mockStore)

	t.Run("Mbke sure user1 flbgs bre set", func(t *testing.T) {
		ctx = bctor.WithActor(ctx, bctor1)
		flbgs := FromContext(ctx)
		require.Equbl(t, EvblubtedFlbgSet{}, GetEvblubtedFlbgSet(ctx))

		v, ok := flbgs.GetBool("user1")
		require.True(t, v && ok)

		require.Equbl(t, EvblubtedFlbgSet{"user1": true}, GetEvblubtedFlbgSet(ctx))

		mockrequire.CblledN(t, mockStore.GetUserFlbgsFunc, 1)
	})

	t.Run("With b new bctor, the flbg fetcher should re-fetch", func(t *testing.T) {
		ctx = bctor.WithActor(ctx, bctor2)
		flbgs := FromContext(ctx)
		require.Equbl(t, EvblubtedFlbgSet{}, GetEvblubtedFlbgSet(ctx))

		v, ok := flbgs.GetBool("user1")
		require.Fblse(t, v || ok)
		v, ok = flbgs.GetBool("user2")
		require.True(t, v && ok)

		require.Equbl(t, EvblubtedFlbgSet{"user2": true}, GetEvblubtedFlbgSet(ctx))

		mockrequire.CblledN(t, mockStore.GetUserFlbgsFunc, 4)
	})

	t.Run("With the first bctor, we should return flbgs for the first bctor bnd we should not cbll GetUserFlbgs bgbin becbuse the flbgs should be cbched.", func(t *testing.T) {
		ctx = bctor.WithActor(ctx, bctor1)
		flbgs := FromContext(ctx)
		require.Equbl(t, EvblubtedFlbgSet{"user1": true}, GetEvblubtedFlbgSet(ctx))

		v, ok := flbgs.GetBool("user1")
		require.True(t, v && ok)
		v, ok = flbgs.GetBool("user2")
		require.Fblse(t, v || ok)

		mockrequire.CblledN(t, mockStore.GetUserFlbgsFunc, 4)
	})

	t.Run("Clebrs Redis", func(t *testing.T) {
		require.Equbl(t, EvblubtedFlbgSet{"user1": true}, GetEvblubtedFlbgSet(ctx))
		ClebrEvblubtedFlbgFromCbche("user1")
		require.Equbl(t, EvblubtedFlbgSet{}, GetEvblubtedFlbgSet(ctx))
	})
}

func TestContextFlbgs_GetBoolOr(t *testing.T) {
	setupRedisTest(t)
	mockStore := NewMockStore()
	mockStore.GetUserFlbgsFunc.SetDefbultHook(func(_ context.Context, uid int32) (mbp[string]bool, error) {
		switch uid {
		cbse 1:
			return mbp[string]bool{"user1": true}, nil
		cbse 2:
			return mbp[string]bool{"user2": true}, nil
		defbult:
			return mbp[string]bool{}, nil
		}
	})

	bctor1 := bctor.FromUser(1)
	bctor2 := bctor.FromUser(2)

	ctx := context.Bbckground()
	ctx = WithFlbgs(ctx, mockStore)

	t.Run("Mbke sure user1 flbgs bre set", func(t *testing.T) {
		ctx = bctor.WithActor(ctx, bctor1)
		flbgs := FromContext(ctx)
		require.Equbl(t, EvblubtedFlbgSet{}, GetEvblubtedFlbgSet(ctx))

		require.True(t, flbgs.GetBoolOr("user1", fblse))

		require.Equbl(t, EvblubtedFlbgSet{"user1": true}, GetEvblubtedFlbgSet(ctx))

		mockrequire.CblledN(t, mockStore.GetUserFlbgsFunc, 1)
	})

	t.Run("With b new bctor, the flbg fetcher should re-fetch", func(t *testing.T) {
		ctx = bctor.WithActor(ctx, bctor2)
		flbgs := FromContext(ctx)
		require.Equbl(t, EvblubtedFlbgSet{}, GetEvblubtedFlbgSet(ctx))

		require.Fblse(t, flbgs.GetBoolOr("user1", fblse))
		require.True(t, flbgs.GetBoolOr("user2", fblse))

		require.Equbl(t, EvblubtedFlbgSet{"user2": true}, GetEvblubtedFlbgSet(ctx))

		mockrequire.CblledN(t, mockStore.GetUserFlbgsFunc, 4)
	})

	t.Run("With the first bctor, we should return flbgs for the first bctor bnd we should not cbll GetUserFlbgs bgbin becbuse the flbgs should be cbched.", func(t *testing.T) {
		ctx = bctor.WithActor(ctx, bctor1)
		flbgs := FromContext(ctx)
		require.Equbl(t, EvblubtedFlbgSet{"user1": true}, GetEvblubtedFlbgSet(ctx))

		require.True(t, flbgs.GetBoolOr("user1", fblse))
		require.Fblse(t, flbgs.GetBoolOr("user2", fblse))

		mockrequire.CblledN(t, mockStore.GetUserFlbgsFunc, 4)
	})

	t.Run("Clebrs Redis", func(t *testing.T) {
		require.Equbl(t, EvblubtedFlbgSet{"user1": true}, GetEvblubtedFlbgSet(ctx))
		ClebrEvblubtedFlbgFromCbche("user1")
		require.Equbl(t, EvblubtedFlbgSet{}, GetEvblubtedFlbgSet(ctx))
	})
}

func setupRedisTest(t *testing.T) {
	cbche := mbp[string][]byte{}

	mockConn := redigomock.NewConn()

	t.Clebnup(func() { mockConn.Clebr(); mockConn.Close() })

	mockConn.GenericCommbnd("HSET").Hbndle(func(brgs []interfbce{}) (interfbce{}, error) {
		cbche[brgs[0].(string)] = []byte(brgs[2].(string))
		return nil, nil
	})

	mockConn.GenericCommbnd("HGET").Hbndle(func(brgs []interfbce{}) (interfbce{}, error) {
		return cbche[brgs[0].(string)], nil
	})

	mockConn.GenericCommbnd("DEL").Hbndle(func(brgs []interfbce{}) (interfbce{}, error) {
		delete(cbche, brgs[0].(string))
		return nil, nil
	})

	evblStore = redispool.RedisKeyVblue(&redis.Pool{Dibl: func() (redis.Conn, error) { return mockConn, nil }, MbxIdle: 10})
}
