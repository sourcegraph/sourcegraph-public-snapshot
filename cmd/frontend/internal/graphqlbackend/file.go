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
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/highlight"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repoupdater"
	repoupdaterprotocol "sourcegraph.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

type fileResolver struct {
	commit *gitCommitResolver

	path string

	// stat is populated by the creator of this fileResolver if it has this
	// information available. Not all creators will have the stat info; in
	// that case, some fileResolver methods have to look up the information
	// on their own.
	stat os.FileInfo
}

func (r *fileResolver) Path() string { return r.path }
func (r *fileResolver) Name() string { return path.Base(r.path) }

func (r *fileResolver) ToDirectory() (*fileResolver, bool) { return r, true }
func (r *fileResolver) ToFile() (*fileResolver, bool)      { return r, true }

func (r *fileResolver) Content(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	vcsrepo := backend.Repos.CachedVCS(r.commit.repo.repo)
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

	vcsrepo := backend.Repos.CachedVCS(r.commit.repo.repo)
	stat, err := vcsrepo.Stat(ctx, api.CommitID(r.commit.oid), r.path)
	if err != nil {
		return false, err
	}

	return stat.IsDir(), nil
}

func (r *fileResolver) Repository(ctx context.Context) (*repositoryResolver, error) {
	return r.commit.Repository(ctx)
}

func (r *fileResolver) URL(ctx context.Context) (string, error) {
	url := r.commit.repoRevURL() + "/-/"

	isDir, err := r.IsDirectory(ctx)
	if err != nil {
		return "", err
	}
	if isDir {
		url += "tree"
	} else {
		url += "blob"
	}
	return url + "/" + r.path, nil
}

func (r *fileResolver) ExternalURL(ctx context.Context) (*string, error) {
	isDir, err := r.IsDirectory(ctx)
	if err != nil {
		return nil, nil
	}

	repo := r.commit.repo.repo
	if repo == nil {
		return nil, nil
	}
	var url *string
	if isDir {
		url, _ = r.treeURL(ctx)
	} else {
		url, _ = r.blobURL(ctx)
	}
	if url != nil {
		return url, nil
	}

	if repo.ExternalRepo != nil {
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
				// TODO(sqs): use rev, not fully resolved commit ID. When we
				// do this, we will need a way for templates to escape rev for
				// path vs query string.
				url := strings.NewReplacer("{rev}", string(r.commit.oid), "{path}", r.path).Replace(urlPattern)
				return &url, nil
			}
		}
	}
	return nil, nil
}

func (r *fileResolver) treeURL(ctx context.Context) (*string, error) {
	repo, err := r.Repository(ctx)
	if err != nil {
		return nil, err
	}
	uri, rev := repo.repo.URI, string(r.commit.oid)
	repoListConfigs := repoListConfigs.Get().(map[api.RepoURI]schema.Repository)
	rc, ok := repoListConfigs[uri]
	if ok && rc.Links != nil && rc.Links.Tree != "" {
		url := strings.Replace(strings.Replace(rc.Links.Tree, "{rev}", rev, 1), "{path}", r.path, 1)
		return &url, nil
	}

	phabRepo, _ := db.Phabricator.GetByURI(context.Background(), uri)
	if phabRepo != nil {
		defaultBranch, err := repo.DefaultBranch(ctx)
		if err != nil {
			return nil, nil
		}
		url := fmt.Sprintf("%s/source/%s/browse/%s/%s;%s", phabRepo.URL, phabRepo.Callsign, *defaultBranch, r.path, rev)
		return &url, nil
	}

	if strings.HasPrefix(string(uri), "github.com/") {
		url := fmt.Sprintf("https://%s/tree/%s/%s", uri, rev, r.path)
		return &url, nil
	}

	host := strings.Split(string(uri), "/")[0]
	if gheURL, ok := conf.GitHubEnterpriseURLs()[host]; ok {
		url := fmt.Sprintf("%s%s/tree/%s/%s", gheURL, strings.TrimPrefix(string(uri), host), rev, r.path)
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
	repoListConfigs := repoListConfigs.Get().(map[api.RepoURI]schema.Repository)
	rc, ok := repoListConfigs[uri]
	if ok && rc.Links != nil && rc.Links.Blob != "" {
		url := strings.Replace(strings.Replace(rc.Links.Blob, "{rev}", rev, 1), "{path}", r.path, 1)
		return &url, nil
	}

	phabRepo, _ := db.Phabricator.GetByURI(context.Background(), uri)
	if phabRepo != nil {
		defaultBranch, err := repo.DefaultBranch(ctx)
		if err != nil {
			return nil, nil
		}
		url := fmt.Sprintf("%s/source/%s/browse/%s/%s;%s", phabRepo.URL, phabRepo.Callsign, *defaultBranch, r.path, rev)
		return &url, nil
	}

	if strings.HasPrefix(string(uri), "github.com/") {
		url := fmt.Sprintf("https://%s/blob/%s/%s", uri, rev, r.path)
		return &url, nil
	}

	host := strings.Split(string(uri), "/")[0]
	if gheURL, ok := conf.GitHubEnterpriseURLs()[host]; ok {
		url := fmt.Sprintf("%s%s/blob/%s/%s", gheURL, strings.TrimPrefix(string(uri), host), rev, r.path)
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
	return renderMarkdown(content), nil
}

func renderMarkdown(content string) string {
	unsafeHTML := gfm.Markdown([]byte(content))

	p := bluemonday.UGCPolicy()
	p.AllowAttrs("name").Matching(bluemonday.SpaceSeparatedTokens).OnElements("a")
	p.AllowAttrs("rel").Matching(regexp.MustCompile(`^nofollow$`)).OnElements("a")
	p.AllowAttrs("class").Matching(regexp.MustCompile(`^anchor$`)).OnElements("a")
	p.AllowAttrs("aria-hidden").Matching(regexp.MustCompile(`^true$`)).OnElements("a")
	p.AllowAttrs("type").Matching(regexp.MustCompile(`^checkbox$`)).OnElements("input")
	p.AllowAttrs("checked", "disabled").Matching(regexp.MustCompile(`^$`)).OnElements("input")
	p.AllowAttrs("class").Matching(regexp.MustCompile("^language-[a-zA-Z0-9]+$")).OnElements("code")
	return string(p.SanitizeBytes(unsafeHTML))
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

	vcsrepo := backend.Repos.CachedVCS(r.commit.repo.repo)
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
	vcsrepo := backend.Repos.CachedVCS(r.commit.repo.repo)
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
	vcsrepo := backend.Repos.CachedVCS(r.commit.repo.repo)
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
	vcsrepo := backend.Repos.CachedVCS(r.commit.repo.repo)
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
