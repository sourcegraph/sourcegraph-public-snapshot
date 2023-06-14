package ctags_config

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ParserType = uint8

const (
	UnknownCtags ParserType = iota
	NoCtags
	UniversalCtags
	ScipCtags
)

func ParserTypeToName(parserType ParserType) string {
	switch parserType {
	case NoCtags:
		return "off"
	case UniversalCtags:
		return "universal-ctags"
	case ScipCtags:
		return "scip-ctags"
	default:
		return "unknown-ctags-type"
	}
}

func ParserNameToParserType(name string) (ParserType, error) {
	switch name {
	case "off":
		return NoCtags, nil
	case "universal-ctags":
		return UniversalCtags, nil
	case "scip-ctags":
		return ScipCtags, nil
	default:
		return UnknownCtags, errors.Errorf("unknown parser type: %s", name)
	}
}

func ParserIsNoop(parserType ParserType) bool {
	return parserType == UnknownCtags || parserType == NoCtags
}

func LanguageSupportsParserType(language string, parserType ParserType) bool {
	switch parserType {
	case ScipCtags:
		_, ok := supportedLanguages[strings.ToLower(language)]
		return ok
	default:
		return true
	}
}

var supportedLanguages = map[string]struct{}{
	"c_sharp":    {},
	"go":         {},
	"java":       {},
	"javascript": {},
	"python":     {},
	"scala":      {},
	"typescript": {},
	"zig":        {},
}
