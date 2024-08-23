package generate

import (
	_ "embed"
	"go/token"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var cfgFilenames = []string{".genqlient.yml", ".genqlient.yaml", "genqlient.yml", "genqlient.yaml"}

// Config represents genqlient's configuration, generally read from
// genqlient.yaml.
//
// Callers must call ValidateAndFillDefaults before using the config.
type Config struct {
	// The following fields are documented at:
	// https://github.com/Khan/genqlient/blob/main/docs/genqlient.yaml
	Schema           StringList              `yaml:"schema"`
	Operations       StringList              `yaml:"operations"`
	Generated        string                  `yaml:"generated"`
	Package          string                  `yaml:"package"`
	ExportOperations string                  `yaml:"export_operations"`
	ContextType      string                  `yaml:"context_type"`
	ClientGetter     string                  `yaml:"client_getter"`
	Bindings         map[string]*TypeBinding `yaml:"bindings"`
	Optional         string                  `yaml:"optional"`
	StructReferences bool                    `yaml:"use_struct_references"`
	Extensions       bool                    `yaml:"use_extensions"`

	// Set to true to use features that aren't fully ready to use.
	//
	// This is primarily intended for genqlient's own tests.  These features
	// are likely BROKEN and come with NO EXPECTATION OF COMPATIBILITY.  Use
	// them at your own risk!
	AllowBrokenFeatures bool `yaml:"allow_broken_features"`

	// The directory of the config-file (relative to which all the other paths
	// are resolved).  Set by ValidateAndFillDefaults.
	baseDir string
}

// A TypeBinding represents a Go type to which genqlient will bind a particular
// GraphQL type, and is documented further at:
// https://github.com/Khan/genqlient/blob/main/docs/genqlient.yaml
type TypeBinding struct {
	Type              string `yaml:"type"`
	ExpectExactFields string `yaml:"expect_exact_fields"`
	Marshaler         string `yaml:"marshaler"`
	Unmarshaler       string `yaml:"unmarshaler"`
}

// pathJoin is like filepath.Join but 1) it only takes two argsuments,
// and b) if the second argument is an absolute path the first argument
// is ignored (similar to how python's os.path.join() works).
func pathJoin(a, b string) string {
	if filepath.IsAbs(b) {
		return b
	}
	return filepath.Join(a, b)
}

// ValidateAndFillDefaults ensures that the configuration is valid, and fills
// in any options that were unspecified.
//
// The argument is the directory relative to which paths will be interpreted,
// typically the directory of the config file.
func (c *Config) ValidateAndFillDefaults(baseDir string) error {
	c.baseDir = baseDir
	for i := range c.Schema {
		c.Schema[i] = pathJoin(baseDir, c.Schema[i])
	}
	for i := range c.Operations {
		c.Operations[i] = pathJoin(baseDir, c.Operations[i])
	}
	c.Generated = pathJoin(baseDir, c.Generated)
	if c.ExportOperations != "" {
		c.ExportOperations = pathJoin(baseDir, c.ExportOperations)
	}

	if c.ContextType == "" {
		c.ContextType = "context.Context"
	}

	if c.Package == "" {
		abs, err := filepath.Abs(c.Generated)
		if err != nil {
			return errorf(nil, "unable to guess package-name: %v", err)
		}

		base := filepath.Base(filepath.Dir(abs))
		if !token.IsIdentifier(base) {
			return errorf(nil, "unable to guess package-name: %v is not a valid identifier", base)
		}

		c.Package = base
	}

	return nil
}

// ReadAndValidateConfig reads the configuration from the given file, validates
// it, and returns it.
func ReadAndValidateConfig(filename string) (*Config, error) {
	text, err := os.ReadFile(filename)
	if err != nil {
		return nil, errorf(nil, "unreadable config file %v: %v", filename, err)
	}

	var config Config
	err = yaml.UnmarshalStrict(text, &config)
	if err != nil {
		return nil, errorf(nil, "invalid config file %v: %v", filename, err)
	}

	err = config.ValidateAndFillDefaults(filepath.Dir(filename))
	if err != nil {
		return nil, errorf(nil, "invalid config file %v: %v", filename, err)
	}

	return &config, nil
}

// ReadAndValidateConfigFromDefaultLocations looks for a config file in the
// current directory, and all parent directories walking up the tree. The
// closest config file will be returned.
func ReadAndValidateConfigFromDefaultLocations() (*Config, error) {
	cfgFile, err := findCfg()
	if err != nil {
		return nil, err
	}
	return ReadAndValidateConfig(cfgFile)
}

//go:embed default_genqlient.yaml
var defaultConfig []byte

func initConfig(filename string) error {
	return os.WriteFile(filename, defaultConfig, 0o644)
}

// findCfg searches for the config file in this directory and all parents up the tree
// looking for the closest match
func findCfg() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", errorf(nil, "unable to get working dir to findCfg: %v", err)
	}

	cfg := findCfgInDir(dir)

	for cfg == "" && dir != filepath.Dir(dir) {
		dir = filepath.Dir(dir)
		cfg = findCfgInDir(dir)
	}

	if cfg == "" {
		return "", os.ErrNotExist
	}

	return cfg, nil
}

func findCfgInDir(dir string) string {
	for _, cfgName := range cfgFilenames {
		path := pathJoin(dir, cfgName)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}
