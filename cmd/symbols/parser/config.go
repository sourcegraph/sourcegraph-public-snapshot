package parser

import (
	"fmt"
	"sync"

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

var parserConfigMutex sync.Mutex
var parserConfig = ParserConfiguration{
	Default: ctags_config.UniversalCtags,
	Engine:  map[string]ctags_config.ParserType{},
}

func init() {
	// Validation only: Do NOT set any values in the configuration in this function.
	conf.ContributeValidator(func(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
		configuration := c.SiteConfig().SyntaxHighlighting
		if configuration == nil || configuration.Symbols == nil {
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
			parserConfigMutex.Lock()
			defer parserConfigMutex.Unlock()

			parserConfig.Engine = ctags_config.CreateEngineMap(conf.Get().SiteConfig())
		})
	}()
}

func GetParserType(language string) ctags_config.ParserType {
	language = languages.NormalizeLanguage(language)

	parserConfigMutex.Lock()
	defer parserConfigMutex.Unlock()

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
