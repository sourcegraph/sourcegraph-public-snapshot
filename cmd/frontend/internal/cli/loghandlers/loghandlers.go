// Package loghandlers contains log15 handlers/filters used by the sourcegraph
// cli
package loghandlers

import (
	"strings"
	"time"

	"github.com/inconshreveable/log15" //nolint:logging // Legacy loghandlers for log15
)

// Trace returns a filter for the given traces that run longer than threshold
func Trace(types []string, threshold time.Duration) func(*log15.Record) bool {
	all := false
	valid := map[string]bool{}
	for _, t := range types {
		valid[t] = true
		if t == "all" {
			all = true
		}
	}
	return func(r *log15.Record) bool {
		if r.Lvl != log15.LvlDebug {
			return true
		}
		if !strings.HasPrefix(r.Msg, "TRACE ") {
			return true
		}
		if !all && !valid[r.Msg[6:]] {
			return false
		}
		for i := 1; i < len(r.Ctx); i += 2 {
			if r.Ctx[i-1] != "duration" {
				continue
			}
			d, ok := r.Ctx[i].(time.Duration)
			return !ok || d >= threshold
		}
		return true
	}
}

// NotNoisey filters out high firing and low signal debug logs
func NotNoisey(r *log15.Record) bool {
	if r.Lvl != log15.LvlDebug {
		return true
	}
	noiseyPrefixes := []string{"repoUpdater: RefreshVCS"}
	for _, prefix := range noiseyPrefixes {
		if strings.HasPrefix(r.Msg, prefix) {
			return false
		}
	}
	if !strings.HasPrefix(r.Msg, "TRACE backend") || len(r.Ctx) < 2 {
		return true
	}
	rpc, ok := r.Ctx[1].(string)
	if !ok {
		return true
	}
	for _, n := range noiseyRPC {
		if rpc == n {
			return false
		}
	}
	return true
}

var noiseyRPC = []string{"MirrorRepos.RefreshVCS"}
