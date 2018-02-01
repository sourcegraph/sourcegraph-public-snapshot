package graphqlbackend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/microcosm-cc/bluemonday"
	gfm "github.com/shurcooL/github_flavored_markdown"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/highlight"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repoupdater"
	repoupdaterprotocol "sourcegraph.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type fileResolver struct {
	commit *gitCommitResolver

	name string
	path string

	// stat is populated by the creator of this fileResolver if it has this
	// information available. Not all creators will have the stat info; in
	// that case, some fileResolver methods have to look up the information
	// on their own.
	stat os.FileInfo
}

func (r *fileResolver) Name() string {
	return r.name
}

func (r *fileResolver) ToDirectory() (*fileResolver, bool) { return r, true }

func (r *fileResolver) ToFile() (*fileResolver, bool) { return r, true }

func (r *fileResolver) Content(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	vcsrepo, err := backend.Repos.OpenVCS(ctx, r.commit.repo.repo)
	if err != nil {
		return "", err
	}

	contents, err := vcsrepo.ReadFile(ctx, api.CommitID(r.commit.oid), r.path)
	if err != nil {
		return "", err
	}

	return string(contents), nil
}

func (r *fileResolver) IsDirectory(ctx context.Context) (bool, error) {
	// Return immediately if we know our stat.
	if r.stat != nil {
		return r.stat.Mode().IsDir(), nil
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	vcsrepo, err := backend.Repos.OpenVCS(ctx, r.commit.repo.repo)
	if err != nil {
		return false, err
	}

	stat, err := vcsrepo.Stat(ctx, api.CommitID(r.commit.oid), r.path)
	if err != nil {
		return false, err
	}

	return stat.IsDir(), nil
}

func (r *fileResolver) Repository(ctx context.Context) (*repositoryResolver, error) {
	return r.commit.Repository(ctx)
}

func (r *fileResolver) URL(ctx context.Context) (*string, error) {
	isDir, err := r.IsDirectory(ctx)
	if err != nil {
		return nil, nil
	}

	// TODO(sqs): don't special-case GitLab, clean this code up
	if repo := r.commit.repo.repo; repo.ExternalRepo != nil && repo.ExternalRepo.ServiceType == "gitlab" {
		info, err := repoupdater.DefaultClient.RepoLookup(ctx, repoupdaterprotocol.RepoLookupArgs{ExternalRepo: repo.ExternalRepo})
		if err != nil {
			return nil, err
		}
		if info.Repo != nil && info.Repo.Links != nil {
			var urlPattern string
			if isDir {
				urlPattern = info.Repo.Links.Tree
			} else {
				urlPattern = info.Repo.Links.Blob
			}
			if urlPattern != "" {
				// TODO(sqs): use rev, not fully resolved commit ID
				url := strings.NewReplacer("{rev}", string(r.commit.oid), "{path}", r.path).Replace(urlPattern)
				return &url, nil
			}
		}
	}

	if isDir {
		return r.treeURL(ctx)
	} else {
		return r.blobURL(ctx)
	}
}

func (r *fileResolver) treeURL(ctx context.Context) (*string, error) {
	repo, err := r.Repository(ctx)
	if err != nil {
		return nil, err
	}
	uri, rev := repo.repo.URI, string(r.commit.oid)
	rc, ok := repoListConfigs[uri]
	if ok && rc.Links != nil && rc.Links.Tree != "" {
		url := strings.Replace(strings.Replace(rc.Links.Tree, "{rev}", rev, 1), "{path}", r.path, 1)
		return &url, nil
	}

	if strings.HasPrefix(string(uri), "github.com/") {
		url := fmt.Sprintf("https://%s/tree/%s/%s", uri, rev, r.path)
		return &url, nil
	}

	host := strings.Split(string(uri), "/")[0]
	if gheURL, ok := githubEnterpriseURLs[host]; ok {
		url := fmt.Sprintf("%s%s/tree/%s/%s", gheURL, strings.TrimPrefix(string(uri), host), rev, r.path)
		return &url, nil
	}

	phabRepo, _ := db.Phabricator.GetByURI(context.Background(), uri)
	if phabRepo != nil {
		url := fmt.Sprintf("%s/source/%s/browse/%s", phabRepo.URL, phabRepo.Callsign, r.path)
		return &url, nil
	}

	return nil, nil
}

func (r *fileResolver) blobURL(ctx context.Context) (*string, error) {
	repo, err := r.Repository(ctx)
	if err != nil {
		return nil, err
	}
	uri, rev := repo.repo.URI, string(r.commit.oid)
	rc, ok := repoListConfigs[uri]
	if ok && rc.Links != nil && rc.Links.Tree != "" {
		url := strings.Replace(strings.Replace(rc.Links.Blob, "{rev}", rev, 1), "{path}", r.path, 1)
		return &url, nil
	}

	if strings.HasPrefix(string(uri), "github.com/") {
		url := fmt.Sprintf("https://%s/blob/%s/%s", uri, rev, r.path)
		return &url, nil
	}

	host := strings.Split(string(uri), "/")[0]
	if gheURL, ok := githubEnterpriseURLs[host]; ok {
		url := fmt.Sprintf("%s%s/blob/%s/%s", gheURL, strings.TrimPrefix(string(uri), host), rev, r.path)
		return &url, nil
	}

	phabRepo, _ := db.Phabricator.GetByURI(context.Background(), uri)
	if phabRepo != nil {
		url := fmt.Sprintf("%s/source/%s/browse/%s", phabRepo.URL, phabRepo.Callsign, r.path)
		return &url, nil
	}

	return nil, nil
}

func (r *fileResolver) RichHTML(ctx context.Context) (string, error) {
	switch path.Ext(r.path) {
	case ".md", ".mdown", ".markdown", ".markdn":
		break
	default:
		return "", nil
	}
	content, err := r.Content(ctx)
	if err != nil {
		return "", err
	}
	return renderMarkdown(content)
}

func renderMarkdown(content string) (string, error) {
	unsafeHTML := gfm.Markdown([]byte(content))

	// The recommended policy at https://github.com/russross/blackfriday#extensions
	p := bluemonday.UGCPolicy()
	p.AllowAttrs("class").Matching(regexp.MustCompile("^language-[a-zA-Z0-9]+$")).OnElements("code")
	return string(p.SanitizeBytes(unsafeHTML)), nil
}

func (r *fileResolver) Binary(ctx context.Context) (bool, error) {
	content, err := r.Content(ctx)
	if err != nil {
		return false, err
	}
	return r.isBinary([]byte(content)), nil
}

// isBinary is a helper to tell if the content of a file is binary or not. It
// is used instead of utf8.Valid in case we ever need to add e.g. extension
// specific checks in addition to checking if the content is valid utf8.
func (r *fileResolver) isBinary(content []byte) bool {
	return !utf8.Valid(content)
}

type highlightedFileResolver struct {
	aborted bool
	html    string
}

func (h *highlightedFileResolver) Aborted() bool { return h.aborted }
func (h *highlightedFileResolver) HTML() string  { return h.html }

func (r *fileResolver) Highlight(ctx context.Context, args *struct {
	DisableTimeout bool
	IsLightTheme   bool
}) (*highlightedFileResolver, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	vcsrepo, err := backend.Repos.OpenVCS(ctx, r.commit.repo.repo)
	if err != nil {
		return nil, err
	}

	code, err := vcsrepo.ReadFile(ctx, api.CommitID(r.commit.oid), r.path)
	if err != nil {
		return nil, err
	}

	// Never pass binary files to the syntax highlighter.
	if r.isBinary(code) {
		return nil, errors.New("cannot render binary file")
	}

	// Highlight the code.
	var (
		html   template.HTML
		result = &highlightedFileResolver{}
	)
	html, result.aborted, err = highlight.Code(ctx, string(code), strings.TrimPrefix(path.Ext(r.path), "."), args.DisableTimeout, args.IsLightTheme)
	if err != nil {
		return nil, err
	}
	result.html = string(html)
	return result, nil
}

func (r *fileResolver) Commits(ctx context.Context) ([]*gitCommitResolver, error) {
	return r.commits(ctx, 10)
}

func (r *fileResolver) commits(ctx context.Context, limit uint) ([]*gitCommitResolver, error) {
	vcsrepo, err := backend.Repos.OpenVCS(ctx, r.commit.repo.repo)
	if err != nil {
		return nil, err
	}

	commits, err := vcsrepo.Commits(ctx, vcs.CommitsOptions{
		Head: api.CommitID(r.commit.oid),
		N:    limit,
		Path: r.path,
	})
	if err != nil {
		return nil, err
	}
	resolvers := make([]*gitCommitResolver, len(commits))
	for i, commit := range commits {
		resolvers[i] = toGitCommitResolver(nil, commit)
		resolvers[i].repo = r.commit.repo
	}

	return resolvers, nil
}

func (r *fileResolver) BlameRaw(ctx context.Context, args *struct {
	StartLine int32
	EndLine   int32
}) (string, error) {
	vcsrepo, err := backend.Repos.OpenVCS(ctx, r.commit.repo.repo)
	if err != nil {
		return "", err
	}

	rawBlame, err := vcsrepo.BlameFileRaw(ctx, r.path, &vcs.BlameOptions{
		NewestCommit: api.CommitID(r.commit.oid),
		StartLine:    int(args.StartLine),
		EndLine:      int(args.EndLine),
	})
	if err != nil {
		return "", err
	}
	return rawBlame, nil
}

func (r *fileResolver) Blame(ctx context.Context,
	args *struct {
		StartLine int32
		EndLine   int32
	}) ([]*hunkResolver, error) {

	vcsrepo, err := backend.Repos.OpenVCS(ctx, r.commit.repo.repo)
	if err != nil {
		return nil, err
	}

	hunks, err := vcsrepo.BlameFile(ctx, r.path, &vcs.BlameOptions{
		NewestCommit: api.CommitID(r.commit.oid),
		StartLine:    int(args.StartLine),
		EndLine:      int(args.EndLine),
	})
	if err != nil {
		return nil, err
	}

	var hunksResolver []*hunkResolver
	for _, hunk := range hunks {
		hunksResolver = append(hunksResolver, &hunkResolver{
			hunk: hunk,
		})
	}

	return hunksResolver, nil
}

func (r *fileResolver) DependencyReferences(ctx context.Context, args *struct {
	Language  string
	Line      int32
	Character int32
}) (*dependencyReferencesResolver, error) {
	depRefs, err := backend.Defs.DependencyReferences(ctx, types.DependencyReferencesOptions{
		RepoID:    r.commit.repositoryDatabaseID(),
		CommitID:  api.CommitID(r.commit.oid),
		Language:  args.Language,
		File:      r.path,
		Line:      int(args.Line),
		Character: int(args.Character),
		Limit:     500,
	})
	if err != nil {
		return nil, err
	}

	var referenceResolver []*dependencyReferenceResolver
	var repos []*repositoryResolver
	var repoIDs []api.RepoID
	for _, ref := range depRefs.References {
		if ref.RepoID == r.commit.repositoryDatabaseID() {
			continue
		}

		repo, err := db.Repos.Get(ctx, ref.RepoID)
		if err != nil {
			return nil, err
		}

		repos = append(repos, &repositoryResolver{repo: repo})
		repoIDs = append(repoIDs, repo.ID)

		depData, err := json.Marshal(ref.DepData)
		if err != nil {
			return nil, err
		}

		hints, err := json.Marshal(ref.Hints)
		if err != nil {
			return nil, err
		}

		referenceResolver = append(referenceResolver, &dependencyReferenceResolver{
			dependencyData: string(depData[:]),
			repo:           ref.RepoID,
			hints:          string(hints)[:],
		})
	}

	loc, err := json.Marshal(depRefs.Location.Location)
	if err != nil {
		return nil, err
	}

	symbol, err := json.Marshal(depRefs.Location.Symbol)
	if err != nil {
		return nil, err
	}

	return &dependencyReferencesResolver{
		dependencyReferenceData: &dependencyReferencesDataResolver{
			references: referenceResolver,
			location: &dependencyLocationResolver{
				location: string(loc[:]),
				symbol:   string(symbol[:]),
			},
		},
		repoData: &repoDataMapResolver{
			repos:   repos,
			repoIDs: repoIDs,
		},
	}, nil
}

func createFileInfo(path string, isDir bool) os.FileInfo {
	return fileInfo{path: path, isDir: isDir}
}

type fileInfo struct {
	path  string
	isDir bool
}

func (f fileInfo) Name() string { return f.path }
func (f fileInfo) Size() int64  { return 0 }
func (f fileInfo) IsDir() bool  { return f.isDir }
func (f fileInfo) Mode() os.FileMode {
	if f.IsDir() {
		return os.ModeDir
	}
	return 0
}
func (f fileInfo) ModTime() time.Time { return time.Now() }
func (f fileInfo) Sys() interface{}   { return interface{}(nil) }
