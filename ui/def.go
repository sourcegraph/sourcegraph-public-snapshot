package ui

import (
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/sourcecode"
	"sourcegraph.com/sourcegraph/sourcegraph/ui/payloads"
	"sourcegraph.com/sourcegraph/sourcegraph/util/eventsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/htmlutil"
)

func serveDef(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	e := json.NewEncoder(w)

	dc, rc, vc, err := handlerutil.GetDefCommon(ctx, mux.Vars(r), nil)
	if err != nil {
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
	}

	if r.Header.Get("X-Definition-Data-Only") != "yes" {
		// This is not a request for definition data only (ie. for the pop-up),
		// but also for the file containing it (ie. navigating to a definition in a
		// different file).
		entry, err := cl.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{Entry: entrySpec, Opt: &sourcegraph.RepoTreeGetOptions{}})

		if err != nil {
			return err
		}

		eventsutil.LogViewDef(ctx, "GoToDefinition")
		if entry.Type == sourcegraph.DirEntry {
			return e.Encode(&handlerutil.URLMovedError{NewURL: d.URL})
		}

		entry.ContentsString = string(entry.Contents)
		entry.Contents = nil

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

	eventsutil.LogViewDef(ctx, "ViewDefPopup")
	return e.Encode(d)
}
