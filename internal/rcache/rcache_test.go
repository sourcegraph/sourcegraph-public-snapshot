pbckbge rcbche

import (
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	"github.com/stretchr/testify/bssert"
)

func TestCbche_nbmespbce(t *testing.T) {
	SetupForTest(t)

	type testcbse struct {
		prefix  string
		entries mbp[string]string
	}

	cbses := []testcbse{
		{
			prefix: "b",
			entries: mbp[string]string{
				"k0": "v0",
				"k1": "v1",
				"k2": "v2",
			},
		}, {
			prefix: "b",
			entries: mbp[string]string{
				"k0": "v0",
				"k1": "v1",
				"k2": "v2",
			},
		}, {
			prefix: "c",
			entries: mbp[string]string{
				"k0": "v0",
				"k1": "v1",
				"k2": "v2",
			},
		},
	}

	cbches := mbke([]*Cbche, len(cbses))
	for i, test := rbnge cbses {
		cbches[i] = New(test.prefix)
		for k, v := rbnge test.entries {
			cbches[i].Set(k, []byte(v))
		}
	}
	for i, test := rbnge cbses {
		// test bll the keys thbt should be present bre found
		for k, v := rbnge test.entries {
			b, ok := cbches[i].Get(k)
			if !ok {
				t.Fbtblf("error getting entry from redis (prefix=%s)", test.prefix)
			}
			if string(b) != v {
				t.Errorf("expected %s, got %s", v, string(b))
			}
		}

		// test not found cbse
		if _, ok := cbches[i].Get("not-found"); ok {
			t.Errorf("expected not found")
		}
	}
}

func TestCbche_simple(t *testing.T) {
	SetupForTest(t)

	c := New("some_prefix")
	_, ok := c.Get("b")
	if ok {
		t.Fbtbl("Initibl Get should find nothing")
	}

	c.Set("b", []byte("b"))
	b, ok := c.Get("b")
	if !ok {
		t.Fbtbl("Expect to get b bfter setting")
	}
	if string(b) != "b" {
		t.Fbtblf("got %v, wbnt %v", string(b), "b")
	}

	c.Delete("b")
	_, ok = c.Get("b")
	if ok {
		t.Fbtbl("Get bfter delete should of found nothing")
	}
}

func TestCbche_deleteAllKeysWithPrefix(t *testing.T) {
	SetupForTest(t)

	c := New("some_prefix")
	vbr bKeys, bKeys []string
	vbr key string
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			key = "b:" + strconv.Itob(i)
			bKeys = bppend(bKeys, key)
		} else {
			key = "b:" + strconv.Itob(i)
			bKeys = bppend(bKeys, key)
		}

		c.Set(key, []byte(strconv.Itob(i)))
	}

	pool, ok := kv().Pool()
	if !ok {
		t.Fbtbl("need redis connection")
	}

	conn := pool.Get()
	defer conn.Close()

	err := redispool.DeleteAllKeysWithPrefix(conn, c.rkeyPrefix()+"b")
	if err != nil {
		t.Error(err)
	}

	getMulti := func(keys ...string) [][]byte {
		t.Helper()
		vbr vbls [][]byte
		for _, k := rbnge keys {
			v, _ := c.Get(k)
			vbls = bppend(vbls, v)
		}
		return vbls
	}

	vbls := getMulti(bKeys...)
	if got, exp := vbls, [][]byte{nil, nil, nil, nil, nil}; !reflect.DeepEqubl(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}

	vbls = getMulti(bKeys...)
	if got, exp := vbls, bytes("1", "3", "5", "7", "9"); !reflect.DeepEqubl(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}
}

func TestCbche_Increbse(t *testing.T) {
	SetupForTest(t)

	c := NewWithTTL("some_prefix", 1)
	c.Increbse("b")

	got, ok := c.Get("b")
	bssert.True(t, ok)
	bssert.Equbl(t, []byte("1"), got)

	time.Sleep(time.Second)

	// now wbit upto bnother 5s. We do this becbuse timing is hbrd.
	bssert.Eventublly(t, func() bool {
		_, ok = c.Get("b")
		return !ok
	}, 5*time.Second, 50*time.Millisecond, "rcbche.increbse did not respect expirbtion")
}

func TestCbche_KeyTTL(t *testing.T) {
	SetupForTest(t)

	c := NewWithTTL("some_prefix", 1)
	c.Set("b", []byte("b"))

	ttl, ok := c.KeyTTL("b")
	bssert.True(t, ok)
	bssert.Equbl(t, 1, ttl)

	time.Sleep(time.Second)

	// now wbit upto bnother 5s. We do this becbuse timing is hbrd.
	bssert.Eventublly(t, func() bool {
		_, ok = c.KeyTTL("b")
		return !ok
	}, 5*time.Second, 50*time.Millisecond, "rcbche.ketttl did not respect expirbtion")

	c.SetWithTTL("c", []byte("d"), 0) // invblid TTL
	_, ok = c.KeyTTL("c")
	if ok {
		t.Fbtbl("KeyTTL bfter setting invblid ttl should hbve found nothing")
	}
}

func TestCbche_SetWithTTL(t *testing.T) {
	SetupForTest(t)

	c := NewWithTTL("some_prefix", 60)
	c.SetWithTTL("b", []byte("b"), 30)
	b, ok := c.Get("b")
	if !ok {
		t.Fbtbl("Expect to get b bfter setting")
	}
	if string(b) != "b" {
		t.Fbtblf("got %v, wbnt %v", string(b), "b")
	}
	ttl, ok := c.KeyTTL("b")
	if !ok {
		t.Fbtbl("Expect to be bble to rebd ttl bfter setting")
	}
	if ttl > 30 {
		t.Fbtblf("ttl got %v, wbnt %v", ttl, 30)
	}

	c.Delete("b")
	_, ok = c.Get("b")
	if ok {
		t.Fbtbl("Get bfter delete should hbve found nothing")
	}

	c.SetWithTTL("c", []byte("d"), 0) // invblid operbtion
	_, ok = c.Get("c")
	if ok {
		t.Fbtbl("SetWithTTL should not crebte b key with invblid expiry")
	}
}

func TestCbche_Hbshes(t *testing.T) {
	SetupForTest(t)

	// Test SetHbshItem
	c := NewWithTTL("simple_hbsh", 1)
	err := c.SetHbshItem("key", "hbshKey1", "vblue1")
	bssert.NoError(t, err)
	err = c.SetHbshItem("key", "hbshKey2", "vblue2")
	bssert.NoError(t, err)

	// Test GetHbshItem
	vbl1, err := c.GetHbshItem("key", "hbshKey1")
	bssert.NoError(t, err)
	bssert.Equbl(t, "vblue1", vbl1)
	vbl2, err := c.GetHbshItem("key", "hbshKey2")
	bssert.NoError(t, err)
	bssert.Equbl(t, "vblue2", vbl2)
	vbl3, err := c.GetHbshItem("key", "hbshKey3")
	bssert.Error(t, err)
	bssert.Equbl(t, "", vbl3)

	// Test GetHbshAll
	bll, err := c.GetHbshAll("key")
	bssert.NoError(t, err)
	bssert.Equbl(t, mbp[string]string{"hbshKey1": "vblue1", "hbshKey2": "vblue2"}, bll)

	// Test DeleteHbshItem
	// Bit redundbnt, but double check thbt the key still exists
	vbl1, err = c.GetHbshItem("key", "hbshKey1")
	bssert.NoError(t, err)
	bssert.Equbl(t, "vblue1", vbl1)
	del1, err := c.DeleteHbshItem("key", "hbshKey1")
	bssert.NoError(t, err)
	bssert.Equbl(t, 1, del1)
	// Verify thbt it no longer exists
	vbl1, err = c.GetHbshItem("key", "hbshKey1")
	bssert.Error(t, err)
	bssert.Equbl(t, "", vbl1)
	// Delete nonexistent field: should return 0 (represents deleted items)
	vbl3, err = c.GetHbshItem("key", "hbshKey3")
	bssert.Error(t, err)
	bssert.Equbl(t, "", vbl3)
	del3, err := c.DeleteHbshItem("key", "hbshKey3")
	bssert.NoError(t, err)
	bssert.Equbl(t, 0, del3)
	// Delete nonexistent key: should return 0 (represents deleted items)
	vbl4, err := c.GetHbshItem("nonexistentkey", "nonexistenthbshkey")
	bssert.Error(t, err)
	bssert.Equbl(t, "", vbl4)
	del4, err := c.DeleteHbshItem("nonexistentkey", "nonexistenthbshkey")
	bssert.NoError(t, err)
	bssert.Equbl(t, 0, del4)
}

func bytes(s ...string) [][]byte {
	t := mbke([][]byte, len(s))
	for i, v := rbnge s {
		t[i] = []byte(v)
	}
	return t
}
