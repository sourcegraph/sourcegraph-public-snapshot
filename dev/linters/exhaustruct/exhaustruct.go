package exhaustruct

import (
	_ "embed"
	"fmt"

	"github.com/GaijinEntertainment/go-exhaustruct/v3/analyzer"
	"golang.org/x/tools/go/analysis"
	"gopkg.in/yaml.v3"

	"github.com/sourcegraph/sourcegraph/dev/linters/nolint"
)

var Analyzer *analysis.Analyzer = nolint.Wrap(createAnalyzer())

//go:embed lint-config.yaml
var lintConfigYAML string

type Config struct {
	IncludeTypes []string `yaml:"include_types"`
	ExcludeTypes []string `yaml:"exclude_types"`
}

func createAnalyzer() *analysis.Analyzer {
	var config Config
	if err := yaml.Unmarshal([]byte(lintConfigYAML), &config); err != nil {
		panic(fmt.Sprintf("Malformed lint-config.yaml: %v", err))
	}

	analyzer, err := analyzer.NewAnalyzer(config.IncludeTypes, config.ExcludeTypes)
	if err != nil {
		panic(err)
	}

	return analyzer
}
