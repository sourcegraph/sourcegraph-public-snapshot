package parser

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/ctags_config"
)

var parserConfig = ctags_config.ParserConfiguration{
	Default: ctags_config.UniversalCtags,
	Engine:  map[string]ctags_config.ParserType{},
}

func init() {
	// Validation only: Do NOT set any values in the configuration in this function.
	conf.ContributeValidator(func(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
		configuration := c.SiteConfig().SyntaxHighlighting
		if configuration == nil {
			return
		}

		for _, engine := range configuration.Symbols.Engine {
			if _, err := ctags_config.ParserNameToParserType(engine); err != nil {
				problems = append(problems, conf.NewSiteProblem(fmt.Sprintf("Not a valid Symbols.Engine: `%s`.", engine)))
			}
		}

		return
	})

	go func() {
		conf.Watch(func() {
			c := conf.Get()

			configuration := c.SiteConfig().SyntaxHighlighting

			// Set the defaults
			parserConfig.Engine = make(map[string]ctags_config.ParserType)
			for lang, engine := range ctags_config.BaseParserConfig.Engine {
				parserConfig.Engine[lang] = engine
			}

			if configuration != nil {
				for lang, engine := range configuration.Symbols.Engine {
					if engine, err := ctags_config.ParserNameToParserType(engine); err != nil {
						parserConfig.Engine[lang] = engine
					}
				}
			}
		})
	}()
}

func GetParserType(language string) ctags_config.ParserType {
	parserType, ok := parserConfig.Engine[language]
	if !ok {
		return parserConfig.Default
	} else {
		return parserType
	}
}
