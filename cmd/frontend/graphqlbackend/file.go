package graphqlbackend

import (
	"context"
	"html/template"
	"path"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/highlight"
	"github.com/sourcegraph/sourcegraph/internal/markdown"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func (r *GitTreeEntryResolver) Content(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cachedRepo, err := backend.CachedGitRepo(ctx, r.commit.repo.repo)
	if err != nil {
		return "", err
	}

	contents, err := git.ReadFile(ctx, *cachedRepo, api.CommitID(r.commit.OID()), r.Path(), 0)
	if err != nil {
		return "", err
	}

	return string(contents), nil
}

func (r *GitTreeEntryResolver) RichHTML(ctx context.Context) (string, error) {
	switch path.Ext(r.Path()) {
	case ".md", ".mdown", ".markdown", ".markdn":
		break
	default:
		return "", nil
	}
	content, err := r.Content(ctx)
	if err != nil {
		return "", err
	}
	return markdown.Render(content), nil
}

type markdownOptions struct {
	AlwaysNil *string
}

func (*schemaResolver) RenderMarkdown(args *struct {
	Markdown string
	Options  *markdownOptions
}) string {
	return markdown.Render(args.Markdown)
}

func (*schemaResolver) HighlightCode(ctx context.Context, args *struct {
	Code           string
	FuzzyLanguage  string
	DisableTimeout bool
	IsLightTheme   bool
}) (string, error) {
	language := highlight.SyntectLanguageMap[strings.ToLower(args.FuzzyLanguage)]
	filePath := "file." + language
	html, _, err := highlight.Code(ctx, highlight.Params{
		Content:        []byte(args.Code),
		Filepath:       filePath,
		DisableTimeout: args.DisableTimeout,
		IsLightTheme:   args.IsLightTheme,
	})
	if err != nil {
		return args.Code, err
	}
	return string(html), nil
}

func (r *GitTreeEntryResolver) Binary(ctx context.Context) (bool, error) {
	content, err := r.Content(ctx)
	if err != nil {
		return false, err
	}
	return highlight.IsBinary([]byte(content)), nil
}

type highlightedFileResolver struct {
	aborted bool
	html    string
}

func (h *highlightedFileResolver) Aborted() bool { return h.aborted }
func (h *highlightedFileResolver) HTML() string  { return h.html }

func (r *GitTreeEntryResolver) Highlight(ctx context.Context, args *HighlightArgs) (*highlightedFileResolver, error) {
	// Timeout for reading file via Git.
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cachedRepo, err := backend.CachedGitRepo(ctx, r.commit.repo.repo)
	if err != nil {
		return nil, err
	}

	content, err := git.ReadFile(ctx, *cachedRepo, api.CommitID(r.commit.OID()), r.Path(), 0)
	if err != nil {
		return nil, err
	}

	// Highlight the content.
	var (
		html   template.HTML
		result = &highlightedFileResolver{}
	)
	simulateTimeout := r.commit.repo.repo.Name == "github.com/sourcegraph/AlwaysHighlightTimeoutTest"
	html, result.aborted, err = highlight.Code(ctx, highlight.Params{
		Content:            content,
		Filepath:           r.Path(),
		DisableTimeout:     args.DisableTimeout,
		IsLightTheme:       args.IsLightTheme,
		HighlightLongLines: args.HighlightLongLines,
		SimulateTimeout:    simulateTimeout,
		Metadata: highlight.Metadata{
			RepoName: string(r.commit.repo.repo.Name),
			Revision: string(r.commit.oid),
		},
	})
	if err != nil {
		return nil, err
	}
	result.html = string(html)
	return result, nil
}
