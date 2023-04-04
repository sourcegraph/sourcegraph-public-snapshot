package redispool

import (
	"bytes"

	"github.com/gomodule/redigo/redis"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// redisGroup is which type of data we have. We use the term group since that
// is what redis uses in its documentation to segregate the different types of
// commands you can run (string, list, hash).
type redisGroup byte

const (
	// redisGroupString doesn't mean the data is a string. This is the
	// original group of command (get, set).
	redisGroupString redisGroup = 's'
	redisGroupList   redisGroup = 'l'
	redisGroupHash   redisGroup = 'h'
)

// redisValue represents what we marshal into a NaiveKeyValueStore.
type redisValue struct {
	// Group is stored so we can enforce WRONGTYPE errors.
	Group redisGroup
	// Reply is the actual value the user wants to set. It is called reply to
	// match what the redis package expects.
	Reply any
	// DeadlineUnix is the unix timestamp of when to expire this value, or 0
	// if no expiry. This is a convenient way to store deadlines since redis
	// only has 1s resolution on TTLs.
	DeadlineUnix int64
}

func (v *redisValue) Marshal() ([]byte, error) {
	var c conn

	// Header, Group, DeadlineUnix
	//
	// Note: this writes a small version header which is just the character !
	// and g. This is enough so we can change the data in the future.
	//
	// We are also gaurenteed to not fail these writes, so we ignore the error
	// for convenience.
	_ = c.bw.WriteByte('!')
	_ = c.bw.WriteByte(byte(v.Group))
	_ = c.writeArg(v.DeadlineUnix)

	// Reply
	switch v.Group {
	case redisGroupString:
		err := c.writeArg(v.Reply)
		if err != nil {
			return nil, err
		}
	case redisGroupList, redisGroupHash:
		vs, err := v.Values()
		if err != nil {
			return nil, err
		}
		_ = c.writeLen('*', len(vs))
		for _, el := range vs {
			err := c.writeArg(el)
			if err != nil {
				return nil, err
			}
		}
	default:
		return nil, errors.Errorf("redis naive internal error: unkown redis group %c", byte(v.Group))
	}

	return c.bw.Bytes(), nil
}

func (v *redisValue) Unmarshal(b []byte) error {
	c := conn{bw: *bytes.NewBuffer(b)}

	// Header, Group
	var header [2]byte
	n, err := c.bw.Read(header[:])
	if err != nil || n != 2 {
		return errors.New("redis naive internal error: failed to parse value header")
	}
	if header[0] != '!' {
		return errors.Errorf("redis naive internal error: expected first byte of value header to be '!' got %q", header[0])
	}
	v.Group = redisGroup(header[1])

	// DeadlineUnix
	v.DeadlineUnix, err = redis.Int64(c.readReply())
	if err != nil {
		return errors.Wrap(err, "redis naive internal error: failed to parse value deadline")
	}

	// Reply
	v.Reply, err = c.readReply()
	if err != nil {
		return err
	}

	// Validation
	switch v.Group {
	case redisGroupString:
		// noop
	case redisGroupList, redisGroupHash:
		_, err := v.Values()
		if err != nil {
			return err
		}
	default:
		return errors.Errorf("redis naive internal error: unkown redis group %c", byte(v.Group))
	}

	return nil
}

// Values will convert v.Reply into values as well as some validation based on
// the v.Group.
func (v *redisValue) Values() ([]any, error) {
	li, ok := v.Reply.([]any)
	if !ok {
		return nil, errors.Errorf("redis naive internal error: non list returned for redis group %c", byte(v.Group))
	}
	if v.Group == redisGroupHash && len(li)%2 != 0 {
		return nil, errors.New("redis naive internal error: hash list is not divisible by 2")
	}
	return li, nil
}
