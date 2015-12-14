package app

import (
	"bytes"
	"encoding/json"
	"html/template"
	htmpl "html/template"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/sourcegraph/mux"

	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveSourceboxDef(w http.ResponseWriter, r *http.Request) error {
	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	dc, bc, _, _, err := handlerutil.GetDefCommon(r, nil)
	if err != nil {
		// Avoid writing a full response, or else the sourcebox will mess with the surrounding page it's embedded in.
		http.Error(w, "", errcode.HTTP(err))
		return nil
	}

	entrySpec := sourcegraph.TreeEntrySpec{RepoRev: bc.BestRevSpec, Path: dc.Def.File}
	opt := sourcegraph.RepoTreeGetOptions{
		TokenizedSource: true,
		GetFileOptions: vcsclient.GetFileOptions{
			FileRange: vcsclient.FileRange{
				StartByte: int64(dc.Def.DefStart), EndByte: int64(dc.Def.DefEnd),
			},
			FullLines: true,
		},
	}

	entry, err := apiclient.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{Entry: entrySpec, Opt: &opt})
	if err != nil {
		return err
	}

	return renderSourcebox(r, w, entry, entrySpec, bc.RepoBuildInfo, dc.Def)
}

func serveSourceboxFile(w http.ResponseWriter, r *http.Request) error {
	opt := sourcegraph.RepoTreeGetOptions{
		TokenizedSource: true,
	}
	tc, _, _, bc, err := handlerutil.GetTreeEntryCommon(r, &opt)
	if err != nil {
		return err
	}

	return renderSourcebox(r, w, tc.Entry, tc.EntrySpec, bc.RepoBuildInfo, nil)
}

func renderSourcebox(r *http.Request, w http.ResponseWriter, entry *sourcegraph.TreeEntry, entrySpec sourcegraph.TreeEntrySpec, buildInfo *sourcegraph.RepoBuildInfo, def *sourcegraph.Def) error {
	ctx := httpctx.FromRequest(r)

	footerStyle := r.FormValue("FooterStyle")
	if footerStyle == "" {
		footerStyle = "full"
	}

	startLine := int(entry.StartLine)
	endLine := int(entry.EndLine)
	// User can specify the start and end lines via URL params.
	if len(r.Form["StartLine"]) > 0 {
		startLine, _ = strconv.Atoi(r.Form["StartLine"][0])
	}
	if len(r.Form["EndLine"]) > 0 {
		endLine, _ = strconv.Atoi(r.Form["EndLine"][0])
		// Shave off the extra line to make the number non-inclusive
		endLine -= 1
	}

	// Render sourcebox.
	var data = struct {
		Lines       []*sourcegraph.SourceCodeLine
		EntrySpec   sourcegraph.TreeEntrySpec
		Def         *sourcegraph.Def
		AppURL      *url.URL
		FooterStyle string
		StartLine   int
		EndLine     int
		LineNumbers bool
	}{
		Lines:       entry.SourceCode.Lines,
		EntrySpec:   entrySpec,
		Def:         def,
		AppURL:      conf.AppURL(ctx),
		FooterStyle: footerStyle,
		StartLine:   startLine,
		EndLine:     endLine,
		LineNumbers: len(r.Form["LineNumbers"]) > 0,
	}

	var buf bytes.Buffer
	if err := tmpl.Get("sourcebox/sourcebox.html").Execute(&buf, &data); err != nil {
		return err
	}

	// Caching
	if buildInfo != nil && buildInfo.LastSuccessful != nil && buildInfo.LastSuccessful.EndedAt != nil {
		w.Header().Set("last-modified", buildInfo.LastSuccessful.EndedAt.Time().Round(time.Second).Format(http.TimeFormat))
		// TODO(sqs): check client if-modified-since and short-circuit sooner to take better advantage of caching
		w.Header().Set("cache-control", "private, max-age=600")
	}

	sb := &sourcegraph.Sourcebox{
		HTML:          template.HTML(buf.String()),
		StylesheetURL: assetAbsURL(ctx, "sourcebox.css").String(),
		ScriptURL:     assetAbsURL(ctx, "sourcebox.js").String(),
	}

	// Render either JS (document.write(...)) or JSON.
	switch mux.Vars(r)["Format"] {
	case "js":
		// JS-escape the HTML string so we can insert it into the
		// template as a JS expr.
		//escapedHTML := htmpl.JS(htmpl.JSEscapeString(string(sb.HTML)))
		escapedHTML, err := json.Marshal(string(sb.HTML))
		if err != nil {
			return err
		}

		w.Header().Set("content-type", "text/javascript; charset=utf-8")
		return tmpl.Get("sourcebox/sourcebox.js").Execute(w, &struct {
			*sourcegraph.Sourcebox

			// EscapedHTML is template.HTML not template.JS because
			// the template package doesn't know that sourcebox.js is
			// a JS file (it assumes the top-level is HTML unless it
			// sees a <script> tag).
			EscapedHTML htmpl.HTML
		}{
			Sourcebox:   sb,
			EscapedHTML: htmpl.HTML(escapedHTML),
		})
	case "json":
		w.Header().Set("Access-Control-Allow-Origin", "*")
		return httputil.WriteJSON(w, sb)
	default:
		http.Error(w, "", http.StatusNotFound)
	}
	return nil
}
