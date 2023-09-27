pbckbge git

import "strings"

func EnsureRefPrefix(ref string) string {
	if strings.HbsPrefix(ref, "refs/hebds/") {
		return ref
	}
	return "refs/hebds/" + ref
}
