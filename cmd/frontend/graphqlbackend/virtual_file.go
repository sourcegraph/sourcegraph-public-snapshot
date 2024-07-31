package graphqlbackend

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/highlight"
	"github.com/sourcegraph/sourcegraph/internal/binary"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/languages"
)

// FileContentFunc is a closure that returns the contents of a file and is used by the VirtualFileResolver.
type FileContentFunc func(ctx context.Context) (string, error)

type VirtualFileResolverOptions struct {
	URL          string
	CanonicalURL string
	ExternalURLs []*externallink.Resolver
}

func NewVirtualFileResolver(stat fs.FileInfo, fileContent FileContentFunc, opts VirtualFileResolverOptions) *VirtualFileResolver {
	return &VirtualFileResolver{
		fileContent: fileContent,
		opts:        opts,
		stat:        stat,
	}
}

type VirtualFileResolver struct {
	fileContent FileContentFunc
	opts        VirtualFileResolverOptions
	// stat is this tree entry's file info. Its Name method must return the full path relative to
	// the root, not the basename.
	stat fs.FileInfo
}

func (r *VirtualFileResolver) Path() string      { return r.stat.Name() }
func (r *VirtualFileResolver) Name() string      { return path.Base(r.stat.Name()) }
func (r *VirtualFileResolver) IsDirectory() bool { return r.stat.Mode().IsDir() }

func (r *VirtualFileResolver) ToGitBlob() (*GitBlobResolver, bool)         { return nil, false }
func (r *VirtualFileResolver) ToVirtualFile() (*VirtualFileResolver, bool) { return r, true }
func (r *VirtualFileResolver) ToBatchSpecWorkspaceFile() (BatchWorkspaceFileResolver, bool) {
	return nil, false
}

func (r *VirtualFileResolver) URL(ctx context.Context) (string, error) {
	return r.opts.URL, nil
}

func (r *VirtualFileResolver) CanonicalURL() string {
	return r.opts.CanonicalURL
}

func (r *VirtualFileResolver) ChangelistURL(_ context.Context) (*string, error) {
	return nil, nil
}

func (r *VirtualFileResolver) ExternalURLs(ctx context.Context) ([]*externallink.Resolver, error) {
	return r.opts.ExternalURLs, nil
}

func (r *VirtualFileResolver) ByteSize(ctx context.Context) (int32, error) {
	content, err := r.Content(ctx, &GitTreeContentPageArgs{})
	if err != nil {
		return 0, err
	}
	return int32(len([]byte(content))), nil
}

func (r *VirtualFileResolver) TotalLines(ctx context.Context) (int32, error) {
	// If it is a binary, return 0
	binary, err := r.Binary(ctx)
	if err != nil || binary {
		return 0, err
	}
	content, err := r.Content(ctx, &GitTreeContentPageArgs{})
	if err != nil {
		return 0, err
	}
	return int32(len(strings.Split(content, "\n"))), nil
}

func (r *VirtualFileResolver) Content(ctx context.Context, args *GitTreeContentPageArgs) (string, error) {
	return r.fileContent(ctx)
}

func (r *VirtualFileResolver) Languages(ctx context.Context) ([]string, error) {
	return languages.GetLanguages(r.Name(), func() ([]byte, error) {
		content, err := r.fileContent(ctx)
		if err != nil {
			return nil, err
		}
		return []byte(content), nil
	})
}

func (r *VirtualFileResolver) RichHTML(ctx context.Context, args *GitTreeContentPageArgs) (string, error) {
	content, err := r.Content(ctx, args)
	if err != nil {
		return "", err
	}
	return richHTML(content, path.Ext(r.Path()))
}

func (r *VirtualFileResolver) Binary(ctx context.Context) (bool, error) {
	content, err := r.Content(ctx, &GitTreeContentPageArgs{})
	if err != nil {
		return false, err
	}
	return binary.IsBinary([]byte(content)), nil
}

var highlightHistogram = promauto.NewHistogram(prometheus.HistogramOpts{
	Name: "virtual_fileserver_highlight_req",
	Help: "This measures the time for highlighting requests",
})

func (r *VirtualFileResolver) Highlight(ctx context.Context, args *HighlightArgs) (*HighlightedFileResolver, error) {
	content, err := r.Content(ctx, &GitTreeContentPageArgs{StartLine: args.StartLine, EndLine: args.EndLine})
	if err != nil {
		return nil, err
	}
	timer := prometheus.NewTimer(highlightHistogram)
	defer timer.ObserveDuration()
	return highlightContent(ctx, args, content, r.Path(), highlight.Metadata{
		// TODO: Use `CanonicalURL` here for where to retrieve the file content, once we have a backend to retrieve such files.
		Revision: fmt.Sprintf("Preview file diff %s", r.stat.Name()),
	})
}
