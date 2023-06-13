package ctags_config

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/languages"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
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
	"java":       {},
	"javascript": {},
	"python":     {},
	"scala":      {},
	"typescript": {},
	"zig":        {},
}

var DefaultEngines = map[string]ParserType{
	// Add the languages we want to turn on by default (you'll need to
	// update the ctags_config module for supported languages as well)
	"zig": ScipCtags,

	// TODO: Turn these on in the next PR, so there is no runtime differences for this PR.
	// "c_sharp": ctags_config.ScipCtags,
	// "python":  ctags_config.ScipCtags,

	// TODO: Not ready to turn on java yet. Worried about not handling enough cases.
	// May wait until after next release
	// "Java":   ScipCtags,
}

func CreateEngineMap(siteConfig schema.SiteConfiguration) map[string]ParserType {
	// Set the defaults
	engines := make(map[string]ParserType)
	for lang, engine := range DefaultEngines {
		lang = languages.NormalizeLanguage(lang)
		engines[lang] = engine
	}

	// Set any relevant overrides
	configuration := siteConfig.SyntaxHighlighting
	if configuration != nil {
		for lang, engine := range configuration.Symbols.Engine {
			lang = languages.NormalizeLanguage(lang)

			if engine, err := ParserNameToParserType(engine); err != nil {
				engines[lang] = engine
			}
		}
	}

	return engines
}
