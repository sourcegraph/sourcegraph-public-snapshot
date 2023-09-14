package config

import (
	"os"
	"strconv"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const DefaultChannel = "#william-buildchecker-webhook-test"

type Config struct {
	BuildkiteToken string
	SlackToken     string
	GithubToken    string
	SlackChannel   string
	Production     bool
	DebugPassword  string
}

func NewFromEnv() (*Config, error) {
	var c Config

	err := envVar("BUILDKITE_WEBHOOK_TOKEN", &c.BuildkiteToken)
	if err != nil {
		return nil, err
	}
	err = envVar("SLACK_TOKEN", &c.SlackToken)
	if err != nil {
		return nil, err
	}
	err = envVar("GITHUB_TOKEN", &c.GithubToken)
	if err != nil {
		return nil, err
	}

	err = envVar("SLACK_CHANNEL", &c.SlackChannel)
	if err != nil {
		c.SlackChannel = DefaultChannel
	}

	err = envVar("BUILDTRACKER_PRODUCTION", &c.Production)
	if err != nil {
		c.Production = false
	}

	if c.Production {
		_ = envVar("BUILDTRACKER_DEBUG_PASSWORD", &c.DebugPassword)
		if c.DebugPassword == "" {
			return nil, errors.New("BUILDTRACKER_DEBUG_PASSWORD is required when BUILDTRACKER_PRODUCTION is true")
		}
	}

	return &c, nil
}

func envVar[T any](name string, target *T) error {
	value, exists := os.LookupEnv(name)
	if !exists {
		return errors.Newf("%s not found in environment", name)
	}

	switch p := any(target).(type) {
	case *bool:
		{
			v, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}

			*p = v
		}
	case *string:
		{
			*p = value
		}
	default:
		panic(errors.Newf("unsuporrted target type %T", target))
	}

	return nil
}
