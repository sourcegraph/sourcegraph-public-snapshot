package command

import (
	"fmt"
	"strings"

	"github.com/kballard/go-shellquote"
)

// flatten combines string values and (non-recursive) string slice values
// into a single string slice.
func flatten(values ...any) []string {
	union := make([]string, 0, len(values))
	for _, value := range values {
		switch v := value.(type) {
		case string:
			union = append(union, v)
		case []string:
			union = append(union, v...)
		}
	}

	return union
}

// intersperse returns a slice following the pattern `flag, v1, flag, v2, ...`.
func intersperse(flag string, values []string) []string {
	interspersed := make([]string, 0, len(values))
	for _, v := range values {
		interspersed = append(interspersed, flag, v)
	}

	return interspersed
}

// quoteEnv returns a slice of env vars in which the values are properly shell quoted.
func quoteEnv(env []string) []string {
	quotedEnv := make([]string, len(env))

	for i, e := range env {
		elems := strings.SplitN(e, "=", 2)
		quotedEnv[i] = fmt.Sprintf("%s=%s", elems[0], shellquote.Join(elems[1]))
	}

	return quotedEnv
}
