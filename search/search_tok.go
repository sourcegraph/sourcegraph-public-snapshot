package search

import (
	"path/filepath"
	"regexp"
	"strings"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

var delims = regexp.MustCompile(`[/.:\$\(\)\*\%\#\@\[\]\{\}]+`)

func BagOfWords(def *graph.Def) map[string]int {
	words := make(map[string]int)

	repoParts := strings.Split(def.Repo, "/")
	if len(repoParts) >= 1 && (strings.HasSuffix(repoParts[0], ".com") || strings.HasSuffix(repoParts[0], ".org")) {
		repoParts = repoParts[1:]
	}
	unitParts := strings.Split(def.Unit, "/")
	defParts := delims.Split(def.Path, -1)
	fileParts := strings.Split(filepath.ToSlash(def.File), "/")
	for i, w := range repoParts {
		words[w]++
		if len(repoParts)-1 == i {
			words[w] += 2 // the last path tends to be the name
		}
	}
	for i, w := range unitParts {
		words[w]++
		if len(unitParts)-1 == i {
			words[w] += 2
		}
	}
	for i, w := range defParts {
		words[w] += 2
		if len(defParts)-1 == i {
			words[w] += 30 // mega extra points for matching the last component of the def path (typically the "name" of the def)
		}
	}
	for i, w := range fileParts {
		words[w]++
		if len(fileParts)-1 == i {
			words[w] += 2
		}
	}

	words[def.Name] += 10
	words[def.Kind] += 1

	return words
}

func UserQueryToksToTSQuery(toks []string) string {
	if len(toks) == 0 {
		return ""
	}
	tokMatch := strings.Join(toks, " & ")
	return tokMatch
}

// strippedQuery is the user query after it has been stripped of special filter terms
func QueryTokens(strippedQuery string) []string {
	prototoks := delims.Split(strippedQuery, -1)
	if len(prototoks) == 0 {
		return nil
	}
	toks := make([]string, 0, len(prototoks))
	for _, tokmaybe := range prototoks {
		if tokmaybe != "" {
			toks = append(toks, tokmaybe)
		}
	}
	return toks
}

func BagOfWordsToTokens(bag map[string]int) []string {
	var v []string
	for word, count := range bag {
		if word == "" {
			continue
		}
		for i := 0; i < count; i++ {
			v = append(v, word)
		}
	}
	return v
}
