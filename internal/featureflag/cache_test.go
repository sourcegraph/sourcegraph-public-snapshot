package featureflag

import (
	"testing"

	"github.com/gomodule/redigo/redis"
)

//go:generate ../../dev/mockgen.sh github.com/gomodule/redigo/redis -o mock_conn_test.go -i Conn

func setupRedisTest(t *testing.T, mockConn redis.Conn) {
	oldPool := pool
	t.Cleanup(func() { pool = oldPool })

	pool = redis.NewPool(func() (redis.Conn, error) {
		return mockConn, nil
	}, 100)
}
