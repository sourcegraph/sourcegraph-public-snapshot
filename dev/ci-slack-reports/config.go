package main

import "github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"

const DefaultChannel = "#william-buildchecker-webhook-test"

type Config struct {
	BigQueryProjectID string
	BigQueryDatasetID string
	SlackToken        string
}

func (c *Config) Load(env *runtime.Env) {
	c.SlackToken = env.Get("SLACK_TOKEN", "", "")
	c.BigQueryProjectID = env.Get("BIGQUERY_PROJECT_ID", "", "")
	c.BigQueryDatasetID = env.Get("BIGQUERY_DATASET_ID", "", "")
}
