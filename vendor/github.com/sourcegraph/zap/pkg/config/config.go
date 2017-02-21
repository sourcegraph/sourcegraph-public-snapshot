package config

import (
	"fmt"
	"os"
	"strings"

	gitconfigfmt "github.com/src-d/go-git/plumbing/format/config"
)

// GlobalConfigPath (if set) enables the use of a global Zap config
// file at the given path. Global config is disabled if this value is
// empty.
var GlobalConfigPath string

// ReadFile attempts to read the config file and decode it.
func ReadFile(path string) (gitconfigfmt.Config, error) {
	// Parse the existing config file.
	var cfg gitconfigfmt.Config
	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()
	if err := gitconfigfmt.NewDecoder(f).Decode(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// WriteFile attempts to encode and write the config file, overwriting the
// existing one.
func WriteFile(cfg *gitconfigfmt.Config, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gitconfigfmt.NewEncoder(f).Encode(cfg)
}

// ReadGlobalFile reads the global config file, returning a zero value Config
// and nil error if it does not exist.
func ReadGlobalFile() (gitconfigfmt.Config, error) {
	var cfg gitconfigfmt.Config
	if GlobalConfigPath == "" {
		return cfg, nil
	}
	cfg, err := ReadFile(GlobalConfigPath)
	if err != nil && !os.IsNotExist(err) {
		return cfg, err
	}
	return cfg, nil
}

// WriteGlobalFile attempts to encode and write the global config file,
// overwriting the existing one.
func WriteGlobalFile(cfg *gitconfigfmt.Config) error {
	if GlobalConfigPath == "" {
		panic("global config is disabled (no GlobalConfigPath set)")
	}
	return WriteFile(cfg, GlobalConfigPath)
}

// KeyPair represents a single configuration key-value pair, where the key is
// either "section.option" or "section.subsection.option".
type KeyPair struct {
	Key, Value string
}

// KeyPairs returns the keypairs for the given configuration.
func KeyPairs(cfg *gitconfigfmt.Config) []KeyPair {
	var pairs []KeyPair
	for _, section := range cfg.Sections {
		for _, option := range section.Options {
			pairs = append(pairs, KeyPair{
				Key:   FormatKey(section.Name, "", option.Key),
				Value: option.Value,
			})
		}
		for _, sub := range section.Subsections {
			for _, option := range sub.Options {
				pairs = append(pairs, KeyPair{
					Key:   FormatKey(section.Name, sub.Name, option.Key),
					Value: option.Value,
				})
			}
		}
	}
	return pairs
}

// FormatKey formats the specified inputs as a key.
func FormatKey(section, subsection, option string) string {
	if subsection == "" {
		return fmt.Sprintf("%s.%s", section, option)
	}
	return fmt.Sprintf("%s.%s.%s", section, subsection, option)
}

// ParseKey parses the specified key into its three components.
func ParseKey(k string) (section, subsection, option string) {
	split := strings.Split(k, ".")
	switch len(split) {
	case 2:
		return split[0], "", split[1]
	case 3:
		return split[0], split[1], split[2]
	default:
		return
	}
}

// EnsureWorkspaceInGlobalConfig ensures the specified directory is in the
// global Zap config file under the "workspaces" section.
func EnsureWorkspaceInGlobalConfig(dir string) error {
	if GlobalConfigPath == "" {
		return nil
	}

	// Read the current config file, if it exists.
	cfg, err := ReadGlobalFile()
	if err != nil {
		return err
	}

	// Append the value to the workspaces list, if there is not already
	// an entry for it.
	section := cfg.Section("workspaces")
	exists := false
	for _, o := range section.Options {
		if o.Value == dir {
			exists = true
			break
		}
	}
	if !exists {
		section.Options = append(section.Options, &gitconfigfmt.Option{Key: "workspace", Value: dir})
	}
	return WriteGlobalFile(&cfg)
}

// EnsureWorkspaceNotInGlobalConfig ensures the specified directory is NOT in the
// global Zap config file under the "workspaces" section.
func EnsureWorkspaceNotInGlobalConfig(dir string) error {
	if GlobalConfigPath == "" {
		return nil
	}

	// Read the current config file, if it exists.
	cfg, err := ReadGlobalFile()
	if err != nil {
		return err
	}
	if len(cfg.Sections) == 0 {
		return nil // No need to write the config file below.
	}

	// Remove the value from the workspaces list, if there is an entry for it.
	section := cfg.Section("workspaces")
	for i, o := range section.Options {
		if o.Value == dir {
			section.Options = append(section.Options[:i], section.Options[i+1:]...)
			break
		}
	}
	if len(section.Options) == 0 {
		cfg.RemoveSection("workspaces")
	}
	return WriteGlobalFile(&cfg)
}
