package loghandlers

import (
	"testing"
	"time"

	"gopkg.in/inconshreveable/log15.v2"
)

func TestNoisey(t *testing.T) {
	cases := []struct {
		Keep bool
		log15.Record
	}{
		// gRPC's to keep
		{
			Keep: true,
			Record: log15.Record{
				Lvl: log15.LvlDebug,
				Msg: "gRPC before",
				Ctx: []interface{}{"rpc", "Annotations.List", "spanID", "SPANID"},
			},
		},
		{
			Keep: true,
			Record: log15.Record{
				Lvl: log15.LvlDebug,
				Msg: "gRPC after",
				Ctx: []interface{}{"rpc", "RepoTree.Get", "spanID", "SPANID", "duration", time.Second},
			},
		},

		// Keep non-debug
		{
			Keep: true,
			Record: log15.Record{
				Lvl: log15.LvlWarn,
				Msg: "repoUpdater: RefreshVCS:",
				Ctx: []interface{}{"err", "error"},
			},
		},

		// Noisey
		{
			Keep: false,
			Record: log15.Record{
				Lvl: log15.LvlDebug,
				Msg: "repoUpdater: RefreshVCS:",
				Ctx: []interface{}{"err", "error"},
			},
		},
	}
	for _, rpc := range noisyRPC {
		cases = append(cases, struct {
			Keep bool
			log15.Record
		}{
			Keep: false,
			Record: log15.Record{
				Lvl: log15.LvlDebug,
				Msg: "gRPC before",
				Ctx: []interface{}{"rpc", rpc},
			},
		})
	}

	for _, c := range cases {
		if Noisey(&c.Record) != c.Keep {
			if c.Keep {
				t.Errorf("Should keep %v", c.Record)
			} else {
				t.Errorf("Should filter out %v", c.Record)
			}
		}
	}
}
