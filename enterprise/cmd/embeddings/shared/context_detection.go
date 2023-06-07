package shared

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

var NO_CONTEXT_MESSAGES_REGEXPS = []*lazyregexp.Regexp{
	lazyregexp.New(`(previous|above)\s+(message|code|text)`),
	lazyregexp.New(
		`(translate|convert|change|for|make|refactor|rewrite|ignore|explain|fix|try|show)\s+(that|this|above|previous|it|again)`,
	),
	lazyregexp.New(
		`(this|that).*?\s+(is|seems|looks)\s+(wrong|incorrect|bad|good)`,
	),
	lazyregexp.New(`^(yes|no|correct|wrong|nope|yep|now|cool)(\s|.|,)`),
	// User provided their own code context in the form of a Markdown code block.
	lazyregexp.New("```"),
}

func isContextRequiredForChatQuery(query string) bool {
	queryTrimmed := strings.TrimSpace(query)
	queryLower := strings.ToLower(queryTrimmed)
	for _, regexp := range NO_CONTEXT_MESSAGES_REGEXPS {
		if regexp.MatchString(queryLower) {
			return false
		}
	}
	return true
}
