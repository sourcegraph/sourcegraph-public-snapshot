package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/notebooks"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewResolver(db database.DB) graphqlbackend.NotebooksResolver {
	return &Resolver{db: db}
}

type Resolver struct {
	db database.DB
}

func (r *Resolver) NodeResolvers() map[string]graphqlbackend.NodeByIDFunc {
	return map[string]graphqlbackend.NodeByIDFunc{
		"Notebook": func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			return r.NotebookByID(ctx, id)
		},
	}
}

const notebookIDKind = "Notebook"

func marshalNotebookID(notebookID int64) graphql.ID {
	return relay.MarshalID(notebookIDKind, notebookID)
}

func unmarshalNotebookID(id graphql.ID) (notebookID int64, err error) {
	if kind := relay.UnmarshalKind(id); kind != notebookIDKind {
		err = errors.Errorf("expected graphql ID to have kind %q; got %q", notebookIDKind, kind)
		return
	}
	err = relay.UnmarshalSpec(id, &notebookID)
	return
}

func (r *Resolver) NotebookByID(ctx context.Context, id graphql.ID) (graphqlbackend.NotebookResolver, error) {
	notebookID, err := unmarshalNotebookID(id)
	if err != nil {
		return nil, err
	}

	notebook, err := notebooks.Notebooks(r.db).GetNotebook(ctx, notebookID)
	if err != nil {
		return nil, err
	}

	return &notebookResolver{notebook, r.db}, nil
}

func convertLineRangeInput(inputLineRage *graphqlbackend.CreateFileBlockLineRangeInput) *notebooks.LineRange {
	if inputLineRage == nil {
		return nil
	}
	return &notebooks.LineRange{StartLine: inputLineRage.StartLine, EndLine: inputLineRage.EndLine}
}

func convertNotebookBlockInput(inputBlock graphqlbackend.CreateNotebookBlockInputArgs) (*notebooks.NotebookBlock, error) {
	block := &notebooks.NotebookBlock{ID: inputBlock.ID}
	switch inputBlock.Type {
	case graphqlbackend.NotebookMarkdownBlockType:
		if inputBlock.MarkdownInput == nil {
			return nil, errors.Errorf("markdown block with id %s is missing input", inputBlock.ID)
		}
		block.Type = notebooks.NotebookMarkdownBlockType
		block.MarkdownInput = &notebooks.NotebookMarkdownBlockInput{Text: *inputBlock.MarkdownInput}
	case graphqlbackend.NotebookQueryBlockType:
		if inputBlock.QueryInput == nil {
			return nil, errors.Errorf("query block with id %s is missing input", inputBlock.ID)
		}
		block.Type = notebooks.NotebookQueryBlockType
		block.QueryInput = &notebooks.NotebookQueryBlockInput{Text: *inputBlock.QueryInput}
	case graphqlbackend.NotebookFileBlockType:
		if inputBlock.FileInput == nil {
			return nil, errors.Errorf("file block with id %s is missing input", inputBlock.ID)
		}
		block.Type = notebooks.NotebookFileBlockType
		block.FileInput = &notebooks.NotebookFileBlockInput{
			RepositoryName: inputBlock.FileInput.RepositoryName,
			FilePath:       inputBlock.FileInput.FilePath,
			Revision:       inputBlock.FileInput.Revision,
			LineRange:      convertLineRangeInput(inputBlock.FileInput.LineRange),
		}
	case graphqlbackend.NotebookSymbolBlockType:
		if inputBlock.SymbolInput == nil {
			return nil, errors.Errorf("symbol block with id %s is missing input", inputBlock.ID)
		}
		block.Type = notebooks.NotebookSymbolBlockType
		block.SymbolInput = &notebooks.NotebookSymbolBlockInput{
			RepositoryName:      inputBlock.SymbolInput.RepositoryName,
			FilePath:            inputBlock.SymbolInput.FilePath,
			Revision:            inputBlock.SymbolInput.Revision,
			LineContext:         inputBlock.SymbolInput.LineContext,
			SymbolName:          inputBlock.SymbolInput.SymbolName,
			SymbolContainerName: inputBlock.SymbolInput.SymbolContainerName,
			SymbolKind:          inputBlock.SymbolInput.SymbolKind,
		}
	default:
		return nil, errors.Newf("invalid block type: %s", inputBlock.Type)
	}
	return block, nil
}

func (r *Resolver) CreateNotebook(ctx context.Context, args graphqlbackend.CreateNotebookInputArgs) (graphqlbackend.NotebookResolver, error) {
	user, err := r.db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, err
	}

	notebookInput := args.Notebook
	blocks := make(notebooks.NotebookBlocks, 0, len(notebookInput.Blocks))
	for _, inputBlock := range notebookInput.Blocks {
		block, err := convertNotebookBlockInput(inputBlock)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, *block)
	}

	notebook := &notebooks.Notebook{
		Title:         notebookInput.Title,
		Public:        notebookInput.Public,
		CreatorUserID: user.ID,
		UpdaterUserID: user.ID,
		Blocks:        blocks,
	}
	err = graphqlbackend.UnmarshalNamespaceID(args.Notebook.Namespace, &notebook.NamespaceUserID, &notebook.NamespaceOrgID)
	if err != nil {
		return nil, err
	}
	err = validateNotebookWritePermissionsForUser(ctx, r.db, notebook, user.ID)
	if err != nil {
		return nil, err
	}

	createdNotebook, err := notebooks.Notebooks(r.db).CreateNotebook(ctx, notebook)
	if err != nil {
		return nil, err
	}
	return &notebookResolver{createdNotebook, r.db}, nil
}

func (r *Resolver) UpdateNotebook(ctx context.Context, args graphqlbackend.UpdateNotebookInputArgs) (graphqlbackend.NotebookResolver, error) {
	user, err := r.db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, err
	}

	id, err := unmarshalNotebookID(args.ID)
	if err != nil {
		return nil, err
	}

	store := notebooks.Notebooks(r.db)
	notebook, err := store.GetNotebook(ctx, id)
	if err != nil {
		return nil, err
	}

	err = validateNotebookWritePermissionsForUser(ctx, r.db, notebook, user.ID)
	if err != nil {
		return nil, err
	}

	notebookInput := args.Notebook
	blocks := make(notebooks.NotebookBlocks, 0, len(notebookInput.Blocks))
	for _, inputBlock := range notebookInput.Blocks {
		block, err := convertNotebookBlockInput(inputBlock)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, *block)
	}

	notebook.Title = notebookInput.Title
	notebook.Public = notebookInput.Public
	notebook.Blocks = blocks
	notebook.UpdaterUserID = user.ID
	var namespaceUserID, namespaceOrgID int32
	err = graphqlbackend.UnmarshalNamespaceID(args.Notebook.Namespace, &namespaceUserID, &namespaceOrgID)
	if err != nil {
		return nil, err
	}
	notebook.NamespaceUserID = namespaceUserID
	notebook.NamespaceOrgID = namespaceOrgID
	// Current user has to have write permissions for both the old and the new namespace.
	err = validateNotebookWritePermissionsForUser(ctx, r.db, notebook, user.ID)
	if err != nil {
		return nil, err
	}

	updatedNotebook, err := store.UpdateNotebook(ctx, notebook)
	if err != nil {
		return nil, err
	}
	return &notebookResolver{updatedNotebook, r.db}, nil
}

func (r *Resolver) DeleteNotebook(ctx context.Context, args graphqlbackend.DeleteNotebookArgs) (*graphqlbackend.EmptyResponse, error) {
	user, err := r.db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, err
	}

	id, err := unmarshalNotebookID(args.ID)
	if err != nil {
		return nil, err
	}

	store := notebooks.Notebooks(r.db)
	notebook, err := store.GetNotebook(ctx, id)
	if err != nil {
		return nil, err
	}

	err = validateNotebookWritePermissionsForUser(ctx, r.db, notebook, user.ID)
	if err != nil {
		return nil, err
	}

	err = store.DeleteNotebook(ctx, id)
	if err != nil {
		return nil, err
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

func marshalNotebookCursor(cursor int64) string {
	return string(relay.MarshalID("NotebookCursor", cursor))
}

func unmarshalNotebookCursor(cursor *string) (int64, error) {
	if cursor == nil {
		return 0, nil
	}
	var after int64
	err := relay.UnmarshalSpec(graphql.ID(*cursor), &after)
	if err != nil {
		return -1, err
	}
	return after, nil
}

func (r *Resolver) Notebooks(ctx context.Context, args graphqlbackend.ListNotebooksArgs) (graphqlbackend.NotebookConnectionResolver, error) {
	orderBy := notebooks.NotebooksOrderByUpdatedAt
	if args.OrderBy == graphqlbackend.NotebookOrderByCreatedAt {
		orderBy = notebooks.NotebooksOrderByCreatedAt
	} else if args.OrderBy == graphqlbackend.NotebookOrderByStarCount {
		orderBy = notebooks.NotebooksOrderByStarCount
	}

	// Request one extra to determine if there are more pages
	newArgs := args
	newArgs.First += 1

	afterCursor, err := unmarshalNotebookCursor(newArgs.After)
	if err != nil {
		return nil, err
	}

	var userID, starredByUserID int32
	if args.CreatorUserID != nil {
		userID, err = graphqlbackend.UnmarshalUserID(*args.CreatorUserID)
		if err != nil {
			return nil, err
		}
	}
	if args.StarredByUserID != nil {
		starredByUserID, err = graphqlbackend.UnmarshalUserID(*args.StarredByUserID)
		if err != nil {
			return nil, err
		}
	}
	var query string
	if args.Query != nil {
		query = *args.Query
	}

	opts := notebooks.ListNotebooksOptions{
		Query:             query,
		CreatorUserID:     userID,
		StarredByUserID:   starredByUserID,
		OrderBy:           orderBy,
		OrderByDescending: args.Descending,
	}
	pageOpts := notebooks.ListNotebooksPageOptions{First: newArgs.First, After: afterCursor}

	if args.Namespace != nil {
		err = graphqlbackend.UnmarshalNamespaceID(*args.Namespace, &opts.NamespaceUserID, &opts.NamespaceOrgID)
		if err != nil {
			return nil, err
		}
	}

	store := notebooks.Notebooks(r.db)
	nbs, err := store.ListNotebooks(ctx, pageOpts, opts)
	if err != nil {
		return nil, err
	}

	count, err := store.CountNotebooks(ctx, opts)
	if err != nil {
		return nil, err
	}

	hasNextPage := false
	if len(nbs) == int(args.First)+1 {
		hasNextPage = true
		nbs = nbs[:len(nbs)-1]
	}

	return &notebookConnectionResolver{
		afterCursor: afterCursor,
		notebooks:   r.notebooksToResolvers(nbs),
		totalCount:  int32(count),
		hasNextPage: hasNextPage,
	}, nil
}

func (r *Resolver) notebooksToResolvers(notebooks []*notebooks.Notebook) []graphqlbackend.NotebookResolver {
	notebookResolvers := make([]graphqlbackend.NotebookResolver, len(notebooks))
	for idx, notebook := range notebooks {
		notebookResolvers[idx] = &notebookResolver{notebook, r.db}
	}
	return notebookResolvers
}

type notebookConnectionResolver struct {
	afterCursor int64
	notebooks   []graphqlbackend.NotebookResolver
	totalCount  int32
	hasNextPage bool
}

func (n *notebookConnectionResolver) Nodes(ctx context.Context) []graphqlbackend.NotebookResolver {
	return n.notebooks
}

func (n *notebookConnectionResolver) TotalCount(ctx context.Context) int32 {
	return n.totalCount
}

func (n *notebookConnectionResolver) PageInfo(ctx context.Context) *gqlutil.PageInfo {
	if len(n.notebooks) == 0 || !n.hasNextPage {
		return gqlutil.HasNextPage(false)
	}
	// The after value (offset) for the next page is computed from the current after value + the number of retrieved notebooks
	return gqlutil.NextPageCursor(marshalNotebookCursor(n.afterCursor + int64(len(n.notebooks))))
}

type notebookResolver struct {
	notebook *notebooks.Notebook
	db       database.DB
}

func (r *notebookResolver) ID() graphql.ID {
	return marshalNotebookID(r.notebook.ID)
}

func (r *notebookResolver) Title(ctx context.Context) string {
	return r.notebook.Title
}

func (r *notebookResolver) Blocks(ctx context.Context) []graphqlbackend.NotebookBlockResolver {
	blockResolvers := make([]graphqlbackend.NotebookBlockResolver, 0, len(r.notebook.Blocks))
	for _, block := range r.notebook.Blocks {
		blockResolvers = append(blockResolvers, &notebookBlockResolver{block})
	}
	return blockResolvers
}

func (r *notebookResolver) Creator(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	if r.notebook.CreatorUserID == 0 {
		return nil, nil
	}
	user, err := graphqlbackend.UserByIDInt32(ctx, r.db, r.notebook.CreatorUserID)
	if err != nil {
		// Handle soft-deleted users
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *notebookResolver) Updater(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	if r.notebook.UpdaterUserID == 0 {
		return nil, nil
	}
	user, err := graphqlbackend.UserByIDInt32(ctx, r.db, r.notebook.UpdaterUserID)
	if err != nil {
		// Handle soft-deleted users
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *notebookResolver) Namespace(ctx context.Context) (*graphqlbackend.NamespaceResolver, error) {
	if r.notebook.NamespaceUserID != 0 {
		n, err := graphqlbackend.NamespaceByID(ctx, r.db, graphqlbackend.MarshalUserID(r.notebook.NamespaceUserID))
		if err != nil {
			// Handle soft-deleted users
			if errcode.IsNotFound(err) {
				return nil, nil
			}
			return nil, err
		}
		return &graphqlbackend.NamespaceResolver{Namespace: n}, nil
	}
	if r.notebook.NamespaceOrgID != 0 {
		n, err := graphqlbackend.NamespaceByID(ctx, r.db, graphqlbackend.MarshalOrgID(r.notebook.NamespaceOrgID))
		if err != nil {
			// On Cloud, the user can have access to an org notebook if it is public. But if the user is not a member of
			// that org, then he does not have access to further information about the org. Instead of returning an error
			// (which would prevent the user from viewing the notebook) we return an empty namespace.
			if dotcom.SourcegraphDotComMode() && errors.HasType[*database.OrgNotFoundError](err) {
				return nil, nil
			}
			return nil, err
		}
		return &graphqlbackend.NamespaceResolver{Namespace: n}, nil
	}
	return nil, nil
}

func (r *notebookResolver) Public(ctx context.Context) bool {
	return r.notebook.Public
}

func (r *notebookResolver) UpdatedAt(ctx context.Context) gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.notebook.UpdatedAt}
}

func (r *notebookResolver) CreatedAt(ctx context.Context) gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.notebook.CreatedAt}
}

func (r *notebookResolver) ViewerCanManage(ctx context.Context) (bool, error) {
	user, err := r.db.Users().GetByCurrentAuthUser(ctx)
	if errors.Is(err, database.ErrNoCurrentUser) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return validateNotebookWritePermissionsForUser(ctx, r.db, r.notebook, user.ID) == nil, nil
}

func (r *notebookResolver) ViewerHasStarred(ctx context.Context) (bool, error) {
	user, err := r.db.Users().GetByCurrentAuthUser(ctx)
	if errors.Is(err, database.ErrNoCurrentUser) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	star, err := notebooks.Notebooks(r.db).GetNotebookStar(ctx, r.notebook.ID, user.ID)
	if errors.Is(err, notebooks.ErrNotebookStarNotFound) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return star != nil, nil
}

type notebookBlockResolver struct {
	block notebooks.NotebookBlock
}

func (r *notebookBlockResolver) ToMarkdownBlock() (graphqlbackend.MarkdownBlockResolver, bool) {
	if r.block.Type == notebooks.NotebookMarkdownBlockType {
		return &markdownBlockResolver{r.block}, true
	}
	return nil, false
}

func (r *notebookBlockResolver) ToQueryBlock() (graphqlbackend.QueryBlockResolver, bool) {
	if r.block.Type == notebooks.NotebookQueryBlockType {
		return &queryBlockResolver{r.block}, true
	}
	return nil, false
}

func (r *notebookBlockResolver) ToFileBlock() (graphqlbackend.FileBlockResolver, bool) {
	if r.block.Type == notebooks.NotebookFileBlockType {
		return &fileBlockResolver{r.block}, true
	}
	return nil, false
}

func (r *notebookBlockResolver) ToSymbolBlock() (graphqlbackend.SymbolBlockResolver, bool) {
	if r.block.Type == notebooks.NotebookSymbolBlockType {
		return &symbolBlockResolver{r.block}, true
	}
	return nil, false
}

func (r *notebookResolver) PatternType(_ context.Context) string {
	return r.notebook.PatternType
}

type markdownBlockResolver struct {
	// block.type == NotebookMarkdownBlockType
	block notebooks.NotebookBlock
}

func (r *markdownBlockResolver) ID() string {
	return r.block.ID
}

func (r *markdownBlockResolver) MarkdownInput() string {
	return r.block.MarkdownInput.Text
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

func (r *fileBlockResolver) FileInput() graphqlbackend.FileBlockInputResolver {
	return &fileBlockInputResolver{*r.block.FileInput}
}

type fileBlockInputResolver struct {
	input notebooks.NotebookFileBlockInput
}

func (r *fileBlockInputResolver) RepositoryName() string {
	return r.input.RepositoryName
}

func (r *fileBlockInputResolver) FilePath() string {
	return r.input.FilePath
}

func (r *fileBlockInputResolver) Revision() *string {
	return r.input.Revision
}

func (r *fileBlockInputResolver) LineRange() graphqlbackend.FileBlockLineRangeResolver {
	if r.input.LineRange == nil {
		return nil
	}
	return &fileBlockLineRangeResolver{*r.input.LineRange}
}

type fileBlockLineRangeResolver struct {
	lineRange notebooks.LineRange
}

func (r *fileBlockLineRangeResolver) StartLine() int32 {
	return r.lineRange.StartLine
}

func (r *fileBlockLineRangeResolver) EndLine() int32 {
	return r.lineRange.EndLine
}

type symbolBlockResolver struct {
	// block.type == NotebookSymbolBlockType
	block notebooks.NotebookBlock
}

func (r *symbolBlockResolver) ID() string {
	return r.block.ID
}

func (r *symbolBlockResolver) SymbolInput() graphqlbackend.SymbolBlockInputResolver {
	return &symbolBlockInputResolver{*r.block.SymbolInput}
}

type symbolBlockInputResolver struct {
	input notebooks.NotebookSymbolBlockInput
}

func (r *symbolBlockInputResolver) RepositoryName() string {
	return r.input.RepositoryName
}

func (r *symbolBlockInputResolver) FilePath() string {
	return r.input.FilePath
}

func (r *symbolBlockInputResolver) Revision() *string {
	return r.input.Revision
}

func (r *symbolBlockInputResolver) LineContext() int32 {
	return r.input.LineContext
}

func (r *symbolBlockInputResolver) SymbolName() string {
	return r.input.SymbolName
}

func (r *symbolBlockInputResolver) SymbolContainerName() string {
	return r.input.SymbolContainerName
}

func (r *symbolBlockInputResolver) SymbolKind() string {
	return r.input.SymbolKind
}
