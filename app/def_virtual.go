package app

import (
	"bytes"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"text/template"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sourcecode"
	"src.sourcegraph.com/sourcegraph/ui/payloads"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
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

func serveDefVirtual(w http.ResponseWriter, r *http.Request, dc *payloads.DefCommon, rc *handlerutil.RepoCommon, vc *handlerutil.RepoRevCommon) error {
	tc, err := virtualTreeEntry(dc.Def, vc.RepoRevSpec)
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

	entrySpec := sourcegraph.TreeEntrySpec{
		RepoRev: rev,
		Path:    def.File,
	}

	entry0 := &vcsclient.FileWithRange{
		TreeEntry: &vcsclient.TreeEntry{
			Name:     filepath.Base(def.File),
			Type:     vcsclient.FileEntry,
			Size_:    int64(len(rawContents)),
			Contents: []byte(rawContents),
		},
		FileRange: vcsclient.FileRange{
			StartLine: 0,
			EndLine:   int64(strings.Count(rawContents, "\n")),
			StartByte: 0,
			EndByte:   int64(len(rawContents)),
		},
	}

	sourceCode, err := parseVirtual(def, entrySpec, entry0)
	if err != nil {
		return nil, err
	}

	entry := &sourcegraph.TreeEntry{
		TreeEntry:  entry0.TreeEntry,
		FileRange:  &entry0.FileRange,
		SourceCode: sourceCode,
	}

	return &handlerutil.TreeEntryCommon{
		EntrySpec: entrySpec,
		Entry:     entry,
	}, nil
}

// parseVirtual returns the parsed tokenized representation of the virtual source code. This closely mirrors what
// sourcecode.Parse returns, but for the fake source code that's generated for virtual defs. It is mostly copied and
// pasted from sourcecode.Parse.
func parseVirtual(def *sourcegraph.Def, entrySpec sourcegraph.TreeEntrySpec, entry *vcsclient.FileWithRange) (*sourcegraph.SourceCode, error) {
	sourceCode := sourcecode.Tokenize(entry)

	refs := virtualEntryRefs(def, entrySpec, entry)
	for _, r := range refs {
		var defURL *url.URL
		if graph.URIEqual(entrySpec.RepoRev.URI, r.DefKey().Repo) {
			defURL = router.Rel.URLToDefAtRev(r.DefKey(), entrySpec.RepoRev.CommitID)
		} else {
			defURL = router.Rel.URLToDef(r.DefKey())
		}

		for _, line := range sourceCode.Lines {
			if r.Start >= uint32(line.StartByte) && r.Start <= uint32(line.EndByte) {
				for k, tok := range line.Tokens {
					start, end := uint32(tok.StartByte), uint32(tok.EndByte)
					if (r.Start >= start && r.Start < end) ||
						(r.End > end && r.Start < start) ||
						(r.End > start && r.End <= end) {
						if tok.URL == nil {
							tok.URL = make([]string, 0, 1)
						}
						tok.URL = append(tok.URL, defURL.String())
						tok.IsDef = r.Def
						line.Tokens[k] = tok
					}
				}
			}
		}
	}

	numRefs := len(refs)
	sourceCode.TooManyRefs = false
	sourceCode.NumRefs = int32(numRefs)

	return sourceCode, nil
}

// virtualEntryRefs returns fake refs for the fake source code generated for a virtual def.
func virtualEntryRefs(def *sourcegraph.Def, entrySpec sourcegraph.TreeEntrySpec, entry *vcsclient.FileWithRange) []*graph.Ref {
	var refs []*graph.Ref
	s := string(entry.Contents)
	for seen, i := 0, strings.Index(s, def.Path); i >= 0; i = strings.Index(s, def.Path) {
		j := i + len(def.Path)

		refs = append(refs, &graph.Ref{
			DefRepo:     def.Repo,
			DefUnitType: def.UnitType,
			DefUnit:     def.Unit,
			DefPath:     def.Path,
			Repo:        def.Repo,
			CommitID:    def.CommitID,
			UnitType:    def.UnitType,
			Unit:        def.Unit,
			Def:         true,
			File:        def.File,
			Start:       uint32(seen + i),
			End:         uint32(seen + j),
		})

		seen += j
		s = s[j:]
	}

	return refs
}
