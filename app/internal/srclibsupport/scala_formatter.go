package srclibsupport

import (
	"encoding/json"
	"fmt"
	"log"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

const scalaSourceUnitType = "scala"

func init() {
	graph.RegisterMakeDefFormatter(scalaSourceUnitType, func(d *graph.Def) graph.DefFormatter {
		var data struct {
			FmtStrings graph.DefFormatStrings
		}
		if err := json.Unmarshal(d.Data, &data); err != nil {
			log.Printf("Scala def formatter JSON error: %v (JSON is: %s)", err, d.Data)
		}
		return scalaFormatter{d, data.FmtStrings}
	})
}

type scalaFormatter struct {
	def *graph.Def
	fs  graph.DefFormatStrings
}

// TODO(sqs): The actual implementation of these methods will depend
// on what data about each def is produced by the grapher and how it's
// stored. Once I see what info the Scala compiler outputs, I can
// update these to produce more useful formatted output.

func (f scalaFormatter) Name(qual graph.Qualification) string {
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

func (f scalaFormatter) Type(qual graph.Qualification) string {
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

func (f scalaFormatter) Language() string             { return f.fs.Language }
func (f scalaFormatter) DefKeyword() string           { return f.fs.DefKeyword }
func (f scalaFormatter) NameAndTypeSeparator() string { return f.fs.NameAndTypeSeparator }
func (f scalaFormatter) Kind() string                 { return f.fs.Kind }
