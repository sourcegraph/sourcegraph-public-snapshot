package git

import "strings"

func EnsureRefPrefix(ref string) string {
	if strings.HasPrefix(ref, "refs/heads/") {
		return ref
	}
	return "refs/heads/" + ref
}
