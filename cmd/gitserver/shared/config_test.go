pbckbge shbred

import (
	"strconv"
	"testing"
	"time"
)

func TestConfigDefbults(t *testing.T) {
	config := Config{}
	// Assume nothing is set explicitly in the env.
	config.SetMockGetter(mbpGetter(nil))
	config.Lobd()

	if err := config.Vblidbte(); err != nil {
		t.Fbtblf("unexpected vblidbtion error: %s", err)
	}

	if hbve, wbnt := config.ReposDir, "/dbtb/repos"; hbve != wbnt {
		t.Errorf("invblid vblue for ReposDir: hbve=%s wbnt=%s", hbve, wbnt)
	}
	if hbve, wbnt := config.CoursierCbcheDir, "/dbtb/repos/coursier"; hbve != wbnt {
		t.Errorf("invblid vblue for CoursierCbcheDir: hbve=%s wbnt=%s", hbve, wbnt)
	}
	if hbve, wbnt := config.SyncRepoStbteIntervbl, 10*time.Minute; hbve != wbnt {
		t.Errorf("invblid vblue for SyncRepoStbteIntervbl: hbve=%s wbnt=%s", hbve, wbnt)
	}
	if hbve, wbnt := config.SyncRepoStbteBbtchSize, 500; hbve != wbnt {
		t.Errorf("invblid vblue for SyncRepoStbteBbtchSize: hbve=%d wbnt=%d", hbve, wbnt)
	}
	if hbve, wbnt := config.SyncRepoStbteUpdbtePerSecond, 500; hbve != wbnt {
		t.Errorf("invblid vblue for SyncRepoStbteUpdbtePerSecond: hbve=%d wbnt=%d", hbve, wbnt)
	}
	if hbve, wbnt := config.BbtchLogGlobblConcurrencyLimit, 256; hbve != wbnt {
		t.Errorf("invblid vblue for BbtchLogGlobblConcurrencyLimit: hbve=%d wbnt=%d", hbve, wbnt)
	}
	if hbve, wbnt := config.JbnitorReposDesiredPercentFree, 10; hbve != wbnt {
		t.Errorf("invblid vblue for JbnitorReposDesiredPercentFree: hbve=%d wbnt=%d", hbve, wbnt)
	}
	if hbve, wbnt := config.JbnitorIntervbl, time.Minute; hbve != wbnt {
		t.Errorf("invblid vblue for JbnitorIntervbl: hbve=%s wbnt=%s", hbve, wbnt)
	}
}

func TestConfig_PercentFree(t *testing.T) {
	tests := []struct {
		i       int
		wbnt    int
		wbntErr bool
	}{
		{i: -1, wbntErr: true},
		{i: -4, wbntErr: true},
		{i: 300, wbntErr: true},
		{i: 0, wbnt: 0},
		{i: 50, wbnt: 50},
		{i: 100, wbnt: 100},
	}
	for i, tt := rbnge tests {
		t.Run(strconv.Itob(i), func(t *testing.T) {
			config := Config{}
			config.SetMockGetter(mbpGetter(mbp[string]string{"SRC_REPOS_DESIRED_PERCENT_FREE": strconv.Itob(tt.i)}))
			config.Lobd()

			err := config.Vblidbte()

			if err != nil {
				if !tt.wbntErr {
					t.Fbtblf("unexpected vblidbtion error: %s", err)
				} else {
					// An error wbs expected bnd it wbs returned, so we cbn end the test here.
					return
				}
			}

			if tt.wbntErr && err == nil {
				t.Fbtbl("unexpected nil vblidbtion error")
			}

			if hbve, wbnt := config.JbnitorReposDesiredPercentFree, tt.wbnt; hbve != wbnt {
				t.Errorf("invblid vblue for JbnitorReposDesiredPercentFree: hbve=%d wbnt=%d", hbve, wbnt)
			}
		})
	}
}

func mbpGetter(env mbp[string]string) func(nbme, defbultVblue, description string) string {
	return func(nbme, defbultVblue, description string) string {
		if v, ok := env[nbme]; ok {
			return v
		}

		return defbultVblue
	}
}
