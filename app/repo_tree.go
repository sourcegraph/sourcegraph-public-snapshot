package app

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/doc"
	"sourcegraph.com/sourcegraph/sourcegraph/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory/filelang"
	"sourcegraph.com/sourcegraph/sourcegraph/util/cacheutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/eventsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
)

// repoTreeTemplate holds data necessary to populate the repository tree templates
// (file and directory)
type repoTreeTemplate struct {
	*handlerutil.RepoCommon
	*handlerutil.RepoRevCommon
	*handlerutil.TreeEntryCommon
	tmpl.Common `json:"-"`

	Definition      *sourcegraph.Def
	RobotsIndex     bool
	Title           string
	MetaDescription string
	EntryPath       string
	Documentation   string
}

// serveRepoTree creates a new response for the code view that contains information
// about the requested tree entry.
func serveRepoTree(w http.ResponseWriter, r *http.Request) error {
	ctx, _ := handlerutil.Client(r)

	opt := sourcegraph.RepoTreeGetOptions{
		GetFileOptions: sourcegraph.GetFileOptions{
			RecurseSingleSubfolderLimit: 200,
		},
	}

	tc, rc, vc, err := handlerutil.GetTreeEntryCommon(ctx, mux.Vars(r), &opt)
	if err != nil {
		return err
	}

	switch tc.Entry.Type {
	case sourcegraph.DirEntry:
		return serveRepoTreeDir(w, r, tc, rc, vc)
	case sourcegraph.FileEntry:
		return serveRepoTreeEntry(w, r, tc, rc, vc, nil)
	}
	return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: errors.New("unrecognized tree entry type")}
}

func serveRepoTreeDir(w http.ResponseWriter, r *http.Request, tc *handlerutil.TreeEntryCommon, rc *handlerutil.RepoCommon, vc *handlerutil.RepoRevCommon) error {
	ctx, cl := handlerutil.Client(r)
	go cacheutil.PrecacheTreeEntry(cl, ctx, tc.Entry, tc.EntrySpec)

	eventsutil.LogBrowseCode(ctx, "dir", tc, rc)
	return tmpl.Exec(r, w, "repo/tree/dir.html", http.StatusOK, nil, &repoTreeTemplate{
		TreeEntryCommon: tc,
		RepoCommon:      rc,
		RepoRevCommon:   vc,
		EntryPath:       tc.EntrySpec.Path,
	})
}

func serveRepoTreeEntry(w http.ResponseWriter, r *http.Request, tc *handlerutil.TreeEntryCommon, rc *handlerutil.RepoCommon, vc *handlerutil.RepoRevCommon, dc *sourcegraph.Def) error {
	var (
		templateFile string
		docs         string
	)
	ctx := httpctx.FromRequest(r)

	switch {
	case tc.Entry.Type == sourcegraph.DirEntry:
		treeURL := router.Rel.URLToRepoTreeEntrySpec(tc.EntrySpec).String()
		http.Redirect(w, r, treeURL, http.StatusFound)
		return nil

	case isDocFile(tc.EntrySpec.Path) && !router.IsRaw(r.URL):
		if tc.Entry.Contents == nil {
			panic("Entry.Contents is nil")
		}
		formatted, err := doc.ToHTML(doc.Format(tc.EntrySpec.Path), tc.Entry.Contents)
		if err != nil {
			return err
		}
		docs = string(formatted)
		eventsutil.LogBrowseCode(ctx, "doc", tc, rc)
		templateFile = "repo/tree/doc.html"

	default:
		eventsutil.LogBrowseCode(ctx, "file", tc, rc)
		templateFile = "repo/tree/file.html"
	}

	tc.Entry.ContentsString = string(tc.Entry.Contents)
	tc.Entry.Contents = nil

	return tmpl.Exec(r, w, templateFile, http.StatusOK, nil, &repoTreeTemplate{
		Definition:      dc,
		TreeEntryCommon: tc,
		RepoCommon:      rc,
		RepoRevCommon:   vc,
		Documentation:   docs,
		EntryPath:       tc.EntrySpec.Path,
	})
}

func isDocFile(filename string) bool {
	langs := filelang.Langs.ByFilename(filename)
	return len(langs) > 0 && langs[0].Type == "prose"
}

// fileToBreadcrumb returns the file link without line number
func fileToBreadcrumb(repoURI string, rev, path string) []*BreadcrumbLink {
	return fileLinesToBreadcrumb(repoURI, rev, path, 0)
}

// fileLinesToBreadcrumb returns the file link with line number
func fileLinesToBreadcrumb(repo string, rev, path string, startLine int) []*BreadcrumbLink {
	return snippetToBreadcrumb(repo, rev, path, startLine, 0)
}

func snippetToBreadcrumb(repo string, rev, path string, startLine int, endLine int) []*BreadcrumbLink {
	return absSnippetToBreadcrumb(nil, repo, rev, path, startLine, endLine, true)
}

type BreadcrumbLink struct {
	Text string
	URL  string
}

// absSnippetToBreadcrumb returns the breadcrumb to a specific set of lines in a file. The URLs are absolute if appURL is given.
func absSnippetToBreadcrumb(appURL *url.URL, repo string, rev, path string, startLine int, endLine int, includeRepo bool) []*BreadcrumbLink {
	curPath := ""
	if appURL != nil {
		curPath = appURL.String()
	}
	curPath += router.Rel.URLToRepoTreeEntry(repo, rev, "").Path

	var links []*BreadcrumbLink

	if includeRepo {
		links = []*BreadcrumbLink{{Text: filepath.Base(repo), URL: curPath}}
	}

	if path == "." {
		return links
	}

	segs := strings.Split(strings.TrimPrefix(path, "/"), "/")
	for i, seg := range segs {
		link := &BreadcrumbLink{Text: seg, URL: curPath + seg}
		if i == len(segs)-1 && startLine != 0 {
			if endLine == 0 {
				endLine = startLine
			}

			link.Text += fmt.Sprintf(":%d", startLine)
			link.URL += fmt.Sprintf("#L%d", startLine)
			if endLine != startLine {
				link.Text += fmt.Sprintf("-%d", endLine)
				link.URL += fmt.Sprintf("-%d", endLine)
			}
		}

		links = append(links, link)
		curPath += seg + "/"
	}

	return links
}
