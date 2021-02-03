package searchcontexts

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

const globalSearchContextName = "global"

type getNamespaceByNameFunc func(ctx context.Context, name string) (*database.Namespace, error)

func ResolveSearchContextSpec(ctx context.Context, searchContextSpec string, getNamespaceByName getNamespaceByNameFunc) (*types.SearchContext, error) {
	if !envvar.SourcegraphDotComMode() {
		return nil, nil
	}

	if IsGlobalSearchContextSpec(searchContextSpec) {
		return GetGlobalSearchContext(), nil
	} else if strings.HasPrefix(searchContextSpec, "@") {
		name := searchContextSpec[1:]
		namespace, err := getNamespaceByName(ctx, name)
		if err != nil {
			return nil, err
		}
		if namespace.User == 0 {
			return nil, fmt.Errorf("search context '%s' not found", searchContextSpec)
		}
		return &types.SearchContext{Name: name, UserID: namespace.User}, nil
	}
	return nil, fmt.Errorf("search context '%s' does not have the correct format (global or @username)", searchContextSpec)
}

func IsGlobalSearchContextSpec(searchContextSpec string) bool {
	// Empty search context spec resolves to global search context
	return searchContextSpec == "" || searchContextSpec == globalSearchContextName
}

func GetGlobalSearchContext() *types.SearchContext {
	return &types.SearchContext{Name: globalSearchContextName}
}

func IsGlobalSearchContext(searchContext *types.SearchContext) bool {
	return searchContext != nil && searchContext.Name == globalSearchContextName
}
