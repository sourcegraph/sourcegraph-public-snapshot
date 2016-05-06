package util

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"

	"github.com/mediocregopher/radix.v2/redis"
)

// LuaEval calls EVAL on the given Cmder for the given script, passing the key
// count and argument list in as well. See http://redis.io/commands/eval for
// more on how EVAL works and for the meaning of the keys argument.
//
// LuaEval will automatically try to call EVALSHA first in order to preserve
// bandwidth, and only falls back on EVAL if the script has never been used
// before.
//
// This method works with any of the Cmder's implemented in radix.v2, including
// Client, Pool, and Cluster.
//
//	r := util.LuaEval(c, `return redis.call('GET', KEYS[1])`, 1, "foo")
//
func LuaEval(c Cmder, script string, keys int, args ...interface{}) *redis.Resp {
	mainKey, _ := redis.KeyFromArgs(args...)

	sumRaw := sha1.Sum([]byte(script))
	sum := hex.EncodeToString(sumRaw[:])

	var r *redis.Resp
	if err := withClientForKey(c, mainKey, func(cc Cmder) {
		r = c.Cmd("EVALSHA", sum, keys, args)
		if r.Err != nil && strings.HasPrefix(r.Err.Error(), "NOSCRIPT") {
			r = c.Cmd("EVAL", script, keys, args)
		}
	}); err != nil {
		return redis.NewResp(err)
	}

	return r
}
