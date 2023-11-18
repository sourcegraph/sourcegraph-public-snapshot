package redispool

import (
	"context"
	"time"

	"github.com/gomodule/redigo/redis"
)

// NaiveValue is the value we send to and from a NaiveKeyValueStore. This
// represents the marshalled value the NaiveKeyValueStore operates on. See the
// unexported redisValue type for more details. However, NaiveKeyValueStore
// should treat this value as opaque.
//
// Note: strings are used to ensure we pass copies around and avoid mutating
// values. They should not be treated as utf8 text.
type NaiveValue string

// NaiveUpdater operates on the value for a key in a NaiveKeyValueStore.
// before is the before value in the store, found is if the key exists in the
// store. after is the new value for it that needs to be stored, or remove is
// true if the key should be removed.
//
// Note: a store should do this update atomically/under concurrency control.
type NaiveUpdater func(before NaiveValue, found bool) (after NaiveValue, remove bool)

// NaiveKeyValueStore is a function on a store which runs f for key.
//
// This minimal function allows us to implement the full functionality of
// KeyValue via FromNaiveKeyValueStore. This does mean for any read on key we
// have to read the full value, and any mutation requires rewriting the full
// value. This is usually fine, but may be an issue when backed by a large
// Hash or List. As such this function is designed with the functionality of
// Cody App in mind (single process, low traffic).
type NaiveKeyValueStore func(ctx context.Context, key string, f NaiveUpdater) error

// FromNaiveKeyValueStore returns a KeyValue based on the store function.
func FromNaiveKeyValueStore(store NaiveKeyValueStore) KeyValue {
	return &naiveKeyValue{
		store: store,
		ctx:   context.Background(),
	}
}

// naiveKeyValue wraps a store to provide the KeyValue interface. Nearly all
// operations go via maybeUpdateGroup method, sink your teeth into that first
// to fully understand how to expand the set of methods provided.
type naiveKeyValue struct {
	store NaiveKeyValueStore
	ctx   context.Context
}

func (kv *naiveKeyValue) Get(key string) Value {
	return kv.maybeUpdateGroup(redisGroupString, key, func(v redisValue, found bool) (redisValue, updaterOp, error) {
		return v, readOnly, nil
	})
}

func (kv *naiveKeyValue) GetSet(key string, value any) Value {
	var oldValue Value
	v := kv.maybeUpdateGroup(redisGroupString, key, func(before redisValue, found bool) (redisValue, updaterOp, error) {
		if found {
			oldValue.reply = before.Reply
		} else {
			oldValue.err = redis.ErrNil
		}

		return redisValue{
			Group: redisGroupString,
			Reply: value,
		}, write, nil
	})
	if v.err != nil {
		return v
	}
	return oldValue
}

func (kv *naiveKeyValue) Set(key string, value any) error {
	return kv.maybeUpdate(key, func(_ redisValue, _ bool) (redisValue, updaterOp, error) {
		return redisValue{
			Group: redisGroupString,
			Reply: value,
		}, write, nil
	}).err
}

func (kv *naiveKeyValue) SetEx(key string, ttlSeconds int, value any) error {
	return kv.maybeUpdate(key, func(_ redisValue, _ bool) (redisValue, updaterOp, error) {
		return redisValue{
			Group:        redisGroupString,
			Reply:        value,
			DeadlineUnix: time.Now().UTC().Unix() + int64(ttlSeconds),
		}, write, nil
	}).err
}

func (kv *naiveKeyValue) SetNx(key string, value any) (bool, error) {
	op := write
	v := kv.maybeUpdate(key, func(_ redisValue, found bool) (redisValue, updaterOp, error) {
		if found {
			op = readOnly
		}
		return redisValue{
			Group: redisGroupString,
			Reply: value,
		}, op, nil
	})
	if v.err != nil {
		return false, v.err
	}
	return op == write, nil
}

func (kv *naiveKeyValue) Incr(key string) (int, error) {
	return kv.maybeUpdateGroup(redisGroupString, key, func(value redisValue, found bool) (redisValue, updaterOp, error) {
		if !found {
			return redisValue{
				Group: redisGroupString,
				Reply: int64(1),
			}, write, nil
		}

		num, err := redis.Int(value.Reply, nil)
		if err != nil {
			return value, readOnly, err
		}

		value.Reply = int64(num + 1)
		return value, write, nil
	}).Int()
}

func (kv *naiveKeyValue) Incrby(key string, val int) (int, error) {
	return kv.maybeUpdateGroup(redisGroupString, key, func(value redisValue, found bool) (redisValue, updaterOp, error) {
		if !found {
			return redisValue{
				Group: redisGroupString,
				Reply: int64(val),
			}, write, nil
		}

		num, err := redis.Int(value.Reply, nil)
		if err != nil {
			return value, readOnly, err
		}

		value.Reply = int64(num + val)
		return value, write, nil
	}).Int()
}

func (kv *naiveKeyValue) Del(key string) error {
	return kv.store(kv.ctx, key, func(_ NaiveValue, _ bool) (NaiveValue, bool) {
		return "", true
	})
}

func (kv *naiveKeyValue) TTL(key string) (int, error) {
	const ttlUnset = -1
	const ttlDoesNotExist = -2
	var ttl int
	err := kv.maybeUpdate(key, func(value redisValue, found bool) (redisValue, updaterOp, error) {
		if !found {
			ttl = ttlDoesNotExist
		} else if value.DeadlineUnix == 0 {
			ttl = ttlUnset
		} else {
			ttl = int(value.DeadlineUnix - time.Now().UTC().Unix())
			// we may have expired since doStore checked
			if ttl <= 0 {
				ttl = ttlDoesNotExist
			}
		}

		return value, readOnly, nil
	}).err

	if err == redis.ErrNil {
		// Already handled above, but just in case lets be explicit
		ttl = ttlDoesNotExist
		err = nil
	}

	return ttl, err
}

func (kv *naiveKeyValue) Expire(key string, ttlSeconds int) error {
	err := kv.maybeUpdate(key, func(value redisValue, found bool) (redisValue, updaterOp, error) {
		if !found {
			return value, readOnly, nil
		}

		value.DeadlineUnix = time.Now().UTC().Unix() + int64(ttlSeconds)
		return value, write, nil
	}).err

	// expire does not error if the key does not exist
	if err == redis.ErrNil {
		err = nil
	}

	return err
}

func (kv *naiveKeyValue) HGet(key, field string) Value {
	var reply any
	err := kv.maybeUpdateValues(redisGroupHash, key, func(li []any) ([]any, updaterOp, error) {
		idx, ok, err := hsetValueIndex(li, field)
		if err != nil {
			return li, readOnly, err
		}
		if !ok {
			return li, readOnly, redis.ErrNil
		}

		reply = li[idx]
		return li, readOnly, nil
	}).err
	return Value{reply: reply, err: err}
}

func (kv *naiveKeyValue) HGetAll(key string) Values {
	return Values(kv.maybeUpdateGroup(redisGroupHash, key, func(value redisValue, found bool) (redisValue, updaterOp, error) {
		return value, readOnly, nil
	}))
}

func (kv *naiveKeyValue) HSet(key, field string, fieldValue any) error {
	return kv.maybeUpdateValues(redisGroupHash, key, func(li []any) ([]any, updaterOp, error) {
		idx, ok, err := hsetValueIndex(li, field)
		if err != nil {
			return li, readOnly, err
		}
		if ok {
			li[idx] = fieldValue
		} else {
			li = append(li, field, fieldValue)
		}

		return li, write, nil
	}).err
}

func (kv *naiveKeyValue) HDel(key, field string) Value {
	var removed int64
	err := kv.maybeUpdateValues(redisGroupHash, key, func(li []any) ([]any, updaterOp, error) {
		idx, ok, err := hsetValueIndex(li, field)
		if err != nil || !ok {
			return li, readOnly, err
		}
		removed = 1
		li = append(li[:idx-1], li[idx+1:]...)
		return li, write, nil
	}).err
	return Value{reply: removed, err: err}
}

func hsetValueIndex(li []any, field string) (int, bool, error) {
	for i := 1; i < len(li); i += 2 {
		if kk, err := redis.String(li[i-1], nil); err != nil {
			return -1, false, err
		} else if kk == field {
			return i, true, nil
		}
	}
	return -1, false, nil
}

func (kv *naiveKeyValue) LPush(key string, value any) error {
	return kv.maybeUpdateValues(redisGroupList, key, func(li []any) ([]any, updaterOp, error) {
		return append([]any{value}, li...), write, nil
	}).err
}

func (kv *naiveKeyValue) LTrim(key string, start, stop int) error {
	return kv.maybeUpdateValues(redisGroupList, key, func(li []any) ([]any, updaterOp, error) {
		beforeLen := len(li)
		li = lrange(li, start, stop)

		op := readOnly
		if len(li) != beforeLen {
			op = write
		}

		return li, op, nil
	}).err
}

func (kv *naiveKeyValue) LLen(key string) (int, error) {
	var innerLi []any
	err := kv.maybeUpdateValues(redisGroupList, key, func(li []any) ([]any, updaterOp, error) {
		innerLi = li
		return li, readOnly, nil
	}).err
	return len(innerLi), err
}

func (kv *naiveKeyValue) LRange(key string, start, stop int) Values {
	var innerLi []any
	err := kv.maybeUpdateValues(redisGroupList, key, func(li []any) ([]any, updaterOp, error) {
		innerLi = li
		return li, readOnly, nil
	}).err
	if err != nil {
		return Values{err: err}
	}
	return Values{reply: lrange(innerLi, start, stop)}
}

func lrange(li []any, start, stop int) []any {
	low, high := rangeOffsetsToHighLow(start, stop, len(li))
	if high <= low {
		return []any(nil)
	}
	return li[low:high]
}

func rangeOffsetsToHighLow(start, stop, size int) (low, high int) {
	if size <= 0 {
		return 0, 0
	}

	start = clampRangeOffset(0, size, start)
	stop = clampRangeOffset(-1, size, stop)

	// Adjust inclusive ending into exclusive for go
	low = start
	high = stop + 1

	return low, high
}

func clampRangeOffset(low, high, offset int) int {
	// negative offset means distance from high
	if offset < 0 {
		offset = high + offset
	}
	if offset < low {
		return low
	}
	if offset >= high {
		return high - 1
	}
	return offset
}

func (kv *naiveKeyValue) WithContext(ctx context.Context) KeyValue {
	return &naiveKeyValue{
		store: kv.store,
		ctx:   ctx,
	}
}

func (kv *naiveKeyValue) Pool() (pool *redis.Pool, ok bool) {
	return nil, false
}

type updaterOp bool

var (
	write    updaterOp = true
	readOnly updaterOp = false
)

// storeUpdater operates on the redisValue for a key and returns its new value
// or error. See doStore for more information.
type storeUpdater func(before redisValue, found bool) (after redisValue, op updaterOp, err error)

// maybeUpdate is a helper for NaiveKeyValueStore and NaiveUpdater. It
// provides consistent behaviour for KeyValue as well as reducing the work
// required for each KeyValue method. It does the following:
//
//   - Marshal NaiveUpdater values to and from redisValue
//   - Handle expiration so updater does not need to.
//   - If a value becomes nil we can delete the key. (redis behaviour)
//   - Handle updaters that only want to read (readOnly updaterOp, error)
func (kv *naiveKeyValue) maybeUpdate(key string, updater storeUpdater) Value {
	var returnValue Value
	storeErr := kv.store(kv.ctx, key, func(beforeRaw NaiveValue, found bool) (NaiveValue, bool) {
		var before redisValue
		defaultDelete := false
		if found {
			// We found the value so we can unmarshal it.
			if err := before.Unmarshal([]byte(beforeRaw)); err != nil {
				// Bad data at key, delete it and return an error
				returnValue.err = err
				return "", true
			}

			// The store won't expire for us, we do it here by checking at
			// read time if the value has expired. If it has pretend we didn't
			// find it and mark that we need to delete the value if we don't
			// get a new one.
			if before.DeadlineUnix != 0 && time.Now().UTC().Unix() >= before.DeadlineUnix {
				found = false
				// We need to inform the store to delete the value, unless we
				// have a new value to takes its place.
				defaultDelete = true
			}
		}

		// Call out to the provided updater to get back what we need to do to
		// the value.
		after, op, err := updater(before, found)
		if err != nil {
			// If updater fails, we tell store to keep the before value (or
			// delete if expired).
			returnValue.err = err
			return beforeRaw, defaultDelete
		}

		// We don't need to update the value, so set the appropriate response
		// values based on what we found at get time.
		if op == readOnly {
			if found {
				returnValue.reply = before.Reply
			} else {
				returnValue.err = redis.ErrNil
			}
			return beforeRaw, defaultDelete
		}

		// Redis will automatically delete keys if some value types become
		// empty.
		if isRedisDeleteValue(after) {
			returnValue.reply = after.Reply
			return "", true
		}

		// Lets convert our redisValue into bytes so we can store the new
		// value.
		afterRaw, err := after.Marshal()
		if err != nil {
			returnValue.err = err
			return beforeRaw, defaultDelete
		}
		returnValue.reply = after.Reply
		return NaiveValue(afterRaw), false
	})
	if storeErr != nil {
		return Value{err: storeErr}
	}
	return returnValue
}

// maybeUpdateGroup is a wrapper of maybeUpdate which additionally will return
// an error if the before is not of the type group.
func (kv *naiveKeyValue) maybeUpdateGroup(group redisGroup, key string, updater storeUpdater) Value {
	return kv.maybeUpdate(key, func(value redisValue, found bool) (redisValue, updaterOp, error) {
		if found && value.Group != group {
			return value, readOnly, redis.Error("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		return updater(value, found)
	})
}

// valuesUpdater takes in the befores. If afters is different op must
// be write so maybeUpdateValues knows to update.
type valuesUpdater func(befores []any) (afters []any, op updaterOp, err error)

// maybeUpdateValues is a specialization of maybeUpdate for all values operations
// on key via updater.
func (kv *naiveKeyValue) maybeUpdateValues(group redisGroup, key string, updater valuesUpdater) Values {
	v := kv.maybeUpdateGroup(group, key, func(value redisValue, found bool) (redisValue, updaterOp, error) {
		var li []any
		if found {
			var err error
			li, err = value.Values()
			if err != nil {
				return value, readOnly, err
			}
		} else {
			value = redisValue{
				Group: group,
			}
		}

		li, op, err := updater(li)
		value.Reply = li
		return value, op, err
	})

	// missing is treated as empty for values
	if v.err == redis.ErrNil {
		return Values{reply: []any(nil)}
	}

	return Values(v)
}

// isRedisDeleteValue returns true if the redisValue is not allowed to be
// stored. An example of this is when a list becomes empty redis will delete
// the key.
func isRedisDeleteValue(v redisValue) bool {
	switch v.Group {
	case redisGroupString:
		return false
	case redisGroupHash, redisGroupList:
		vs, _ := v.Reply.([]any)
		return len(vs) == 0
	default:
		return false
	}
}
