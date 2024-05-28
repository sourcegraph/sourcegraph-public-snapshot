package service

import (
	"github.com/sourcegraph/sourcegraph/cmd/msp-example/internal/httpapi"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

type Config struct {
	StatelessMode bool
	HTTPAPI       httpapi.Config
}

func (c *Config) Load(env *runtime.Env) {
	c.StatelessMode = env.GetBool("STATELESS_MODE", "false", "if true, disable dependencies")
	c.HTTPAPI.Variable = env.Get("VARIABLE", "13", "variable value")
}
