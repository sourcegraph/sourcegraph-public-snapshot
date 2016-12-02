package java_def

import (
	"encoding/json"
	"strings"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

func init() {
	graph.RegisterMakeDefFormatter("Java", newDefFormatter)
	graph.RegisterMakeDefFormatter("JavaArtifact", newDefFormatter)
}

// DefData should be kept in sync with the def 'Data' field emitted by the Java
// grapher.
type DefData struct {
	JavaKind       string
	TypeExpression string
	Modifiers      []string
	Package        string
}

func (s *DefData) isLocalVar() bool {
	return s.JavaKind == "LOCAL_VARIABLE"
}

func newDefFormatter(s *graph.Def) graph.DefFormatter {
	var si DefData
	if len(s.Data) > 0 {
		if err := json.Unmarshal(s.Data, &si); err != nil {
			panic("unmarshal Java def data: " + err.Error())
		}
	}
	return defFormatter{s, &si}
}

type defFormatter struct {
	def  *graph.Def
	data *DefData
}

func (f defFormatter) Language() string { return "Java" }

func (f defFormatter) DefKeyword() string {
	switch f.data.JavaKind {
	case "CLASS":
		return "class"
	case "INTERFACE":
		return "interface"
	case "ENUM":
		return "enum"
	case "PACKAGE":
		return "package"
	case "METHOD":
		return "method"
	case "CONSTRUCTOR":
		return "constructor"
	default:
		return ""
	}
}

func (f defFormatter) Kind() string { return f.data.JavaKind }

func (f defFormatter) Name(qual graph.Qualification) string {

	switch qual {
	case graph.Unqualified:
		return f.def.Name
	case graph.ScopeQualified:
		return f.scopePrefix() + f.def.Name
	case graph.DepQualified, graph.RepositoryWideQualified:
		return f.widePrefix() + f.def.Name
	case graph.LanguageWideQualified:
		return string(f.def.Repo) + "/" + f.Name(graph.RepositoryWideQualified)
	}
	panic("Name: unhandled qual " + string(qual))
}

func (f defFormatter) Type(qual graph.Qualification) string {
	switch f.data.JavaKind {
	case "CLASS", "INTERFACE", "ENUM", "PACKAGE":
		return ""
	default:
		return strings.Replace(strings.Replace(f.data.TypeExpression, ",", ", ", -1), ")", ") ", -1)
	}
}

func (f defFormatter) NameAndTypeSeparator() string {
	if f.data.JavaKind == "CONSTRUCTOR" || f.data.JavaKind == "METHOD" {
		return ""
	}
	return " "
}

// Constructs scope-level prefix which consists for classes, interfaces, enums (including constants),
// and methods (including constructors) from path elements after package name concatenated by "."
// - for class foo.bar.baz.Qux scope prefix is empty string
// - for method foo.bar.baz.Qux.norf() scope prefix is Qux.
// - for package foo.bar.baz scope prefix is empty string
// - for inner enum foo.bar.baz.Qux.Norf scope prefix is Qux.
// - for inner enum's constant foo.bar.baz.Qux.Norf.A scope prefix is Qux.Norf.
func (f defFormatter) scopePrefix() string {
	switch f.data.JavaKind {
	case "CLASS", "INTERFACE", "METHOD", "CONSTRUCTOR", "ENUM", "ENUM_CONSTANT":
		pathComponents := strings.Split(strings.TrimPrefix(string(f.def.Path), strings.Replace(f.data.Package, ".", "/", -1)+"/"), "/")
		l := len(pathComponents)
		if l > 1 {
			pathComponents = pathComponents[l-2 : l-1]
			return prefix(pathComponents)
		}
	}
	return ""
}

// Constructs wide-level prefix which consists for classes, interfaces, enums (including constants),
// and methods (including constructors) from path elements except the last one concatenated by "."
// - for class foo.bar.baz.Qux prefix is foo.bar.baz.
// - for method foo.bar.baz.Qux.norf() prefix is foo.bar.baz.Qux.
// - for package foo.bar.baz prefix is foo.bar.
// - for inner enum foo.bar.baz.Qux.Norf prefix is foo.bar.baz.Qux
// - for inner enum's constant foo.bar.baz.Qux.Norf.A prefix is foo.bar.baz.Qux.Norf.
func (f defFormatter) widePrefix() string {
	switch f.data.JavaKind {
	case "CLASS", "INTERFACE", "METHOD", "CONSTRUCTOR", "ENUM", "ENUM_CONSTANT":
		pathComponents := strings.Split(string(f.def.Path), "/")
		pathComponents = pathComponents[:len(pathComponents)-1]
		return prefix(pathComponents)
	}
	return ""
}

// Concatenates path components by "." to form a prefix.
// Returns empty string if there are no path components available
func prefix(pathComponents []string) string {
	for i, component := range pathComponents {
		pathComponents[i] = strings.Replace(component, ":type", "", -1)
	}
	if len(pathComponents) > 0 {
		return strings.Join(pathComponents, ".") + "."
	}
	return ""
}
