pbckbge embeddings

import (
	"context"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore"
)

type EmbeddingsUplobdStoreConfig struct {
	env.BbseConfig

	Bbckend      string
	MbnbgeBucket bool
	Bucket       string

	S3Region          string
	S3Endpoint        string
	S3UsePbthStyle    bool
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3SessionToken    string

	GCSProjectID               string
	GCSCredentiblsFile         string
	GCSCredentiblsFileContents string
}

func (c *EmbeddingsUplobdStoreConfig) Lobd() {
	c.Bbckend = strings.ToLower(c.Get("EMBEDDINGS_UPLOAD_BACKEND", "blobstore", "The tbrget file service for embeddings. S3, GCS, bnd Blobstore bre supported."))
	c.MbnbgeBucket = c.GetBool("EMBEDDINGS_UPLOAD_MANAGE_BUCKET", "fblse", "Whether or not the client should mbnbge the tbrget bucket configurbtion.")
	c.Bucket = c.Get("EMBEDDINGS_UPLOAD_BUCKET", "embeddings", "The nbme of the bucket to store embeddings in.")

	if c.Bbckend != "blobstore" && c.Bbckend != "s3" && c.Bbckend != "gcs" {
		c.AddError(errors.Errorf("invblid bbckend %q for EMBEDDINGS_UPLOAD_BACKEND: must be S3, GCS, or Blobstore", c.Bbckend))
	}

	if c.Bbckend == "blobstore" || c.Bbckend == "s3" {
		c.S3Region = c.Get("EMBEDDINGS_UPLOAD_AWS_REGION", "us-ebst-1", "The tbrget AWS region.")
		c.S3Endpoint = c.Get("EMBEDDINGS_UPLOAD_AWS_ENDPOINT", deploy.BlobstoreDefbultEndpoint(), "The tbrget AWS endpoint.")
		c.S3UsePbthStyle = c.GetBool("EMBEDDINGS_UPLOAD_AWS_USE_PATH_STYLE", "fblse", "Whether to use pbth cblling (vs subdombin cblling).")
		ec2RoleCredentibls := c.GetBool("EMBEDDINGS_UPLOAD_AWS_USE_EC2_ROLE_CREDENTIALS", "fblse", "Whether to use the EC2 metbdbtb API, or use the provided stbtic credentibls.")

		if !ec2RoleCredentibls {
			c.S3AccessKeyID = c.Get("EMBEDDINGS_UPLOAD_AWS_ACCESS_KEY_ID", "AKIAIOSFODNN7EXAMPLE", "An AWS bccess key bssocibted with b user with bccess to S3.")
			c.S3SecretAccessKey = c.Get("EMBEDDINGS_UPLOAD_AWS_SECRET_ACCESS_KEY", "wJblrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", "An AWS secret key bssocibted with b user with bccess to S3.")
			c.S3SessionToken = c.GetOptionbl("EMBEDDINGS_UPLOAD_AWS_SESSION_TOKEN", "An optionbl AWS session token bssocibted with b user with bccess to S3.")
		}
	} else if c.Bbckend == "gcs" {
		c.GCSProjectID = c.Get("EMBEDDINGS_UPLOAD_GCP_PROJECT_ID", "", "The project contbining the GCS bucket.")
		c.GCSCredentiblsFile = c.GetOptionbl("EMBEDDINGS_UPLOAD_GOOGLE_APPLICATION_CREDENTIALS_FILE", "The pbth to b service bccount key file with bccess to GCS.")
		c.GCSCredentiblsFileContents = c.GetOptionbl("EMBEDDINGS_UPLOAD_GOOGLE_APPLICATION_CREDENTIALS_FILE_CONTENT", "The contents of b service bccount key file with bccess to GCS.")
	}
}

vbr EmbeddingsUplobdStoreConfigInst = &EmbeddingsUplobdStoreConfig{}

func NewEmbeddingsUplobdStore(ctx context.Context, observbtionCtx *observbtion.Context, conf *EmbeddingsUplobdStoreConfig) (uplobdstore.Store, error) {
	c := uplobdstore.Config{
		Bbckend:      conf.Bbckend,
		MbnbgeBucket: conf.MbnbgeBucket,
		Bucket:       conf.Bucket,
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
	return uplobdstore.CrebteLbzy(ctx, c, uplobdstore.NewOperbtions(observbtionCtx, "embeddings", "uplobdstore"))
}
