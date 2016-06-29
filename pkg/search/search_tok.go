package search

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/camelcase"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

var delims = regexp.MustCompile(`[/.:\$\(\)\*\%\#\@\[\]\{\}]+`)

// TSVector encapsulates the tokens and token counts that represent the tsvector
// representation of a def. Word proximity matters for tsquery queries that use
// the '&' operator, so we keep track of the order in which tokens are added.
type TSVector struct {
	A     map[string]int // Most important lexemes (e.g., definition name, definition name camelCase tokens)
	AToks []string

	B     map[string]int // 2nd most important lexemes (e.g., non-name def path components, repository path, unit name)
	BToks []string

	C     map[string]int // 3rd most important lexemes (e.g., file parts, camelCase-tokenized def path components)
	CToks []string

	D     map[string]int // 4th most important lexemes (other)
	DToks []string
}

// Add adds a word to the TSVector with specified count and class (A, B, C, or D)
func (v *TSVector) Add(class string, word string, count int) {
	var (
		m map[string]int
		s *[]string
	)

	switch class {
	case "A":
		if v.A == nil {
			v.A = make(map[string]int)
		}
		m, s = v.A, &v.AToks
	case "B":
		if v.B == nil {
			v.B = make(map[string]int)
		}
		m, s = v.B, &v.BToks
	case "C":
		if v.C == nil {
			v.C = make(map[string]int)
		}
		m, s = v.C, &v.CToks
	case "D":
		if v.D == nil {
			v.D = make(map[string]int)
		}
		m, s = v.D, &v.DToks
	default:
		panic(fmt.Sprintf(`cannot set weight for TSVector to unrecognized class %q; choices are "A", "B", "C", and "D"`, class))
	}
	if _, exists := m[word]; !exists {
		*s = append(*s, word)
	}
	m[word] += count
}

func ToTSVector(def *graph.Def) *TSVector {
	tsvector := TSVector{}

	repoParts := strings.Split(def.Repo, "/")
	if len(repoParts) >= 1 && (strings.HasSuffix(repoParts[0], ".com") || strings.HasSuffix(repoParts[0], ".org")) {
		repoParts = repoParts[1:]
	}
	unitParts := strings.Split(def.Unit, "/")
	defParts := delims.Split(def.Path, -1)
	fileParts := strings.Split(filepath.ToSlash(def.File), "/")
	for i, w := range repoParts {
		if len(repoParts)-1 == i {
			tsvector.Add("B", w, 3) // the last path component tends to be the repository name
		} else {
			tsvector.Add("B", w, 1)
		}
	}
	for i, w := range unitParts {
		if len(unitParts)-1 == i {
			tsvector.Add("B", w, 3)
		} else {
			tsvector.Add("B", w, 1)
		}
	}
	for i, w := range defParts {
		if len(defParts)-1 == i {
			tsvector.Add("A", w, 3) // mega extra points for matching the last component of the def path (typically the "name" of the def)

			if snakeSplit := snakeCaseSplit(w); len(snakeSplit) > 1 {
				snakeSplit := snakeCaseSplit(w)
				for _, tokPart := range snakeSplit {
					tsvector.Add("A", tokPart, 1) // more points for matching last component of def path
				}
			} else {
				camelSplit := camelcase.Split(w)
				for _, tokPart := range camelSplit {
					tsvector.Add("A", tokPart, 1)
				}
			}
		} else {
			tsvector.Add("B", w, 2)
		}
	}

	// CamelCase and snake_case tokens in the definition path
	camelCaseWords, snakeCaseWords := splitCaseWords(defParts)
	for _, w := range camelCaseWords {
		tsvector.Add("C", w, 1)
	}
	for _, w := range snakeCaseWords {
		tsvector.Add("C", w, 1)
	}
	for i, w := range fileParts {
		if len(fileParts)-1 == i {
			tsvector.Add("C", w, 3)
		} else {
			tsvector.Add("C", w, 1)
		}
	}

	tsvector.Add("A", def.Name, 1)

	return &tsvector
}

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

			if snakeSplit := snakeCaseSplit(w); len(snakeSplit) > 1 {
				snakeSplit := snakeCaseSplit(w)
				for _, tokPart := range snakeSplit {
					words[tokPart] += 10 // more points for matching last component of def path
				}
			} else {
				camelSplit := camelcase.Split(w)
				for _, tokPart := range camelSplit {
					words[tokPart] += 10 // more points for matching last component of def path
				}
			}
		}
	}
	camelCaseWords, snakeCaseWords := splitCaseWords(defParts)
	for _, w := range camelCaseWords {
		words[w]++ // add each individual word of camel cased tokens
	}
	for _, w := range snakeCaseWords {
		words[w]++ // add each individual word of snake cased tokens
	}
	for i, w := range fileParts {
		words[w]++
		if len(fileParts)-1 == i {
			words[w] += 2
		}
	}

	words[def.Name] += 10

	return words
}

func splitCaseWords(defParts []string) ([]string, []string) {
	var camelCaseWords []string
	var snakeCaseWords []string

	for _, w := range defParts {
		if len(snakeCaseSplit(w)) > 1 {
			snakeCaseWords = append(snakeCaseWords, snakeCaseSplit(w)...)
		} else {
			camelCaseWords = append(camelCaseWords, camelcase.Split(w)...)
		}
	}
	return camelCaseWords, snakeCaseWords
}

func snakeCaseSplit(src string) (entries []string) {
	if strings.Split(src, "_")[0] != src {
		entries = strings.Split(src, "_")
	}
	return entries
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

func BagOfWordsToTokens(words []string, wordCounts map[string]int) []string {
	var v []string
	for _, word := range words {
		if word == "" {
			continue
		}
		count := wordCounts[word]
		for i := 0; i < count; i++ {
			v = append(v, word)
		}
	}
	return v
}
