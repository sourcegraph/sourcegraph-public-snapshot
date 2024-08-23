package ff

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// Parse the flags in the flag set from the provided (presumably commandline)
// args. Additional options may be provided to parse from a config file and/or
// environment variables in that priority order.
func Parse(fs *flag.FlagSet, args []string, options ...Option) error {
	var c Context
	for _, option := range options {
		option(&c)
	}

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("error parsing commandline args: %w", err)
	}

	provided := map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		provided[f.Name] = true
	})

	if c.configFile == "" && c.configFileFlagName != "" {
		if f := fs.Lookup(c.configFileFlagName); f != nil {
			c.configFile = f.Value.String()
		}
	}

	if c.configFile != "" && c.configFileParser != nil {
		f, err := os.Open(c.configFile)
		if err == nil {
			defer f.Close()
			err = c.configFileParser(f, func(name, value string) error {
				if fs.Lookup(name) == nil {
					if c.ignoreUndefined {
						return nil
					}
					return fmt.Errorf("config file flag %q not defined in flag set", name)
				}

				if provided[name] {
					return nil // commandline args take precedence
				}

				if err := fs.Set(name, value); err != nil {
					return fmt.Errorf("error setting flag %q from config file: %v", name, err)
				}

				return nil
			})
			if err != nil {
				return err
			}
		} else if err != nil && !c.allowMissingConfigFile {
			return err
		}
	}

	fs.Visit(func(f *flag.Flag) {
		provided[f.Name] = true
	})

	if c.envVarPrefix != "" || c.envVarNoPrefix {
		var errs []string
		fs.VisitAll(func(f *flag.Flag) {
			if provided[f.Name] {
				return // commandline args and config file take precedence
			}

			var key string
			{
				key = strings.ToUpper(f.Name)
				key = envVarReplacer.Replace(key)
				if !c.envVarNoPrefix {
					key = strings.ToUpper(c.envVarPrefix) + "_" + key
				}
			}
			if value := os.Getenv(key); value != "" {
				var values []string
				if c.envVarIgnoreCommas {
					values = []string{value}
				} else {
					values = strings.Split(value, ",")
				}
				for _, v := range values {
					if err := fs.Set(f.Name, v); err != nil {
						errs = append(errs, fmt.Sprintf("error setting flag %q from env var %q: %v", f.Name, key, err))
					}
				}
			}
		})
		if len(errs) > 0 {
			return fmt.Errorf("error parsing env vars: %s", strings.Join(errs, "; "))
		}
	}

	return nil
}

// Context contains private fields used during parsing.
type Context struct {
	configFile             string
	configFileFlagName     string
	configFileParser       ConfigFileParser
	allowMissingConfigFile bool
	envVarPrefix           string
	envVarNoPrefix         bool
	envVarIgnoreCommas     bool
	ignoreUndefined        bool
}

// Option controls some aspect of parse behavior.
type Option func(*Context)

// WithConfigFile tells parse to read the provided filename as a config file.
// Requires WithConfigFileParser, and overrides WithConfigFileFlag.
func WithConfigFile(filename string) Option {
	return func(c *Context) {
		c.configFile = filename
	}
}

// WithConfigFileFlag tells parse to treat the flag with the given name as a
// config file. Requires WithConfigFileParser, and is overridden by WithConfigFile.
func WithConfigFileFlag(flagname string) Option {
	return func(c *Context) {
		c.configFileFlagName = flagname
	}
}

// WithConfigFileParser tells parse how to interpret the config file provided via
// WithConfigFile or WithConfigFileFlag.
func WithConfigFileParser(p ConfigFileParser) Option {
	return func(c *Context) {
		c.configFileParser = p
	}
}

// WithAllowMissingConfigFile will permit parse to succeed, even if a provided
// config file doesn't exist.
func WithAllowMissingConfigFile(allow bool) Option {
	return func(c *Context) {
		c.allowMissingConfigFile = allow
	}
}

// WithEnvVarPrefix tells parse to look in the environment for variables with
// the given prefix. Flag names are converted to environment variables by
// capitalizing them, and replacing separator characters like periods or hyphens
// with underscores.
func WithEnvVarPrefix(prefix string) Option {
	return func(c *Context) {
		c.envVarPrefix = prefix
	}
}

// WithEnvVarNoPrefix tells parse to look in the environment for variables with
// no prefix. See WithEnvVarPrefix for an explanation of how flag names are
// converted to environment variables names.
func WithEnvVarNoPrefix() Option {
	return func(c *Context) {
		c.envVarNoPrefix = true
	}
}

// WithEnvVarIgnoreCommas tells parse to ignore commas in environment variable
// values, treating the complete value as a single string passed to the
// associated flag. By default, if an environment variable's value contains
// commas, each comma-delimited token is treated as a separate instance of the
// associated flag.
func WithEnvVarIgnoreCommas(ignore bool) Option {
	return func(c *Context) {
		c.envVarIgnoreCommas = ignore
	}
}

// WithIgnoreUndefined tells parse to ignore undefined flags that it encounters,
// which would normally throw an error.
func WithIgnoreUndefined(ignore bool) Option {
	return func(c *Context) {
		c.ignoreUndefined = ignore
	}
}

// ConfigFileParser interprets the config file represented by the reader
// and calls the set function for each parsed flag pair.
type ConfigFileParser func(r io.Reader, set func(name, value string) error) error

// PlainParser is a parser for config files in an extremely simple format. Each
// line is tokenized as a single key/value pair. The first whitespace-delimited
// token in the line is interpreted as the flag name, and all remaining tokens
// are interpreted as the value. Any leading hyphens on the flag name are
// ignored.
func PlainParser(r io.Reader, set func(name, value string) error) error {
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue // skip empties
		}

		if line[0] == '#' {
			continue // skip comments
		}

		var (
			name  string
			value string
			index = strings.IndexRune(line, ' ')
		)
		if index < 0 {
			name, value = line, "true" // boolean option
		} else {
			name, value = line[:index], strings.TrimSpace(line[index:])
		}

		if i := strings.Index(value, " #"); i >= 0 {
			value = strings.TrimSpace(value[:i])
		}

		if err := set(name, value); err != nil {
			return err
		}
	}
	return nil
}

var envVarReplacer = strings.NewReplacer(
	"-", "_",
	".", "_",
	"/", "_",
)
