package graphqlbackend

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"path"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/highlight"
	"github.com/sourcegraph/sourcegraph/pkg/markdown"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func (r *gitTreeEntryResolver) Content(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	contents, err := git.ReadFile(ctx, backend.CachedGitRepo(r.commit.repo.repo), api.CommitID(r.commit.oid), r.path)
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
	return markdown.Render(content, nil)
}

type markdownOptions struct {
	AlwaysNil *string
}

func (*schemaResolver) RenderMarkdown(args *struct {
	Markdown string
	Options  *markdownOptions
}) (string, error) {
	return markdown.Render(args.Markdown, nil)
}

func (r *gitTreeEntryResolver) Binary(ctx context.Context) (bool, error) {
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

func (r *gitTreeEntryResolver) Highlight(ctx context.Context, args *struct {
	DisableTimeout bool
	IsLightTheme   bool
}) (*highlightedFileResolver, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	code, err := git.ReadFile(ctx, backend.CachedGitRepo(r.commit.repo.repo), api.CommitID(r.commit.oid), r.path)
	if err != nil {
		return nil, err
	}

	// Never pass binary files to the syntax highlighter.
	if isBinary(code) {
		return nil, errors.New("cannot render binary file")
	}

	// Highlight the code.
	var (
		html   template.HTML
		result = &highlightedFileResolver{}
	)
	html, result.aborted, err = highlight.Code(ctx, string(code), r.path, args.DisableTimeout, args.IsLightTheme)
	if err != nil {
		return nil, err
	}
	result.html = string(html)
	return result, nil
}
