package searchquery

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	reOperationPrefix        = `(?P<Operation>-?[a-z0-9]+)`
	reEscapedQuoteOrNotQuote = `((\\")|[^"])`
	reAnyValueInQuote        = `(.` + reEscapedQuoteOrNotQuote + `*)`
	reQuotedValue            = `("` + reAnyValueInQuote + `")`
	reUnquotedValue          = `([^ ]+(\s|$))`
	reValue                  = `(?P<Value>` + reQuotedValue + `|` + reUnquotedValue + `)`
	reOperationAndValue      = `(?P<OperationAndValue>` + reOperationPrefix + `:` + reValue + `)`
	re                       = regexp.MustCompile(`\s?` + reOperationAndValue + `\s?`)
)

// Parse parses a search query. See the tests for examples of what this looks like.
func Parse(q string) (remaining string, operations [][2]string) {
	for _, match := range re.FindAllStringSubmatch(q, -1) {
		for i := 0; i < len(match); i++ {
			name := re.SubexpNames()[i]
			if name == "OperationAndValue" {
				operation := match[i+1]
				value := match[i+2]
				if []rune(value)[0] == '"' {
					value, _ = strconv.Unquote(value)
				} else {
					value = strings.TrimSpace(value)
				}
				operations = append(operations, [2]string{operation, value})
				i += 2
			}
		}
	}
	var remainingItems []string
	for _, item := range re.Split(q, -1) {
		item = strings.TrimSpace(item)
		if item != "" {
			remainingItems = append(remainingItems, item)
		}
	}
	remaining = strings.Join(remainingItems, " ")
	remaining = strings.Replace(remaining, `\:`, `:`, -1)
	return
}
