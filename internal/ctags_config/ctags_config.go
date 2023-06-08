package ctags_config

import "github.com/sourcegraph/sourcegraph/lib/errors"

type ParserType = uint8

const (
	UnknownCtags ParserType = iota
	NoCtags
	UniversalCtags
	ScipCtags
)

func ParserTypeToName(pType ParserType) string {
	switch pType {
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

type ParserConfiguration struct {
	Default ParserType
	Engine  map[string]ParserType
}

var SupportLanguages = map[string]struct{}{
	"Zig":    {},
	"Python": {},
	"C#":     {},
	"Java":   {},
}

var BaseParserConfig = ParserConfiguration{
	Engine: map[string]ParserType{
		// TODO: put our other languages here
		// TODO: also list the languages we support
		"Zig":    ScipCtags,
		"Python": ScipCtags,
		"C#":     ScipCtags,

		// TODO: Not ready to turn on java yet. Worried about not handling enough cases.
		// May wait until after next release
		// "Java":   ScipCtags,
	},
}
