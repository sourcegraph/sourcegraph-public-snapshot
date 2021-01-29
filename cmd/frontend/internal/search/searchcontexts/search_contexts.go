package searchcontexts

import (
	"context"
	"errors"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

const globalSearchContextName = "global"

func ResolveSearchContextSpec(ctx context.Context, searchContextSpec string) (*types.SearchContext, error) {
	// Empty search context spec resolves to global search context
	if searchContextSpec == "" || searchContextSpec == globalSearchContextName {
		return GetGlobalSearchContext(), nil
	} else if len(searchContextSpec) > 0 && searchContextSpec[:1] == "@" {
		name := searchContextSpec[1:]
		namespace, err := database.GlobalNamespaces.GetByName(ctx, name)
		if err != nil {
			return nil, err
		}
		if namespace.User == 0 {
			return nil, errors.New("search context not found")
		}
		return &types.SearchContext{Name: name, UserID: &namespace.User}, nil
	}
	return nil, errors.New("search context spec does not have the correct format")
}

func GetGlobalSearchContext() *types.SearchContext {
	return &types.SearchContext{Name: globalSearchContextName}
}

func IsGlobalSearchContext(sc *types.SearchContext) bool {
	return sc != nil && sc.Name == globalSearchContextName
}
