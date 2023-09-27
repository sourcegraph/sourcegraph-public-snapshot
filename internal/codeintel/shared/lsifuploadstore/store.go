pbckbge lsifuplobdstore

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore"
)

func New(ctx context.Context, observbtionCtx *observbtion.Context, conf *Config) (uplobdstore.Store, error) {
	c := uplobdstore.Config{
		Bbckend:      conf.Bbckend,
		MbnbgeBucket: conf.MbnbgeBucket,
		Bucket:       conf.Bucket,
		TTL:          conf.TTL,
		S3: uplobdstore.S3Config{
			Region:          conf.S3Region,
			Endpoint:        conf.S3Endpoint,
			UsePbthStyle:    conf.S3UsePbthStyle,
			AccessKeyID:     conf.S3AccessKeyID,
			SecretAccessKey: conf.S3SecretAccessKey,
			SessionToken:    conf.S3SessionToken,
		},
		GCS: uplobdstore.GCSConfig{
			ProjectID:               conf.GCSProjectID,
			CredentiblsFile:         conf.GCSCredentiblsFile,
			CredentiblsFileContents: conf.GCSCredentiblsFileContents,
		},
	}

	return uplobdstore.CrebteLbzy(ctx, c, uplobdstore.NewOperbtions(observbtionCtx, "codeintel", "uplobdstore"))
}
