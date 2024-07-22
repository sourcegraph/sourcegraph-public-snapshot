package search

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/object"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ObjectStorageConfigInst = &ObjectStorageConfig{}

type ObjectStorageConfig struct {
	env.BaseConfig

	Backend      string
	ManageBucket bool
	Bucket       string

	S3Region          string
	S3Endpoint        string
	S3UsePathStyle    bool
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3SessionToken    string

	GCSProjectID               string
	GCSCredentialsFile         string
	GCSCredentialsFileContents string
}

func (c *ObjectStorageConfig) Load() {
	c.Backend = strings.ToLower(c.Get("SEARCH_JOBS_UPLOAD_BACKEND", "blobstore", "The target file service for search jobs. S3, GCS, and Blobstore are supported."))
	c.ManageBucket = c.GetBool("SEARCH_JOBS_UPLOAD_MANAGE_BUCKET", "false", "Whether or not the client should manage the target bucket configuration.")
	c.Bucket = c.Get("SEARCH_JOBS_UPLOAD_BUCKET", "search-jobs", "The name of the bucket to store search job results in.")

	if c.Backend != "blobstore" && c.Backend != "s3" && c.Backend != "gcs" {
		c.AddError(errors.Errorf("invalid backend %q for SEARCH_JOBS_UPLOAD_BACKEND: must be S3, GCS, or Blobstore", c.Backend))
	}

	if c.Backend == "blobstore" || c.Backend == "s3" {
		c.S3Region = c.Get("SEARCH_JOBS_UPLOAD_AWS_REGION", "us-east-1", "The target AWS region.")
		c.S3Endpoint = c.Get("SEARCH_JOBS_UPLOAD_AWS_ENDPOINT", deploy.BlobstoreDefaultEndpoint(), "The target AWS endpoint.")
		c.S3UsePathStyle = c.GetBool("SEARCH_JOBS_UPLOAD_AWS_USE_PATH_STYLE", "false", "Whether to use path calling (vs subdomain calling).")
		ec2RoleCredentials := c.GetBool("SEARCH_JOBS_UPLOAD_AWS_USE_EC2_ROLE_CREDENTIALS", "false", "Whether to use the EC2 metadata API, or use the provided static credentials.")

		if !ec2RoleCredentials {
			c.S3AccessKeyID = c.Get("SEARCH_JOBS_UPLOAD_AWS_ACCESS_KEY_ID", "AKIAIOSFODNN7EXAMPLE", "An AWS access key associated with a user with access to S3.")
			c.S3SecretAccessKey = c.Get("SEARCH_JOBS_UPLOAD_AWS_SECRET_ACCESS_KEY", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", "An AWS secret key associated with a user with access to S3.")
			c.S3SessionToken = c.GetOptional("SEARCH_JOBS_UPLOAD_AWS_SESSION_TOKEN", "An optional AWS session token associated with a user with access to S3.")
		}
	} else if c.Backend == "gcs" {
		c.GCSProjectID = c.Get("SEARCH_JOBS_UPLOAD_GCP_PROJECT_ID", "", "The project containing the GCS bucket.")
		c.GCSCredentialsFile = c.GetOptional("SEARCH_JOBS_UPLOAD_GOOGLE_APPLICATION_CREDENTIALS_FILE", "The path to a service account key file with access to GCS.")
		c.GCSCredentialsFileContents = c.GetOptional("SEARCH_JOBS_UPLOAD_GOOGLE_APPLICATION_CREDENTIALS_FILE_CONTENT", "The contents of a service account key file with access to GCS.")
	}
}

var ConfigInst = &ObjectStorageConfig{}

func NewObjectStorage(ctx context.Context, observationCtx *observation.Context, conf *ObjectStorageConfig) (object.Storage, error) {
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
	return object.CreateLazyStorage(ctx, c, object.NewOperations(observationCtx, "search_jobs", "uploadstore"))
}
