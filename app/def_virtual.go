package app

import (
	"bytes"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

// isVirtual returns true if the definition does not exist anywhere in
// source code. This signals to the UI that it should not try to fetch
// the source file and should instead display an artificial page with
// the docstring and examples of the definition. Examples of virtual
// definitions include definitions in vendored JARs for which the
// source repository cannot be inferred and definitions found in
// auto-generated source files that aren't checked into the repository
// (e.g., Thrift definitions).
//
// This is a stopgap until we have better support for resolving
// external JARs and auto-generated source files.
//
// Use of this function should be limited to the app and ui packages.
// The rest of the code should not treat virtual definitions any
// differently from normal definitions.
func isVirtual(def graph.DefKey) bool {
	if strings.HasPrefix(def.Path, "__virtual__/") {
		return true
	}
	return false
}

func serveDefVirtual(w http.ResponseWriter, r *http.Request, dc *sourcegraph.Def, rc *handlerutil.RepoCommon, vc *handlerutil.RepoRevCommon) error {
	tc, err := virtualTreeEntry(dc, vc.RepoRevSpec)
	if err != nil {
		return err
	}
	return serveRepoTreeEntry(w, r, tc, rc, vc, dc)
}

var virtualFileTemplate = template.Must(template.New("").Parse(`
// This is an auto-generated file for a definition
// that does not exist in source code.
{{.Path}}
`))

// virtualTreeEntry returns the fake source code for a virtual def.
func virtualTreeEntry(def *sourcegraph.Def, rev sourcegraph.RepoRevSpec) (*handlerutil.TreeEntryCommon, error) {

	var buf bytes.Buffer
	err := virtualFileTemplate.Execute(&buf, def)
	if err != nil {
		return nil, err
	}
	rawContents := buf.String()

	return &handlerutil.TreeEntryCommon{
		EntrySpec: sourcegraph.TreeEntrySpec{
			RepoRev: rev,
			Path:    def.File,
		},
		Entry: &sourcegraph.TreeEntry{
			BasicTreeEntry: &sourcegraph.BasicTreeEntry{
				Name:     filepath.Base(def.File),
				Type:     sourcegraph.FileEntry,
				Contents: []byte(rawContents),
			},
			FileRange: &sourcegraph.FileRange{
				StartLine: 0,
				EndLine:   int64(strings.Count(rawContents, "\n")),
				StartByte: 0,
				EndByte:   int64(len(rawContents)),
			},
			ContentsString: rawContents,
		},
	}, nil
}
