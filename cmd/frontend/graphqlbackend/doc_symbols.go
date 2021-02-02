package graphqlbackend

import (
	"context"
	"log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type docSymbolsArgs struct {
	LSIFSymbolsArgs
	First *int
}

func (r *GitTreeEntryResolver) DocSymbols(ctx context.Context, args *docSymbolsArgs) (*DocSymbolConnectionResolver, error) {
	// MARK

	lsifResolver, err := r.LSIF(ctx, &struct{ ToolName *string }{})
	if err != nil {
		return nil, err
	}
	symbolsConnection, err := lsifResolver.Symbols(ctx, &args.LSIFSymbolsArgs)
	if err != nil {
		return nil, err
	}
	log.Printf("# symbolsConnection %T", symbolsConnection)
	return nil, nil
}

type DocSymbolConnectionResolver struct {
	first   *int32
	symbols []*docSymbolResolver
}

func (r *DocSymbolConnectionResolver) Nodes(ctx context.Context) ([]*docSymbolResolver, error) {
	return nil, nil
}

func (r *DocSymbolConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return nil, nil
}

type docSymbolResolver struct{}

func (r *docSymbolResolver) Name(ctx context.Context) (string, error) {
	return "", nil
}

func (r *docSymbolResolver) Children(ctx context.Context) ([]*docSymbolResolver, error) {
	return nil, nil
}
