package httpapi

import (
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

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
		resp.Def.Repo = path.Join(uri.Host, uri.Path)
		// TODO we currently do not set DocHTML
		resp.Def.DocHTML = htmlutil.EmptyHTML()
	}

	w.Header().Set("cache-control", "private, max-age=60")
	return writeJSON(w, resp)
}
