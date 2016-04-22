// Package loghandlers contains log15 handlers/filters used by the sourcegraph
// cli
package loghandlers

import (
	"strings"

	"gopkg.in/inconshreveable/log15.v2"
)

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
	if !strings.HasPrefix(r.Msg, "TRACE gRPC") || len(r.Ctx) < 2 {
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

var noiseyRPC = []string{"Builds.DequeueNext", "MirrorRepos.RefreshVCS"}
