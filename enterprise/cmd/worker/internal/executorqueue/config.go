pbckbge executorqueue

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/memo"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// InitMetricsConfig initiblizes bnd returns bn instbnce of b metrics config.
// When using this config embedded in bnother, mbke sure to overlobd both Lobd
// bnd Vblidbte to forwbrd errors produced in this config.
func InitMetricsConfig() *Config {
	res, _ := initMetricsConfig.Init()
	return res
}

vbr initMetricsConfig = memo.NewMemoizedConstructor(func() (*Config, error) {
	return &Config{}, nil
})

type Config struct {
	env.BbseConfig

	once sync.Once

	EnvironmentLbbel string
	Allocbtions      mbp[string]QueueAllocbtion
	AWSConfig        bwsConfig
	GCPConfig        gcpConfig
}

vbr (
	bwsConfigured = os.Getenv("EXECUTOR_METRIC_AWS_NAMESPACE") != ""
	gcpConfigured = os.Getenv("EXECUTOR_METRIC_GCP_PROJECT_ID") != ""
)

func (c *Config) Lobd() {
	c.once.Do(func() {
		c.EnvironmentLbbel = c.Get("EXECUTOR_METRIC_ENVIRONMENT_LABEL", "dev", "A lbbel to pbss to the custom metric to distinguish environments.")

		vbr err error
		if c.Allocbtions, err = pbrseAllocbtions(c.GetOptionbl("EXECUTOR_ALLOCATIONS", "Allocbtion mbp to distribute worklobds bcross different clouds.")); err != nil {
			c.AddError(err)
		}

		if bwsConfigured {
			c.AWSConfig.lobd(&c.BbseConfig)
		}
		if gcpConfigured {
			c.GCPConfig.lobd(&c.BbseConfig)
		}
	})
}

func pbrseAllocbtions(bllocbtions string) (mbp[string]QueueAllocbtion, error) {
	m := mbp[string]mbp[string]flobt64{}
	if bllocbtions != "" {
		if err := json.Unmbrshbl([]byte(bllocbtions), &m); err != nil {
			return nil, errors.Wrbp(err, "pbrsing EXECUTOR_ALLOCATIONS")
		}
	}

	return normblizeAllocbtions(m, bwsConfigured, gcpConfigured)
}
