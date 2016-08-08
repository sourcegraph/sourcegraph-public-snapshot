package httpapi

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sqs/pbtypes"
)

var lpClient *langp.Client

func init() {
	if !feature.Features.Universe {
		return
	}
	var err error
	lpClient, err = langp.NewClient(os.Getenv("SG_LANGUAGE_PROCESSOR"))
	if err != nil {
		log.Fatal("$SG_LANGUAGE_PROCESSOR", err)
	}
}

func serveRepoHoverInfo(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	repo, repoRev, err := handlerutil.GetRepoAndRev(ctx, mux.Vars(r))
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
		Title string
		Def   *sourcegraph.Def `json:"def"`
	}{}

	if feature.IsUniverseRepo(repo.URI) {
		hover, err := lpClient.Hover(&langp.Position{
			Repo:      repo.URI,
			Commit:    repoRev.CommitID,
			File:      file,
			Line:      line,
			Character: character,
		})
		if err != nil {
			return err
		}
		if len(hover.Contents) > 0 { // TODO: We don't handle this case in the frontend.
			resp.Title = hover.Contents[0].Value
			desc := ""
			for _, content := range hover.Contents[1:] {
				desc += fmt.Sprintf("%s<br>", template.HTMLEscapeString(content.Value))
			}
			// Fake the definition.
			resp.Def = &sourcegraph.Def{
				Def: graph.Def{
					DefKey: graph.DefKey{Repo: repo.URI},
				},
				DocHTML: &pbtypes.HTML{HTML: desc},
			}
		}

		// TODO: We don't handle the case of no contents in the frontend from
		// an error handling perspective, so this is here.
		if len(hover.Contents) == 0 {
			resp.Def = &sourcegraph.Def{
				Def: graph.Def{
					DefKey: graph.DefKey{Repo: repo.URI},
				},
			}
		}
		return writeJSON(w, resp)
	}

	defSpec, err := cl.Annotations.GetDefAtPos(ctx, &sourcegraph.AnnotationsGetDefAtPosOptions{
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

	resp.Def, err = cl.Defs.Get(ctx, &sourcegraph.DefsGetOp{
		Def: *defSpec,
		Opt: &sourcegraph.DefGetOptions{
			Doc: true,
		},
	})
	if err != nil {
		return err
	}

	return writeJSON(w, resp)
}
