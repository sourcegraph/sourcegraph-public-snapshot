package keyword

import "strings"

var keywordPrefixes = []string{
	"func ",
	"function ",
	"const ",
	"def ",
	"import ",
	"class ",
	"type ",
	"var ",
	"export ",
}

func hasKeywordPrefix(line string) bool {
	for _, prefix := range keywordPrefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}
