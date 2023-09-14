package encryption

import (
	"context"
	"fmt"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type recordEncrypter struct {
	store   *database.RecordEncrypter
	decrypt bool
	metrics *metrics
	logger  log.Logger
}

var (
	_ goroutine.Handler      = &recordEncrypter{}
	_ goroutine.ErrorHandler = &recordEncrypter{}
)

func (e *recordEncrypter) Handle(ctx context.Context) (err error) {
	for _, config := range database.EncryptionConfigs {
		if handleErr := e.handleBatch(ctx, config); handleErr != nil {
			err = errors.CombineErrors(err, handleErr)
		}
	}

	return err
}

func (e *recordEncrypter) handleBatch(ctx context.Context, config database.EncryptionConfig) error {
	if e.decrypt {
		return e.handleDecryptBatch(ctx, config)
	}

	return e.handleEncryptBatch(ctx, config)
}

func (e *recordEncrypter) handleEncryptBatch(ctx context.Context, config database.EncryptionConfig) error {
	count, err := e.store.EncryptBatch(ctx, config)
	if err != nil || count == 0 {
		return err
	}

	e.metrics.numRecordsEncrypted.WithLabelValues(config.TableName).Add(float64(count))
	e.logger.Debug("encrypted records", log.String("tableName", config.TableName), log.Int("count", count))
	return nil
}

func (e *recordEncrypter) handleDecryptBatch(ctx context.Context, config database.EncryptionConfig) error {
	count, err := e.store.DecryptBatch(ctx, config)
	if err != nil || count == 0 {
		return err
	}

	e.metrics.numRecordsDecrypted.WithLabelValues(config.TableName).Add(float64(count))
	e.logger.Debug("decrypted records", log.String("tableName", config.TableName), log.Int("count", count))
	return nil
}

func (e *recordEncrypter) HandleError(err error) {
	verb := "encrypt"
	if e.decrypt {
		verb = "decrypt"
	}

	e.metrics.numErrors.Add(1)
	e.logger.Error(fmt.Sprintf("failed to %s batch of records", verb), log.Error(err))
}
