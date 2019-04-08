package redispool

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseRedisConnectionUrl(t *testing.T) {
	redisConnectionConfiguration, err := ParseRedisConnectionUrl("")
	assert.Error(t, err)

	redisConnectionConfiguration, err = ParseRedisConnectionUrl("localhost"); if err != nil {
		t.Fatalf("expected to parse localhost")
	}
	assert.Equal(t, "localhost", redisConnectionConfiguration.Host)
	assert.Equal(t, "6379", redisConnectionConfiguration.Port)
	assert.Equal(t, "", redisConnectionConfiguration.Password)
	assert.Equal(t, "", redisConnectionConfiguration.Db)


	redisConnectionConfiguration, err = ParseRedisConnectionUrl("localhost:8080"); if err != nil {
		t.Fatalf("expected to parse localhost")
	}
	assert.Equal(t, "localhost", redisConnectionConfiguration.Host)
	assert.Equal(t, "8080", redisConnectionConfiguration.Port)
	assert.Equal(t, "", redisConnectionConfiguration.Password)
	assert.Equal(t, "", redisConnectionConfiguration.Db)


	redisConnectionConfiguration, err = ParseRedisConnectionUrl("localhost:8080/5"); if err != nil {
		t.Fatalf("expected to parse localhost")
	}
	assert.Equal(t, "localhost", redisConnectionConfiguration.Host)
	assert.Equal(t, "8080", redisConnectionConfiguration.Port)
	assert.Equal(t, "", redisConnectionConfiguration.Password)
	assert.Equal(t, "5", redisConnectionConfiguration.Db)

	redisConnectionConfiguration, err = ParseRedisConnectionUrl("notredis://localhost:8080/5"); if err != nil {
		t.Fatalf("expected to parse localhost")
	}
	assert.Equal(t, "notredis", redisConnectionConfiguration.Host)
	assert.Equal(t, "6379", redisConnectionConfiguration.Port)
	assert.Equal(t, "", redisConnectionConfiguration.Password)
	assert.Equal(t, "", redisConnectionConfiguration.Db)

	redisConnectionConfiguration, err = ParseRedisConnectionUrl("redis://localhost"); if err != nil {
		t.Fatalf("expected to parse localhost")
	}
	assert.Equal(t, "localhost", redisConnectionConfiguration.Host)
	assert.Equal(t, "6379", redisConnectionConfiguration.Port)
	assert.Equal(t, "", redisConnectionConfiguration.Password)
	assert.Equal(t, "", redisConnectionConfiguration.Db)

	redisConnectionConfiguration, err = ParseRedisConnectionUrl("redis://192.168.1.1"); if err != nil {
		t.Fatalf("expected to parse localhost")
	}
	assert.Equal(t, "192.168.1.1", redisConnectionConfiguration.Host)
	assert.Equal(t, "6379", redisConnectionConfiguration.Port)
	assert.Equal(t, "", redisConnectionConfiguration.Password)
	assert.Equal(t, "", redisConnectionConfiguration.Db)

	redisConnectionConfiguration, err = ParseRedisConnectionUrl("redis://192.168.1.1:1000"); if err != nil {
		t.Fatalf("expected to parse localhost")
	}
	assert.Equal(t, "192.168.1.1", redisConnectionConfiguration.Host)
	assert.Equal(t, "1000", redisConnectionConfiguration.Port)
	assert.Equal(t, "", redisConnectionConfiguration.Password)
	assert.Equal(t, "", redisConnectionConfiguration.Db)

	redisConnectionConfiguration, err = ParseRedisConnectionUrl("redis://192.168.1.1:1000/5"); if err != nil {
		t.Fatalf("expected to parse localhost")
	}
	assert.Equal(t, "192.168.1.1", redisConnectionConfiguration.Host)
	assert.Equal(t, "1000", redisConnectionConfiguration.Port)
	assert.Equal(t, "", redisConnectionConfiguration.Password)
	assert.Equal(t, "5", redisConnectionConfiguration.Db)

	redisConnectionConfiguration, err = ParseRedisConnectionUrl("redis://:test@192.168.1.1:1000/5"); if err != nil {
		t.Fatalf("expected to parse localhost")
	}
	assert.Equal(t, "192.168.1.1", redisConnectionConfiguration.Host)
	assert.Equal(t, "1000", redisConnectionConfiguration.Port)
	assert.Equal(t, "test", redisConnectionConfiguration.Password)
	assert.Equal(t, "5", redisConnectionConfiguration.Db)

	redisConnectionConfiguration, err = ParseRedisConnectionUrl("redis://hello:test@192.168.1.1:1000/5"); if err != nil {
		t.Fatalf("expected to parse localhost")
	}
	assert.Equal(t, "192.168.1.1", redisConnectionConfiguration.Host)
	assert.Equal(t, "1000", redisConnectionConfiguration.Port)
	assert.Equal(t, "test", redisConnectionConfiguration.Password)
	assert.Equal(t, "5", redisConnectionConfiguration.Db)
}
