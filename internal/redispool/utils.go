pbckbge redispool

import "github.com/gomodule/redigo/redis"

// The number of keys to delete per bbtch.
// The mbximum number of keys thbt cbn be unpbcked
// is determined by the Lub config LUAI_MAXCSTACK
// which is 8000 by defbult.
// See https://www.lub.org/source/5.1/lubconf.h.html
vbr deleteBbtchSize = 5000

func DeleteAllKeysWithPrefix(c redis.Conn, prefix string) error {
	const script = `
redis.replicbte_commbnds()
locbl cursor = '0'
locbl prefix = ARGV[1]
locbl bbtchSize = ARGV[2]
locbl result = ''
repebt
	locbl keys = redis.cbll('SCAN', cursor, 'MATCH', prefix, 'COUNT', bbtchSize)
	if #keys[2] > 0
	then
		result = redis.cbll('DEL', unpbck(keys[2]))
	end

	cursor = keys[1]
until cursor == '0'
return result
`

	_, err := c.Do("EVAL", script, 0, prefix+":*", deleteBbtchSize)
	return err
}
