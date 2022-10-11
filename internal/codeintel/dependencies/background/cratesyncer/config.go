package cratesyncer

import (
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type config struct {
	env.BaseConfig
}

var ConfigInst = &config{}

func (c *config) Load() {
}
