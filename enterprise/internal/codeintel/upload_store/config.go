package uploadstore

import (
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	Backend      string
	ManageBucket bool
	S3           S3Config
	GCS          GCSConfig
}

func (c *Config) Load() {
	c.Backend = c.GetOptional(
		"PRECISE_CODE_INTEL_UPLOAD_BACKEND",
		"The target file service for code intelligence uploads. S3 and GCS are supported.",
	)

	c.ManageBucket = c.GetBool(
		"PRECISE_CODE_INTEL_UPLOAD_MANAGE_BUCKET",
		"true",
		"Whether or not the client should manage the target bucket configuration.",
	)

	config, ok := c.loaders()[c.Backend]
	if !ok {
		c.AddError(fmt.Errorf("invalid backend %q for PRECISE_CODE_INTEL_UPLOAD_BACKEND: must be S3 or GCS", c.Backend))
		return
	}

	config.load(&c.BaseConfig)
}

type loader interface {
	load(parent *env.BaseConfig)
}

func (c *Config) loaders() map[string]loader {
	return map[string]loader{
		"S3":  &c.S3,
		"GCS": &c.GCS,
	}
}

type commonConfig struct {
	// Bucket is the target bucket for LSIF uploads.
	Bucket string

	// TTL is the maximum age of an upload before deletion in the configured bucket.
	TTL time.Duration
}

type S3Config struct {
	commonConfig
}

func (c *S3Config) load(parent *env.BaseConfig) {
	c.Bucket = parent.Get("PRECISE_CODE_INTEL_UPLOAD_BUCKET", "", "The name of the bucket to store LSIF uploads in.")
	c.TTL = parent.GetInterval("PRECISE_CODE_INTEL_UPLOAD_TTL", "168h", "The maximum age of an upload before deletion.")

	// Environment variables used by underlying libraries, here for documentation
	_ = env.Get("AWS_ACCESS_KEY_ID", "", "An AWS access key associated with a user with access to S3.")
	_ = env.Get("AWS_SECRET_ACCESS_KEY", "", "An AWS secret key associated with a user with access to S3.")
	_ = env.Get("AWS_SHARED_CREDENTIALS_FILE", " ~/.aws/credentials", "The path to an AWS credentials file.")
	_ = env.Get("AWS_PROFILE", "default", "The name within the shared credentials file to use.")
	_ = env.Get("AWS_ENDPOINT", "", "Specifies the URL of the AWS API. Used to target a MinIO instance instead of S3.")
	_ = env.Get("AWS_REGION", "", "Specifies the AWS Region to send the request to.")
	_ = env.Get("AWS_S3_FORCE_PATH_STYLE", "", "If set, S3 virtual host request path are not used. Set this to target a MinIO instance.")

}

type GCSConfig struct {
	commonConfig

	// ProjectID specifies the project containing the GCS bucket.
	ProjectID string
}

func (c *GCSConfig) load(parent *env.BaseConfig) {
	c.Bucket = parent.Get("PRECISE_CODE_INTEL_UPLOAD_BUCKET", "", "The name of the bucket to store LSIF uploads in.")
	c.TTL = parent.GetInterval("PRECISE_CODE_INTEL_UPLOAD_TTL", "168h", "The maximum age of an upload before deletion.")
	c.ProjectID = parent.Get("PRECISE_CODE_INTEL_UPLOAD_GCP_PROJECT_ID", "", "The project containing the GCS bucket.")

	// Environment variables used by underlying libraries, here for documentation
	_ = env.Get("GOOGLE_APPLICATION_CREDENTIALS", "", "The path to a service account key file with access to GCS.")
	_ = env.Get("GOOGLE_APPLICATION_CREDENTIALS_JSON", "", "The contents of a service account key file with access to GCS.")
}
