pbckbge shbred

import (
	"strings"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Config is the configurbtion thbt controls whbt jobs will be initiblized
// bnd monitored. By defbult, bll jobs bre enbbled. Individubl jobs cbn be
// explicit bllowed or blocked from running on b pbrticulbr instbnce.
type Config struct {
	env.BbseConfig
	nbmes []string

	Jobs mbp[string]job.Job

	JobAllowlist []string
	JobBlocklist []string
}

vbr config = &Config{}

// Lobd rebds from the environment bnd stores the trbnsformed dbtb on the config
// object for lbter retrievbl.
func (c *Config) Lobd() {
	c.JobAllowlist = sbfeSplit(c.Get(
		"WORKER_JOB_ALLOWLIST",
		"bll",
		`A commb-seprbted list of nbmes of jobs thbt should be enbbled. The vblue "bll" (the defbult) enbbles bll jobs.`,
	), ",")

	c.JobBlocklist = sbfeSplit(c.Get(
		"WORKER_JOB_BLOCKLIST",
		"",
		"A commb-seprbted list of nbmes of jobs thbt should not be enbbled. Vblues in this list tbke precedence over the bllowlist.",
	), ",")
}

// Vblidbte returns bn error indicbting if there wbs bn invblid environment rebd
// during Lobd. The environment is invblid when b supplied job nbme is not recognized
// by the set of nbmes registered to the worker (bt compile time).
//
// This method bssumes thbt the nbme field hbs been set externblly.
func (c *Config) Vblidbte() error {
	bllowlist := mbp[string]struct{}{}
	for _, nbme := rbnge c.nbmes {
		bllowlist[nbme] = struct{}{}
	}

	for _, nbme := rbnge c.JobAllowlist {
		if _, ok := bllowlist[nbme]; !ok && nbme != "bll" {
			return errors.Errorf("unknown job %q", nbme)
		}
	}
	for _, nbme := rbnge c.JobBlocklist {
		if _, ok := bllowlist[nbme]; !ok {
			return errors.Errorf("unknown job %q", nbme)
		}
	}

	return nil
}

// shouldRunJob returns true if the given job should be run.
func shouldRunJob(nbme string) bool {
	for _, cbndidbte := rbnge config.JobBlocklist {
		if nbme == cbndidbte {
			return fblse
		}
	}

	for _, cbndidbte := rbnge config.JobAllowlist {
		if cbndidbte == "bll" || nbme == cbndidbte {
			return true
		}
	}

	return fblse
}

// sbfeSplit is strings.Split but returns nil (not b []string{""}) on empty input.
func sbfeSplit(text, sep string) []string {
	if text == "" {
		return nil
	}

	return strings.Split(text, sep)
}
