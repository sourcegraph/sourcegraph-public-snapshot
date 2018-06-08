package graphqlbackend

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/microcosm-cc/bluemonday"
	gfm "github.com/shurcooL/github_flavored_markdown"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/graphqlbackend/externallink"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/highlight"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
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

	contents, err := backend.CachedGitRepoTmp(r.commit.repo.repo).ReadFile(ctx, api.CommitID(r.commit.oid), r.path)
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

	stat, err := backend.CachedGitRepoTmp(r.commit.repo.repo).Stat(ctx, api.CommitID(r.commit.oid), r.path)
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

func (r *fileResolver) ExternalURLs(ctx context.Context) ([]*externallink.Resolver, error) {
	isDir, err := r.IsDirectory(ctx)
	if err != nil {
		return nil, nil
	}
	return externallink.FileOrDir(ctx, r.commit.repo.repo, r.commit.revForURL(), r.path, isDir)
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
	return isBinary([]byte(content)), nil
}

// isBinary is a helper to tell if the content of a file is binary or not.
func isBinary(content []byte) bool {
	// We first check if the file is valid UTF8, since we always consider that
	// to be non-binary.
	//
	// Secondly, if the file is not valid UTF8, we check if the detected HTTP
	// content type is text, which covers a whole slew of other non-UTF8 text
	// encodings for us.
	return !utf8.Valid(content) && !strings.HasPrefix(http.DetectContentType(content), "text/")
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

	code, err := backend.CachedGitRepoTmp(r.commit.repo.repo).ReadFile(ctx, api.CommitID(r.commit.oid), r.path)
	if err != nil {
		return nil, err
	}

	// Never pass binary files to the syntax highlighter.
	if isBinary(code) {
		return nil, errors.New("cannot render binary file")
	}

	extension := strings.TrimPrefix(path.Ext(r.path), ".")
	if extension == "" {
		// When the file does not have an extension, fall back to the basename
		// (e.g. for Dockerfile).
		extension = path.Base(r.path)
	}

	// Highlight the code.
	var (
		html   template.HTML
		result = &highlightedFileResolver{}
	)
	html, result.aborted, err = highlight.Code(ctx, string(code), extension, args.DisableTimeout, args.IsLightTheme)
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
	commits, err := backend.CachedGitRepoTmp(r.commit.repo.repo).Commits(ctx, git.CommitsOptions{
		Range: string(r.commit.oid),
		N:     limit,
		Path:  r.path,
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

func (r *fileResolver) Blame(ctx context.Context,
	args *struct {
		StartLine int32
		EndLine   int32
	}) ([]*hunkResolver, error) {
	hunks, err := git.BlameFile(ctx, gitserver.Repo{Name: r.commit.repo.repo.URI}, r.path, &git.BlameOptions{
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
