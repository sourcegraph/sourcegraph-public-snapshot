package shared

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig
	names []string

	SetupHookAllowlist []string
	SetupHookBlocklist []string
}

var config = &Config{}

func (c *Config) Load() {
	c.SetupHookAllowlist = safeSplit(c.Get(
		"WORKER_SETUP_HOOK_ALLOWLIST",
		"all",
		`A comma-seprated list of names of hooks that should be enabled. The value "all" (the default) enables everything.`,
	), ",")

	c.SetupHookBlocklist = safeSplit(c.Get(
		"WORKER_SETUP_HOOK_BLOCKLIST",
		"",
		"A comma-seprated list of names of hooks that should not be enabled. Values in this list take precedence over the allowlist.",
	), ",")
}

func (c *Config) Validate() error {
	allowlist := map[string]struct{}{}
	for _, name := range c.names {
		allowlist[name] = struct{}{}
	}

	for _, name := range c.SetupHookAllowlist {
		if _, ok := allowlist[name]; !ok && name != "all" {
			return fmt.Errorf("unknown setup hook %q", name)
		}
	}
	for _, name := range c.SetupHookBlocklist {
		if _, ok := allowlist[name]; !ok {
			return fmt.Errorf("unknown setup hook %q", name)
		}
	}

	return nil
}

func safeSplit(text, sep string) []string {
	if text == "" {
		return nil
	}

	return strings.Split(text, sep)
}
