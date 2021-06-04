package shared

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

// Config is the configuration that controls what tasks will be initialized
// and monitored. By default, all tasks are enabled. Individual tasks can be
// explicit allowed or blocked from running on a particular instance.
type Config struct {
	env.BaseConfig
	names []string

	TaskAllowlist []string
	TaskBlocklist []string
}

var config = &Config{}

// Load reads from the environment and stores the transformed data on the config
// object for later retrieval.
func (c *Config) Load() {
	c.TaskAllowlist = safeSplit(c.Get(
		"WORKER_TASK_ALLOWLIST",
		"all",
		`A comma-seprated list of names of tasks that should be enabled. The value "all" (the default) enables all tasks.`,
	), ",")

	c.TaskBlocklist = safeSplit(c.Get(
		"WORKER_TASK_BLOCKLIST",
		"",
		"A comma-seprated list of names of tasks that should not be enabled. Values in this list take precedence over the allowlist.",
	), ",")
}

// Validate returns an error indicating if there was an invalid environment read
// during Load. The environment is invalid when a supplied task name is not recognized
// by the set of names registered to the worker (at compile time).
//
// This method assumes that the name field has been set externally.
func (c *Config) Validate() error {
	allowlist := map[string]struct{}{}
	for _, name := range c.names {
		allowlist[name] = struct{}{}
	}

	for _, name := range c.TaskAllowlist {
		if _, ok := allowlist[name]; !ok && name != "all" {
			return fmt.Errorf("unknown †ask %q", name)
		}
	}
	for _, name := range c.TaskBlocklist {
		if _, ok := allowlist[name]; !ok {
			return fmt.Errorf("unknown †ask %q", name)
		}
	}

	return nil
}

// shouldRunTask returns true if the given task should be run.
func shouldRunTask(name string) bool {
	for _, candidate := range config.TaskBlocklist {
		if name == candidate {
			return false
		}
	}

	for _, candidate := range config.TaskAllowlist {
		if candidate == "all" || name == candidate {
			return true
		}
	}

	return false
}

// safeSplit is strings.Split but returns nil (not a []string{""}) on empty input.
func safeSplit(text, sep string) []string {
	if text == "" {
		return nil
	}

	return strings.Split(text, sep)
}
