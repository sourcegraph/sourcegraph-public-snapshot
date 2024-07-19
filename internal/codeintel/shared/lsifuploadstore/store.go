package lsifuploadstore

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/object"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func New(ctx context.Context, observationCtx *observation.Context, conf *Config) (object.Storage, error) {
	c := object.StorageConfig{
		Backend:      conf.Backend,
		ManageBucket: conf.ManageBucket,
		Bucket:       conf.Bucket,
		S3: object.S3Config{
			Region:          conf.S3Region,
			Endpoint:        conf.S3Endpoint,
			UsePathStyle:    conf.S3UsePathStyle,
			AccessKeyID:     conf.S3AccessKeyID,
			SecretAccessKey: conf.S3SecretAccessKey,
			SessionToken:    conf.S3SessionToken,
		},
		GCS: object.GCSConfig{
			ProjectID:               conf.GCSProjectID,
			CredentialsFile:         conf.GCSCredentialsFile,
			CredentialsFileContents: conf.GCSCredentialsFileContents,
		},
	}

	return object.CreateLazyStorage(ctx, c, object.NewOperations(observationCtx, "codeintel", "uploadstore"))
}
