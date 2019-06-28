package graphqlbackend

import (
	"context"
	"html/template"
	"path"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/markdown"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/highlight"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func (r *gitTreeEntryResolver) Content(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cachedRepo, err := backend.CachedGitRepo(ctx, r.commit.repo.repo.TODO())
	if err != nil {
		return "", err
	}

	contents, err := git.ReadFile(ctx, *cachedRepo, api.CommitID(r.commit.OID()), r.path)
	if err != nil {
		return "", err
	}

	return string(contents), nil
}

func (r *gitTreeEntryResolver) RichHTML(ctx context.Context) (string, error) {
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
	return markdown.Render(content, nil), nil
}

type markdownOptions struct {
	AlwaysNil *string
}

func (*schemaResolver) RenderMarkdown(args *struct {
	Markdown string
	Options  *markdownOptions
}) string {
	return markdown.Render(args.Markdown, nil)
}

func (*schemaResolver) HighlightCode(ctx context.Context, args *struct {
	Code           string
	FuzzyLanguage  string
	DisableTimeout bool
	IsLightTheme   bool
}) (string, error) {
	language := highlight.SyntectLanguageMap[strings.ToLower(args.FuzzyLanguage)]
	filePath := "file." + language
	html, _, err := highlight.Code(ctx, []byte(args.Code), filePath, args.DisableTimeout, args.IsLightTheme)
	if err != nil {
		return args.Code, err
	}
	return string(html), nil
}

func (r *gitTreeEntryResolver) Binary(ctx context.Context) (bool, error) {
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

func (r *gitTreeEntryResolver) Highlight(ctx context.Context, args *struct {
	DisableTimeout bool
	IsLightTheme   bool
}) (*highlightedFileResolver, error) {
	// Timeout for reading file via Git.
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cachedRepo, err := backend.CachedGitRepo(ctx, r.commit.repo.repo.TODO())
	if err != nil {
		return nil, err
	}

	content, err := git.ReadFile(ctx, *cachedRepo, api.CommitID(r.commit.OID()), r.path)
	if err != nil {
		return nil, err
	}

	// Highlight the content.
	var (
		html   template.HTML
		result = &highlightedFileResolver{}
	)
	html, result.aborted, err = highlight.Code(ctx, content, r.path, args.DisableTimeout, args.IsLightTheme)
	if err != nil {
		return nil, err
	}
	result.html = string(html)
	return result, nil
}
