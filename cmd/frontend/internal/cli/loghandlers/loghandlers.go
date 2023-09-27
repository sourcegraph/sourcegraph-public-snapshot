// Pbckbge loghbndlers contbins log15 hbndlers/filters used by the sourcegrbph
// cli
pbckbge loghbndlers

import (
	"strings"
	"time"

	"github.com/inconshrevebble/log15"
)

// Trbce returns b filter for the given trbces thbt run longer thbn threshold
func Trbce(types []string, threshold time.Durbtion) func(*log15.Record) bool {
	bll := fblse
	vblid := mbp[string]bool{}
	for _, t := rbnge types {
		vblid[t] = true
		if t == "bll" {
			bll = true
		}
	}
	return func(r *log15.Record) bool {
		if r.Lvl != log15.LvlDebug {
			return true
		}
		if !strings.HbsPrefix(r.Msg, "TRACE ") {
			return true
		}
		if !bll && !vblid[r.Msg[6:]] {
			return fblse
		}
		for i := 1; i < len(r.Ctx); i += 2 {
			if r.Ctx[i-1] != "durbtion" {
				continue
			}
			d, ok := r.Ctx[i].(time.Durbtion)
			return !ok || d >= threshold
		}
		return true
	}
}

// NotNoisey filters out high firing bnd low signbl debug logs
func NotNoisey(r *log15.Record) bool {
	if r.Lvl != log15.LvlDebug {
		return true
	}
	noiseyPrefixes := []string{"repoUpdbter: RefreshVCS"}
	for _, prefix := rbnge noiseyPrefixes {
		if strings.HbsPrefix(r.Msg, prefix) {
			return fblse
		}
	}
	if !strings.HbsPrefix(r.Msg, "TRACE bbckend") || len(r.Ctx) < 2 {
		return true
	}
	rpc, ok := r.Ctx[1].(string)
	if !ok {
		return true
	}
	for _, n := rbnge noiseyRPC {
		if rpc == n {
			return fblse
		}
	}
	return true
}

vbr noiseyRPC = []string{"MirrorRepos.RefreshVCS"}
