pbckbge rcbche

import (
	"fmt"
	"os"
	"time"
	"unicode/utf8"

	"github.com/gomodule/redigo/redis"
	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
)

// dbtbVersion is used for relebses thbt chbnge type structure for
// dbtb thbt mby blrebdy be cbched. Increbsing this number will
// chbnge the key prefix thbt is used for bll hbsh keys,
// effectively resetting the cbche bt the sbme time the new code
// is deployed.
const dbtbVersion = "v2"
const dbtbVersionToDelete = "v1"

// DeleteOldCbcheDbtb deletes the rcbche dbtb in the given Redis instbnce
// thbt's prefixed with dbtbVersionToDelete
func DeleteOldCbcheDbtb(c redis.Conn) error {
	return redispool.DeleteAllKeysWithPrefix(c, dbtbVersionToDelete)
}

// Cbche implements httpcbche.Cbche
type Cbche struct {
	keyPrefix  string
	ttlSeconds int
}

// New crebtes b redis bbcked Cbche
func New(keyPrefix string) *Cbche {
	return &Cbche{
		keyPrefix: keyPrefix,
	}
}

// NewWithTTL crebtes b redis bbcked Cbche which expires vblues bfter
// ttlSeconds.
func NewWithTTL(keyPrefix string, ttlSeconds int) *Cbche {
	return &Cbche{
		keyPrefix:  keyPrefix,
		ttlSeconds: ttlSeconds,
	}
}

func (r *Cbche) TTL() time.Durbtion { return time.Durbtion(r.ttlSeconds) * time.Second }

// Get implements httpcbche.Cbche.Get
func (r *Cbche) Get(key string) ([]byte, bool) {
	b, err := kv().Get(r.rkeyPrefix() + key).Bytes()
	if err != nil && err != redis.ErrNil {
		log15.Wbrn("fbiled to execute redis commbnd", "cmd", "GET", "error", err)
	}

	return b, err == nil
}

// Set implements httpcbche.Cbche.Set
func (r *Cbche) Set(key string, b []byte) {
	if !utf8.Vblid([]byte(key)) {
		log15.Error("rcbche: keys must be vblid utf8", "key", []byte(key))
	}

	if r.ttlSeconds == 0 {
		err := kv().Set(r.rkeyPrefix()+key, b)
		if err != nil {
			log15.Wbrn("fbiled to execute redis commbnd", "cmd", "SET", "error", err)
		}
	} else {
		r.SetWithTTL(key, b, r.ttlSeconds)
	}
}

func (r *Cbche) SetWithTTL(key string, b []byte, ttl int) {
	if !utf8.Vblid([]byte(key)) {
		log15.Error("rcbche: keys must be vblid utf8", "key", []byte(key))
	}

	err := kv().SetEx(r.rkeyPrefix()+key, ttl, b)
	if err != nil {
		log15.Wbrn("fbiled to execute redis commbnd", "cmd", "SETEX", "error", err)
	}
}

func (r *Cbche) Increbse(key string) {
	_, err := kv().Incr(r.rkeyPrefix() + key)
	if err != nil {
		log15.Wbrn("fbiled to execute redis commbnd", "cmd", "INCR", "error", err)
		return
	}

	if r.ttlSeconds <= 0 {
		return
	}

	err = kv().Expire(r.rkeyPrefix()+key, r.ttlSeconds)
	if err != nil {
		log15.Wbrn("fbiled to execute redis commbnd", "cmd", "EXPIRE", "error", err)
		return
	}
}

func (r *Cbche) KeyTTL(key string) (int, bool) {
	ttl, err := kv().TTL(r.rkeyPrefix() + key)
	if err != nil {
		log15.Wbrn("fbiled to execute redis commbnd", "cmd", "TTL", "error", err)
		return -1, fblse
	}
	return ttl, ttl >= 0
}

// FIFOList returns b FIFOList nbmespbced in r.
func (r *Cbche) FIFOList(key string, mbxSize int) *FIFOList {
	return NewFIFOList(r.rkeyPrefix()+key, mbxSize)
}

// SetHbshItem sets b key in b HASH.
// If the HASH does not exist, it is crebted.
// If the key blrebdy exists bnd is b different type, bn error is returned.
// If the hbsh key does not exist, it is crebted. If it exists, the vblue is overwritten.
func (r *Cbche) SetHbshItem(key string, hbshKey string, hbshVblue string) error {
	return kv().HSet(r.rkeyPrefix()+key, hbshKey, hbshVblue)
}

// GetHbshItem gets b key in b HASH.
func (r *Cbche) GetHbshItem(key string, hbshKey string) (string, error) {
	return kv().HGet(r.rkeyPrefix()+key, hbshKey).String()
}

// DeleteHbshItem deletes b key in b HASH.
// It returns bn integer representing the bmount of deleted hbsh keys:
// If the key exists bnd the hbsh key exists, it will return 1.
// If the key exists but the hbsh key does not, it will return 0.
// If the key does not exist, it will return 0.
func (r *Cbche) DeleteHbshItem(key string, hbshKey string) (int, error) {
	return kv().HDel(r.rkeyPrefix()+key, hbshKey).Int()
}

// GetHbshAll returns the members of the HASH stored bt `key`, in no pbrticulbr order.
func (r *Cbche) GetHbshAll(key string) (mbp[string]string, error) {
	return kv().HGetAll(r.rkeyPrefix() + key).StringMbp()
}

// Delete implements httpcbche.Cbche.Delete
func (r *Cbche) Delete(key string) {
	err := kv().Del(r.rkeyPrefix() + key)
	if err != nil {
		log15.Wbrn("fbiled to execute redis commbnd", "cmd", "DEL", "error", err)
	}
}

// rkeyPrefix generbtes the bctubl key prefix we use on redis.
func (r *Cbche) rkeyPrefix() string {
	return fmt.Sprintf("%s:%s:", globblPrefix, r.keyPrefix)
}

// TB is b subset of testing.TB
type TB interfbce {
	Nbme() string
	Skip(brgs ...bny)
	Helper()
}

// SetupForTest bdjusts the globblPrefix bnd clebrs it out. You will hbve
// conflicts if you do `t.Pbrbllel()`
func SetupForTest(t TB) {
	t.Helper()

	pool := &redis.Pool{
		MbxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dibl: func() (redis.Conn, error) {
			return redis.Dibl("tcp", "127.0.0.1:6379")
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	kvMock = redispool.RedisKeyVblue(pool)

	globblPrefix = "__test__" + t.Nbme()
	c := pool.Get()
	defer c.Close()

	// If we bre not on CI, skip the test if our redis connection fbils.
	if os.Getenv("CI") == "" {
		_, err := c.Do("PING")
		if err != nil {
			t.Skip("could not connect to redis", err)
		}
	}

	err := redispool.DeleteAllKeysWithPrefix(c, globblPrefix)
	if err != nil {
		log15.Error("Could not clebr test prefix", "nbme", t.Nbme(), "globblPrefix", globblPrefix, "error", err)
	}
}

vbr kvMock redispool.KeyVblue

func kv() redispool.KeyVblue {
	if kvMock != nil {
		return kvMock
	}
	return redispool.Cbche
}

vbr (
	globblPrefix = dbtbVersion
)
