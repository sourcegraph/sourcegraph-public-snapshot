package python

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

// defData is stored in graph.Def's Data field as JSON.
//
// TODO(beyang): I copied this from grapher_old.go and it's not being set right
// now. Pls update this formatter to work with whatever "Data" struct you do use.
type defData struct {
	Kind          string
	FuncSignature string
}

func init() {
	graph.RegisterMakeDefFormatter("PipPackage", newDefFormatter)
	graph.RegisterMakeDefFormatter("PythonProgram", newDefFormatter)
}

func newDefFormatter(s *graph.Def) graph.DefFormatter {
	var si defData
	if len(s.Data) > 0 {
		if err := json.Unmarshal(s.Data, &si); err != nil {
			panic("unmarshal Python def data: " + err.Error())
		}
	}
	return defFormatter{s, &si}
}

type defFormatter struct {
	def  *graph.Def
	data *defData
}

func (f defFormatter) Language() string { return "Python" }

func (f defFormatter) DefKeyword() string {
	if f.isFunc() {
		return "def"
	}
	if f.data.Kind == "class" {
		return "class"
	}
	if f.data.Kind == "module" {
		return "module"
	}
	if f.data.Kind == "package" {
		return "package"
	}
	return ""
}

func (f defFormatter) Kind() string { return f.data.Kind }

func dotted(slashed string) string { return strings.Replace(slashed, "/", ".", -1) }

func (f defFormatter) Name(qual graph.Qualification) string {
	if qual == graph.Unqualified {
		return f.def.Name
	}

	// Get the name of the containing package or module
	var containerName string
	if filename := filepath.Base(f.def.File); filename == "__init__.py" {
		containerName = filepath.Base(filepath.Dir(f.def.File))
	} else if strings.HasSuffix(filename, ".py") {
		containerName = filename[:len(filename)-len(".py")]
	} else if strings.HasSuffix(filename, ".c") {
		// Special case for Standard Lib C extensions
		return dotted(string(f.def.TreePath))
	} else {
		// Should never reach here, but fall back to TreePath if we do
		return string(f.def.TreePath)
	}

	// Compute the path relative to the containing package or module
	var treePathCmps = strings.Split(string(f.def.TreePath), "/")
	// Note(kludge): The first occurrence of the container name in the treepath may not be the correct occurrence.
	containerCmpIdx := -1
	for t, component := range treePathCmps {
		if component == containerName {
			containerCmpIdx = t
			break
		}
	}
	var relTreePath string
	if containerCmpIdx != -1 {
		relTreePath = strings.Join(treePathCmps[containerCmpIdx+1:], "/")
		if relTreePath == "" {
			relTreePath = "."
		}
	} else {
		// Should never reach here, but fall back to the unqualified name if we do
		relTreePath = f.def.Name
	}

	switch qual {
	case graph.ScopeQualified:
		return dotted(relTreePath)
	case graph.DepQualified:
		return dotted(filepath.Join(containerName, relTreePath))
	case graph.RepositoryWideQualified:
		return dotted(string(f.def.TreePath))
	case graph.LanguageWideQualified:
		return string(f.def.Repo) + "/" + f.Name(graph.RepositoryWideQualified)
	}
	panic("Name: unhandled qual " + string(qual))
}

func (f defFormatter) isFunc() bool {
	k := f.data.Kind
	return k == "function" || k == "method" || k == "constructor"
}

func (f defFormatter) NameAndTypeSeparator() string {
	if f.isFunc() {
		return ""
	}
	return " "
}

func (f defFormatter) Type(qual graph.Qualification) string {
	fullSig := f.data.FuncSignature
	if strings.Contains(fullSig, ")") { // kludge to get rid of extra type info (very noisy)
		return fullSig[:strings.Index(fullSig, ")")+1]
	}
	return fullSig
}
