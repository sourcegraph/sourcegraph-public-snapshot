package resolvers

import (
	"regexp"
	"strings"
)

func joinQueryParts(parts []string) string {
	return "((" + strings.Join(parts, ") OR (") + "))"
}

func joinPathPrefixRegexps(paths []string) string {
	var parts []string
	for _, path := range paths {
		if path == "" || path == "." || path == "/" {
			continue
		}
		parts = append(parts, "^"+regexp.QuoteMeta(path)+"($|/)")
	}
	return strings.Join(parts, "|")
}
