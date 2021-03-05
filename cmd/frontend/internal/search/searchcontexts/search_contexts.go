package searchcontexts

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

const (
	GlobalSearchContextName = "global"
	searchContextSpecPrefix = "@"
)

type getNamespaceByNameFunc func(ctx context.Context, name string) (*database.Namespace, error)

func ResolveSearchContextSpec(ctx context.Context, searchContextSpec string, getNamespaceByName getNamespaceByNameFunc) (*types.SearchContext, error) {
	if !envvar.SourcegraphDotComMode() {
		return nil, nil
	}

	if IsGlobalSearchContextSpec(searchContextSpec) {
		return GetGlobalSearchContext(), nil
	} else if strings.HasPrefix(searchContextSpec, searchContextSpecPrefix) {
		name := searchContextSpec[1:]
		namespace, err := getNamespaceByName(ctx, name)
		if err != nil {
			return nil, err
		}
		if namespace.User == 0 {
			return nil, fmt.Errorf("search context '%s' not found", searchContextSpec)
		}
		return GetUserSearchContext(name, namespace.User), nil
	}
	return nil, fmt.Errorf("search context '%s' does not have the correct format (global or @username)", searchContextSpec)
}

func GetUsersSearchContexts(ctx context.Context) ([]*types.SearchContext, error) {
	searchContexts := []*types.SearchContext{GetGlobalSearchContext()}
	a := actor.FromContext(ctx)
	if a.IsAuthenticated() {
		user, err := database.GlobalUsers.GetByID(ctx, a.UID)
		if err != nil {
			return nil, err
		}
		searchContexts = append(searchContexts, GetUserSearchContext(user.Username, a.UID))
	}
	return searchContexts, nil
}

func IsGlobalSearchContextSpec(searchContextSpec string) bool {
	// Empty search context spec resolves to global search context
	return searchContextSpec == "" || searchContextSpec == GlobalSearchContextName
}

func IsGlobalSearchContext(searchContext *types.SearchContext) bool {
	return searchContext != nil && searchContext.Name == GlobalSearchContextName
}

func GetUserSearchContext(name string, userID int32) *types.SearchContext {
	return &types.SearchContext{Name: name, Description: "Your repositories on Sourcegraph", UserID: userID}
}

func GetGlobalSearchContext() *types.SearchContext {
	return &types.SearchContext{Name: GlobalSearchContextName, Description: "All repositories on Sourcegraph"}
}

func GetSearchContextSpec(searchContext *types.SearchContext) string {
	if IsGlobalSearchContext(searchContext) {
		return searchContext.Name
	}
	return searchContextSpecPrefix + searchContext.Name
}
