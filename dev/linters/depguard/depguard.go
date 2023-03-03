package depguard

import (
	"encoding/json"
	"strings"

	"github.com/OpenPeeDeeP/depguard/v2"
	"golang.org/x/tools/go/analysis"
)

var Analyzer *analysis.Analyzer = &analysis.Analyzer{
	Name: "depguard",
	Doc:  "Go linter that checks if package imports are in a list of acceptable packages",
	Run: func(p *analysis.Pass) (interface{}, error) {
		return nil, nil
	},
	RunDespiteErrors: false,
}

var settings = `{
  "main": {
    "files": [
      "$all",
      "!$test"
    ],
    "allow": [
      "$gostd",
      "github.com/OpenPeeDeeP"
    ],
    "deny": {
      "reflect": "Who needs reflection"
    }
  },
  "tests": {
    "files": [
      "$test"
    ],
    "deny": {
      "github.com/stretchr/testify": "Please use standard library for tests"
    }
  }
}

`

func init() {
	var depSettings depguard.LinterSettings
	if err := json.NewDecoder(strings.NewReader(settings)).Decode(&depSettings); err != nil {
		Analyzer.Run = func(p *analysis.Pass) (interface{}, error) {
			return nil, err
		}
		return
	}

	analyzer, err := depguard.NewAnalyzer(&depSettings)
	if err != nil {
		Analyzer.Run = func(p *analysis.Pass) (interface{}, error) {
			return nil, err
		}
	} else {
		Analyzer = analyzer
	}
}
