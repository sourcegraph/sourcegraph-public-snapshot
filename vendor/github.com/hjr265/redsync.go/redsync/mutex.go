// Package redsync provides a Redis-based distributed mutual exclusion lock implementation as described in the blog post http://antirez.com/news/77.
//
// Values containing the types defined in this package should not be copied.
package redsync

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
	// DefaultExpiry is used when Mutex Duration is 0
	DefaultExpiry = 8 * time.Second
	// DefaultTries is used when Mutex Duration is 0
	DefaultTries = 16
	// DefaultDelay is used when Mutex Delay is 0
	DefaultDelay = 512 * time.Millisecond
	// DefaultFactor is used when Mutex Factor is 0
	DefaultFactor = 0.01
)

var (
	// ErrFailed is returned when lock cannot be acquired
	ErrFailed = errors.New("failed to acquire lock")
)

// Locker interface with Lock returning an error when lock cannot be aquired
type Locker interface {
	Lock() error
	Touch() bool
	Unlock() bool
}

// Pool is a generic connection pool
type Pool interface {
	Get() redis.Conn
}

var _ = Pool(&redis.Pool{})

// A Mutex is a mutual exclusion lock.
//
// Fields of a Mutex must not be changed after first use.
type Mutex struct {
	Name   string        // Resouce name
	Expiry time.Duration // Duration for which the lock is valid, DefaultExpiry if 0

	Tries int           // Number of attempts to acquire lock before admitting failure, DefaultTries if 0
	Delay time.Duration // Delay between two attempts to acquire lock, DefaultDelay if 0

	Factor float64 // Drift factor, DefaultFactor if 0

	Quorum int // Quorum for the lock, set to len(addrs)/2+1 by NewMutex()

	value string
	until time.Time

	nodes []Pool
	nodem sync.Mutex
}

var _ = Locker(&Mutex{})

// NewMutex returns a new Mutex on a named resource connected to the Redis instances at given addresses.
func NewMutex(name string, addrs []net.Addr) (*Mutex, error) {
	if len(addrs) == 0 {
		panic("redsync: addrs is empty")
	}

	nodes := make([]Pool, len(addrs))
	for i, addr := range addrs {
		dialTo := addr
		node := &redis.Pool{
			MaxActive: 1,
			Wait:      true,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", dialTo.String())
			},
		}
		nodes[i] = Pool(node)
	}

	return NewMutexWithGenericPool(name, nodes)
}

// NewMutexWithPool returns a new Mutex on a named resource connected to the Redis instances at given redis Pools.
func NewMutexWithPool(name string, nodes []*redis.Pool) (*Mutex, error) {
	if len(nodes) == 0 {
		panic("redsync: nodes is empty")
	}

	genericNodes := make([]Pool, len(nodes))
	for i, node := range nodes {
		genericNodes[i] = Pool(node)
	}

	return &Mutex{
		Name:   name,
		Quorum: len(genericNodes)/2 + 1,
		nodes:  genericNodes,
	}, nil
}

// NewMutexWithGenericPool returns a new Mutex on a named resource connected to the Redis instances at given generic Pools.
// different from NewMutexWithPool to maintain backwards compatibility
func NewMutexWithGenericPool(name string, genericNodes []Pool) (*Mutex, error) {
	if len(genericNodes) == 0 {
		panic("redsync: genericNodes is empty")
	}

	return &Mutex{
		Name:   name,
		Quorum: len(genericNodes)/2 + 1,
		nodes:  genericNodes,
	}, nil
}

// RedSync provides mutex handling via a multiple Redis connection pools.
type RedSync struct {
	pools []Pool
}

// New creates and returns a new RedSync instance from given network addresses.
func New(addrs []net.Addr) *RedSync {
	pools := make([]Pool, len(addrs))
	for i, addr := range addrs {
		dialTo := addr
		node := &redis.Pool{
			MaxActive: 1,
			Wait:      true,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", dialTo.String())
			},
		}
		pools[i] = Pool(node)
	}

	return &RedSync{pools}
}

// NewWithGenericPool creates and returns a new RedSync instance from given generic Pools.
func NewWithGenericPool(genericNodes []Pool) *RedSync {
	if len(genericNodes) == 0 {
		panic("redsync: genericNodes is empty")
	}

	return &RedSync{
		pools: genericNodes,
	}
}

// NewMutex returns a new Mutex with the given name.
func (r *RedSync) NewMutex(name string) *Mutex {
	return &Mutex{
		Name:   name,
		Quorum: len(r.pools)/2 + 1,
		nodes:  r.pools,
	}
}

// Lock locks m.
// In case it returns an error on failure, you may retry to acquire the lock by calling this method again.
func (m *Mutex) Lock() error {
	m.nodem.Lock()
	defer m.nodem.Unlock()

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return err
	}
	value := base64.StdEncoding.EncodeToString(b)

	expiry := m.Expiry
	if expiry == 0 {
		expiry = DefaultExpiry
	}

	retries := m.Tries
	if retries == 0 {
		retries = DefaultTries
	}

	for i := 0; i < retries; i++ {
		n := 0
		start := time.Now()
		for _, node := range m.nodes {
			if node == nil {
				continue
			}

			conn := node.Get()
			reply, err := redis.String(conn.Do("set", m.Name, value, "nx", "px", int(expiry/time.Millisecond)))
			conn.Close()
			if err != nil {
				continue
			}
			if reply != "OK" {
				continue
			}
			n++
		}

		factor := m.Factor
		if factor == 0 {
			factor = DefaultFactor
		}

		until := time.Now().Add(expiry - time.Now().Sub(start) - time.Duration(int64(float64(expiry)*factor)) + 2*time.Millisecond)
		if n >= m.Quorum && time.Now().Before(until) {
			m.value = value
			m.until = until
			return nil
		}
		for _, node := range m.nodes {
			if node == nil {
				continue
			}

			conn := node.Get()
			_, err := delScript.Do(conn, m.Name, value)
			conn.Close()
			if err != nil {
				continue
			}
		}

		// Have no delay on the last try so we can return ErrFailed sooner.
		if i == retries-1 {
			continue
		}

		delay := m.Delay
		if delay == 0 {
			delay = DefaultDelay
		}
		time.Sleep(delay)
	}

	return ErrFailed
}

// Touch resets m's expiry to the expiry value.
// It is a run-time error if m is not locked on entry to Touch.
// It returns the status of the touch
func (m *Mutex) Touch() bool {
	m.nodem.Lock()
	defer m.nodem.Unlock()

	value := m.value
	if value == "" {
		panic("redsync: touch of unlocked mutex")
	}

	expiry := m.Expiry
	if expiry == 0 {
		expiry = DefaultExpiry
	}
	reset := int(expiry / time.Millisecond)

	n := 0
	for _, node := range m.nodes {
		if node == nil {
			continue
		}

		conn := node.Get()
		reply, err := touchScript.Do(conn, m.Name, value, reset)
		conn.Close()
		if err != nil {
			continue
		}
		if reply != "OK" {
			continue
		}
		n++
	}
	if n >= m.Quorum {
		return true
	}
	return false
}

// Unlock unlocks m.
// It is a run-time error if m is not locked on entry to Unlock.
// It returns the status of the unlock
func (m *Mutex) Unlock() bool {
	m.nodem.Lock()
	defer m.nodem.Unlock()

	value := m.value
	if value == "" {
		panic("redsync: unlock of unlocked mutex")
	}

	m.value = ""
	m.until = time.Unix(0, 0)

	n := 0
	for _, node := range m.nodes {
		if node == nil {
			continue
		}

		conn := node.Get()
		status, err := delScript.Do(conn, m.Name, value)
		conn.Close()
		if err != nil {
			continue
		}
		if status == 0 {
			continue
		}
		n++
	}
	if n >= m.Quorum {
		return true
	}
	return false
}

var delScript = redis.NewScript(1, `
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
else
	return 0
end`)

var touchScript = redis.NewScript(1, `
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("set", KEYS[1], ARGV[1], "xx", "px", ARGV[2])
else
	return "ERR"
end`)
