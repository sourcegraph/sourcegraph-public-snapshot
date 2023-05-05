package parser

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

type parserConfiguration struct {
	Default ParserType
	Engine  map[string]ParserType
}

var supportLanguages = map[string]struct{}{
	"rust": {},
}

var baseParserConfig = parserConfiguration{
	Engine: map[string]ParserType{
		// TODO: put our other languages here
		// TODO: also list the languages we support
		"rust": ScipCtags,
	},
}

var parserConfig = parserConfiguration{
	Default: UniversalCtags,
	Engine:  map[string]ParserType{},
}

func init() {
	// Validation only: Do NOT set any values in the configuration in this function.
	conf.ContributeValidator(func(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
		configuration := c.SiteConfig().SyntaxHighlighting
		if configuration == nil {
			return
		}

		for _, engine := range configuration.Symbols.Engine {
			if _, err := paserNameToParserType(engine); err != nil {
				problems = append(problems, conf.NewSiteProblem(fmt.Sprintf("Not a valid Symbols.Engine: `%s`.", engine)))
			}
		}

		return
	})

	go func() {
		conf.Watch(func() {
			c := conf.Get()

			configuration := c.SiteConfig().SyntaxHighlighting
			if configuration == nil {
				return
			}

			// Set the defaults
			parserConfig.Engine = make(map[string]ParserType)
			for lang, engine := range baseParserConfig.Engine {
				parserConfig.Engine[lang] = engine
			}

			for lang, engine := range configuration.Symbols.Engine {
				if engine, err := paserNameToParserType(engine); err != nil {
					parserConfig.Engine[lang] = engine
				}
			}
		})
	}()
}

func GetParserType(language string) ParserType {
	parserType, ok := parserConfig.Engine[language]
	if !ok {
		return parserConfig.Default
	} else {
		return parserType
	}
}
