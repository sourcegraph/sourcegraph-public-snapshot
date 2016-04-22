package loghandlers

import (
	"testing"
	"time"

	"gopkg.in/inconshreveable/log15.v2"
)

func TestNotNoisey(t *testing.T) {
	keep := []log15.Record{
		mkRecord(log15.LvlDebug, "TRACE gRPC", "rpc", "Annotations.List", "spanID", "SPANID"),
		mkRecord(log15.LvlDebug, "TRACE gRPC", "rpc", "RepoTree.Get", "spanID", "SPANID", "duration", time.Second),
		mkRecord(log15.LvlWarn, "repoUpdater: RefreshVCS:", "err", "error"),
	}
	noisey := []log15.Record{mkRecord(log15.LvlDebug, "repoUpdater: RefreshVCS:", "err", "error")}
	for _, rpc := range noiseyRPC {
		noisey = append(noisey, mkRecord(log15.LvlDebug, "TRACE gRPC", "rpc", rpc))
	}

	for _, r := range keep {
		if !NotNoisey(&r) {
			t.Errorf("Should keep %v", r)
		}
	}
	for _, r := range noisey {
		if NotNoisey(&r) {
			t.Errorf("Should filter out %v", r)
		}
	}
}

func mkRecord(lvl log15.Lvl, msg string, ctx ...interface{}) log15.Record {
	return log15.Record{
		Lvl: lvl,
		Msg: msg,
		Ctx: ctx,
	}
}
