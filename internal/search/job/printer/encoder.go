package printer

import (
	"strings"
)

func trimmedUpperName(name string) string {
	return strings.ToUpper(strings.TrimSuffix(name, "Job"))
}
