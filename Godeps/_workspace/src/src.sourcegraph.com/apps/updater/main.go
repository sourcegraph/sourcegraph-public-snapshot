package updater

import (
	"fmt"
	"html"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/shurcooL/Go-Package-Store/pkg"
	"github.com/shurcooL/Go-Package-Store/presenter/github"
	"github.com/shurcooL/Go-Package-Store/repo"
	"github.com/shurcooL/go/gzip_file_server"
	"github.com/shurcooL/httpfs/html/vfstemplate"
	"github.com/sourcegraph/httpcache"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/platform/pctx"
	"src.sourcegraph.com/sourcegraph/platform/putil"
	"src.sourcegraph.com/sourcegraph/platform/storage"
)

// appName is the app name used for Sourcegraph platform storage.
const appName = "updater"

type app struct {
	http.Handler

	// appCtx is the app context with high priveldge. It's used to access the Sourcegraph platform storage
	// (on behalf of users that may not have write access). This service implementation is responsible for doing
	// authorization checks.
	appCtx context.Context

	gitHubRegister sync.Once
}

// TODO: Remove globals.
var (
	// updater is set based on the source of Go packages. If nil, it means
	// we don't have support to update Go packages from the current source.
	// It's used to update repos in the backend, and to disable the frontend UI
	// for updating packages.
	updater repo.Updater
)

// New returns an updater app http.Handler.
func New(ctx context.Context) http.Handler {
	err := loadTemplates()
	if err != nil {
		log.Fatalln("loadTemplates:", err)
	}

	app := app{
		appCtx: ctx,
	}

	h := http.NewServeMux()
	h.HandleFunc("/mock.html", mockHandler)
	h.HandleFunc("/debug.html", debugHandler)
	h.HandleFunc("/", app.mainHandler)
	assetsFileServer := gzip_file_server.New(Assets)
	assetsFileServer = passThrough{assetsFileServer}
	h.Handle("/assets/", assetsFileServer)

	app.Handler = h

	return app
}

var t *template.Template

func loadTemplates() error {
	var err error
	t = template.New("").Funcs(template.FuncMap{
		"updateSupported": func() bool { return updater != nil },
		"commitID":        func(commitID string) string { return commitID[:8] },
	})
	t, err = vfstemplate.ParseGlob(Assets, t, "/assets/*.tmpl")
	return err
}

// shouldPresentUpdate determines if the given goPackage should be presented as an available update.
// It checks that the Go package is on default branch, does not have a dirty working tree, and does not have the remote revision.
func shouldPresentUpdate(repo *pkg.Repo) bool {
	// TODO: Replicate the previous behavior fully, then remove this commented out code:
	//return status.PlumbingPresenterV2(goPackage)[:3] == "  +" // Ignore stash.

	if repo.RemoteURL == "" || repo.Local.Revision == "" || repo.Remote.Revision == "" {
		return false
	}

	if repo.VCS != nil {
		if b, err := repo.VCS.Branch(repo.Path); err != nil || b != repo.VCS.DefaultBranch() {
			return false
		}
		if s, err := repo.VCS.Status(repo.Path); err != nil || s != "" {
			return false
		}
		if c, err := repo.VCS.Contains(repo.Path, repo.Remote.Revision); err != nil || c {
			return false
		}
	}

	return repo.Local.Revision != repo.Remote.Revision
}

func getGodeps(ctx context.Context) (*sourcegraph.TreeEntry, error) {
	sg, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, err
	}
	repoRevSpec, _ := pctx.RepoRevSpec(ctx)
	return sg.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{
		Entry: sourcegraph.TreeEntrySpec{
			RepoRev: repoRevSpec,
			Path:    "/Godeps/Godeps.json",
		},
		Opt: &sourcegraph.RepoTreeGetOptions{
			GetFileOptions: sourcegraph.GetFileOptions{
				EntireFile: true,
			},
		},
	})
}

func (a *app) mainHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method should be GET", http.StatusMethodNotAllowed)
		return
	}

	if err := loadTemplates(); err != nil {
		fmt.Fprintln(w, "loadTemplates:", err)
		return
	}

	a.gitHubRegister.Do(func() {
		// Use appCtx for cache, so that users without write access can still update cache.
		kv := storage.Namespace(a.appCtx, appName, "") // THINK: Consider per-repo cache vs global cache.
		cache := kvCache{kv: kv}
		cacheTransport := httpcache.NewTransport(cache)

		// Optionally, perform GitHub API authentication with provided token.
		if token := os.Getenv("REPO_UPDATER_APP_GITHUB_TOKEN"); token != "" {
			authTransport := &oauth2.Transport{
				Source: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}),
			}
			cacheTransport.Transport = authTransport
		}

		github.SetClient(&http.Client{Transport: cacheTransport})
	})

	ctx := putil.Context(req)
	f, err := getGodeps(ctx)
	if grpc.Code(err) == codes.NotFound {
		fmt.Fprintf(w, `<h2 style="text-align: center;">%s</h2>`, html.EscapeString("No Godeps.json file."))
		return
	} else if err != nil {
		fmt.Fprintf(w, `<div>Error: %s</div>`, html.EscapeString(err.Error()))
		return
	}

	g, err := parseGodeps(f.Contents)
	if err != nil {
		panic(err)
	}

	pipeline := NewWorkspace()
	go func() { // This needs to happen in the background because sending input will be blocked on processing.
		for _, dependency := range g.Deps {
			pipeline.AddRevision(dependency.ImportPath, dependency.Rev)
		}
		pipeline.Done()
	}()
	updater = nil

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	data := struct {
		BaseURI string
	}{
		BaseURI: pctx.BaseURI(ctx),
	}
	err = t.ExecuteTemplate(w, "head.html.tmpl", data)
	if err != nil {
		log.Println("ExecuteTemplate head.html.tmpl:", err)
		return
	}

	updatesAvailable := 0

	for presented := range pipeline.Presented() {
		updatesAvailable++

		err := t.ExecuteTemplate(w, "repo.html.tmpl", presented)
		if err != nil {
			log.Println("ExecuteTemplate repo.html.tmpl:", err)
			return
		}
	}

	if updatesAvailable == 0 {
		io.WriteString(w, `<script>document.getElementById("no_updates").style.display = "";</script>`)
	}

	err = t.ExecuteTemplate(w, "tail.html.tmpl", nil)
	if err != nil {
		log.Println("ExecuteTemplate tail.html.tmpl:", err)
		return
	}
}

func debugHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method should be GET", http.StatusMethodNotAllowed)
		return
	}

	if err := loadTemplates(); err != nil {
		fmt.Fprintln(w, "loadTemplates:", err)
		return
	}

	ctx := putil.Context(req)

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	f, err := getGodeps(ctx)
	if grpc.Code(err) == codes.NotFound {
		fmt.Fprintf(w, `<div>%s</div>`, html.EscapeString("Godeps.json file doesn't exist, etc."))
		return
	} else if err != nil {
		fmt.Fprintf(w, `<div>Some error: %s</div>`, html.EscapeString(err.Error()))
		return
	}

	g, err := parseGodeps(f.Contents)
	if err != nil {
		panic(err)
	}
	g.Deps = g.Deps[:20] // HACK.

	err = t.ExecuteTemplate(w, "wip.html.tmpl", g)
	if err != nil {
		log.Println("t.Execute:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
