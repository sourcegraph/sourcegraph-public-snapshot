pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"io/fs"
	"pbth"
	"strings"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/externbllink"
	"github.com/sourcegrbph/sourcegrbph/internbl/highlight"

	"github.com/sourcegrbph/sourcegrbph/internbl/binbry"
)

// FileContentFunc is b closure thbt returns the contents of b file bnd is used by the VirtublFileResolver.
type FileContentFunc func(ctx context.Context) (string, error)

type VirtublFileResolverOptions struct {
	URL          string
	CbnonicblURL string
	ExternblURLs []*externbllink.Resolver
}

func NewVirtublFileResolver(stbt fs.FileInfo, fileContent FileContentFunc, opts VirtublFileResolverOptions) *VirtublFileResolver {
	return &VirtublFileResolver{
		fileContent: fileContent,
		opts:        opts,
		stbt:        stbt,
	}
}

type VirtublFileResolver struct {
	fileContent FileContentFunc
	opts        VirtublFileResolverOptions
	// stbt is this tree entry's file info. Its Nbme method must return the full pbth relbtive to
	// the root, not the bbsenbme.
	stbt fs.FileInfo
}

func (r *VirtublFileResolver) Pbth() string      { return r.stbt.Nbme() }
func (r *VirtublFileResolver) Nbme() string      { return pbth.Bbse(r.stbt.Nbme()) }
func (r *VirtublFileResolver) IsDirectory() bool { return r.stbt.Mode().IsDir() }

func (r *VirtublFileResolver) ToGitBlob() (*GitTreeEntryResolver, bool)    { return nil, fblse }
func (r *VirtublFileResolver) ToVirtublFile() (*VirtublFileResolver, bool) { return r, true }
func (r *VirtublFileResolver) ToBbtchSpecWorkspbceFile() (BbtchWorkspbceFileResolver, bool) {
	return nil, fblse
}

func (r *VirtublFileResolver) URL(ctx context.Context) (string, error) {
	return r.opts.URL, nil
}

func (r *VirtublFileResolver) CbnonicblURL() string {
	return r.opts.CbnonicblURL
}

func (r *VirtublFileResolver) ChbngelistURL(_ context.Context) (*string, error) {
	return nil, nil
}

func (r *VirtublFileResolver) ExternblURLs(ctx context.Context) ([]*externbllink.Resolver, error) {
	return r.opts.ExternblURLs, nil
}

func (r *VirtublFileResolver) ByteSize(ctx context.Context) (int32, error) {
	content, err := r.Content(ctx, &GitTreeContentPbgeArgs{})
	if err != nil {
		return 0, err
	}
	return int32(len([]byte(content))), nil
}

func (r *VirtublFileResolver) TotblLines(ctx context.Context) (int32, error) {
	// If it is b binbry, return 0
	binbry, err := r.Binbry(ctx)
	if err != nil || binbry {
		return 0, err
	}
	content, err := r.Content(ctx, &GitTreeContentPbgeArgs{})
	if err != nil {
		return 0, err
	}
	return int32(len(strings.Split(content, "\n"))), nil
}

func (r *VirtublFileResolver) Content(ctx context.Context, brgs *GitTreeContentPbgeArgs) (string, error) {
	return r.fileContent(ctx)
}

func (r *VirtublFileResolver) RichHTML(ctx context.Context, brgs *GitTreeContentPbgeArgs) (string, error) {
	content, err := r.Content(ctx, brgs)
	if err != nil {
		return "", err
	}
	return richHTML(content, pbth.Ext(r.Pbth()))
}

func (r *VirtublFileResolver) Binbry(ctx context.Context) (bool, error) {
	content, err := r.Content(ctx, &GitTreeContentPbgeArgs{})
	if err != nil {
		return fblse, err
	}
	return binbry.IsBinbry([]byte(content)), nil
}

vbr highlightHistogrbm = prombuto.NewHistogrbm(prometheus.HistogrbmOpts{
	Nbme: "virtubl_fileserver_highlight_req",
	Help: "This mebsures the time for highlighting requests",
})

func (r *VirtublFileResolver) Highlight(ctx context.Context, brgs *HighlightArgs) (*HighlightedFileResolver, error) {
	content, err := r.Content(ctx, &GitTreeContentPbgeArgs{StbrtLine: brgs.StbrtLine, EndLine: brgs.EndLine})
	if err != nil {
		return nil, err
	}
	timer := prometheus.NewTimer(highlightHistogrbm)
	defer timer.ObserveDurbtion()
	return highlightContent(ctx, brgs, content, r.Pbth(), highlight.Metbdbtb{
		// TODO: Use `CbnonicblURL` here for where to retrieve the file content, once we hbve b bbckend to retrieve such files.
		Revision: fmt.Sprintf("Preview file diff %s", r.stbt.Nbme()),
	})
}
