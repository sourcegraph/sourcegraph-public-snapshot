package httpapi

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/universe"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
)

func serveJumpToDef(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.Client(r)

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

	var response lsp.Location

	if universe.EnabledFile(file) {
		defRange, err := langp.DefaultClient.Definition(r.Context(), &langp.Position{
			Repo:      repo.URI,
			Commit:    repoRev.CommitID,
			File:      file,
			Line:      line,
			Character: character,
		})
		universeObserve("Definition", err)
		if err != nil {
			return err
		}

		// Don't add an ugly commit ID to the URL in the address bar if
		// the user is on a named branch or default branch. Also, never
		// add a rev if the jump is cross-repo.
		var rev string
		if v := routevar.ToRepoRev(mux.Vars(r)); defRange.Repo == v.Repo {
			rev = v.Rev
		}

		if defRange.Empty() {
			// Nothing to do.
		} else if defRange.File == "." {
			// Special case the top level directory
			response.URI = makeLSPURI(defRange.Repo, rev, "")
		} else {
			// We increment the line number by 1 because the blob view is not zero-indexed.
			response.URI = makeLSPURI(defRange.Repo, rev, defRange.File)
			response.Range = lsp.Range{
				Start: lsp.Position{Line: defRange.StartLine, Character: defRange.StartCharacter},
				End:   lsp.Position{Line: defRange.EndLine, Character: defRange.EndCharacter},
			}
		}
		w.Header().Set("cache-control", "private, max-age=60")
		return writeJSON(w, response)
	}

	defSpec, err := cl.Annotations.GetDefAtPos(r.Context(), &sourcegraph.AnnotationsGetDefAtPosOptions{
		Entry: sourcegraph.TreeEntrySpec{
			RepoRev: repoRev,
			Path:    file,
		},
		Line:      uint32(line),
		Character: uint32(character),
	})
	if err != nil {
		return err
	}
	def, err := cl.Defs.Get(r.Context(),
		&sourcegraph.DefsGetOp{
			Def: sourcegraph.DefSpec{
				Repo:     defSpec.Repo,
				CommitID: defSpec.CommitID,
				UnitType: defSpec.UnitType,
				Unit:     defSpec.Unit,
				Path:     defSpec.Path},
			Opt: &sourcegraph.DefGetOptions{ComputeLineRange: true}})
	if err != nil {
		return err
	}

	// Don't add an ugly commit ID to the URL in the address bar if
	// the user is on a named branch or default branch. Also, never
	// add a rev if the jump is cross-repo.
	var rev string
	if v := routevar.ToRepoRev(mux.Vars(r)); def.Repo == v.Repo {
		rev = v.Rev
	}

	response.URI = makeLSPURI(def.Repo, rev, def.File)
	response.Range = lsp.Range{
		Start: lsp.Position{Line: int(def.StartLine) - 1},
	}
	w.Header().Set("cache-control", "private, max-age=60")
	return writeJSON(w, response)
}

// makeLSPURI returns a file URI for the LSP response that Monaco on
// the frontend knows how to interpret.
func makeLSPURI(repo, rev, file string) string {
	return fmt.Sprintf("git://%s?%s#%s", repo, rev, file)
}
