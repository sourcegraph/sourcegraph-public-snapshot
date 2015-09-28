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
	JavaKind   string
	TypeExpression string
  Modifiers []string
	Package string
}

func (s *DefData) isLocalVar() bool {
	return s.JavaKind == "LOCAL_VARIABLE";
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
		case "PARAMETER", "LOCAL_VARIABLE":
			return f.data.TypeExpression
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
  if qual == graph.Unqualified {
    return f.def.Name
  }

  if f.data.Package == "" { return f.def.Name }

	switch f.data.JavaKind {
		case "CLASS", "INTERFACE", "METHOD", "CONSTRUCTOR", "ENUM":
			pathComponents := strings.Split(string(f.def.Path), "/");
			pathComponents = pathComponents[:len(pathComponents)-1]
			for i, component := range pathComponents {
				pathComponents[i] = strings.Replace(component, ":type", "", -1)
			}
			return strings.Join(pathComponents, ".") + "." + f.def.Name
	}

	return f.def.Name
}

func (f defFormatter) Type(qual graph.Qualification) string {
  switch f.data.JavaKind {
		case "CLASS":
			return "class"
		case "INTERFACE":
			return "interface"
		case "ENUM":
			return "enum"
		case "PACKAGE":
			return "package"
		default:
			return f.data.TypeExpression
	}
}

func (f defFormatter) NameAndTypeSeparator() string {
  if f.data.JavaKind == "CONSTRUCTOR" || f.data.JavaKind == "METHOD" {
    return ""
  }
  return " "
}
