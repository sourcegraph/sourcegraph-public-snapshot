package javascript_def

import (
	"encoding/json"
	"path/filepath"

	"strings"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

func init() {
	graph.RegisterMakeDefFormatter("CommonJSPackage", newDefFormatter)
}

// jsgDefData is the "data" field output by jsg.
//
// It must be kept in sync with what the JavaScript code emits.
type jsgDefData struct {
	NodeJS *struct {
		ModuleExports bool
	}
	AMD *struct {
		Module bool
	}
}

// defData is stored in the graph.Def's Data field as JSON.
//
// It must be kept in sync with what the JavaScript code emits.
type defData struct {
	Kind string
	Key  DefPath
	*jsgDefData
	Type   string
	IsFunc bool
}

// DefPath must be kept in sync with what the JavaScript code emits.
type DefPath struct {
	Namespace string
	Module    string
	Path      string
}

func newDefFormatter(s *graph.Def) graph.DefFormatter {
	var si defData
	if len(s.Data) > 0 {
		if err := json.Unmarshal(s.Data, &si); err != nil {
			panic("unmarshal JavaScript def data: " + err.Error())
		}
	}
	return defFormatter{s, &si}
}

type defFormatter struct {
	def  *graph.Def
	data *defData
}

func (f defFormatter) Language() string { return "JavaScript" }

func (f defFormatter) DefKeyword() string {
	switch f.data.Kind {
	case Func:
		return "function"
	case Var:
		return "var"
	}
	return ""
}

func (f defFormatter) Kind() string { return f.data.Kind }

func (f defFormatter) Name(qual graph.Qualification) string {
	if f.data.Key.Namespace == "global" || f.data.Key.Namespace == "file" {
		return scopePathComponentsAfterAtSign(f.data.Key.Path)
	}
	switch qual {
	case graph.Unqualified:
		return f.def.Name
	case graph.ScopeQualified:
		return f.data.Key.Path
	case graph.DepQualified:
		return strings.TrimSuffix(filepath.Base(f.data.Key.Module), ".js") + "." + f.Name(graph.ScopeQualified)
	case graph.RepositoryWideQualified:
		return filepath.Join(f.def.Unit, strings.TrimSuffix(f.data.Key.Module, ".js")) + "." + f.Name(graph.ScopeQualified)
	case graph.LanguageWideQualified:
		return string(f.def.Repo) + "/" + f.Name(graph.RepositoryWideQualified)
	}
	panic("Name: unhandled qual " + string(qual))
}

func (f defFormatter) NameAndTypeSeparator() string {
	if f.data.IsFunc {
		return ""
	}
	return " "
}

func (f defFormatter) Type(qual graph.Qualification) string {
	var ts string
	if f.data.IsFunc {
		ts = strings.Replace(strings.TrimPrefix(f.data.Type, "fn"), ") -> ", ") ", -1)
	} else {
		ts = f.data.Type
	}

	ts = strings.Replace(ts, ": ?", "", -1)
	return ts
}

const (
	// Kinds
	NPMPackage      = "npm_package"
	CommonJSModule  = "commonjs_module"
	AMDModule       = "amd_module"
	Func            = "func"
	ConstructorFunc = "constructor_func"
	Var             = "var"
	Property        = "property"
	Prototype       = "prototype"
)

func scopePathComponentsAfterAtSign(path string) string {
	parts := strings.Split(path, ".")
	if len(parts) == 1 {
		return path
	}
	for i := len(parts) - 2; i >= 0; i-- {
		part := parts[i]
		if strings.Contains(part, "@") {
			return strings.Join(parts[i+1:], ".")
		}
	}
	return path
}
