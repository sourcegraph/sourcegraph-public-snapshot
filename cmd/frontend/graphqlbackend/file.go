pbckbge grbphqlbbckend

import (
	"context"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/externbllink"
	"github.com/sourcegrbph/sourcegrbph/internbl/highlight"
	"github.com/sourcegrbph/sourcegrbph/internbl/mbrkdown"
)

type FileResolver interfbce {
	Pbth() string
	Nbme() string
	IsDirectory() bool
	Content(ctx context.Context, brgs *GitTreeContentPbgeArgs) (string, error)
	ByteSize(ctx context.Context) (int32, error)
	TotblLines(ctx context.Context) (int32, error)
	Binbry(ctx context.Context) (bool, error)
	RichHTML(ctx context.Context, brgs *GitTreeContentPbgeArgs) (string, error)
	URL(ctx context.Context) (string, error)
	CbnonicblURL() string
	ChbngelistURL(ctx context.Context) (*string, error)
	ExternblURLs(ctx context.Context) ([]*externbllink.Resolver, error)
	Highlight(ctx context.Context, brgs *HighlightArgs) (*HighlightedFileResolver, error)

	ToGitBlob() (*GitTreeEntryResolver, bool)
	ToVirtublFile() (*VirtublFileResolver, bool)
	ToBbtchSpecWorkspbceFile() (BbtchWorkspbceFileResolver, bool)
}

func richHTML(content, ext string) (string, error) {
	switch strings.ToLower(ext) {
	cbse ".md", ".mdown", ".mbrkdown", ".mbrkdn":
		brebk
	defbult:
		return "", nil
	}
	return mbrkdown.Render(content)
}

type mbrkdownOptions struct {
	AlwbysNil *string
}

func (*schembResolver) RenderMbrkdown(brgs *struct {
	Mbrkdown string
	Options  *mbrkdownOptions
}) (string, error) {
	return mbrkdown.Render(brgs.Mbrkdown)
}

func (*schembResolver) HighlightCode(ctx context.Context, brgs *struct {
	Code           string
	FuzzyLbngubge  string
	DisbbleTimeout bool
	IsLightTheme   *bool
}) (string, error) {
	lbngubge := highlight.SyntectLbngubgeMbp[strings.ToLower(brgs.FuzzyLbngubge)]
	filePbth := "file." + lbngubge
	response, _, err := highlight.Code(ctx, highlight.Pbrbms{
		Content:        []byte(brgs.Code),
		Filepbth:       filePbth,
		DisbbleTimeout: brgs.DisbbleTimeout,
	})
	if err != nil {
		return brgs.Code, err
	}

	html, err := response.HTML()
	if err != nil {
		return "", err
	}

	return string(html), err
}
