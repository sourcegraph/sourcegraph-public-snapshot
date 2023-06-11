package parser

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/ctags_config"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/languages"
)

type ParserType = ctags_config.ParserType

type ParserConfiguration struct {
	Default ParserType
	Engine  map[string]ParserType
}

var parserConfig = ParserConfiguration{
	Default: ctags_config.UniversalCtags,
	Engine:  map[string]ctags_config.ParserType{},
}

var defaultParserConfig = ParserConfiguration{
	Engine: map[string]ParserType{
		// Add the languages we want to turn on by default (you'll need to
		// update the ctags_config module for supported languages as well)
		"zig": ctags_config.ScipCtags,

		// TODO: Turn these on in the next PR, so there is no runtime differences for this PR.
		// "c_sharp": ctags_config.ScipCtags,
		// "python":  ctags_config.ScipCtags,

		// TODO: Not ready to turn on java yet. Worried about not handling enough cases.
		// May wait until after next release
		// "Java":   ScipCtags,
	},
}

func init() {
	// Validation only: Do NOT set any values in the configuration in this function.
	conf.ContributeValidator(func(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
		configuration := c.SiteConfig().SyntaxHighlighting
		if configuration == nil {
			return nil
		}

		for language, engine := range configuration.Symbols.Engine {
			parser_engine, err := ctags_config.ParserNameToParserType(engine)
			if err != nil {
				return conf.NewSiteProblems(fmt.Sprintf("Not a valid Symbols.Engine: `%s`.", engine))
			}

			language = languages.NormalizeLanguage(language)
			if !ctags_config.LanguageSupportsParserType(language, parser_engine) {
				return conf.NewSiteProblems(fmt.Sprintf("Not a valid Symbols.Engine for language: %s `%s`.", language, engine))
			}

		}

		return nil
	})

	// Update parserConfig here
	go func() {
		conf.Watch(func() {
			// Set the defaults
			parserConfig.Engine = make(map[string]ctags_config.ParserType)
			for lang, engine := range defaultParserConfig.Engine {
				lang = languages.NormalizeLanguage(lang)
				parserConfig.Engine[lang] = engine
			}

			// Set any relevant overrides
			c := conf.Get()
			configuration := c.SiteConfig().SyntaxHighlighting
			if configuration != nil {
				for lang, engine := range configuration.Symbols.Engine {
					lang = languages.NormalizeLanguage(lang)

					if engine, err := ctags_config.ParserNameToParserType(engine); err != nil {
						parserConfig.Engine[lang] = engine
					}
				}
			}
		})
	}()
}

func GetParserType(language string) ctags_config.ParserType {
	language = languages.NormalizeLanguage(language)

	parserType, ok := parserConfig.Engine[language]
	if !ok {
		parserType = parserConfig.Default
	}

	// Default back to UniversalCtags if somehow we've got an unsupported
	// type by this time. (I don't think it's possible)
	if !ctags_config.LanguageSupportsParserType(language, parserType) {
		return ctags_config.UniversalCtags
	}

	return parserType
}
