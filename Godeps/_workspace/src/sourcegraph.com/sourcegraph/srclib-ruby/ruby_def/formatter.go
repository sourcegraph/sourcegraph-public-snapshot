package ruby_def

import (
	"encoding/json"
	"strings"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

func init() {
	graph.RegisterMakeDefFormatter("rubygem", newDefFormatter)
	graph.RegisterMakeDefFormatter("ruby", newDefFormatter)
}

// DefData should be kept in sync with the def 'Data' field emitted by the Ruby
// grapher.
type DefData struct {
	RubyKind   string
	TypeString string
	Module     string
	RubyPath   string
	Signature  string
	ReturnType string
}

func (s *DefData) isLocalVar() bool {
	return strings.Contains(s.RubyPath, ">_local_")
}

func newDefFormatter(s *graph.Def) graph.DefFormatter {
	var si DefData
	if len(s.Data) > 0 {
		if err := json.Unmarshal(s.Data, &si); err != nil {
			panic("unmarshal Ruby def data: " + err.Error())
		}
	}
	return defFormatter{s, &si}
}

type defFormatter struct {
	def  *graph.Def
	data *DefData
}

func (f defFormatter) Language() string { return "JavaScript" }

func (f defFormatter) DefKeyword() string {
	switch f.data.RubyKind {
	case "method":
		return "def"
	case "class", "module":
		return f.data.RubyKind
	}
	return ""
}

func (f defFormatter) Kind() string { return f.data.RubyKind }

func (f defFormatter) Name(qual graph.Qualification) string {
	if f.data.isLocalVar() {
		return f.def.Name
	}

	switch qual {
	case graph.Unqualified:
		return f.def.Name
	case graph.ScopeQualified:
		return f.data.RubyPath
	case graph.DepQualified:
		return f.data.RubyPath
	case graph.RepositoryWideQualified:
		return f.data.RubyPath
	case graph.LanguageWideQualified:
		return f.data.RubyPath
	}
	panic("Name: unhandled qual " + string(qual))
}

func (f defFormatter) NameAndTypeSeparator() string {
	if f.data.RubyKind == "method" {
		return ""
	}
	return " "
}

func (f defFormatter) Type(qual graph.Qualification) string {
	var ts string
	if f.data.RubyKind == "method" {
		if i := strings.Index(f.data.Signature, "("); i != -1 {
			ts = f.data.Signature[i:]
		}
		ts += " " + cleanType(f.data.ReturnType)
	} else {
		ts = cleanType(f.data.TypeString)
	}
	return strings.TrimPrefix(ts, "::")
}

func cleanType(t string) string {
	t = strings.TrimSuffix(t, "#")
	switch t {
	case "NilClass":
		return "nil"
	case "TrueClass":
		return "true"
	case "FalseClass":
		return "false"
	}
	return t
}
