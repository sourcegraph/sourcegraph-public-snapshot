package htmlutil

import (
	"fmt"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
)

func init() {
	origTypes := chroma.StandardTypes
	sourcegraphTypes := map[chroma.TokenType]string{}
	for k, v := range origTypes {
		if k == chroma.PreWrapper {
			sourcegraphTypes[k] = v
		} else {
			sourcegraphTypes[k] = fmt.Sprintf("chroma-%s", v)
		}
	}
	chroma.StandardTypes = sourcegraphTypes
}

// SyntaxHighlightingOptions
func SyntaxHighlightingOptions() []html.Option {
	return []html.Option{
		html.WithClasses(true),
		html.WithLineNumbers(false),
	}
}
