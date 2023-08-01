package redactor

import (
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

var urlRegex = lazyregexp.New(`((https?|ssh|git)://[^:@]+:)[^@]+(@)`)

func RedactCommandArgs(args []string) []string {
	var redactedArgs = make([]string, len(args))
	for i, arg := range args {
		redactedArgs[i] = urlRegex.ReplaceAllString(arg, "$1<REDACTED>$3")
	}
	return redactedArgs
}
