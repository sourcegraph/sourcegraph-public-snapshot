package redispool

import (
	"context"
	"fmt"

	"github.com/gomodule/redigo/redis"

	"github.com/sourcegraph/sourcegraph/internal/tenant"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// The number of keys to delete per batch.
// The maximum number of keys that can be unpacked
// is determined by the Lua config LUAI_MAXCSTACK
// which is 8000 by default.
// See https://www.lua.org/source/5.1/luaconf.h.html
var deleteBatchSize = 5000

func DeleteAllKeysWithPrefix(ctx context.Context, c redis.Conn, prefix string) error {
	const script = `
redis.replicate_commands()
local cursor = '0'
local prefix = ARGV[1]
local batchSize = ARGV[2]
local result = ''
repeat
	local keys = redis.call('SCAN', cursor, 'MATCH', prefix, 'COUNT', batchSize)
	if #keys[2] > 0
	then
		result = redis.call('DEL', unpack(keys[2]))
	end

	cursor = keys[1]
until cursor == '0'
return result
`

	t := tenant.FromContext(ctx)
	if t.ID() == 0 {
		return errors.New("tenant not set")
	}

	_, err := c.Do("EVAL", script, 0, fmt.Sprintf("tnt_%d:", t.ID())+prefix+":*", deleteBatchSize)
	return err
}
