package app

import (
	"net/http"
	"net/url"
	pathpkg "path"
	"path/filepath"

	"github.com/sourcegraph/gddo/doc"
	"github.com/sourcegraph/gddo/gosrc"
	"github.com/sourcegraph/mux"
	"golang.org/x/net/context"

	"strings"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"src.sourcegraph.com/sourcegraph/app/internal/godocsupport"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func PkgGoDocURL(pkg string) (*url.URL, error) {
	var repo, subdir string
	if gosrc.IsGoRepoPath(pkg) {
		repo = "github.com/golang/go"
		subdir = pathpkg.Join("src", pkg)
	} else if strings.Count(pkg, "/") >= 2 {
		parts := strings.Split(pkg, "/")
		repo = strings.Join(parts[:3], "/")
		subdir = strings.Join(parts[3:], "/")
	}
	return router.Rel.URLToRepoGoDoc(repo, "", subdir)
}

func serveGoDoc(w http.ResponseWriter, r *http.Request) error {
	return tmpl.Exec(r, w, "godoc/home.html", http.StatusOK, nil, &struct{ tmpl.Common }{})
}

func serveRepoGoDoc(w http.ResponseWriter, r *http.Request) error {
	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	rc, vc, err := handlerutil.GetRepoAndRevCommon(r, nil)
	if err != nil {
		return err
	}

	if vc.RepoCommit == nil {
		return renderRepoNoVCSDataTemplate(w, r, rc)
	}

	bc, err := handlerutil.GetRepoBuildCommon(r, rc, vc, nil)
	if err != nil {
		return err
	}
	vc.RepoRevSpec = bc.BestRevSpec // Remove after getRepo refactor.

	v := mux.Vars(r)
	dir, err := getGodocDir(ctx, apiclient, rc.Repo, bc.BestRevSpec, pathpkg.Clean(v["Path"]))
	if err != nil {
		return err
	}

	pkg, err := doc.NewPackage(dir)
	if err != nil {
		return err
	}

	subpkgs := make([]*godocsupport.Package, len(dir.Subdirectories))
	for i, subdir := range dir.Subdirectories {
		subpkgs[i] = &godocsupport.Package{
			Path:     subdir,
			Synopsis: "",
		}
	}

	return tmpl.Exec(r, w, "repo/godoc.html", http.StatusOK, nil, &struct {
		handlerutil.RepoCommon
		handlerutil.RepoRevCommon
		handlerutil.RepoBuildCommon

		RepoIsBuilt bool

		Pkg     *doc.Package
		Subpkgs []*godocsupport.Package
		PDoc    *godocsupport.TDoc

		RobotsIndex bool
		tmpl.Common
	}{
		RepoCommon:      *rc,
		RepoRevCommon:   *vc,
		RepoBuildCommon: bc,

		RepoIsBuilt: bc.RepoBuildInfo != nil && bc.RepoBuildInfo.LastSuccessful != nil,

		Pkg:     pkg,
		Subpkgs: subpkgs,
		PDoc:    godocsupport.NewTDoc(pkg),

		RobotsIndex: true,
	})
}

func getGodocDir(ctx context.Context, cl *sourcegraph.Client, repo *sourcegraph.Repo, repoRevSpec sourcegraph.RepoRevSpec, subdir string) (*gosrc.Directory, error) {
	var importPath string
	if repoRevSpec.URI == "github.com/golang/go" {
		importPath = strings.TrimPrefix(subdir, "src/")
	} else {
		importPath = pathpkg.Join(repoRevSpec.URI, subdir)
	}

	d := &gosrc.Directory{
		ImportPath:  importPath,
		ProjectRoot: repoRevSpec.URI,
		ProjectName: repo.Name,
		ProjectURL:  string(repo.HomepageURL),
		VCS:         repo.VCS,
		DeadEndFork: repo.Fork, /* TODO(repomd): && repo.GitHubStars < 3,*/
		BrowseURL:   router.Rel.URLToRepoTreeEntry(repo.URI, repoRevSpec.CommitID, subdir).String(),
		LineFmt:     "%s#startline=%d&endline=%[2]d",
	}

	entrySpec := sourcegraph.TreeEntrySpec{
		RepoRev: repoRevSpec,
		Path:    subdir,
	}
	dirEntry, err := cl.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{Entry: entrySpec, Opt: nil})
	if err != nil {
		return nil, err
	}
	for _, entry := range dirEntry.Entries {
		path := filepath.Join(subdir, entry.Name)
		switch entry.Type {
		case vcsclient.FileEntry:
			if filepath.Ext(entry.Name) == ".go" {
				file, err := cl.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{Entry: sourcegraph.TreeEntrySpec{RepoRev: repoRevSpec, Path: path}, Opt: nil})
				if err != nil {
					return nil, err
				}
				d.Files = append(d.Files, &gosrc.File{
					Name:      entry.Name,
					Data:      file.Contents,
					BrowseURL: router.Rel.URLToRepoTreeEntry(repo.URI, repoRevSpec.CommitID, path).String(),
				})
			}
		case vcsclient.DirEntry:
			d.Subdirectories = append(d.Subdirectories, path)
		}
	}

	return d, nil
}
