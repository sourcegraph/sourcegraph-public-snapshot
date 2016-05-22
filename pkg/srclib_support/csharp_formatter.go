package srclib_support

import (
	"encoding/json"
	"fmt"
	"log"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

const csharpSourceUnitType = "NugetPackage"

func init() {
	graph.RegisterMakeDefFormatter(csharpSourceUnitType, func(d *graph.Def) graph.DefFormatter {
		var data struct {
			FmtStrings graph.DefFormatStrings
		}
		if err := json.Unmarshal(d.Data, &data); err != nil {
			log.Printf("C# def formatter JSON error: %v (JSON is: %s)", err, d.Data)
		}
		return csharpFormatter{d, data.FmtStrings}
	})
}

type csharpFormatter struct {
	def *graph.Def
	fs  graph.DefFormatStrings
}

func (f csharpFormatter) Name(qual graph.Qualification) string {
	switch qual {
	case graph.Unqualified:
		return f.fs.Name.Unqualified
	case graph.ScopeQualified:
		return f.fs.Name.ScopeQualified
	case graph.DepQualified:
		return f.fs.Name.DepQualified
	case graph.RepositoryWideQualified:
		return f.fs.Name.RepositoryWideQualified
	case graph.LanguageWideQualified:
		return f.fs.Name.LanguageWideQualified
	}
	log.Println("Name: unrecognized Qualification: " + fmt.Sprint(qual))
	return ""
}

func (f csharpFormatter) Type(qual graph.Qualification) string {
	switch qual {
	case graph.Unqualified:
		return f.fs.Type.Unqualified
	case graph.ScopeQualified:
		return f.fs.Type.ScopeQualified
	case graph.DepQualified:
		return f.fs.Type.DepQualified
	case graph.RepositoryWideQualified:
		return f.fs.Type.RepositoryWideQualified
	case graph.LanguageWideQualified:
		return f.fs.Type.LanguageWideQualified
	}
	log.Println("Type: unrecognized Qualification: " + fmt.Sprint(qual))
	return ""
}

func (f csharpFormatter) Language() string             { return f.fs.Language }
func (f csharpFormatter) DefKeyword() string           { return f.fs.DefKeyword }
func (f csharpFormatter) NameAndTypeSeparator() string { return f.fs.NameAndTypeSeparator }
func (f csharpFormatter) Kind() string                 { return f.fs.Kind }
