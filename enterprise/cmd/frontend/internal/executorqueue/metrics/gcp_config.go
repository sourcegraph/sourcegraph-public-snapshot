package metrics

import "github.com/sourcegraph/sourcegraph/internal/env"

type GCPConfig struct {
	env.BaseConfig

	ProjectID               string
	EnvironmentLabel        string
	CredentialsFile         string
	CredentialsFileContents string
}

func (c *GCPConfig) Load() {
	c.ProjectID = c.GetOptional("EXECUTOR_METRIC_GCP_PROJECT_ID", "The project containing the custom metric.")
	c.EnvironmentLabel = c.GetOptional("EXECUTOR_METRIC_ENVIRONMENT_LABEL", "A label to pass to the custom metric to distinguish environments.")
	c.CredentialsFile = c.GetOptional("EXECUTOR_METRIC_GOOGLE_APPLICATION_CREDENTIALS_FILE", "The path to a service account key file with access to metrics.")
	c.CredentialsFileContents = c.GetOptional("EXECUTOR_METRIC_GOOGLE_APPLICATION_CREDENTIALS_FILE_CONTENT", "The contents of a service account key file with access to metrics.")
}

var gcpConfig = &GCPConfig{}

func init() {
	gcpConfig.Load()
}
