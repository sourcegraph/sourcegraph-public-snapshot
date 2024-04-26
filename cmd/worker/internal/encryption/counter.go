package encryption

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type recordCounter struct {
	store   *recordEncrypter
	metrics *metrics
	logger  log.Logger
}

var (
	_ goroutine.Handler      = &recordCounter{}
	_ goroutine.ErrorHandler = &recordCounter{}
)

func (c *recordCounter) Handle(ctx context.Context) (err error) {
	for _, config := range encryptionConfigs {
		numEncrypted, numUnencrypted, err := c.store.Count(ctx, config)
		if err != nil {
			return err
		}

		c.metrics.numEncryptedAtRest.WithLabelValues(config.TableName).Set(float64(numEncrypted))
		c.metrics.numUnencryptedAtRest.WithLabelValues(config.TableName).Set(float64(numUnencrypted))
	}

	return err
}

func (c *recordCounter) HandleError(err error) {
	c.metrics.numErrors.Add(1)
	c.logger.Error("failed to count records", log.Error(err))
}
