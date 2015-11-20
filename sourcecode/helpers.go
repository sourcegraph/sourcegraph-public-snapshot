package sourcecode

import (
	"fmt"
	"html/template"
	"path/filepath"
	"regexp"
	"strings"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

var wordBreaks = regexp.MustCompile(`([\./:])`)
var wordBreakSentinel = "wb_c3642b62"
var defPathTokenizer = regexp.MustCompile(`[^\.\(\)\*\s:#]+|[\.\(\)\*\s:#]+`)

func htmlEscapeStringWithCodeBreaks(code string) string {
	code = wordBreaks.ReplaceAllString(code, "${1}"+wordBreakSentinel)
	code = template.HTMLEscapeString(code)
	code = strings.Replace(code, wordBreakSentinel, "<wbr>", -1)
	return code
}

func DefQualifiedNameAndType(s *sourcegraph.Def, qualStr string) template.HTML {
	sf := graph.PrintFormatter(&s.Def)
	qual := graph.Qualification(qualStr)
	sepAndType := htmlEscapeStringWithCodeBreaks(sf.NameAndTypeSeparator() + sf.Type(qual))
	return DefQualifiedName(s, qualStr) + template.HTML(sepAndType)
}

func DefQualifiedName(def *sourcegraph.Def, qualStr string) template.HTML {
	sf := graph.PrintFormatter(&def.Def)
	qual := graph.Qualification(qualStr)
	qualName := htmlEscapeStringWithCodeBreaks(sf.Name(qual))
	escapedName := htmlEscapeStringWithCodeBreaks(def.Name)
	wrappedName := fmt.Sprintf(`<wbr><span class="name">%s</span>`, escapedName)
	cmps := defPathTokenizer.FindAllString(qualName, -1)
	for c, cmp := range cmps {
		if cmp == escapedName || cmp == "<wbr>"+escapedName {
			cmps[c] = wrappedName
		}
	}
	return template.HTML(strings.Join(cmps, ""))
}

// DefNameFromSpec should only be used when the Def is missing for whatever
// reason
func DefNameFromSpec(defSpec *sourcegraph.DefSpec) template.HTML {
	name := fmt.Sprintf("%s %s", defSpec.Unit, defSpec.Path)
	escapedName := htmlEscapeStringWithCodeBreaks(name)
	wrappedName := fmt.Sprintf(`<wbr><span class="name">%s</span>`, escapedName)
	return template.HTML(wrappedName)
}

// IsLikelyCodeFile returns whether the given path is likely a source code file
// that Sourcegraph might have processed.
func IsLikelyCodeFile(path string) bool {
	base := filepath.Base(strings.ToLower(path))
	if _, ok := exceptions[base]; ok {
		return true
	}
	if strings.HasPrefix(base, ".") {
		return false
	}
	if _, ok := nonCodeNames[base]; ok {
		return false
	}
	if _, ok := nonCodeExts[filepath.Ext(base)]; ok {
		return false
	}
	return true
}

var (
	nonCodeNames = map[string]struct{}{
		"readme":       {},
		"license":      {},
		"authors":      {},
		"contributors": {},
		"changelog":    {},
		"install":      {},
	}
	nonCodeExts = map[string]struct{}{
		".md":       {},
		".txt":      {},
		".rst":      {},
		".mdown":    {},
		".markdown": {},
		".ascii":    {},
		".rdoc":     {},
	}
	exceptions = map[string]struct{}{
		".aurora": {},
	}
)
