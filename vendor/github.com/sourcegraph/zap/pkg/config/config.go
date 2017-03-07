package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	gitconfigfmt "github.com/src-d/go-git/plumbing/format/config"
)

var (
	// globalMu serializes concurrent writes to the global config.
	globalMu sync.Mutex

	// globalPath (if set) enables the use of a global Zap config file at
	// the given path. Global config is disabled if this value is empty.
	globalPath string
)

// SetGlobalConfigPath enables the use of a global Zap config file at the
// given path. Global config is disabled if this is never called.
//
// It will create the global config if it is missing.
func SetGlobalConfigPath(path string) error {
	globalPath = path

	if globalPath == "" {
		return nil
	}
	if _, err := os.Stat(globalPath); !os.IsNotExist(err) {
		return nil // file exists
	}

	globalMu.Lock()
	defer globalMu.Unlock()

	var cfg gitconfigfmt.Config
	cfg.Section("default").SetOption("remote", "wss://sourcegraph.com/.api/zap")
	return WriteFile(&cfg, globalPath)
}

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
	// We do not grab globalMu since that is only to protect concurrent
	// writes.
	var cfg gitconfigfmt.Config
	if globalPath == "" {
		return cfg, nil
	}
	cfg, err := ReadFile(globalPath)
	if err != nil && !os.IsNotExist(err) {
		return cfg, err
	}
	return cfg, nil
}

// UpdateGlobalFile will update the global config in a concurrency safe
// way. The apply function remutates the cfg struct and returns true if it
// wants the result written.
//
// Note: While apply is running it holds a lock. Avoid calling any global
// config functions in apply to avoid deadlock.
func UpdateGlobalFile(apply func(cfg *gitconfigfmt.Config) bool) error {
	if globalPath == "" {
		panic("global config is disabled (no global config path set)")
	}
	globalMu.Lock()
	defer globalMu.Unlock()
	cfg, err := ReadFile(globalPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if apply(&cfg) {
		return WriteFile(&cfg, globalPath)
	}
	return nil
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
	if globalPath == "" {
		return nil
	}

	return UpdateGlobalFile(func(cfg *gitconfigfmt.Config) bool {
		// Append the value to the workspaces list, if there is not already
		// an entry for it.
		section := cfg.Section("workspaces")
		for _, o := range section.Options {
			if o.Value == dir {
				return false
			}
		}
		section.Options = append(section.Options, &gitconfigfmt.Option{Key: "workspace", Value: dir})
		return true
	})
}

// EnsureWorkspaceNotInGlobalConfig ensures the specified directory is NOT in the
// global Zap config file under the "workspaces" section.
func EnsureWorkspaceNotInGlobalConfig(dir string) error {
	if globalPath == "" {
		return nil
	}

	return UpdateGlobalFile(func(cfg *gitconfigfmt.Config) bool {
		if len(cfg.Sections) == 0 {
			return false
		}

		// Remove the value from the workspaces list, if there is an entry for it.
		didMutate := false
		section := cfg.Section("workspaces")
		for i, o := range section.Options {
			if o.Value == dir {
				section.Options = append(section.Options[:i], section.Options[i+1:]...)
				didMutate = true
				break
			}
		}
		if len(section.Options) == 0 {
			didMutate = true
			cfg.RemoveSection("workspaces")
		}
		return didMutate
	})
}
