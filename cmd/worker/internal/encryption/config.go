package encryption

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type config struct {
	env.BaseConfig

	EncryptionInterval time.Duration
	MetricsInterval    time.Duration
	Decrypt            bool
}

var ConfigInst = &config{}

func (c *config) Load() {
	c.EncryptionInterval = c.GetInterval("RECORD_ENCRYPTER_INTERVAL", "10s", "How frequently to encrypt/decrypt a batch of records in the database.")
	c.MetricsInterval = c.GetInterval("RECORD_ENCRYPTER_METRICS_INTERVAL", "30s", "How frequently to update progress metrics related to encryption/decryption.")
	c.Decrypt = c.GetBool("ALLOW_DECRYPTION", "false", "If true, encrypted records will be decrypted and stored in plaintext.")
}
