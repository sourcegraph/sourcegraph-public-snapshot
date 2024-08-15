package redispool

import (
	"context"
	"time"

	"github.com/gomodule/redigo/redis"
)

// KeyValue is a key value store modeled after the most common usage we have
// of redis in the Sourcegraph codebase.
//
// The purpose of KeyValue is to provide a more ergonomic way to interact with
// a key value store. Additionally it makes it possible to replace the store
// with something which is not redis.
//
// To understand the behaviour of a method in this interface view the
// corresponding redis documentation at https://redis.io/commands/COMMANDNAME/
// eg https://redis.io/commands/GetSet/
type KeyValue interface {
	Get(key string) Value
	GetSet(key string, value any) Value
	MGet(keys []string) Values

	Set(key string, value any) error
	SetEx(key string, ttlSeconds int, value any) error
	SetNx(key string, value any) (bool, error)
	Incr(key string) (int, error)
	Incrby(key string, value int) (int, error)
	IncrByInt64(key string, value int64) (int64, error)
	DecrByInt64(key string, value int64) (int64, error)
	Del(key string) error

	TTL(key string) (int, error)
	Expire(key string, ttlSeconds int) error

	HGet(key, field string) Value
	HGetAll(key string) Values
	HSet(key, field string, value any) error
	HDel(key, field string) Value

	LPush(ctx context.Context, key string, value any) error
	LTrim(ctx context.Context, key string, start, stop int) error
	LLen(ctx context.Context, key string) (int, error)
	LRange(ctx context.Context, key string, start, stop int) Values

	// Ping checks the connection to the redis server.
	Ping(ctx context.Context) error

	// Keys returns all keys matching the glob pattern. NOTE: this command takes time
	// linear in the number of keys, and should not be run over large keyspaces.
	Keys(pattern string) ([]string, error)

	// WithContext will return a KeyValue that should respect ctx for all blocking operations.
	//
	// DEPRECATED: We are in the process of adding context to each individual operation (Get, Set, etc.)
	// Once that is complete, this method will be removed.
	WithContext(ctx context.Context) KeyValue
	WithLatencyRecorder(r LatencyRecorder) KeyValue

	// WithPrefix wraps r to return a RedisKeyValue that prefixes all keys with 'prefix:'
	WithPrefix(prefix string) KeyValue

	// Pool returns the underlying redis pool.
	// The intention of this API is Pool is only for advanced use cases and the caller
	// should consider if they need to use it. Pool is very hard to mock, while
	// the other functions on this interface are trivial to mock.
	Pool() *redis.Pool
}

// Value is a response from an operation on KeyValue. It provides convenient
// methods to get at the underlying value of the reply.
//
// Note: the available methods are based on current need. If you need to add
// another helper go for it.
type Value struct {
	reply any
	err   error
}

// NewValue returns a new Value for the given reply and err. Useful in tests using NewMockKeyValue.
func NewValue(reply any, err error) Value {
	return Value{reply: reply, err: err}
}

func (v Value) Bool() (bool, error) {
	return redis.Bool(v.reply, v.err)
}

func (v Value) Bytes() ([]byte, error) {
	return redis.Bytes(v.reply, v.err)
}

func (v Value) Int() (int, error) {
	return redis.Int(v.reply, v.err)
}

func (v Value) Int64() (int64, error) {
	return redis.Int64(v.reply, v.err)
}

func (v Value) String() (string, error) {
	return redis.String(v.reply, v.err)
}

func (v Value) IsNil() bool {
	return v.reply == nil
}

// Values is a response from an operation on KeyValue which returns multiple
// items. It provides convenient methods to get at the underlying value of the
// reply.
//
// Note: the available methods are based on current need. If you need to add
// another helper go for it.
type Values struct {
	reply interface{}
	err   error
}

func (v Values) ByteSlices() ([][]byte, error) {
	return redis.ByteSlices(redis.Values(v.reply, v.err))
}

func (v Values) Strings() ([]string, error) {
	return redis.Strings(v.reply, v.err)
}

func (v Values) StringMap() (map[string]string, error) {
	return redis.StringMap(v.reply, v.err)
}

type LatencyRecorder func(call string, latency time.Duration, err error)

// NewKeyValue returns a KeyValue for addr.
//
// poolOpts is a required argument which sets defaults in the case we connect
// to redis. If used we only override TestOnBorrow and Dial.
func NewKeyValue(addr string, poolOpts *redis.Pool) KeyValue {
	poolOpts.TestOnBorrow = func(c redis.Conn, t time.Time) error {
		_, err := c.Do("PING")
		return err
	}
	poolOpts.Dial = func() (redis.Conn, error) {
		return dialRedis(addr)
	}
	return &redisKeyValue{pool: poolOpts}
}

// NewTestKeyValue returns a KeyValue connected to a local Redis server for integration tests.
func NewTestKeyValue() KeyValue {
	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "127.0.0.1:6379")
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	return &redisKeyValue{pool: pool}
}

// redisKeyValue is a KeyValue backed by pool
type redisKeyValue struct {
	pool     *redis.Pool
	ctx      context.Context
	prefix   string
	recorder *LatencyRecorder
}

func (r *redisKeyValue) Ping(ctx context.Context) error {
	// The 'ping' command takes no arguments
	return r.do(ctx, "PING", []string{}, []any{}).err
}

func (r *redisKeyValue) Get(key string) Value {
	return r.doWithoutContext("GET", []string{key}, []any{})
}

func (r *redisKeyValue) GetSet(key string, val any) Value {
	return r.doWithoutContext("GETSET", []string{key}, []any{val})
}

func (r *redisKeyValue) MGet(keys []string) Values {
	return Values(r.doWithoutContext("MGET", keys, []any{}))
}

func (r *redisKeyValue) Set(key string, val any) error {
	return r.doWithoutContext("SET", []string{key}, []any{val}).err
}

func (r *redisKeyValue) SetEx(key string, ttlSeconds int, val any) error {
	return r.doWithoutContext("SETEX", []string{key}, []any{ttlSeconds, val}).err
}

func (r *redisKeyValue) SetNx(key string, val any) (bool, error) {
	_, err := r.doWithoutContext("SET", []string{key}, []any{val, "NX"}).String()
	if err == redis.ErrNil {
		return false, nil
	}
	return true, err
}

func (r *redisKeyValue) Incr(key string) (int, error) {
	return r.doWithoutContext("INCR", []string{key}, []any{}).Int()
}

func (r *redisKeyValue) Incrby(key string, value int) (int, error) {
	return r.doWithoutContext("INCRBY", []string{key}, []any{value}).Int()
}

func (r *redisKeyValue) IncrByInt64(key string, value int64) (int64, error) {
	return r.doWithoutContext("INCRBY", []string{key}, []any{value}).Int64()
}

func (r *redisKeyValue) DecrByInt64(key string, value int64) (int64, error) {
	return r.doWithoutContext("DECRBY", []string{key}, []any{value}).Int64()
}

func (r *redisKeyValue) Del(key string) error {
	return r.doWithoutContext("DEL", []string{key}, []any{}).err
}

func (r *redisKeyValue) TTL(key string) (int, error) {
	return r.doWithoutContext("TTL", []string{key}, []any{}).Int()
}

func (r *redisKeyValue) Expire(key string, ttlSeconds int) error {
	return r.doWithoutContext("EXPIRE", []string{key}, []any{ttlSeconds}).err
}

func (r *redisKeyValue) HGet(key, field string) Value {
	return r.doWithoutContext("HGET", []string{key}, []any{field})
}

func (r *redisKeyValue) HGetAll(key string) Values {
	return Values(r.doWithoutContext("HGETALL", []string{key}, []any{}))
}

func (r *redisKeyValue) HSet(key, field string, val any) error {
	return r.doWithoutContext("HSET", []string{key}, []any{field, val}).err
}

func (r *redisKeyValue) HDel(key, field string) Value {
	return r.doWithoutContext("HDEL", []string{key}, []any{field})
}

func (r *redisKeyValue) LPush(ctx context.Context, key string, value any) error {
	return r.do(ctx, "LPUSH", []string{key}, []any{value}).err
}
func (r *redisKeyValue) LTrim(ctx context.Context, key string, start, stop int) error {
	return r.do(ctx, "LTRIM", []string{key}, []any{start, stop}).err
}
func (r *redisKeyValue) LLen(ctx context.Context, key string) (int, error) {
	raw := r.do(ctx, "LLEN", []string{key}, []any{})
	return redis.Int(raw.reply, raw.err)
}

func (r *redisKeyValue) LRange(ctx context.Context, key string, start, stop int) Values {
	return Values(r.do(ctx, "LRANGE", []string{key}, []any{start, stop}))
}

func (r *redisKeyValue) WithContext(ctx context.Context) KeyValue {
	return &redisKeyValue{
		pool:   r.pool,
		ctx:    ctx,
		prefix: r.prefix,
	}
}

func (r *redisKeyValue) WithLatencyRecorder(rec LatencyRecorder) KeyValue {
	return &redisKeyValue{
		pool:     r.pool,
		ctx:      r.ctx,
		prefix:   r.prefix,
		recorder: &rec,
	}
}

func (r *redisKeyValue) WithPrefix(prefix string) KeyValue {
	return &redisKeyValue{
		pool:   r.pool,
		ctx:    r.ctx,
		prefix: r.prefix + prefix + ":",
	}
}

func (r *redisKeyValue) Keys(pattern string) ([]string, error) {
	return Values(r.doWithoutContext("KEYS", []string{pattern}, []any{})).Strings()
}

func (r *redisKeyValue) Pool() *redis.Pool {
	return r.pool
}

func (r *redisKeyValue) doWithoutContext(commandName string, keys []string, args []any) Value {
	// Fall back to the context field, or context.Background if it is not set.
	// Note: we will remove this logic (including r.ctx) once all operations take context directly.
	ctx := r.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return r.do(nil, commandName, keys, args)
}

func (r *redisKeyValue) do(ctx context.Context, commandName string, keys []string, args []any) Value {
	c, err := r.pool.GetContext(ctx)
	if err != nil {
		return Value{err: err}
	}
	defer c.Close()

	var start time.Time
	if r.recorder != nil {
		start = time.Now()
	}

	prefixedKeys := make([]any, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = r.prefix + key
	}

	reply, err := c.Do(commandName, append(prefixedKeys, args...)...)

	if r.recorder != nil {
		elapsed := time.Since(start)
		(*r.recorder)(commandName, elapsed, err)
	}
	return Value{
		reply: reply,
		err:   err,
	}
}
