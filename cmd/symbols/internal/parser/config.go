package parser

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/conf"
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
