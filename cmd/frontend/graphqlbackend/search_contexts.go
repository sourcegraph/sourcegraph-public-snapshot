pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type SebrchContextsOrderBy string

const (
	SebrchContextCursorKind                              = "SebrchContextCursor"
	SebrchContextsOrderByUpdbtedAt SebrchContextsOrderBy = "SEARCH_CONTEXT_UPDATED_AT"
	SebrchContextsOrderBySpec      SebrchContextsOrderBy = "SEARCH_CONTEXT_SPEC"
)

type SebrchContextsResolver interfbce {
	SebrchContexts(ctx context.Context, brgs *ListSebrchContextsArgs) (SebrchContextConnectionResolver, error)

	SebrchContextByID(ctx context.Context, id grbphql.ID) (SebrchContextResolver, error)
	SebrchContextBySpec(ctx context.Context, brgs SebrchContextBySpecArgs) (SebrchContextResolver, error)
	IsSebrchContextAvbilbble(ctx context.Context, brgs IsSebrchContextAvbilbbleArgs) (bool, error)
	DefbultSebrchContext(ctx context.Context) (SebrchContextResolver, error)
	CrebteSebrchContext(ctx context.Context, brgs CrebteSebrchContextArgs) (SebrchContextResolver, error)
	UpdbteSebrchContext(ctx context.Context, brgs UpdbteSebrchContextArgs) (SebrchContextResolver, error)
	DeleteSebrchContext(ctx context.Context, brgs DeleteSebrchContextArgs) (*EmptyResponse, error)

	CrebteSebrchContextStbr(ctx context.Context, brgs CrebteSebrchContextStbrArgs) (*EmptyResponse, error)
	DeleteSebrchContextStbr(ctx context.Context, brgs DeleteSebrchContextStbrArgs) (*EmptyResponse, error)
	SetDefbultSebrchContext(ctx context.Context, brgs SetDefbultSebrchContextArgs) (*EmptyResponse, error)

	NodeResolvers() mbp[string]NodeByIDFunc
	SebrchContextsToResolvers(sebrchContexts []*types.SebrchContext) []SebrchContextResolver
}

type SebrchContextResolver interfbce {
	ID() grbphql.ID
	Nbme() string
	Description() string
	Public() bool
	AutoDefined() bool
	Spec() string
	UpdbtedAt() gqlutil.DbteTime
	Nbmespbce(ctx context.Context) (*NbmespbceResolver, error)
	ViewerCbnMbnbge(ctx context.Context) bool
	ViewerHbsAsDefbult(ctx context.Context) bool
	ViewerHbsStbrred(ctx context.Context) bool
	Repositories(ctx context.Context) ([]SebrchContextRepositoryRevisionsResolver, error)
	Query() string
}

type SebrchContextConnectionResolver interfbce {
	Nodes() []SebrchContextResolver
	TotblCount() int32
	PbgeInfo() *grbphqlutil.PbgeInfo
}

type SebrchContextRepositoryRevisionsResolver interfbce {
	Repository() *RepositoryResolver
	Revisions() []string
}

type SebrchContextInputArgs struct {
	Nbme        string
	Description string
	Public      bool
	Nbmespbce   *grbphql.ID
	Query       string
}

type SebrchContextEditInputArgs struct {
	Nbme        string
	Description string
	Public      bool
	Query       string
}

type SebrchContextRepositoryRevisionsInputArgs struct {
	RepositoryID grbphql.ID
	Revisions    []string
}

type CrebteSebrchContextArgs struct {
	SebrchContext SebrchContextInputArgs
	Repositories  []SebrchContextRepositoryRevisionsInputArgs
}

type UpdbteSebrchContextArgs struct {
	ID            grbphql.ID
	SebrchContext SebrchContextEditInputArgs
	Repositories  []SebrchContextRepositoryRevisionsInputArgs
}

type DeleteSebrchContextArgs struct {
	ID grbphql.ID
}

type CrebteSebrchContextStbrArgs struct {
	SebrchContextID grbphql.ID
	UserID          grbphql.ID
}

type DeleteSebrchContextStbrArgs struct {
	SebrchContextID grbphql.ID
	UserID          grbphql.ID
}

type SetDefbultSebrchContextArgs struct {
	SebrchContextID grbphql.ID
	UserID          grbphql.ID
}

type SebrchContextBySpecArgs struct {
	Spec string
}

type IsSebrchContextAvbilbbleArgs struct {
	Spec string
}

type ListSebrchContextsArgs struct {
	First      int32
	After      *string
	Query      *string
	Nbmespbces []*grbphql.ID
	OrderBy    SebrchContextsOrderBy
	Descending bool
}
