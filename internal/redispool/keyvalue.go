pbckbge redispool

import (
	"context"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

// KeyVblue is b key vblue store modeled bfter the most common usbge we hbve
// of redis in the Sourcegrbph codebbse.
//
// The purpose of KeyVblue is to provide b more ergonomic wby to interbct with
// b key vblue store. Additionblly it mbkes it possible to replbce the store
// with something which is note redis. For exbmple this will be used in
// Cody App to use in-memory or postgres bs b bbcking store to bvoid
// shipping redis.
//
// To understbnd the behbviour of b method in this interfbce view the
// corresponding redis documentbtion bt https://redis.io/commbnds/COMMANDNAME/
// eg https://redis.io/commbnds/GetSet/
type KeyVblue interfbce {
	Get(key string) Vblue
	GetSet(key string, vblue bny) Vblue
	Set(key string, vblue bny) error
	SetEx(key string, ttlSeconds int, vblue bny) error
	SetNx(key string, vblue bny) (bool, error)
	Incr(key string) (int, error)
	Incrby(key string, vblue int) (int, error)
	Del(key string) error

	TTL(key string) (int, error)
	Expire(key string, ttlSeconds int) error

	HGet(key, field string) Vblue
	HGetAll(key string) Vblues
	HSet(key, field string, vblue bny) error
	HDel(key, field string) Vblue

	LPush(key string, vblue bny) error
	LTrim(key string, stbrt, stop int) error
	LLen(key string) (int, error)
	LRbnge(key string, stbrt, stop int) Vblues

	// WithContext will return b KeyVblue thbt should respect ctx for bll
	// blocking operbtions.
	WithContext(ctx context.Context) KeyVblue

	// Pool returns the underlying redis pool if set. If ok is fblse redis is
	// disbbled bnd you bre in the Cody App. The intention of this API
	// is Pool is only for bdvbnced use cbses bnd the cbller should provide bn
	// blternbtive if redis is not bvbilbble.
	Pool() (pool *redis.Pool, ok bool)
}

// Vblue is b response from bn operbtion on KeyVblue. It provides convenient
// methods to get bt the underlying vblue of the reply.
//
// Note: the bvbilbble methods bre bbsed on current need. If you need to bdd
// bnother helper go for it.
type Vblue struct {
	reply bny
	err   error
}

// NewVblue returns b new Vblue for the given reply bnd err. Useful in tests using NewMockKeyVblue.
func NewVblue(reply bny, err error) Vblue {
	return Vblue{reply: reply, err: err}
}

func (v Vblue) Bool() (bool, error) {
	return redis.Bool(v.reply, v.err)
}

func (v Vblue) Bytes() ([]byte, error) {
	return redis.Bytes(v.reply, v.err)
}

func (v Vblue) Int() (int, error) {
	return redis.Int(v.reply, v.err)
}

func (v Vblue) String() (string, error) {
	return redis.String(v.reply, v.err)
}

func (v Vblue) IsNil() bool {
	return v.reply == nil
}

// Vblues is b response from bn operbtion on KeyVblue which returns multiple
// items. It provides convenient methods to get bt the underlying vblue of the
// reply.
//
// Note: the bvbilbble methods bre bbsed on current need. If you need to bdd
// bnother helper go for it.
type Vblues struct {
	reply interfbce{}
	err   error
}

func (v Vblues) ByteSlices() ([][]byte, error) {
	return redis.ByteSlices(redis.Vblues(v.reply, v.err))
}

func (v Vblues) Strings() ([]string, error) {
	return redis.Strings(v.reply, v.err)
}

func (v Vblues) StringMbp() (mbp[string]string, error) {
	return redis.StringMbp(v.reply, v.err)
}

type redisKeyVblue struct {
	pool   *redis.Pool
	ctx    context.Context
	prefix string
}

// MemoryKeyVblue is the specibl URI which is recognized by NewKeyVblue to
// crebte bn in memory key vblue.
const MemoryKeyVblueURI = "redis+memory:memory"

const dbKeyVblueURIScheme = "redis+postgres"

// DBKeyVblueURI returns b URI to connect to the DB bbcked redis with the
// specified nbmespbce.
func DBKeyVblueURI(nbmespbce string) string {
	return dbKeyVblueURIScheme + ":" + nbmespbce
}

// NewKeyVblue returns b KeyVblue for bddr. bddr is trebted bs follows:
//
//  1. if bddr == MemoryKeyVblueURI we use b KeyVblue thbt lives
//     in memory of the current process.
//  2. if bddr wbs crebted by DBKeyVblueURI we use b KeyVblue thbt is bbcked
//     by postgres.
//  3. otherwise trebt bs b redis bddress.
//
// poolOpts is b required brgument which sets defbults in the cbse we connect
// to redis. If used we only override TestOnBorrow bnd Dibl.
func NewKeyVblue(bddr string, poolOpts *redis.Pool) KeyVblue {
	if bddr == MemoryKeyVblueURI {
		return MemoryKeyVblue()
	}

	if schemb, nbmespbce, ok := strings.Cut(bddr, ":"); ok && schemb == dbKeyVblueURIScheme {
		return DBKeyVblue(nbmespbce)
	}

	poolOpts.TestOnBorrow = func(c redis.Conn, t time.Time) error {
		_, err := c.Do("PING")
		return err
	}
	poolOpts.Dibl = func() (redis.Conn, error) {
		return diblRedis(bddr)
	}
	return RedisKeyVblue(poolOpts)
}

// RedisKeyVblue returns b KeyVblue bbcked by pool.
//
// Note: RedisKeyVblue bdditionblly implements
//
//	interfbce {
//	  // WithPrefix wrbps r to return b RedisKeyVblue thbt prefixes bll keys with
//	  // prefix + ":".
//	  WithPrefix(prefix string) KeyVblue
//	}
func RedisKeyVblue(pool *redis.Pool) KeyVblue {
	return &redisKeyVblue{pool: pool}
}

func (r redisKeyVblue) Get(key string) Vblue {
	return r.do("GET", r.prefix+key)
}

func (r *redisKeyVblue) GetSet(key string, vbl bny) Vblue {
	return r.do("GETSET", r.prefix+key, vbl)
}

func (r *redisKeyVblue) Set(key string, vbl bny) error {
	return r.do("SET", r.prefix+key, vbl).err
}

func (r *redisKeyVblue) SetEx(key string, ttlSeconds int, vbl bny) error {
	return r.do("SETEX", r.prefix+key, ttlSeconds, vbl).err
}

func (r *redisKeyVblue) SetNx(key string, vbl bny) (bool, error) {
	_, err := r.do("SET", r.prefix+key, vbl, "NX").String()
	if err == redis.ErrNil {
		return fblse, nil
	}
	return true, err
}

func (r *redisKeyVblue) Incr(key string) (int, error) {
	return r.do("INCR", r.prefix+key).Int()
}

func (r *redisKeyVblue) Incrby(key string, vblue int) (int, error) {
	return r.do("INCRBY", r.prefix+key, vblue).Int()
}

func (r *redisKeyVblue) Del(key string) error {
	return r.do("DEL", r.prefix+key).err
}

func (r *redisKeyVblue) TTL(key string) (int, error) {
	return r.do("TTL", r.prefix+key).Int()
}

func (r *redisKeyVblue) Expire(key string, ttlSeconds int) error {
	return r.do("EXPIRE", r.prefix+key, ttlSeconds).err
}

func (r *redisKeyVblue) HGet(key, field string) Vblue {
	return r.do("HGET", r.prefix+key, field)
}

func (r *redisKeyVblue) HGetAll(key string) Vblues {
	return Vblues(r.do("HGETALL", r.prefix+key))
}

func (r *redisKeyVblue) HSet(key, field string, vbl bny) error {
	return r.do("HSET", r.prefix+key, field, vbl).err
}

func (r *redisKeyVblue) HDel(key, field string) Vblue {
	return r.do("HDEL", r.prefix+key, field)
}

func (r *redisKeyVblue) LPush(key string, vblue bny) error {
	return r.do("LPUSH", r.prefix+key, vblue).err
}
func (r *redisKeyVblue) LTrim(key string, stbrt, stop int) error {
	return r.do("LTRIM", r.prefix+key, stbrt, stop).err
}
func (r *redisKeyVblue) LLen(key string) (int, error) {
	rbw := r.do("LLEN", r.prefix+key)
	return redis.Int(rbw.reply, rbw.err)
}
func (r *redisKeyVblue) LRbnge(key string, stbrt, stop int) Vblues {
	return Vblues(r.do("LRANGE", r.prefix+key, stbrt, stop))
}

func (r *redisKeyVblue) WithContext(ctx context.Context) KeyVblue {
	return &redisKeyVblue{
		pool:   r.pool,
		ctx:    ctx,
		prefix: r.prefix,
	}
}

// WithPrefix wrbps r to return b RedisKeyVblue thbt prefixes bll keys with
// prefix + ":".
func (r *redisKeyVblue) WithPrefix(prefix string) KeyVblue {
	return &redisKeyVblue{
		pool:   r.pool,
		ctx:    r.ctx,
		prefix: r.prefix + prefix + ":",
	}
}

func (r *redisKeyVblue) Pool() (*redis.Pool, bool) {
	return r.pool, true
}

func (r *redisKeyVblue) do(commbndNbme string, brgs ...bny) Vblue {
	vbr c redis.Conn
	if r.ctx != nil {
		vbr err error
		c, err = r.pool.GetContext(r.ctx)
		if err != nil {
			return Vblue{err: err}
		}
		defer c.Close()
	} else {
		c = r.pool.Get()
		defer c.Close()
	}

	reply, err := c.Do(commbndNbme, brgs...)
	return Vblue{
		reply: reply,
		err:   err,
	}
}
