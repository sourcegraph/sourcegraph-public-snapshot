package httpapi

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/sourcegraph-go/pkg/lsp"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/htmlutil"
)

func serveRepoHoverInfo(w http.ResponseWriter, r *http.Request) error {
	repo, repoRev, err := handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
	if err != nil {
		return err
	}

	file := r.URL.Query().Get("file")

	line, err := strconv.Atoi(r.URL.Query().Get("line"))
	if err != nil {
		return err
	}

	character, err := strconv.Atoi(r.URL.Query().Get("character"))
	if err != nil {
		return err
	}

	var resp = &struct {
		Title      string
		Def        *sourcegraph.Def `json:"def"`
		Unresolved bool
	}{}

	// TODO deprecate this endpoint. For now we send a request to our
	// xlang endpoint.
	var hover lsp.Hover
	err = textDocumentPositionRequest{
		Method:    "textDocument/hover",
		Repo:      repo.URI,
		Commit:    repoRev.CommitID,
		File:      file,
		Line:      line,
		Character: character,
	}.Serve(r.Context(), &hover)
	if err != nil {
		return err
	}
	resp.Unresolved = true
	if len(hover.Contents) >= 1 {
		resp.Unresolved = false
		resp.Title = hover.Contents[0].Value
		// Cleanup title to not be as long. Go specific
		if i := strings.Index(resp.Title, "{"); i > 0 {
			resp.Title = strings.TrimSpace(resp.Title[:i])
		}
	}

	// We also need to do a definition so we can say which repo the def is
	// defined in.
	var locations []lsp.Location
	err = textDocumentPositionRequest{
		Method:    "textDocument/definition",
		Repo:      repo.URI,
		Commit:    repoRev.CommitID,
		File:      file,
		Line:      line,
		Character: character,
	}.Serve(r.Context(), &locations)
	if err == nil && len(locations) >= 1 {
		uri, err := url.Parse(locations[0].URI)
		if err != nil {
			return err
		}
		resp.Def = &sourcegraph.Def{}

		// HACK: The frontend relies on this information to redirect to def
		// landing pages. This will fail for custom import paths, and maybe
		// other cases too.
		unit := path.Join(uri.Host, uri.Path, path.Dir(uri.Fragment))
		resp.Def.Unit = unit
		resp.Def.UnitType = "GoPackage"
		resp.Def.File = "none"
		resp.Def.Kind = "none"
		resp.Def.Path = hackParseDefKeyPath(resp.Title)
		if resp.Def.Path == "" {
			// Note: can't 404 here because it would break the chrome ext in most
			// cases that don't have a valid resp.Def.Path -- luckily that ext
			// doesn't rely on it.
			log15.Crit(fmt.Sprintf("hackParseDefKeyPath failed to parse %q", resp.Title))
		}

		resp.Def.Repo = path.Join(uri.Host, uri.Path)
		// TODO we currently do not set DocHTML
		resp.Def.DocHTML = htmlutil.EmptyHTML()
	}

	w.Header().Set("cache-control", "private, max-age=60")
	return writeJSON(w, resp)
}

func hackParseDefKeyPath(s string) string {
	fields := strings.Fields(s)
	switch {
	case len(fields) >= 2 && fields[0] == "func" && strings.Contains(fields[1], "(") && strings.Contains(fields[1], "."):
		// "func (*SomeStruct).Foobar(..."
		fields := strings.Split(fields[1], "(") // ["" "*SomeStruct).Foobar"]
		if len(fields) < 2 {
			return ""
		}
		fields = strings.Split(fields[1], ")") // ["*Router" ".Match"]
		typ := strings.TrimPrefix(fields[0], "*")
		method := strings.TrimPrefix(fields[1], ".")
		return typ + "/" + method
	case len(fields) >= 2 && fields[0] == "func" && strings.Contains(fields[1], "("):
		// "func Foobar(..."
		return strings.Split(fields[1], "(")[0]
	case len(fields) >= 2 && fields[0] == "type":
		// "type Foobar {..."
		return fields[1]
	}
	return ""
}
