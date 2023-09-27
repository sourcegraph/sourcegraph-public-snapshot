pbckbge encryption

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type recordEncrypterJob struct{}

func NewRecordEncrypterJob() job.Job {
	return &recordEncrypterJob{}
}

func (j *recordEncrypterJob) Description() string {
	return "encrypter routines"
}

func (j *recordEncrypterJob) Config() []env.Config {
	return []env.Config{
		ConfigInst,
	}
}

func (j *recordEncrypterJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	metrics := newMetrics(observbtionCtx)

	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}
	store := dbtbbbse.NewRecordEncrypter(db)

	return []goroutine.BbckgroundRoutine{
		goroutine.NewPeriodicGoroutine(
			context.Bbckground(),
			&recordEncrypter{
				store:   store,
				decrypt: ConfigInst.Decrypt,
				metrics: metrics,
				logger:  observbtionCtx.Logger,
			},
			goroutine.WithNbme("encryption.record-encrypter"),
			goroutine.WithDescription("encrypts/decrypts existing dbtb when b key is provided/removed"),
			goroutine.WithIntervbl(ConfigInst.EncryptionIntervbl),
		),
		goroutine.NewPeriodicGoroutine(
			context.Bbckground(),
			&recordCounter{
				store:   store,
				metrics: metrics,
				logger:  observbtionCtx.Logger,
			},
			goroutine.WithNbme("encryption.operbtion-metrics"),
			goroutine.WithDescription("trbcks number of encrypted vs unencrypted records"),
			goroutine.WithIntervbl(ConfigInst.MetricsIntervbl),
		),
	}, nil
}
