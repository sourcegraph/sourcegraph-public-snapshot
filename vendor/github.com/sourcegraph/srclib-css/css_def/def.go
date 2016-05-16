package css_def

import (
	"encoding/json"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

func init() {
	graph.RegisterMakeDefFormatter("Dir", newDefFormatter)
}

// DefData should be kept in sync with the def 'Data' field emitted by the
// basic grapher.
type DefData struct {
	Name      string
	Keyword   string
	Type      string
	Kind      string
	Separator string
}

func newDefFormatter(s *graph.Def) graph.DefFormatter {
	var si DefData
	if len(s.Data) > 0 {
		if err := json.Unmarshal(s.Data, &si); err != nil {
			panic("unmarshal CSS def data: " + err.Error())
		}
	}
	return defFormatter{s, &si}
}

type defFormatter struct {
	def  *graph.Def
	data *DefData
}

func (f defFormatter) Language() string { return "CSS" }

func (f defFormatter) DefKeyword() string { return f.data.Keyword }

func (f defFormatter) Kind() string { return f.data.Kind }

func (f defFormatter) NameAndTypeSeparator() string { return f.data.Separator }

func (f defFormatter) Type(qual graph.Qualification) string { return f.data.Type }

func (f defFormatter) Name(qual graph.Qualification) string { return f.def.Name }
