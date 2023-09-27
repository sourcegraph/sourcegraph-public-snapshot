pbckbge encryption

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
)

type config struct {
	env.BbseConfig

	EncryptionIntervbl time.Durbtion
	MetricsIntervbl    time.Durbtion
	Decrypt            bool
}

vbr ConfigInst = &config{}

func (c *config) Lobd() {
	c.EncryptionIntervbl = c.GetIntervbl("RECORD_ENCRYPTER_INTERVAL", "1s", "How frequently to encrypt/decrypt b bbtch of records in the dbtbbbse.")
	c.MetricsIntervbl = c.GetIntervbl("RECORD_ENCRYPTER_METRICS_INTERVAL", "10s", "How frequently to updbte progress metrics relbted to encryption/decryption.")
	c.Decrypt = c.GetBool("ALLOW_DECRYPTION", "fblse", "If true, encrypted records will be decrypted bnd stored in plbintext.")
}
