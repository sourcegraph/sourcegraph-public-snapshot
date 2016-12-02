// Package dockerfiledef provides a srclib def formatter for Dockerfile defs.
package dockerfiledef

import (
	"fmt"
	"path/filepath"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

const dockerfileSourceUnitType = "Dockerfile"

func init() {
	graph.RegisterMakeDefFormatter(dockerfileSourceUnitType, func(d *graph.Def) graph.DefFormatter { return dockerfileFormatter{d} })
}

type dockerfileFormatter struct {
	d *graph.Def
}

func (f dockerfileFormatter) Name(qual graph.Qualification) string {
	// TODO(sqs): use data stored
	switch qual {
	case graph.Unqualified:
		return f.d.Name
	case graph.ScopeQualified:
		return f.d.Name
	case graph.DepQualified:
		return fmt.Sprintf("%s %s", filepath.Base(level(f.d.File)), f.d.Name)
	case graph.RepositoryWideQualified:
		return fmt.Sprintf("%s %s", level(f.d.File), f.d.Name)
	case graph.LanguageWideQualified:
		return fmt.Sprintf("%s %s", filepath.Join(f.d.Repo, level(f.d.File)), f.d.Name)
	}
	panic("Name: unrecognized Qualification: " + fmt.Sprint(qual))
}

func level(file string) string {
	dir := filepath.Dir(file)
	if dir == "." {
		return "Root"
	}
	return dir
}

func (_ dockerfileFormatter) Type(qual graph.Qualification) string {
	switch qual {
	case graph.Unqualified:
		return ""
	case graph.ScopeQualified:
		return ""
	case graph.DepQualified:
		return ""
	case graph.RepositoryWideQualified:
		return ""
	case graph.LanguageWideQualified:
		return ""
	}
	panic("Type: unrecognized Qualification: " + fmt.Sprint(qual))
}

func (_ dockerfileFormatter) Language() string             { return "Dockerfile" }
func (_ dockerfileFormatter) DefKeyword() string           { return "" }
func (_ dockerfileFormatter) NameAndTypeSeparator() string { return "" }
func (f dockerfileFormatter) Kind() string                 { return "Dockerfile" }
