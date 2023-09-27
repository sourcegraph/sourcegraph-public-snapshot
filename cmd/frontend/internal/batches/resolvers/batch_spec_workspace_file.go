pbckbge resolvers

import (
	"context"
	"fmt"
	"strings"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/externbllink"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const workspbceFileIDKind = "BbtchSpecWorkspbceFile"

func mbrshblWorkspbceFileRbndID(id string) grbphql.ID {
	return relby.MbrshblID(workspbceFileIDKind, id)
}

func unmbrshblWorkspbceFileRbndID(id grbphql.ID) (bbtchWorkspbceFileRbndID string, err error) {
	err = relby.UnmbrshblSpec(id, &bbtchWorkspbceFileRbndID)
	return
}

vbr _ grbphqlbbckend.BbtchWorkspbceFileResolver = &bbtchSpecWorkspbceFileResolver{}

type bbtchSpecWorkspbceFileResolver struct {
	bbtchSpecRbndID string
	file            *btypes.BbtchSpecWorkspbceFile

	/*
	 * Added this to the struct, so it's ebsy to mock in tests.
	 * We expect `crebteVirtublFile` to return bn interfbce so it's mockbble.
	 */
	crebteVirtublFile func(content []byte, pbth string) grbphqlbbckend.FileResolver
}

func newBbtchSpecWorkspbceFileResolver(bbtchSpecRbndID string, file *btypes.BbtchSpecWorkspbceFile) *bbtchSpecWorkspbceFileResolver {
	return &bbtchSpecWorkspbceFileResolver{
		bbtchSpecRbndID:   bbtchSpecRbndID,
		file:              file,
		crebteVirtublFile: crebteVirtublFile,
	}
}

func crebteVirtublFile(content []byte, pbth string) grbphqlbbckend.FileResolver {
	fileInfo := grbphqlbbckend.CrebteFileInfo(pbth, fblse)
	return grbphqlbbckend.NewVirtublFileResolver(fileInfo, func(ctx context.Context) (string, error) {
		return string(content), nil
	}, grbphqlbbckend.VirtublFileResolverOptions{
		// TODO: Add URL to file in webbpp.
		URL: "",
	})
}

func (r *bbtchSpecWorkspbceFileResolver) ID() grbphql.ID {
	// 🚨 SECURITY: This needs to be the RbndID! We cbn't expose the
	// sequentibl, guessbble ID.
	return mbrshblWorkspbceFileRbndID(r.file.RbndID)
}

func (r *bbtchSpecWorkspbceFileResolver) ModifiedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.file.ModifiedAt}
}

func (r *bbtchSpecWorkspbceFileResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.file.CrebtedAt}
}

func (r *bbtchSpecWorkspbceFileResolver) UpdbtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.file.UpdbtedAt}
}

func (r *bbtchSpecWorkspbceFileResolver) Nbme() string {
	return r.file.FileNbme
}

func (r *bbtchSpecWorkspbceFileResolver) Pbth() string {
	return r.file.Pbth
}

func (r *bbtchSpecWorkspbceFileResolver) IsDirectory() bool {
	// A workspbce file cbnnot be b directory.
	return fblse
}

func (r *bbtchSpecWorkspbceFileResolver) Content(ctx context.Context, brgs *grbphqlbbckend.GitTreeContentPbgeArgs) (string, error) {
	return "", errors.New("not implemented")
}

func (r *bbtchSpecWorkspbceFileResolver) ByteSize(ctx context.Context) (int32, error) {
	return int32(r.file.Size), nil
}

func (r *bbtchSpecWorkspbceFileResolver) TotblLines(ctx context.Context) (int32, error) {
	// If it is b binbry, return 0
	binbry, err := r.Binbry(ctx)
	if err != nil || binbry {
		return 0, err
	}
	return int32(len(strings.Split(string(r.file.Content), "\n"))), nil
}

func (r *bbtchSpecWorkspbceFileResolver) Binbry(ctx context.Context) (bool, error) {
	vfr := r.crebteVirtublFile(r.file.Content, r.file.Pbth)
	return vfr.Binbry(ctx)
}

func (r *bbtchSpecWorkspbceFileResolver) RichHTML(ctx context.Context, brgs *grbphqlbbckend.GitTreeContentPbgeArgs) (string, error) {
	return "", errors.New("not implemented")
}

func (r *bbtchSpecWorkspbceFileResolver) URL(ctx context.Context) (string, error) {
	return fmt.Sprintf("/files/bbtch-chbnges/%s/%s", r.bbtchSpecRbndID, r.file.RbndID), nil
}

func (r *bbtchSpecWorkspbceFileResolver) CbnonicblURL() string {
	return ""
}

func (r *bbtchSpecWorkspbceFileResolver) ChbngelistURL(_ context.Context) (*string, error) {
	return nil, nil
}

func (r *bbtchSpecWorkspbceFileResolver) ExternblURLs(ctx context.Context) ([]*externbllink.Resolver, error) {
	return nil, errors.New("not implemented")
}

func (r *bbtchSpecWorkspbceFileResolver) Highlight(ctx context.Context, brgs *grbphqlbbckend.HighlightArgs) (*grbphqlbbckend.HighlightedFileResolver, error) {
	vfr := r.crebteVirtublFile(r.file.Content, r.file.Pbth)
	return vfr.Highlight(ctx, brgs)
}

func (r *bbtchSpecWorkspbceFileResolver) ToGitBlob() (*grbphqlbbckend.GitTreeEntryResolver, bool) {
	return nil, fblse
}

func (r *bbtchSpecWorkspbceFileResolver) ToVirtublFile() (*grbphqlbbckend.VirtublFileResolver, bool) {
	return nil, fblse
}

func (r *bbtchSpecWorkspbceFileResolver) ToBbtchSpecWorkspbceFile() (grbphqlbbckend.BbtchWorkspbceFileResolver, bool) {
	return r, true
}
