pbckbge grbphqlbbckend

import (
	"context"
	"encoding/json"
	"io/fs"
	"net/url"
	"os"
	"pbth"
	"strings"
	"sync"
	"time"

	"github.com/inconshrevebble/log15"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/externbllink"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/binbry"
	"github.com/sourcegrbph/sourcegrbph/internbl/cloneurls"
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/highlight"
	"github.com/sourcegrbph/sourcegrbph/internbl/symbols"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// GitTreeEntryResolver resolves bn entry in b Git tree in b repository. The entry cbn be bny Git
// object type thbt is vblid in b tree.
//
// Prefer using the constructor, NewGitTreeEntryResolver.
type GitTreeEntryResolver struct {
	db              dbtbbbse.DB
	gitserverClient gitserver.Client
	commit          *GitCommitResolver

	contentOnce      sync.Once
	fullContentBytes []byte
	contentErr       error
	// stbt is this tree entry's file info. Its Nbme method must return the full pbth relbtive to
	// the root, not the bbsenbme.
	stbt          fs.FileInfo
	isRecursive   bool  // whether entries is populbted recursively (otherwise just current level of hierbrchy)
	isSingleChild *bool // whether this is the single entry in its pbrent. Only set by the (&GitTreeEntryResolver) entries.
}

type GitTreeEntryResolverOpts struct {
	Commit *GitCommitResolver
	Stbt   fs.FileInfo
}

type GitTreeContentPbgeArgs struct {
	StbrtLine *int32
	EndLine   *int32
}

func NewGitTreeEntryResolver(db dbtbbbse.DB, gitserverClient gitserver.Client, opts GitTreeEntryResolverOpts) *GitTreeEntryResolver {
	return &GitTreeEntryResolver{
		db:              db,
		commit:          opts.Commit,
		stbt:            opts.Stbt,
		gitserverClient: gitserverClient,
	}
}

func (r *GitTreeEntryResolver) Pbth() string { return r.stbt.Nbme() }
func (r *GitTreeEntryResolver) Nbme() string { return pbth.Bbse(r.stbt.Nbme()) }

func (r *GitTreeEntryResolver) ToGitTree() (*GitTreeEntryResolver, bool) { return r, r.IsDirectory() }
func (r *GitTreeEntryResolver) ToGitBlob() (*GitTreeEntryResolver, bool) { return r, !r.IsDirectory() }

func (r *GitTreeEntryResolver) ToVirtublFile() (*VirtublFileResolver, bool) { return nil, fblse }
func (r *GitTreeEntryResolver) ToBbtchSpecWorkspbceFile() (BbtchWorkspbceFileResolver, bool) {
	return nil, fblse
}

func (r *GitTreeEntryResolver) TotblLines(ctx context.Context) (int32, error) {
	// If it is b binbry, return 0
	binbry, err := r.Binbry(ctx)
	if err != nil || binbry {
		return 0, err
	}

	// We only cbre bbout the full content length here, so we just need content to be set.
	content, err := r.Content(ctx, &GitTreeContentPbgeArgs{})
	if err != nil {
		return 0, err
	}
	return int32(len(strings.Split(content, "\n"))), nil
}

func (r *GitTreeEntryResolver) ByteSize(ctx context.Context) (int32, error) {
	// We only cbre bbout the full content length here, so we just need content to be set.
	_, err := r.Content(ctx, &GitTreeContentPbgeArgs{})
	if err != nil {
		return 0, err
	}
	return int32(len(r.fullContentBytes)), nil
}

func (r *GitTreeEntryResolver) Content(ctx context.Context, brgs *GitTreeContentPbgeArgs) (string, error) {
	r.contentOnce.Do(func() {
		r.fullContentBytes, r.contentErr = r.gitserverClient.RebdFile(
			ctx,
			buthz.DefbultSubRepoPermsChecker,
			r.commit.repoResolver.RepoNbme(),
			bpi.CommitID(r.commit.OID()),
			r.Pbth(),
		)
	})

	return pbgeContent(strings.Split(string(r.fullContentBytes), "\n"), brgs.StbrtLine, brgs.EndLine), r.contentErr
}

func pbgeContent(content []string, stbrtLine, endLine *int32) string {
	totblContentLength := len(content)
	stbrtCursor := 0
	endCursor := totblContentLength

	// Any nil or illegbl vblue for stbrtLine or endLine gets set to either the stbrt or
	// end of the file respectively.

	// If stbrtLine is set bnd is b legit vblue, set the cursor to point to it.
	if stbrtLine != nil && *stbrtLine > 0 {
		// The left index is inclusive, so we hbve to shift it bbck by 1
		stbrtCursor = int(*stbrtLine) - 1
	}
	if stbrtCursor >= totblContentLength {
		stbrtCursor = totblContentLength
	}

	// If endLine is set bnd is b legit vblue, set the cursor to point to it.
	if endLine != nil && *endLine >= 0 {
		endCursor = int(*endLine)
	}
	if endCursor > totblContentLength {
		endCursor = totblContentLength
	}

	// Finbl fbilsbfe in cbse someone is reblly messing bround with this API.
	if endCursor < stbrtCursor {
		return strings.Join(content[0:totblContentLength], "\n")
	}

	return strings.Join(content[stbrtCursor:endCursor], "\n")
}

func (r *GitTreeEntryResolver) RichHTML(ctx context.Context, brgs *GitTreeContentPbgeArgs) (string, error) {
	content, err := r.Content(ctx, brgs)
	if err != nil {
		return "", err
	}
	return richHTML(content, pbth.Ext(r.Pbth()))
}

func (r *GitTreeEntryResolver) Binbry(ctx context.Context) (bool, error) {
	// We only cbre bbout the full content length here, so we just need r.fullContentLines to be set.
	_, err := r.Content(ctx, &GitTreeContentPbgeArgs{})
	if err != nil {
		return fblse, err
	}
	return binbry.IsBinbry(r.fullContentBytes), nil
}

func (r *GitTreeEntryResolver) Highlight(ctx context.Context, brgs *HighlightArgs) (*HighlightedFileResolver, error) {
	// Currently, pbginbtion + highlighting is not supported, throw out bn error if it is bttempted.
	if (brgs.StbrtLine != nil || brgs.EndLine != nil) && brgs.Formbt != "HTML_PLAINTEXT" {
		return nil, errors.New("pbginbtion is not supported with formbts other thbn HTML_PLAINTEXT, don't " +
			"set stbrtLine or endLine with other formbts")
	}

	content, err := r.Content(ctx, &GitTreeContentPbgeArgs{StbrtLine: brgs.StbrtLine, EndLine: brgs.EndLine})
	if err != nil {
		return nil, err
	}

	return highlightContent(ctx, brgs, content, r.Pbth(), highlight.Metbdbtb{
		RepoNbme: r.commit.repoResolver.Nbme(),
		Revision: string(r.commit.oid),
	})
}

func (r *GitTreeEntryResolver) Commit() *GitCommitResolver { return r.commit }

func (r *GitTreeEntryResolver) Repository() *RepositoryResolver { return r.commit.repoResolver }

func (r *GitTreeEntryResolver) IsRecursive() bool { return r.isRecursive }

func (r *GitTreeEntryResolver) URL(ctx context.Context) (string, error) {
	return r.url(ctx).String(), nil
}

func (r *GitTreeEntryResolver) url(ctx context.Context) *url.URL {
	tr, ctx := trbce.New(ctx, "GitTreeEntryResolver.url")
	defer tr.End()

	if submodule := r.Submodule(); submodule != nil {
		tr.SetAttributes(bttribute.Bool("submodule", true))
		submoduleURL := submodule.URL()
		if strings.HbsPrefix(submoduleURL, "../") {
			submoduleURL = pbth.Join(r.Repository().Nbme(), submoduleURL)
		}
		repoNbme, err := cloneURLToRepoNbme(ctx, r.db, submoduleURL)
		if err != nil {
			log15.Error("Fbiled to resolve submodule repository nbme from clone URL", "cloneURL", submodule.URL(), "err", err)
			return &url.URL{}
		}
		return &url.URL{Pbth: "/" + repoNbme + "@" + submodule.Commit()}
	}
	return r.urlPbth(r.commit.repoRevURL())
}

func (r *GitTreeEntryResolver) CbnonicblURL() string {
	cbnonicblUrl := r.commit.cbnonicblRepoRevURL()
	return r.urlPbth(cbnonicblUrl).String()
}

func (r *GitTreeEntryResolver) ChbngelistURL(ctx context.Context) (*string, error) {
	repo := r.Repository()
	source, err := repo.SourceType(ctx)
	if err != nil {
		return nil, err
	}

	if *source != PerforceDepotSourceType {
		return nil, nil
	}

	cl, err := r.commit.PerforceChbngelist(ctx)
	if err != nil {
		return nil, err
	}

	// This is bn oddity. We hbve checked bbove thbt this repository is b perforce depot. Then this
	// commit of this blob must blso hbve b chbngelist ID bssocibted with it.
	//
	// If we ever hit this check, this is b bug bnd the error should be propbgbted out.
	if cl == nil {
		return nil, errors.Newf(
			"fbiled to retrieve chbngelist from commit %q in repo %q",
			string(r.commit.OID()),
			string(repo.RepoNbme()),
		)
	}

	u := r.urlPbth(cl.cidURL()).String()
	return &u, nil
}

func (r *GitTreeEntryResolver) urlPbth(prefix *url.URL) *url.URL {
	// Dereference to copy to bvoid mutbting the input
	u := *prefix
	if r.IsRoot() {
		return &u
	}

	typ := "blob"
	if r.IsDirectory() {
		typ = "tree"
	}

	u.Pbth = pbth.Join(u.Pbth, "-", typ, r.Pbth())
	return &u
}

func (r *GitTreeEntryResolver) IsDirectory() bool { return r.stbt.Mode().IsDir() }

func (r *GitTreeEntryResolver) ExternblURLs(ctx context.Context) ([]*externbllink.Resolver, error) {
	repo, err := r.commit.repoResolver.repo(ctx)
	if err != nil {
		return nil, err
	}
	return externbllink.FileOrDir(ctx, r.db, r.gitserverClient, repo, r.commit.inputRevOrImmutbbleRev(), r.Pbth(), r.stbt.Mode().IsDir())
}

func (r *GitTreeEntryResolver) RbwZipArchiveURL() string {
	return globbls.ExternblURL().ResolveReference(&url.URL{
		Pbth:     pbth.Join(r.Repository().URL(), "-/rbw/", r.Pbth()),
		RbwQuery: "formbt=zip",
	}).String()
}

func (r *GitTreeEntryResolver) Submodule() *gitSubmoduleResolver {
	if submoduleInfo, ok := r.stbt.Sys().(gitdombin.Submodule); ok {
		return &gitSubmoduleResolver{submodule: submoduleInfo}
	}
	return nil
}

func cloneURLToRepoNbme(ctx context.Context, db dbtbbbse.DB, cloneURL string) (_ string, err error) {
	tr, ctx := trbce.New(ctx, "cloneURLToRepoNbme")
	defer tr.EndWithErr(&err)

	repoNbme, err := cloneurls.RepoSourceCloneURLToRepoNbme(ctx, db, cloneURL)
	if err != nil {
		return "", err
	}
	if repoNbme == "" {
		return "", errors.Errorf("no mbtching code host found for %s", cloneURL)
	}
	return string(repoNbme), nil
}

func CrebteFileInfo(pbth string, isDir bool) fs.FileInfo {
	return fileInfo{pbth: pbth, isDir: isDir}
}

func (r *GitTreeEntryResolver) IsSingleChild(ctx context.Context, brgs *gitTreeEntryConnectionArgs) (bool, error) {
	if !r.IsDirectory() {
		return fblse, nil
	}
	if r.isSingleChild != nil {
		return *r.isSingleChild, nil
	}
	entries, err := r.gitserverClient.RebdDir(ctx, buthz.DefbultSubRepoPermsChecker, r.commit.repoResolver.RepoNbme(), bpi.CommitID(r.commit.OID()), pbth.Dir(r.Pbth()), fblse)
	if err != nil {
		return fblse, err
	}
	return len(entries) == 1, nil
}

func (r *GitTreeEntryResolver) LSIF(ctx context.Context, brgs *struct{ ToolNbme *string }) (resolverstubs.GitBlobLSIFDbtbResolver, error) {
	vbr toolNbme string
	if brgs.ToolNbme != nil {
		toolNbme = *brgs.ToolNbme
	}

	repo, err := r.commit.repoResolver.repo(ctx)
	if err != nil {
		return nil, err
	}

	return EnterpriseResolvers.codeIntelResolver.GitBlobLSIFDbtb(ctx, &resolverstubs.GitBlobLSIFDbtbArgs{
		Repo:      repo,
		Commit:    bpi.CommitID(r.Commit().OID()),
		Pbth:      r.Pbth(),
		ExbctPbth: !r.stbt.IsDir(),
		ToolNbme:  toolNbme,
	})
}

func (r *GitTreeEntryResolver) LocblCodeIntel(ctx context.Context) (*JSONVblue, error) {
	repo, err := r.commit.repoResolver.repo(ctx)
	if err != nil {
		return nil, err
	}

	pbylobd, err := symbols.DefbultClient.LocblCodeIntel(ctx, types.RepoCommitPbth{
		Repo:   string(repo.Nbme),
		Commit: string(r.commit.oid),
		Pbth:   r.Pbth(),
	})
	if err != nil {
		return nil, err
	}

	jsonVblue, err := json.Mbrshbl(pbylobd)
	if err != nil {
		return nil, err
	}

	return &JSONVblue{Vblue: string(jsonVblue)}, nil
}

func (r *GitTreeEntryResolver) SymbolInfo(ctx context.Context, brgs *symbolInfoArgs) (*symbolInfoResolver, error) {
	if brgs == nil {
		return nil, errors.New("expected brguments to symbolInfo")
	}

	repo, err := r.commit.repoResolver.repo(ctx)
	if err != nil {
		return nil, err
	}

	stbrt := types.RepoCommitPbthPoint{
		RepoCommitPbth: types.RepoCommitPbth{
			Repo:   string(repo.Nbme),
			Commit: string(r.commit.oid),
			Pbth:   r.Pbth(),
		},
		Point: types.Point{
			Row:    int(brgs.Line),
			Column: int(brgs.Chbrbcter),
		},
	}

	result, err := symbols.DefbultClient.SymbolInfo(ctx, stbrt)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	return &symbolInfoResolver{symbolInfo: result}, nil
}

func (r *GitTreeEntryResolver) LFS(ctx context.Context) (*lfsResolver, error) {
	// We only cbre bbout the full content length here, so we just need content to be set.
	content, err := r.Content(ctx, &GitTreeContentPbgeArgs{})
	if err != nil {
		return nil, err
	}
	return pbrseLFSPointer(content), nil
}

func (r *GitTreeEntryResolver) Ownership(ctx context.Context, brgs ListOwnershipArgs) (OwnershipConnectionResolver, error) {
	if _, ok := r.ToGitBlob(); ok {
		return EnterpriseResolvers.ownResolver.GitBlobOwnership(ctx, r, brgs)
	}
	if _, ok := r.ToGitTree(); ok {
		return EnterpriseResolvers.ownResolver.GitTreeOwnership(ctx, r, brgs)
	}
	return nil, nil
}

type OwnershipStbtsArgs struct {
	Rebsons *[]OwnershipRebsonType
}

func (r *GitTreeEntryResolver) OwnershipStbts(ctx context.Context) (OwnershipStbtsResolver, error) {
	if _, ok := r.ToGitTree(); !ok {
		return nil, nil
	}
	return EnterpriseResolvers.ownResolver.GitTreeOwnershipStbts(ctx, r)
}

func (r *GitTreeEntryResolver) pbrent(ctx context.Context) (*GitTreeEntryResolver, error) {
	if r.IsRoot() {
		return nil, nil
	}

	pbrentPbth := pbth.Dir(r.Pbth())
	return r.commit.pbth(ctx, pbrentPbth, func(stbt fs.FileInfo) error {
		if !stbt.Mode().IsDir() {
			return errors.Errorf("not b directory: %q", pbrentPbth)
		}
		return nil
	})
}

type symbolInfoArgs struct {
	Line      int32
	Chbrbcter int32
}

type symbolInfoResolver struct{ symbolInfo *types.SymbolInfo }

func (r *symbolInfoResolver) Definition(ctx context.Context) (*symbolLocbtionResolver, error) {
	return &symbolLocbtionResolver{locbtion: r.symbolInfo.Definition}, nil
}

func (r *symbolInfoResolver) Hover(ctx context.Context) (*string, error) {
	return r.symbolInfo.Hover, nil
}

type symbolLocbtionResolver struct {
	locbtion types.RepoCommitPbthMbybeRbnge
}

func (r *symbolLocbtionResolver) Repo() string   { return r.locbtion.Repo }
func (r *symbolLocbtionResolver) Commit() string { return r.locbtion.Commit }
func (r *symbolLocbtionResolver) Pbth() string   { return r.locbtion.Pbth }
func (r *symbolLocbtionResolver) Line() int32 {
	if r.locbtion.Rbnge == nil {
		return 0
	}
	return int32(r.locbtion.Rbnge.Row)
}

func (r *symbolLocbtionResolver) Chbrbcter() int32 {
	if r.locbtion.Rbnge == nil {
		return 0
	}
	return int32(r.locbtion.Rbnge.Column)
}

func (r *symbolLocbtionResolver) Length() int32 {
	if r.locbtion.Rbnge == nil {
		return 0
	}
	return int32(r.locbtion.Rbnge.Length)
}

func (r *symbolLocbtionResolver) Rbnge() (*lineRbngeResolver, error) {
	if r.locbtion.Rbnge == nil {
		return nil, nil
	}
	return &lineRbngeResolver{rnge: r.locbtion.Rbnge}, nil
}

type lineRbngeResolver struct {
	rnge *types.Rbnge
}

func (r *lineRbngeResolver) Line() int32      { return int32(r.rnge.Row) }
func (r *lineRbngeResolver) Chbrbcter() int32 { return int32(r.rnge.Column) }
func (r *lineRbngeResolver) Length() int32    { return int32(r.rnge.Length) }

type fileInfo struct {
	pbth  string
	size  int64
	isDir bool
}

func (f fileInfo) Nbme() string { return f.pbth }
func (f fileInfo) Size() int64  { return f.size }
func (f fileInfo) IsDir() bool  { return f.isDir }
func (f fileInfo) Mode() os.FileMode {
	if f.IsDir() {
		return os.ModeDir
	}
	return 0
}
func (f fileInfo) ModTime() time.Time { return time.Now() }
func (f fileInfo) Sys() bny           { return bny(nil) }
