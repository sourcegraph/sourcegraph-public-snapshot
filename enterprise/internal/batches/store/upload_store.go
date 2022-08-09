package store

import (
	"context"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Config struct {
	env.BaseConfig

	Backend      string
	ManageBucket bool
	Bucket       string
	TTL          time.Duration

	S3Region          string
	S3Endpoint        string
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3SessionToken    string

	GCSProjectID               string
	GCSCredentialsFile         string
	GCSCredentialsFileContents string
}

func (c *Config) Load() {
	c.Backend = strings.ToLower(c.Get("BATCH_CHANGES_UPLOAD_BACKEND", "MinIO", "The target file service for batch changes uploads. S3, GCS, and MinIO are supported."))
	c.ManageBucket = c.GetBool("BATCH_CHANGES_UPLOAD_MANAGE_BUCKET", "false", "Whether or not the client should manage the target bucket configuration.")
	c.Bucket = c.Get("BATCH_CHANGES_UPLOAD_BUCKET", "batches-uploads", "The name of the bucket to store LSIF uploads in.")
	// TODO: TTL?
	c.TTL = c.GetInterval("BATCH_CHANGES_UPLOAD_TTL", "8640h", "The maximum age of an upload before deletion.")

	if c.Backend != "minio" && c.Backend != "s3" && c.Backend != "gcs" {
		c.AddError(errors.Errorf("invalid backend %q for BATCH_CHANGES_UPLOAD_BACKEND: must be S3, GCS, or MinIO", c.Backend))
	}

	if c.Backend == "minio" || c.Backend == "s3" {
		c.S3Region = c.Get("BATCH_CHANGES_UPLOAD_AWS_REGION", "us-east-1", "The target AWS region.")
		c.S3Endpoint = c.Get("BATCH_CHANGES_UPLOAD_AWS_ENDPOINT", "http://minio:9000", "The target AWS endpoint.")
		ec2RoleCredentials := c.GetBool("BATCH_CHANGES_UPLOAD_AWS_USE_EC2_ROLE_CREDENTIALS", "false", "Whether to use the EC2 metadata API, or use the provided static credentials.")

		if !ec2RoleCredentials {
			c.S3AccessKeyID = c.Get("BATCH_CHANGES_UPLOAD_AWS_ACCESS_KEY_ID", "AKIAIOSFODNN7EXAMPLE", "An AWS access key associated with a user with access to S3.")
			c.S3SecretAccessKey = c.Get("BATCH_CHANGES_UPLOAD_AWS_SECRET_ACCESS_KEY", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", "An AWS secret key associated with a user with access to S3.")
			c.S3SessionToken = c.GetOptional("BATCH_CHANGES_UPLOAD_AWS_SESSION_TOKEN", "An optional AWS session token associated with a user with access to S3.")
		}
	} else if c.Backend == "gcs" {
		c.GCSProjectID = c.Get("BATCH_CHANGES_UPLOAD_GCP_PROJECT_ID", "", "The project containing the GCS bucket.")
		c.GCSCredentialsFile = c.GetOptional("BATCH_CHANGES_UPLOAD_GOOGLE_APPLICATION_CREDENTIALS_FILE", "The path to a service account key file with access to GCS.")
		c.GCSCredentialsFileContents = c.GetOptional("BATCH_CHANGES_UPLOAD_GOOGLE_APPLICATION_CREDENTIALS_FILE_CONTENT", "The contents of a service account key file with access to GCS.")
	}
}

func NewUploadStore(ctx context.Context, conf *Config, observationContext *observation.Context) (uploadstore.Store, error) {
	c := uploadstore.Config{
		Backend:      conf.Backend,
		ManageBucket: conf.ManageBucket,
		Bucket:       conf.Bucket,
		TTL:          conf.TTL,
		S3: uploadstore.S3Config{
			Region:          conf.S3Region,
			Endpoint:        conf.S3Endpoint,
			AccessKeyID:     conf.S3AccessKeyID,
			SecretAccessKey: conf.S3SecretAccessKey,
			SessionToken:    conf.S3SessionToken,
		},
		GCS: uploadstore.GCSConfig{
			ProjectID:               conf.GCSProjectID,
			CredentialsFile:         conf.GCSCredentialsFile,
			CredentialsFileContents: conf.GCSCredentialsFileContents,
		},
	}

	return uploadstore.CreateLazy(ctx, c, uploadstore.NewOperations(observationContext, "batchchanges", "uploadstore"))
}
