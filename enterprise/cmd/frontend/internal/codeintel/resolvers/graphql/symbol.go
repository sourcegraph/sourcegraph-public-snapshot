package graphql

import (
	"fmt"
	"strings"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
)

func NewDocSymbolResolver(symbols []*resolvers.AdjustedSymbol, id string, locationResolver *CachedLocationResolver) (gql.DocSymbolResolver, error) {
	foundSymbol := findSymbol(symbols, id)
	if foundSymbol == nil {
		return nil, fmt.Errorf("Failed to find symbol with id %s", id)
	}
	return newDocSymbolResolver(foundSymbol, locationResolver), nil
}

func findSymbol(symbols []*resolvers.AdjustedSymbol, id string) *resolvers.AdjustedSymbol {
	for _, symbol := range symbols {
		if symbol.Identifier == id {
			return symbol
		}
		if !strings.HasPrefix(id, symbol.Identifier) {
			continue
		}
		foundSymbol := findSymbol(symbol.Children, id)
		if foundSymbol != nil {
			return foundSymbol
		}
	}
	return nil
}
