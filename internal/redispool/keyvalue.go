package redispool

import (
	"context"

	"github.com/gomodule/redigo/redis"
)

// KeyValue is a key value store modeled after the most common usage we have
// of redis in the Sourcegraph codebase.
//
// The purpose of KeyValue is to provide a more ergonomic way to interact with
// a key value store. Additionally it makes it possible to replace the store
// with something which is note redis. For example this will be used in
// Sourcegraph App to use in-memory or postgres as a backing store to avoid
// shipping redis.
//
// To understand the behaviour of a method in this interface view the
// corresponding redis documentation at https://redis.io/commands/COMMANDNAME/
// eg https://redis.io/commands/GetSet/
type KeyValue interface {
	Get(key string) Value
	GetSet(key string, value any) Value
	Set(key string, value any) error
	Del(key string) error

	HGet(key, field string) Value
	HSet(key, field string, value any) error

	Expire(key string, seconds int) error

	// WithContext will return a KeyValue that should respect ctx for all
	// blocking operations.
	WithContext(ctx context.Context) KeyValue

	// Pool returns the underlying redis pool if set. If ok is false redis is
	// disabled and you are in the Sourcegraph App. The intention of this API
	// is Pool is only for advanced use cases and the caller should provide an
	// alternative if redis is not available.
	Pool() (pool *redis.Pool, ok bool)
}

// Value is a response from an operation on KeyValue. It provides convenient
// methods to get at the underlying value of the reply.
//
// Note: the available methods are based on current need. If you need to add
// another helper go for it.
type Value struct {
	reply interface{}
	err   error
}

func (v Value) Bool() (bool, error) {
	return redis.Bool(v.reply, v.err)
}

func (v Value) Bytes() ([]byte, error) {
	return redis.Bytes(v.reply, v.err)
}

func (v Value) String() (string, error) {
	return redis.String(v.reply, v.err)
}

type redisKeyValue struct {
	pool   *redis.Pool
	ctx    context.Context
	prefix string
}

// RedisKeyValue returns a KeyValue backed by pool.
//
// Note: RedisKeyValue additionally implements
//
//	interface {
//	  // WithPrefix wraps r to return a RedisKeyValue that prefixes all keys with
//	  // prefix + ":".
//	  WithPrefix(prefix string) KeyValue
//	}
func RedisKeyValue(pool *redis.Pool) KeyValue {
	return &redisKeyValue{pool: pool}
}

func (r redisKeyValue) Get(key string) Value {
	return r.do("GET", r.prefix+key)
}

func (r *redisKeyValue) GetSet(key string, val any) Value {
	return r.do("GETSET", r.prefix+key, val)
}

func (r *redisKeyValue) Set(key string, val any) error {
	return r.do("SET", r.prefix+key, val).err
}

func (r *redisKeyValue) Del(key string) error {
	return r.do("DEL", r.prefix+key).err
}

func (r *redisKeyValue) HGet(key, field string) Value {
	return r.do("HGET", r.prefix+key, field)
}

func (r *redisKeyValue) HSet(key, field string, val any) error {
	return r.do("HSET", r.prefix+key, field, val).err
}

func (r *redisKeyValue) Expire(key string, seconds int) error {
	return r.do("EXPIRE", r.prefix+key, seconds).err
}

func (r *redisKeyValue) WithContext(ctx context.Context) KeyValue {
	return &redisKeyValue{
		pool:   r.pool,
		ctx:    ctx,
		prefix: r.prefix,
	}
}

// WithPrefix wraps r to return a RedisKeyValue that prefixes all keys with
// prefix + ":".
func (r *redisKeyValue) WithPrefix(prefix string) KeyValue {
	return &redisKeyValue{
		pool:   r.pool,
		ctx:    r.ctx,
		prefix: r.prefix + prefix + ":",
	}
}

func (r *redisKeyValue) Pool() (*redis.Pool, bool) {
	return r.pool, true
}

func (r *redisKeyValue) do(commandName string, args ...any) Value {
	var c redis.Conn
	if r.ctx != nil {
		var err error
		c, err = r.pool.GetContext(r.ctx)
		if err != nil {
			return Value{err: err}
		}
		defer c.Close()
	} else {
		c = r.pool.Get()
		defer c.Close()
	}

	reply, err := c.Do(commandName, args...)
	return Value{
		reply: reply,
		err:   err,
	}
}
