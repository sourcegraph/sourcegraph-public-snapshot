package ui

import (
	"encoding/json"
	"html/template"
	"net/http"

	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sourcecode"
	"src.sourcegraph.com/sourcegraph/ui/payloads"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/htmlutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveDef(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)
	e := json.NewEncoder(w)

	dc, rc, vc, err := handlerutil.GetDefCommon(r, nil)
	if err != nil {
		if urlErr, ok := err.(*handlerutil.URLMovedError); ok {
			return e.Encode(urlErr)
		} else if dc != nil && dc.Def != nil {
			// Create a fake minimal definition based on def spec in request. This is the best we can do
			// here since the actual def hasn't been indexed and isn't available.
			// TODO: This is hacky and should be refactored.
			defKey := dc.Def.DefKey

			// TODO: Refactor to reuse of sourcecode.DefQualifiedName(def, "scope") for rendering the fake minimal definition.
			qualifiedName := template.HTML(template.HTMLEscapeString(defKey.Unit))
			if defKey.Path != "" && defKey.Path != "." {
				if qualifiedName != "" {
					qualifiedName += "."
				}
				qualifiedName += `<span class="name">` + template.HTML(template.HTMLEscapeString(defKey.Path)) + "</span>"
			}
			qualifiedName = sourcecode.OverrideStyleViaRegexpFlags(qualifiedName)
			return e.Encode(payloads.DefCommon{
				QualifiedName: htmlutil.SanitizeForPB(string(qualifiedName)),
				URL:           router.Rel.URLToDef(defKey).String(),
				Found:         false,
			})
		}
		return err
	}

	def := dc.Def
	entrySpec := sourcegraph.TreeEntrySpec{RepoRev: vc.RepoRevSpec, Path: def.File}
	qualifiedName := sourcecode.DefQualifiedNameAndType(def, "scope")
	qualifiedName = sourcecode.OverrideStyleViaRegexpFlags(qualifiedName)
	d := payloads.DefCommon{
		Def:               def,
		QualifiedName:     htmlutil.SanitizeForPB(string(qualifiedName)),
		URL:               router.Rel.URLToDefAtRev(def.DefKey, vc.RepoRevSpec.Rev).String(),
		File:              entrySpec,
		ByteStartPosition: def.DefStart,
		ByteEndPosition:   def.DefEnd,
		Found:             true,
	}

	if r.Header.Get("X-Definition-Data-Only") != "yes" {
		// This is not a request for definition data only (ie. for the pop-up),
		// but also for the file containing it (ie. navigating to a definition in a
		// different file).
		entry, err := cl.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{Entry: entrySpec, Opt: &sourcegraph.RepoTreeGetOptions{
			TokenizedSource: sourcecode.IsLikelyCodeFile(entrySpec.Path),
		}})

		if err != nil {
			return err
		}
		if entry.Type == vcsclient.DirEntry {
			return e.Encode(&handlerutil.URLMovedError{NewURL: d.URL})
		}
		return e.Encode(&struct {
			*payloads.CodeFile
			Model *payloads.DefCommon
		}{
			CodeFile: &payloads.CodeFile{
				Repo:              rc.Repo,
				RepoCommit:        vc.RepoCommit,
				EntrySpec:         entrySpec,
				SrclibDataVersion: &sourcegraph.SrclibDataVersion{CommitID: vc.RepoRevSpec.CommitID},
				Entry:             entry,
			},
			Model: &d,
		})
	}
	return e.Encode(d)
}
