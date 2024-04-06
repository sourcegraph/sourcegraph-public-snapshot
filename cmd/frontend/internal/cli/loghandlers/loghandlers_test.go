package loghandlers

import (
	"testing"
	"time"

	"github.com/inconshreveable/log15" //nolint:logging // TODO move all logging to sourcegraph/log
)

func TestNotNoisey(t *testing.T) {
	keep := []log15.Record{
		mkRecord(log15.LvlDebug, "TRACE backend", "rpc", "Annotations.List", "spanID", "SPANID"),
		mkRecord(log15.LvlDebug, "TRACE backend", "rpc", "RepoTree.Get", "spanID", "SPANID", "duration", time.Second),
		mkRecord(log15.LvlWarn, "repoUpdater: RefreshVCS:", "err", "error"),
	}
	noisey := []log15.Record{mkRecord(log15.LvlDebug, "repoUpdater: RefreshVCS:", "err", "error")}
	for _, rpc := range noiseyRPC {
		noisey = append(noisey, mkRecord(log15.LvlDebug, "TRACE backend", "rpc", rpc))
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

var traces = []log15.Record{
	mkRecord(log15.LvlDebug, "TRACE backend", "rpc", "RepoTree.Get", "duration", time.Second),
	mkRecord(log15.LvlDebug, "TRACE HTTP", "routename", "repo.resolve", "duration", time.Second/3),
	mkRecord(log15.LvlDebug, "TRACE HTTP", "routename", "repo.resolve", "duration", 2*time.Second),
}

func TestTrace_All(t *testing.T) {
	f := Trace([]string{"all"}, 0)
	for _, r := range traces {
		if !f(&r) {
			t.Errorf("Should allow %v", r)
		}
	}
}

func TestTrace_None(t *testing.T) {
	f := Trace([]string{}, 0)
	for _, r := range traces {
		if f(&r) {
			t.Errorf("Should filter %v", r)
		}
	}
}

func TestTrace_Specific(t *testing.T) {
	f := Trace([]string{"HTTP"}, 0)
	for _, r := range traces {
		keep := r.Msg == "TRACE HTTP"
		if f(&r) == keep {
			continue
		} else if keep {
			t.Errorf("Should keep %v", r)
		} else {
			t.Errorf("Should filter %v", r)
		}
	}
}

func TestTrace_Threshold(t *testing.T) {
	threshold := time.Second
	f := Trace([]string{"all"}, threshold)
	for _, r := range traces {
		keep := r.Ctx[len(r.Ctx)-1].(time.Duration) >= threshold
		if f(&r) == keep {
			continue
		} else if keep {
			t.Errorf("Should keep %v for threshold %s", r, threshold)
		} else {
			t.Errorf("Should filter %v for threshold %s", r, threshold)
		}
	}
}

func mkRecord(lvl log15.Lvl, msg string, ctx ...any) log15.Record {
	return log15.Record{
		Lvl: lvl,
		Msg: msg,
		Ctx: ctx,
	}
}
