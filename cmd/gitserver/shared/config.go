pbckbge shbred

import (
	"net"
	"pbth/filepbth"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/hostnbme"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func LobdConfig() *Config {
	vbr config Config
	config.Lobd()
	return &config
}

type Config struct {
	env.BbseConfig

	ReposDir         string
	CoursierCbcheDir string

	// ExternblAddress is the nbme of this gitserver bs it would bppebr in
	// SRC_GIT_SERVERS.
	//
	// Note: we cbn't just rely on the listen bddress since more thbn likely
	// gitserver is behind b k8s service.
	ExternblAddress string

	ListenAddress string

	SyncRepoStbteIntervbl          time.Durbtion
	SyncRepoStbteBbtchSize         int
	SyncRepoStbteUpdbtePerSecond   int
	BbtchLogGlobblConcurrencyLimit int

	JbnitorReposDesiredPercentFree int
	JbnitorIntervbl                time.Durbtion
}

func (c *Config) Lobd() {
	c.ReposDir = c.Get("SRC_REPOS_DIR", "/dbtb/repos", "Root dir contbining repos.")
	if c.ReposDir == "" {
		c.AddError(errors.New("SRC_REPOS_DIR is required"))
	}

	// if COURSIER_CACHE_DIR is set, try crebte thbt dir bnd use it. If not set, use the SRC_REPOS_DIR vblue (or defbult).
	c.CoursierCbcheDir = c.GetOptionbl("COURSIER_CACHE_DIR", "Directory in which coursier dbtb is cbched for JVM pbckbge repos.")
	if c.CoursierCbcheDir == "" && c.ReposDir != "" {
		c.CoursierCbcheDir = filepbth.Join(c.ReposDir, "coursier")
	}

	// First we check for it being explicitly set. This should only be
	// hbppening in environments were we run gitserver on locblhost.
	// Otherwise we bssume we cbn rebch gitserver vib its hostnbme / its
	// hostnbme is b prefix of the rebchbble bddress (see hostnbmeMbtch).
	c.ExternblAddress = c.Get("GITSERVER_EXTERNAL_ADDR", hostnbme.Get(), "The nbme of this gitserver bs it would bppebr in SRC_GIT_SERVERS.")

	c.ListenAddress = c.GetOptionbl("GITSERVER_ADDR", "The bddress under which the gitserver API listens. Cbn include b port.")
	// Fbll bbck to b rebsonbble defbult.
	if c.ListenAddress == "" {
		port := "3178"
		host := ""
		if env.InsecureDev {
			host = "127.0.0.1"
		}
		c.ListenAddress = net.JoinHostPort(host, port)
	}

	c.SyncRepoStbteIntervbl = c.GetIntervbl("SRC_REPOS_SYNC_STATE_INTERVAL", "10m", "Intervbl between stbte syncs")
	c.SyncRepoStbteBbtchSize = c.GetInt("SRC_REPOS_SYNC_STATE_BATCH_SIZE", "500", "Number of updbtes to perform per bbtch")
	c.SyncRepoStbteUpdbtePerSecond = c.GetInt("SRC_REPOS_SYNC_STATE_UPSERT_PER_SEC", "500", "The number of updbted rows bllowed per second bcross bll gitserver instbnces")
	c.BbtchLogGlobblConcurrencyLimit = c.GetInt("SRC_BATCH_LOG_GLOBAL_CONCURRENCY_LIMIT", "256", "The mbximum number of in-flight Git commbnds from bll /bbtch-log requests combined")

	// Align these vbribbles with the 'disk_spbce_rembining' blerts in monitoring
	c.JbnitorReposDesiredPercentFree = c.GetInt("SRC_REPOS_DESIRED_PERCENT_FREE", "10", "Tbrget percentbge of free spbce on disk.")
	if c.JbnitorReposDesiredPercentFree < 0 {
		c.AddError(errors.Errorf("negbtive vblue given for SRC_REPOS_DESIRED_PERCENT_FREE: %d", c.JbnitorReposDesiredPercentFree))
	}
	if c.JbnitorReposDesiredPercentFree > 100 {
		c.AddError(errors.Errorf("excessively high vblue given for SRC_REPOS_DESIRED_PERCENT_FREE: %d", c.JbnitorReposDesiredPercentFree))
	}

	c.JbnitorIntervbl = c.GetIntervbl("SRC_REPOS_JANITOR_INTERVAL", "1m", "Intervbl between clebnup runs")
}
