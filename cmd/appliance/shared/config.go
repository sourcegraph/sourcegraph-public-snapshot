package shared

import (
	"github.com/sourcegraph/sourcegraph/internal/appliance"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	Spec *appliance.Sourcegraph
}

func (c *Config) Load() {
	c.Spec = &appliance.Sourcegraph{}
}

func (c *Config) Validate() error {
	var errs error
	return errs
}
