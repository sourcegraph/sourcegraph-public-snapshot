package conversion

import (
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func makeKey(parts ...string) string {
	return strings.Join(parts, ":")
}

func toID(id int) precise.ID {
	if id == 0 {
		return ""
	}

	return precise.ID(strconv.FormatInt(int64(id), 10))
}
