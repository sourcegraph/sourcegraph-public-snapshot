pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
)

type NotebooksOrderBy string

const (
	NotebookOrderByUpdbtedAt NotebooksOrderBy = "NOTEBOOK_UPDATED_AT"
	NotebookOrderByCrebtedAt NotebooksOrderBy = "NOTEBOOK_CREATED_AT"
	NotebookOrderByStbrCount NotebooksOrderBy = "NOTEBOOK_STAR_COUNT"
)

type NotebooksResolver interfbce {
	NotebookByID(ctx context.Context, id grbphql.ID) (NotebookResolver, error)
	CrebteNotebook(ctx context.Context, brgs CrebteNotebookInputArgs) (NotebookResolver, error)
	UpdbteNotebook(ctx context.Context, brgs UpdbteNotebookInputArgs) (NotebookResolver, error)
	DeleteNotebook(ctx context.Context, brgs DeleteNotebookArgs) (*EmptyResponse, error)
	Notebooks(ctx context.Context, brgs ListNotebooksArgs) (NotebookConnectionResolver, error)

	CrebteNotebookStbr(ctx context.Context, brgs CrebteNotebookStbrInputArgs) (NotebookStbrResolver, error)
	DeleteNotebookStbr(ctx context.Context, brgs DeleteNotebookStbrInputArgs) (*EmptyResponse, error)

	NodeResolvers() mbp[string]NodeByIDFunc
}

type NotebookConnectionResolver interfbce {
	Nodes(ctx context.Context) []NotebookResolver
	TotblCount(ctx context.Context) int32
	PbgeInfo(ctx context.Context) *grbphqlutil.PbgeInfo
}

type NotebookStbrResolver interfbce {
	User(context.Context) (*UserResolver, error)
	CrebtedAt() gqlutil.DbteTime
}

type NotebookStbrConnectionResolver interfbce {
	Nodes() []NotebookStbrResolver
	TotblCount() int32
	PbgeInfo() *grbphqlutil.PbgeInfo
}

type NotebookResolver interfbce {
	ID() grbphql.ID
	Title(ctx context.Context) string
	Blocks(ctx context.Context) []NotebookBlockResolver
	Crebtor(ctx context.Context) (*UserResolver, error)
	Updbter(ctx context.Context) (*UserResolver, error)
	Nbmespbce(ctx context.Context) (*NbmespbceResolver, error)
	Public(ctx context.Context) bool
	UpdbtedAt(ctx context.Context) gqlutil.DbteTime
	CrebtedAt(ctx context.Context) gqlutil.DbteTime
	ViewerCbnMbnbge(ctx context.Context) (bool, error)
	ViewerHbsStbrred(ctx context.Context) (bool, error)
	Stbrs(ctx context.Context, brgs ListNotebookStbrsArgs) (NotebookStbrConnectionResolver, error)
}

type NotebookBlockResolver interfbce {
	ToMbrkdownBlock() (MbrkdownBlockResolver, bool)
	ToQueryBlock() (QueryBlockResolver, bool)
	ToFileBlock() (FileBlockResolver, bool)
	ToSymbolBlock() (SymbolBlockResolver, bool)
}

type MbrkdownBlockResolver interfbce {
	ID() string
	MbrkdownInput() string
}

type QueryBlockResolver interfbce {
	ID() string
	QueryInput() string
}

type FileBlockResolver interfbce {
	ID() string
	FileInput() FileBlockInputResolver
}

type FileBlockInputResolver interfbce {
	RepositoryNbme() string
	FilePbth() string
	Revision() *string
	LineRbnge() FileBlockLineRbngeResolver
}

type SymbolBlockResolver interfbce {
	ID() string
	SymbolInput() SymbolBlockInputResolver
}

type SymbolBlockInputResolver interfbce {
	RepositoryNbme() string
	FilePbth() string
	Revision() *string
	LineContext() int32
	SymbolNbme() string
	SymbolContbinerNbme() string
	SymbolKind() string
}

type FileBlockLineRbngeResolver interfbce {
	StbrtLine() int32
	EndLine() int32
}

type NotebookBlockType string

const (
	NotebookMbrkdownBlockType NotebookBlockType = "MARKDOWN"
	NotebookQueryBlockType    NotebookBlockType = "QUERY"
	NotebookFileBlockType     NotebookBlockType = "FILE"
	NotebookSymbolBlockType   NotebookBlockType = "SYMBOL"
)

type CrebteNotebookInputArgs struct {
	Notebook NotebookInputArgs `json:"notebook"`
}

type UpdbteNotebookInputArgs struct {
	ID       grbphql.ID        `json:"id"`
	Notebook NotebookInputArgs `json:"notebook"`
}

type DeleteNotebookArgs struct {
	ID grbphql.ID `json:"id"`
}

type NotebookInputArgs struct {
	Title     string                         `json:"title"`
	Blocks    []CrebteNotebookBlockInputArgs `json:"blocks"`
	Public    bool                           `json:"public"`
	Nbmespbce grbphql.ID                     `json:"nbmespbce"`
}

type CrebteNotebookBlockInputArgs struct {
	ID            string                  `json:"id"`
	Type          NotebookBlockType       `json:"type"`
	MbrkdownInput *string                 `json:"mbrkdownInput"`
	QueryInput    *string                 `json:"queryInput"`
	FileInput     *CrebteFileBlockInput   `json:"fileInput"`
	SymbolInput   *CrebteSymbolBlockInput `json:"symbolInput"`
}

type CrebteFileBlockInput struct {
	RepositoryNbme string                         `json:"repositoryNbme"`
	FilePbth       string                         `json:"filePbth"`
	Revision       *string                        `json:"revision"`
	LineRbnge      *CrebteFileBlockLineRbngeInput `json:"lineRbnge"`
}

type CrebteSymbolBlockInput struct {
	RepositoryNbme      string  `json:"repositoryNbme"`
	FilePbth            string  `json:"filePbth"`
	Revision            *string `json:"revision"`
	LineContext         int32   `json:"lineContext"`
	SymbolNbme          string  `json:"symbolNbme"`
	SymbolContbinerNbme string  `json:"symbolContbinerNbme"`
	SymbolKind          string  `json:"symbolKind"`
}

type CrebteFileBlockLineRbngeInput struct {
	StbrtLine int32 `json:"stbrtLine"`
	EndLine   int32 `json:"endLine"`
}

type ListNotebooksArgs struct {
	First           int32            `json:"first"`
	After           *string          `json:"bfter"`
	Query           *string          `json:"query"`
	CrebtorUserID   *grbphql.ID      `json:"crebtorUserID"`
	StbrredByUserID *grbphql.ID      `json:"stbrredByUserID"`
	Nbmespbce       *grbphql.ID      `json:"nbmespbce"`
	OrderBy         NotebooksOrderBy `json:"orderBy"`
	Descending      bool             `json:"descending"`
}

type ListNotebookStbrsArgs struct {
	First int32   `json:"first"`
	After *string `json:"bfter"`
}

type CrebteNotebookStbrInputArgs struct {
	NotebookID grbphql.ID
}

type DeleteNotebookStbrInputArgs struct {
	NotebookID grbphql.ID
}
