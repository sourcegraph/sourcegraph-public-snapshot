package search

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/camelcase"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

var delims = regexp.MustCompile(`[/.:\$\(\)\*\%\#\@\[\]\{\}]+`)

func ToTSVector(def *graph.Def) (aToks []string, bToks []string, cToks []string, dToks []string) {
	repoParts := strings.Split(def.Repo, "/")
	if len(repoParts) >= 1 && (strings.HasSuffix(repoParts[0], ".com") || strings.HasSuffix(repoParts[0], ".org")) {
		repoParts = repoParts[1:]
	}
	for _, w := range repoParts {
		bToks = appendRepeated(bToks, w, 1)
	}
	bToks = appendRepeated(bToks, repoParts[len(repoParts)-1], 2) // the last path component tends to be the repository name

	unitParts := strings.Split(def.Unit, "/")
	for _, w := range unitParts {
		bToks = appendRepeated(bToks, w, 1)
	}
	bToks = appendRepeated(bToks, unitParts[len(unitParts)-1], 2)

	defParts := delims.Split(def.Path, -1)
	for _, w := range defParts {
		bToks = appendRepeated(bToks, w, 2)
	}
	lastDefPart := defParts[len(defParts)-1]
	aToks = appendRepeated(aToks, lastDefPart, 3) // mega extra points for matching the last component of the def path (typically the "name" of the def)
	for _, w := range splitCaseWords(lastDefPart) {
		aToks = appendRepeated(aToks, w, 1) // more points for matching last component of def path
	}
	// CamelCase and snake_case tokens in the definition path
	for _, part := range defParts {
		for _, w := range splitCaseWords(part) {
			cToks = appendRepeated(cToks, w, 1)
		}
	}

	fileParts := strings.Split(filepath.ToSlash(def.File), "/")
	for _, w := range fileParts {
		cToks = appendRepeated(cToks, w, 1)
	}
	cToks = appendRepeated(cToks, fileParts[len(fileParts)-1], 2)

	aToks = appendRepeated(aToks, def.Name, 1)

	return
}

func splitCaseWords(w string) []string {
	if strings.Contains(w, "_") {
		return strings.Split(w, "_")
	}
	return camelcase.Split(w)
}

func appendRepeated(s []string, w string, count int) []string {
	for i := 0; i < count; i++ {
		s = append(s, w)
	}
	return s
}
