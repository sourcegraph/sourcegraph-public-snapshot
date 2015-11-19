package app

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"src.sourcegraph.com/sourcegraph/app/internal/schemautil"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/sourcecode"
	"src.sourcegraph.com/sourcegraph/ui/payloads"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
	"src.sourcegraph.com/sourcegraph/util/textutil"
)

// serveDef creates a new response for the code view that contains information
// about a definition. Additionally, it may also contain information about the
// tree entry that contains that definition.
func serveDef(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	dc, bc, rc, vc, err := handlerutil.GetDefCommon(r, &sourcegraph.DefGetOptions{Doc: true})
	if err != nil {
		return err
	}

	if isVirtual(dc.Def.DefKey) {
		return serveDefVirtual(w, r, dc, bc, rc, vc)
	}

	tc := &handlerutil.TreeEntryCommon{
		EntrySpec: sourcegraph.TreeEntrySpec{
			RepoRev: bc.BestRevSpec,
			Path:    dc.Def.File,
		},
	}
	tc.Entry, err = cl.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{Entry: tc.EntrySpec, Opt: &sourcegraph.RepoTreeGetOptions{
		TokenizedSource: sourcecode.IsLikelyCodeFile(tc.EntrySpec.Path),
	}})

	if err != nil {
		return err
	}

	return serveRepoTreeEntry(w, r, tc, rc, vc, bc, dc)
}

func serveDefExamples(w http.ResponseWriter, r *http.Request) error {
	var opt sourcegraph.DefListExamplesOptions
	err := schemautil.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}
	opt.Formatted = true

	ctx := httpctx.FromRequest(r)
	apiclient := handlerutil.APIClient(r)

	dc, bc, rc, vc, err := handlerutil.GetDefCommon(r, nil)
	if err != nil {
		return err
	}

	// Get actual list of examples
	examples, err := apiclient.Defs.ListExamples(ctx, &sourcegraph.DefsListExamplesOp{Def: dc.Def.DefSpec(), Opt: &opt})
	if err != nil {
		return err
	}

	// Highlight this def in examples.
	u0 := router.Rel.URLToDefAtRev(dc.Def.DefKey, bc.BestRevSpec.CommitID) // internal
	u1 := router.Rel.URLToDefAtRev(dc.Def.DefKey, "")                      // external
	for _, x := range examples.Examples {
		x.SrcHTML = strings.Replace(string(x.SrcHTML), u0.String()+`" class="`, u0.String()+`" class="highlight highlight-primary `, -1)
		x.SrcHTML = strings.Replace(string(x.SrcHTML), u1.String()+`" class="`, u1.String()+`" class="highlight highlight-primary `, -1)
		x.SrcHTML = strings.Replace(string(x.SrcHTML), "class=\"", "class=\"defn-popover ", -1)
	}

	pg, err := paginatePrevNext(opt, examples.StreamResponse)
	if err != nil {
		return err
	}

	return tmpl.Exec(r, w, "def/examples.html", http.StatusOK, nil, &struct {
		handlerutil.RepoCommon
		handlerutil.RepoRevCommon
		payloads.DefCommon
		Examples  []*sourcegraph.Example
		PageLinks []pageLink
		Options   sourcegraph.DefListExamplesOptions
		tmpl.Common
	}{
		RepoCommon:    *rc,
		RepoRevCommon: *vc,
		DefCommon:     *dc,
		Examples:      examples.Examples,
		PageLinks:     pg,
		Options:       opt,
	})
}

func serveDefPopover(w http.ResponseWriter, r *http.Request) error {
	// viewState holds the repository and file that the user is currently
	// viewing.
	type viewState struct {
		CurrentRepo string
		CurrentFile string
	}
	var opt viewState
	err := schemautil.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}
	dc, _, _, _, err := handlerutil.GetDefCommon(r, &sourcegraph.DefGetOptions{Doc: true})
	if err != nil {
		// TODO(gbbr): Set up custom responses for each scenario.
		// All of the below errors will cause full page HTML pages or redirects, if
		// bubbled up the chain, so we return nil instead.
		// Temporarily StatusNotFound with empty body will be returned.
		switch e := err.(type) {
		case *handlerutil.URLMovedError, *handlerutil.RepoNotEnabledError,
			*handlerutil.NoVCSDataError, *handlerutil.NoBuildError:
			http.Error(w, fmt.Sprintf("Not found (%#v)", e), http.StatusNotFound)
			return nil
		}
		return err
	}

	hdr := http.Header{
		"access-control-allow-origin":  []string{"*"},
		"access-control-allow-methods": []string{"GET"},
	}

	return tmpl.Exec(r, w, "def/popover.html", http.StatusOK, hdr, &struct {
		Def         *sourcegraph.Def
		CurrentRepo string
		CurrentFile string
		tmpl.Common
	}{
		Def:         dc.Def,
		CurrentRepo: opt.CurrentRepo,
		CurrentFile: opt.CurrentFile,
	})
}

func defMetaDescription(def *sourcegraph.Def) string {
	docText := strings.TrimSpace(textutil.TextFromHTML(string(def.DocHTML.HTML)))
	var desc string
	if docText == "" {
		desc += fmt.Sprintf("%s in %s", def.Fmt().Name(graph.ScopeQualified), repoBasename(def.Def.Repo))
	} else {
		desc += textutil.Truncate(300, collapseSpace(docText))
	}

	return desc
}

var spaceRE = regexp.MustCompile(`\s+`)

func collapseSpace(s string) string {
	return spaceRE.ReplaceAllString(s, " ")
}

// defRobotsIndex returns a boolean indicating whether the page
// corresponding to this def should be indexed by robots (e.g.,
// Googlebot).
func defRobotsIndex(repo *sourcegraph.Repo, def *sourcegraph.Def) bool {
	return !repo.Private && def.Exported && !def.Test
}
