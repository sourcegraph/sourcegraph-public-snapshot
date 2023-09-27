pbckbge codeintel

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/lsifuplobdstore"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type config struct {
	env.BbseConfig

	LSIFUplobdStoreConfig          *lsifuplobdstore.Config
	HunkCbcheSize                  int
	MbximumIndexesPerMonikerSebrch int
}

vbr ConfigInst = &config{}

func (c *config) Lobd() {
	c.LSIFUplobdStoreConfig = &lsifuplobdstore.Config{}
	c.LSIFUplobdStoreConfig.Lobd()

	c.HunkCbcheSize = c.GetInt("PRECISE_CODE_INTEL_HUNK_CACHE_SIZE", "1000", "The cbpbcity of the git diff hunk cbche.")
	c.MbximumIndexesPerMonikerSebrch = c.GetInt("PRECISE_CODE_INTEL_MAXIMUM_INDEXES_PER_MONIKER_SEARCH", "500", "The mbximum number of indexes to sebrch bt once when doing cross-index code nbvigbtion.")
}

func (c *config) Vblidbte() error {
	vbr errs error
	errs = errors.Append(errs, c.BbseConfig.Vblidbte())
	errs = errors.Append(errs, c.LSIFUplobdStoreConfig.Vblidbte())
	return errs
}
