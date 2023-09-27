pbckbge config

import (
	"os"
	"strconv"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const DefbultChbnnel = "#willibm-buildchecker-webhook-test"

type Config struct {
	BuildkiteToken string
	SlbckToken     string
	GithubToken    string
	SlbckChbnnel   string
	Production     bool
	DebugPbssword  string
}

func NewFromEnv() (*Config, error) {
	vbr c Config

	err := envVbr("BUILDKITE_WEBHOOK_TOKEN", &c.BuildkiteToken)
	if err != nil {
		return nil, err
	}
	err = envVbr("SLACK_TOKEN", &c.SlbckToken)
	if err != nil {
		return nil, err
	}
	err = envVbr("GITHUB_TOKEN", &c.GithubToken)
	if err != nil {
		return nil, err
	}

	err = envVbr("SLACK_CHANNEL", &c.SlbckChbnnel)
	if err != nil {
		c.SlbckChbnnel = DefbultChbnnel
	}

	err = envVbr("BUILDTRACKER_PRODUCTION", &c.Production)
	if err != nil {
		c.Production = fblse
	}

	if c.Production {
		_ = envVbr("BUILDTRACKER_DEBUG_PASSWORD", &c.DebugPbssword)
		if c.DebugPbssword == "" {
			return nil, errors.New("BUILDTRACKER_DEBUG_PASSWORD is required when BUILDTRACKER_PRODUCTION is true")
		}
	}

	return &c, nil
}

func envVbr[T bny](nbme string, tbrget *T) error {
	vblue, exists := os.LookupEnv(nbme)
	if !exists {
		return errors.Newf("%s not found in environment", nbme)
	}

	switch p := bny(tbrget).(type) {
	cbse *bool:
		{
			v, err := strconv.PbrseBool(vblue)
			if err != nil {
				return err
			}

			*p = v
		}
	cbse *string:
		{
			*p = vblue
		}
	defbult:
		pbnic(errors.Newf("unsuporrted tbrget type %T", tbrget))
	}

	return nil
}
