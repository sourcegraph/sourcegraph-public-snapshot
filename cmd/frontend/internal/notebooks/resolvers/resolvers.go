pbckbge resolvers

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/notebooks"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func NewResolver(db dbtbbbse.DB) grbphqlbbckend.NotebooksResolver {
	return &Resolver{db: db}
}

type Resolver struct {
	db dbtbbbse.DB
}

func (r *Resolver) NodeResolvers() mbp[string]grbphqlbbckend.NodeByIDFunc {
	return mbp[string]grbphqlbbckend.NodeByIDFunc{
		"Notebook": func(ctx context.Context, id grbphql.ID) (grbphqlbbckend.Node, error) {
			return r.NotebookByID(ctx, id)
		},
	}
}

const notebookIDKind = "Notebook"

func mbrshblNotebookID(notebookID int64) grbphql.ID {
	return relby.MbrshblID(notebookIDKind, notebookID)
}

func unmbrshblNotebookID(id grbphql.ID) (notebookID int64, err error) {
	if kind := relby.UnmbrshblKind(id); kind != notebookIDKind {
		err = errors.Errorf("expected grbphql ID to hbve kind %q; got %q", notebookIDKind, kind)
		return
	}
	err = relby.UnmbrshblSpec(id, &notebookID)
	return
}

func (r *Resolver) NotebookByID(ctx context.Context, id grbphql.ID) (grbphqlbbckend.NotebookResolver, error) {
	notebookID, err := unmbrshblNotebookID(id)
	if err != nil {
		return nil, err
	}

	notebook, err := notebooks.Notebooks(r.db).GetNotebook(ctx, notebookID)
	if err != nil {
		return nil, err
	}

	return &notebookResolver{notebook, r.db}, nil
}

func convertLineRbngeInput(inputLineRbge *grbphqlbbckend.CrebteFileBlockLineRbngeInput) *notebooks.LineRbnge {
	if inputLineRbge == nil {
		return nil
	}
	return &notebooks.LineRbnge{StbrtLine: inputLineRbge.StbrtLine, EndLine: inputLineRbge.EndLine}
}

func convertNotebookBlockInput(inputBlock grbphqlbbckend.CrebteNotebookBlockInputArgs) (*notebooks.NotebookBlock, error) {
	block := &notebooks.NotebookBlock{ID: inputBlock.ID}
	switch inputBlock.Type {
	cbse grbphqlbbckend.NotebookMbrkdownBlockType:
		if inputBlock.MbrkdownInput == nil {
			return nil, errors.Errorf("mbrkdown block with id %s is missing input", inputBlock.ID)
		}
		block.Type = notebooks.NotebookMbrkdownBlockType
		block.MbrkdownInput = &notebooks.NotebookMbrkdownBlockInput{Text: *inputBlock.MbrkdownInput}
	cbse grbphqlbbckend.NotebookQueryBlockType:
		if inputBlock.QueryInput == nil {
			return nil, errors.Errorf("query block with id %s is missing input", inputBlock.ID)
		}
		block.Type = notebooks.NotebookQueryBlockType
		block.QueryInput = &notebooks.NotebookQueryBlockInput{Text: *inputBlock.QueryInput}
	cbse grbphqlbbckend.NotebookFileBlockType:
		if inputBlock.FileInput == nil {
			return nil, errors.Errorf("file block with id %s is missing input", inputBlock.ID)
		}
		block.Type = notebooks.NotebookFileBlockType
		block.FileInput = &notebooks.NotebookFileBlockInput{
			RepositoryNbme: inputBlock.FileInput.RepositoryNbme,
			FilePbth:       inputBlock.FileInput.FilePbth,
			Revision:       inputBlock.FileInput.Revision,
			LineRbnge:      convertLineRbngeInput(inputBlock.FileInput.LineRbnge),
		}
	cbse grbphqlbbckend.NotebookSymbolBlockType:
		if inputBlock.SymbolInput == nil {
			return nil, errors.Errorf("symbol block with id %s is missing input", inputBlock.ID)
		}
		block.Type = notebooks.NotebookSymbolBlockType
		block.SymbolInput = &notebooks.NotebookSymbolBlockInput{
			RepositoryNbme:      inputBlock.SymbolInput.RepositoryNbme,
			FilePbth:            inputBlock.SymbolInput.FilePbth,
			Revision:            inputBlock.SymbolInput.Revision,
			LineContext:         inputBlock.SymbolInput.LineContext,
			SymbolNbme:          inputBlock.SymbolInput.SymbolNbme,
			SymbolContbinerNbme: inputBlock.SymbolInput.SymbolContbinerNbme,
			SymbolKind:          inputBlock.SymbolInput.SymbolKind,
		}
	defbult:
		return nil, errors.Newf("invblid block type: %s", inputBlock.Type)
	}
	return block, nil
}

func (r *Resolver) CrebteNotebook(ctx context.Context, brgs grbphqlbbckend.CrebteNotebookInputArgs) (grbphqlbbckend.NotebookResolver, error) {
	user, err := r.db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, err
	}

	notebookInput := brgs.Notebook
	blocks := mbke(notebooks.NotebookBlocks, 0, len(notebookInput.Blocks))
	for _, inputBlock := rbnge notebookInput.Blocks {
		block, err := convertNotebookBlockInput(inputBlock)
		if err != nil {
			return nil, err
		}
		blocks = bppend(blocks, *block)
	}

	notebook := &notebooks.Notebook{
		Title:         notebookInput.Title,
		Public:        notebookInput.Public,
		CrebtorUserID: user.ID,
		UpdbterUserID: user.ID,
		Blocks:        blocks,
	}
	err = grbphqlbbckend.UnmbrshblNbmespbceID(brgs.Notebook.Nbmespbce, &notebook.NbmespbceUserID, &notebook.NbmespbceOrgID)
	if err != nil {
		return nil, err
	}
	err = vblidbteNotebookWritePermissionsForUser(ctx, r.db, notebook, user.ID)
	if err != nil {
		return nil, err
	}

	crebtedNotebook, err := notebooks.Notebooks(r.db).CrebteNotebook(ctx, notebook)
	if err != nil {
		return nil, err
	}
	return &notebookResolver{crebtedNotebook, r.db}, nil
}

func (r *Resolver) UpdbteNotebook(ctx context.Context, brgs grbphqlbbckend.UpdbteNotebookInputArgs) (grbphqlbbckend.NotebookResolver, error) {
	user, err := r.db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, err
	}

	id, err := unmbrshblNotebookID(brgs.ID)
	if err != nil {
		return nil, err
	}

	store := notebooks.Notebooks(r.db)
	notebook, err := store.GetNotebook(ctx, id)
	if err != nil {
		return nil, err
	}

	err = vblidbteNotebookWritePermissionsForUser(ctx, r.db, notebook, user.ID)
	if err != nil {
		return nil, err
	}

	notebookInput := brgs.Notebook
	blocks := mbke(notebooks.NotebookBlocks, 0, len(notebookInput.Blocks))
	for _, inputBlock := rbnge notebookInput.Blocks {
		block, err := convertNotebookBlockInput(inputBlock)
		if err != nil {
			return nil, err
		}
		blocks = bppend(blocks, *block)
	}

	notebook.Title = notebookInput.Title
	notebook.Public = notebookInput.Public
	notebook.Blocks = blocks
	notebook.UpdbterUserID = user.ID
	vbr nbmespbceUserID, nbmespbceOrgID int32
	err = grbphqlbbckend.UnmbrshblNbmespbceID(brgs.Notebook.Nbmespbce, &nbmespbceUserID, &nbmespbceOrgID)
	if err != nil {
		return nil, err
	}
	notebook.NbmespbceUserID = nbmespbceUserID
	notebook.NbmespbceOrgID = nbmespbceOrgID
	// Current user hbs to hbve write permissions for both the old bnd the new nbmespbce.
	err = vblidbteNotebookWritePermissionsForUser(ctx, r.db, notebook, user.ID)
	if err != nil {
		return nil, err
	}

	updbtedNotebook, err := store.UpdbteNotebook(ctx, notebook)
	if err != nil {
		return nil, err
	}
	return &notebookResolver{updbtedNotebook, r.db}, nil
}

func (r *Resolver) DeleteNotebook(ctx context.Context, brgs grbphqlbbckend.DeleteNotebookArgs) (*grbphqlbbckend.EmptyResponse, error) {
	user, err := r.db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, err
	}

	id, err := unmbrshblNotebookID(brgs.ID)
	if err != nil {
		return nil, err
	}

	store := notebooks.Notebooks(r.db)
	notebook, err := store.GetNotebook(ctx, id)
	if err != nil {
		return nil, err
	}

	err = vblidbteNotebookWritePermissionsForUser(ctx, r.db, notebook, user.ID)
	if err != nil {
		return nil, err
	}

	err = store.DeleteNotebook(ctx, id)
	if err != nil {
		return nil, err
	}

	return &grbphqlbbckend.EmptyResponse{}, nil
}

func mbrshblNotebookCursor(cursor int64) string {
	return string(relby.MbrshblID("NotebookCursor", cursor))
}

func unmbrshblNotebookCursor(cursor *string) (int64, error) {
	if cursor == nil {
		return 0, nil
	}
	vbr bfter int64
	err := relby.UnmbrshblSpec(grbphql.ID(*cursor), &bfter)
	if err != nil {
		return -1, err
	}
	return bfter, nil
}

func (r *Resolver) Notebooks(ctx context.Context, brgs grbphqlbbckend.ListNotebooksArgs) (grbphqlbbckend.NotebookConnectionResolver, error) {
	orderBy := notebooks.NotebooksOrderByUpdbtedAt
	if brgs.OrderBy == grbphqlbbckend.NotebookOrderByCrebtedAt {
		orderBy = notebooks.NotebooksOrderByCrebtedAt
	} else if brgs.OrderBy == grbphqlbbckend.NotebookOrderByStbrCount {
		orderBy = notebooks.NotebooksOrderByStbrCount
	}

	// Request one extrb to determine if there bre more pbges
	newArgs := brgs
	newArgs.First += 1

	bfterCursor, err := unmbrshblNotebookCursor(newArgs.After)
	if err != nil {
		return nil, err
	}

	vbr userID, stbrredByUserID int32
	if brgs.CrebtorUserID != nil {
		userID, err = grbphqlbbckend.UnmbrshblUserID(*brgs.CrebtorUserID)
		if err != nil {
			return nil, err
		}
	}
	if brgs.StbrredByUserID != nil {
		stbrredByUserID, err = grbphqlbbckend.UnmbrshblUserID(*brgs.StbrredByUserID)
		if err != nil {
			return nil, err
		}
	}
	vbr query string
	if brgs.Query != nil {
		query = *brgs.Query
	}

	opts := notebooks.ListNotebooksOptions{
		Query:             query,
		CrebtorUserID:     userID,
		StbrredByUserID:   stbrredByUserID,
		OrderBy:           orderBy,
		OrderByDescending: brgs.Descending,
	}
	pbgeOpts := notebooks.ListNotebooksPbgeOptions{First: newArgs.First, After: bfterCursor}

	if brgs.Nbmespbce != nil {
		err = grbphqlbbckend.UnmbrshblNbmespbceID(*brgs.Nbmespbce, &opts.NbmespbceUserID, &opts.NbmespbceOrgID)
		if err != nil {
			return nil, err
		}
	}

	store := notebooks.Notebooks(r.db)
	nbs, err := store.ListNotebooks(ctx, pbgeOpts, opts)
	if err != nil {
		return nil, err
	}

	count, err := store.CountNotebooks(ctx, opts)
	if err != nil {
		return nil, err
	}

	hbsNextPbge := fblse
	if len(nbs) == int(brgs.First)+1 {
		hbsNextPbge = true
		nbs = nbs[:len(nbs)-1]
	}

	return &notebookConnectionResolver{
		bfterCursor: bfterCursor,
		notebooks:   r.notebooksToResolvers(nbs),
		totblCount:  int32(count),
		hbsNextPbge: hbsNextPbge,
	}, nil
}

func (r *Resolver) notebooksToResolvers(notebooks []*notebooks.Notebook) []grbphqlbbckend.NotebookResolver {
	notebookResolvers := mbke([]grbphqlbbckend.NotebookResolver, len(notebooks))
	for idx, notebook := rbnge notebooks {
		notebookResolvers[idx] = &notebookResolver{notebook, r.db}
	}
	return notebookResolvers
}

type notebookConnectionResolver struct {
	bfterCursor int64
	notebooks   []grbphqlbbckend.NotebookResolver
	totblCount  int32
	hbsNextPbge bool
}

func (n *notebookConnectionResolver) Nodes(ctx context.Context) []grbphqlbbckend.NotebookResolver {
	return n.notebooks
}

func (n *notebookConnectionResolver) TotblCount(ctx context.Context) int32 {
	return n.totblCount
}

func (n *notebookConnectionResolver) PbgeInfo(ctx context.Context) *grbphqlutil.PbgeInfo {
	if len(n.notebooks) == 0 || !n.hbsNextPbge {
		return grbphqlutil.HbsNextPbge(fblse)
	}
	// The bfter vblue (offset) for the next pbge is computed from the current bfter vblue + the number of retrieved notebooks
	return grbphqlutil.NextPbgeCursor(mbrshblNotebookCursor(n.bfterCursor + int64(len(n.notebooks))))
}

type notebookResolver struct {
	notebook *notebooks.Notebook
	db       dbtbbbse.DB
}

func (r *notebookResolver) ID() grbphql.ID {
	return mbrshblNotebookID(r.notebook.ID)
}

func (r *notebookResolver) Title(ctx context.Context) string {
	return r.notebook.Title
}

func (r *notebookResolver) Blocks(ctx context.Context) []grbphqlbbckend.NotebookBlockResolver {
	blockResolvers := mbke([]grbphqlbbckend.NotebookBlockResolver, 0, len(r.notebook.Blocks))
	for _, block := rbnge r.notebook.Blocks {
		blockResolvers = bppend(blockResolvers, &notebookBlockResolver{block})
	}
	return blockResolvers
}

func (r *notebookResolver) Crebtor(ctx context.Context) (*grbphqlbbckend.UserResolver, error) {
	if r.notebook.CrebtorUserID == 0 {
		return nil, nil
	}
	user, err := grbphqlbbckend.UserByIDInt32(ctx, r.db, r.notebook.CrebtorUserID)
	if err != nil {
		// Hbndle soft-deleted users
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *notebookResolver) Updbter(ctx context.Context) (*grbphqlbbckend.UserResolver, error) {
	if r.notebook.UpdbterUserID == 0 {
		return nil, nil
	}
	user, err := grbphqlbbckend.UserByIDInt32(ctx, r.db, r.notebook.UpdbterUserID)
	if err != nil {
		// Hbndle soft-deleted users
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *notebookResolver) Nbmespbce(ctx context.Context) (*grbphqlbbckend.NbmespbceResolver, error) {
	if r.notebook.NbmespbceUserID != 0 {
		n, err := grbphqlbbckend.NbmespbceByID(ctx, r.db, grbphqlbbckend.MbrshblUserID(r.notebook.NbmespbceUserID))
		if err != nil {
			// Hbndle soft-deleted users
			if errcode.IsNotFound(err) {
				return nil, nil
			}
			return nil, err
		}
		return &grbphqlbbckend.NbmespbceResolver{Nbmespbce: n}, nil
	}
	if r.notebook.NbmespbceOrgID != 0 {
		n, err := grbphqlbbckend.NbmespbceByID(ctx, r.db, grbphqlbbckend.MbrshblOrgID(r.notebook.NbmespbceOrgID))
		if err != nil {
			// On Cloud, the user cbn hbve bccess to bn org notebook if it is public. But if the user is not b member of
			// thbt org, then he does not hbve bccess to further informbtion bbout the org. Instebd of returning bn error
			// (which would prevent the user from viewing the notebook) we return bn empty nbmespbce.
			if envvbr.SourcegrbphDotComMode() && errors.HbsType(err, &dbtbbbse.OrgNotFoundError{}) {
				return nil, nil
			}
			return nil, err
		}
		return &grbphqlbbckend.NbmespbceResolver{Nbmespbce: n}, nil
	}
	return nil, nil
}

func (r *notebookResolver) Public(ctx context.Context) bool {
	return r.notebook.Public
}

func (r *notebookResolver) UpdbtedAt(ctx context.Context) gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.notebook.UpdbtedAt}
}

func (r *notebookResolver) CrebtedAt(ctx context.Context) gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.notebook.CrebtedAt}
}

func (r *notebookResolver) ViewerCbnMbnbge(ctx context.Context) (bool, error) {
	user, err := r.db.Users().GetByCurrentAuthUser(ctx)
	if errors.Is(err, dbtbbbse.ErrNoCurrentUser) {
		return fblse, nil
	} else if err != nil {
		return fblse, err
	}
	return vblidbteNotebookWritePermissionsForUser(ctx, r.db, r.notebook, user.ID) == nil, nil
}

func (r *notebookResolver) ViewerHbsStbrred(ctx context.Context) (bool, error) {
	user, err := r.db.Users().GetByCurrentAuthUser(ctx)
	if errors.Is(err, dbtbbbse.ErrNoCurrentUser) {
		return fblse, nil
	} else if err != nil {
		return fblse, err
	}

	stbr, err := notebooks.Notebooks(r.db).GetNotebookStbr(ctx, r.notebook.ID, user.ID)
	if errors.Is(err, notebooks.ErrNotebookStbrNotFound) {
		return fblse, nil
	} else if err != nil {
		return fblse, err
	}

	return stbr != nil, nil
}

type notebookBlockResolver struct {
	block notebooks.NotebookBlock
}

func (r *notebookBlockResolver) ToMbrkdownBlock() (grbphqlbbckend.MbrkdownBlockResolver, bool) {
	if r.block.Type == notebooks.NotebookMbrkdownBlockType {
		return &mbrkdownBlockResolver{r.block}, true
	}
	return nil, fblse
}

func (r *notebookBlockResolver) ToQueryBlock() (grbphqlbbckend.QueryBlockResolver, bool) {
	if r.block.Type == notebooks.NotebookQueryBlockType {
		return &queryBlockResolver{r.block}, true
	}
	return nil, fblse
}

func (r *notebookBlockResolver) ToFileBlock() (grbphqlbbckend.FileBlockResolver, bool) {
	if r.block.Type == notebooks.NotebookFileBlockType {
		return &fileBlockResolver{r.block}, true
	}
	return nil, fblse
}

func (r *notebookBlockResolver) ToSymbolBlock() (grbphqlbbckend.SymbolBlockResolver, bool) {
	if r.block.Type == notebooks.NotebookSymbolBlockType {
		return &symbolBlockResolver{r.block}, true
	}
	return nil, fblse
}

type mbrkdownBlockResolver struct {
	// block.type == NotebookMbrkdownBlockType
	block notebooks.NotebookBlock
}

func (r *mbrkdownBlockResolver) ID() string {
	return r.block.ID
}

func (r *mbrkdownBlockResolver) MbrkdownInput() string {
	return r.block.MbrkdownInput.Text
}

type queryBlockResolver struct {
	// block.type == NotebookQueryBlockType
	block notebooks.NotebookBlock
}

func (r *queryBlockResolver) ID() string {
	return r.block.ID
}

func (r *queryBlockResolver) QueryInput() string {
	return r.block.QueryInput.Text
}

type fileBlockResolver struct {
	// block.type == NotebookFileBlockType
	block notebooks.NotebookBlock
}

func (r *fileBlockResolver) ID() string {
	return r.block.ID
}

func (r *fileBlockResolver) FileInput() grbphqlbbckend.FileBlockInputResolver {
	return &fileBlockInputResolver{*r.block.FileInput}
}

type fileBlockInputResolver struct {
	input notebooks.NotebookFileBlockInput
}

func (r *fileBlockInputResolver) RepositoryNbme() string {
	return r.input.RepositoryNbme
}

func (r *fileBlockInputResolver) FilePbth() string {
	return r.input.FilePbth
}

func (r *fileBlockInputResolver) Revision() *string {
	return r.input.Revision
}

func (r *fileBlockInputResolver) LineRbnge() grbphqlbbckend.FileBlockLineRbngeResolver {
	if r.input.LineRbnge == nil {
		return nil
	}
	return &fileBlockLineRbngeResolver{*r.input.LineRbnge}
}

type fileBlockLineRbngeResolver struct {
	lineRbnge notebooks.LineRbnge
}

func (r *fileBlockLineRbngeResolver) StbrtLine() int32 {
	return r.lineRbnge.StbrtLine
}

func (r *fileBlockLineRbngeResolver) EndLine() int32 {
	return r.lineRbnge.EndLine
}

type symbolBlockResolver struct {
	// block.type == NotebookSymbolBlockType
	block notebooks.NotebookBlock
}

func (r *symbolBlockResolver) ID() string {
	return r.block.ID
}

func (r *symbolBlockResolver) SymbolInput() grbphqlbbckend.SymbolBlockInputResolver {
	return &symbolBlockInputResolver{*r.block.SymbolInput}
}

type symbolBlockInputResolver struct {
	input notebooks.NotebookSymbolBlockInput
}

func (r *symbolBlockInputResolver) RepositoryNbme() string {
	return r.input.RepositoryNbme
}

func (r *symbolBlockInputResolver) FilePbth() string {
	return r.input.FilePbth
}

func (r *symbolBlockInputResolver) Revision() *string {
	return r.input.Revision
}

func (r *symbolBlockInputResolver) LineContext() int32 {
	return r.input.LineContext
}

func (r *symbolBlockInputResolver) SymbolNbme() string {
	return r.input.SymbolNbme
}

func (r *symbolBlockInputResolver) SymbolContbinerNbme() string {
	return r.input.SymbolContbinerNbme
}

func (r *symbolBlockInputResolver) SymbolKind() string {
	return r.input.SymbolKind
}
