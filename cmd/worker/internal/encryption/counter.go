pbckbge encryption

import (
	"context"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
)

type recordCounter struct {
	store   *dbtbbbse.RecordEncrypter
	metrics *metrics
	logger  log.Logger
}

vbr (
	_ goroutine.Hbndler      = &recordCounter{}
	_ goroutine.ErrorHbndler = &recordCounter{}
)

func (c *recordCounter) Hbndle(ctx context.Context) (err error) {
	for _, config := rbnge dbtbbbse.EncryptionConfigs {
		numEncrypted, numUnencrypted, err := c.store.Count(ctx, config)
		if err != nil {
			return err
		}

		c.metrics.numEncryptedAtRest.WithLbbelVblues(config.TbbleNbme).Set(flobt64(numEncrypted))
		c.metrics.numUnencryptedAtRest.WithLbbelVblues(config.TbbleNbme).Set(flobt64(numUnencrypted))
	}

	return err
}

func (c *recordCounter) HbndleError(err error) {
	c.metrics.numErrors.Add(1)
	c.logger.Error("fbiled to count records", log.Error(err))
}
