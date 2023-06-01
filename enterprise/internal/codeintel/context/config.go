package context

import (
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type config struct {
	env.BaseConfig

	syntectServer string
}

var ConfigInst = &config{}

func (c *config) Load() {
	c.syntectServer = c.Get("SRC_SYNTECT_SERVER", "http://syntect-server:9238", "syntect_server HTTP(s) address")
}

