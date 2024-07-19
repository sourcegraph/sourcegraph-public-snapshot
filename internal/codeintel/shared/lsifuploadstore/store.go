package lsifuploadstore

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/kv"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func New(ctx context.Context, observationCtx *observation.Context, conf *Config) (kv.Store, error) {
	c := kv.Config{
		Backend:      conf.Backend,
		ManageBucket: conf.ManageBucket,
		Bucket:       conf.Bucket,
		TTL:          conf.TTL,
		S3: kv.S3Config{
			Region:          conf.S3Region,
			Endpoint:        conf.S3Endpoint,
			UsePathStyle:    conf.S3UsePathStyle,
			AccessKeyID:     conf.S3AccessKeyID,
			SecretAccessKey: conf.S3SecretAccessKey,
			SessionToken:    conf.S3SessionToken,
		},
		GCS: kv.GCSConfig{
			ProjectID:               conf.GCSProjectID,
			CredentialsFile:         conf.GCSCredentialsFile,
			CredentialsFileContents: conf.GCSCredentialsFileContents,
		},
	}

	return kv.CreateLazy(ctx, c, kv.NewOperations(observationCtx, "codeintel", "uploadstore"))
}
