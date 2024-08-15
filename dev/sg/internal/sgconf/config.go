package sgconf

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/env"
	"gopkg.in/yaml.v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func parseConfigFile(name string, isBaseConfig bool) (*Config, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open file %q", name)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.Wrap(err, "reading configuration file")
	}

	return parseConfig(data, isBaseConfig)
}

func parseConfig(data []byte, isBaseConfig bool) (*Config, error) {
	var conf Config
	conf.isBaseConfig = isBaseConfig
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, err
	}

	rr, err := root.RepositoryRoot()
	if err != nil {
		return nil, err
	}

	for name, cmd := range conf.BazelCommands {
		cmd.Config.Name = name
		cmd.Config.RepositoryRoot = rr
	}

	for name, cmd := range conf.DockerCommands {
		cmd.Config.Name = name
		cmd.Config.RepositoryRoot = rr
	}

	for name, cmd := range conf.Commands {
		cmd.Config.Name = name
		cmd.Config.RepositoryRoot = rr
		normalizeCmd(cmd)
	}

	for name, cmd := range conf.Commandsets {
		cmd.Name = name
		conf.Commandsets[name] = cmd
	}

	for name, cmd := range conf.Tests {
		cmd.Config.Name = name
		cmd.Config.RepositoryRoot = rr
		normalizeCmd(cmd)
		conf.Tests[name] = cmd
	}

	return &conf, nil
}

func normalizeCmd(cmd *run.Command) {
	// Trim trailing whitespace so extra args apply to last command (instead of being interpreted as
	// a new shell command on a separate line).
	cmd.Cmd = strings.TrimSpace(cmd.Cmd)
}

type Commandset struct {
	Name           string            `yaml:"-"`
	Commands       []string          `yaml:"commands"`
	BazelCommands  []string          `yaml:"bazelCommands"`
	DockerCommands []string          `yaml:"dockerCommands"`
	Checks         []string          `yaml:"checks"`
	Env            map[string]string `yaml:"env"`
	NewEnv         map[string]env.EnvVar
	Deprecated     string `yaml:"deprecated"`
}

// UnmarshalYAML implements the Unmarshaler interface.
func (c *Commandset) UnmarshalYAML(unmarshal func(any) error) error {
	// To be backwards compatible we first try to unmarshal as a simple list.
	var list []string
	if err := unmarshal(&list); err == nil {
		c.Commands = list
		return nil
	}

	// If it's not a list we try to unmarshal it as a Commandset. In order to
	// not recurse infinitely (calling UnmarshalYAML over and over) we create a
	// temporary type alias.
	type rawCommandset Commandset
	if err := unmarshal((*rawCommandset)(c)); err != nil {
		return err
	}

	return nil
}

func (c *Commandset) IsDeprecated() bool {
	return c.Deprecated != ""
}

func (c *Commandset) Merge(other *Commandset) *Commandset {
	merged := c

	if other.Name != merged.Name && other.Name != "" {
		merged.Name = other.Name
	}

	if len(other.Commands) != 0 {
		merged.Commands = other.Commands
	}

	if len(other.Checks) != 0 {
		merged.Checks = other.Checks
	}

	if len(other.BazelCommands) != 0 {
		merged.BazelCommands = other.BazelCommands
	}

	if len(other.DockerCommands) != 0 {
		merged.DockerCommands = other.DockerCommands
	}

	for k, v := range other.Env {
		merged.Env[k] = v
	}

	for k, v := range other.NewEnv {
		mv, exists := merged.NewEnv[k]
		if exists {
			merged.NewEnv[k] = env.CompareByPriority(mv, v)
		}
		merged.NewEnv[k] = v
	}

	return merged
}

// If you add an entry here, remember to add it to the merge function.
type Config struct {
	Env               map[string]string `yaml:"env"`
	NewEnv            map[string]env.EnvVar
	Commands          map[string]*run.Command       `yaml:"commands"`
	BazelCommands     map[string]*run.BazelCommand  `yaml:"bazelCommands"`
	DockerCommands    map[string]*run.DockerCommand `yaml:"dockerCommands"`
	Commandsets       map[string]*Commandset        `yaml:"commandsets"`
	DefaultCommandset string                        `yaml:"defaultCommandset"`
	Tests             map[string]*run.Command       `yaml:"tests"`

	isBaseConfig bool
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// We create an alias to avoid recursion
	type alias Config
	var temp alias
	temp.isBaseConfig = c.isBaseConfig

	if err := unmarshal(&temp); err != nil {
		return err
	}

	*c = Config(temp)
	c.populateEnvField()
	return nil
}

func (c *Config) populateEnvField() {
	c.PopulateNewEnv(c.isBaseConfig)
}

func (c *Config) PopulateNewEnv(isBaseConfig bool) {
	gep := env.GlobalEnvPriority
	cep := env.BaseCommandEnvPriority
	csep := env.BaseCommandsetEnvPriority
	if !isBaseConfig {
		gep = env.OverrideGlobalEnvPriority
		cep = env.OverrideCommandEnvPriority
		csep = env.OverrideCommandsetEnvPriority
	}

	c.NewEnv = env.ConvertEnvMap(c.Env, gep)

	for name, cmd := range c.Commands {
		cmd.Config.NewEnv = env.ConvertEnvMap(cmd.Config.Env, cep)
		c.Commands[name] = cmd
	}

	for _, dcmd := range c.DockerCommands {
		dcmd.Config.NewEnv = env.ConvertEnvMap(dcmd.Config.Env, cep)
	}

	for _, bcmd := range c.BazelCommands {
		bcmd.Config.NewEnv = env.ConvertEnvMap(bcmd.Config.Env, cep)
	}

	for _, cmds := range c.Commandsets {
		cmds.NewEnv = env.ConvertEnvMap(cmds.Env, csep)
	}

	for _, ts := range c.Tests {
		ts.Config.NewEnv = env.ConvertEnvMap(ts.Config.Env, cep)
	}
}

// Merge merges the top-level entries of two Config objects, using the
// values from `other` if they are set as overrides and returns a new config
func (c *Config) Merge(other *Config) *Config {
	merged := *c
	for k, v := range other.Env {
		merged.Env[k] = v
	}

	for k, v := range other.NewEnv {
		mv, exists := merged.NewEnv[k]
		if exists {
			merged.NewEnv[k] = env.CompareByPriority(mv, v)
		}
		merged.NewEnv[k] = v
	}

	for name, override := range other.Commands {
		if original, ok := merged.Commands[name]; ok {
			merged.Commands[name] = pointers.Ptr(original.Merge(*override))
		} else {
			fmt.Println("not ok")
			merged.Commands[name] = override
		}
	}

	for name, override := range other.BazelCommands {
		if original, ok := merged.BazelCommands[name]; ok {
			merged.BazelCommands[name] = pointers.Ptr(original.Merge(*override))
		} else {
			merged.BazelCommands[name] = override
		}
	}

	for name, override := range other.DockerCommands {
		if original, ok := merged.DockerCommands[name]; ok {
			merged.DockerCommands[name] = pointers.Ptr(original.Merge(*override))
		} else {
			merged.DockerCommands[name] = override
		}
	}

	for name, override := range other.Commandsets {
		if original, ok := merged.Commandsets[name]; ok {
			merged.Commandsets[name] = original.Merge(override)
		} else {
			merged.Commandsets[name] = override
		}
	}

	if other.DefaultCommandset != "" {
		merged.DefaultCommandset = other.DefaultCommandset
	}

	for name, override := range other.Tests {
		if original, ok := merged.Tests[name]; ok {
			merged.Tests[name] = pointers.Ptr(original.Merge(*override))
		} else {
			merged.Tests[name] = override
		}
	}

	return &merged
}

func (c *Config) GetEnv(key string) string {
	// First look into process env, emulating the logic in makeEnv used
	// in internal/run/run.go
	val, ok := os.LookupEnv(key)
	if ok {
		return val
	}
	// Otherwise check in globalConf.Env and *expand* the key, because a value might refer to another env var.
	return os.Expand(c.NewEnv[key].Value, func(lookup string) string {
		if lookup == key {
			return os.Getenv(lookup)
		}

		if e, ok := c.NewEnv[lookup]; ok {
			return e.Value
		}
		return os.Getenv(lookup)
	})
}
