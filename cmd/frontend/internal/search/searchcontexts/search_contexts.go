package searchcontexts

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

const (
	GlobalSearchContextName = "global"
	searchContextSpecPrefix = "@"
)

var namespacedSearchContextSpecRegexp = regexp.MustCompile(searchContextSpecPrefix + `(.*?)\/(.*)`)

func ResolveSearchContextSpec(ctx context.Context, db dbutil.DB, searchContextSpec string) (*types.SearchContext, error) {
	if IsGlobalSearchContextSpec(searchContextSpec) {
		return GetGlobalSearchContext(), nil
	} else if submatches := namespacedSearchContextSpecRegexp.FindStringSubmatch(searchContextSpec); len(submatches) == 3 {
		// We expect 3 submatches, because FindStringSubmatch returns entire string as first submatch, and 2 captured groups
		// as additional submatches
		namespaceName, searchContextName := submatches[1], submatches[2]
		namespace, err := database.Namespaces(db).GetByName(ctx, namespaceName)
		if err != nil {
			return nil, err
		}
		return database.SearchContexts(db).GetSearchContext(ctx, database.GetSearchContextOptions{
			Name:            searchContextName,
			NamespaceUserID: namespace.User,
			NamespaceOrgID:  namespace.Organization,
		})
	} else if strings.HasPrefix(searchContextSpec, searchContextSpecPrefix) {
		namespaceName := searchContextSpec[1:]
		namespace, err := database.Namespaces(db).GetByName(ctx, namespaceName)
		if err != nil {
			return nil, err
		}
		if namespace.User == 0 {
			return nil, fmt.Errorf("search context '%s' not found", searchContextSpec)
		}
		return GetUserSearchContext(namespaceName, namespace.User), nil
	}
	// Check if instance-level context
	searchContext, err := database.SearchContexts(db).GetSearchContext(ctx, database.GetSearchContextOptions{Name: searchContextSpec})
	if err != nil {
		return nil, err
	}
	return searchContext, nil
}

func CreateSearchContextWithRepositoryRevisions(ctx context.Context, db dbutil.DB, searchContext *types.SearchContext, repositoryRevisions []*types.SearchContextRepositoryRevisions) (*types.SearchContext, error) {
	if IsGlobalSearchContext(searchContext) {
		return nil, errors.New("cannot override global search context")
	}

	searchContext, err := database.SearchContexts(db).CreateSearchContextWithRepositoryRevisions(ctx, searchContext, repositoryRevisions)
	if err != nil {
		return nil, err
	}
	return searchContext, nil
}

func GetAutoDefinedSearchContexts(ctx context.Context, db dbutil.DB) ([]*types.SearchContext, error) {
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

func GetRepositoryRevisions(ctx context.Context, db dbutil.DB, searchContextID int64) ([]*search.RepositoryRevisions, error) {
	searchContextRepositoryRevisions, err := database.SearchContexts(db).GetSearchContextRepositoryRevisions(ctx, searchContextID)
	if err != nil {
		return nil, err
	}

	repositoryRevisions := make([]*search.RepositoryRevisions, 0, len(searchContextRepositoryRevisions))
	for _, searchContextRepositoryRevision := range searchContextRepositoryRevisions {
		revisionSpecs := make([]search.RevisionSpecifier, 0, len(searchContextRepositoryRevision.Revisions))
		for _, revision := range searchContextRepositoryRevision.Revisions {
			revisionSpecs = append(revisionSpecs, search.RevisionSpecifier{RevSpec: revision})
		}
		repositoryRevisions = append(repositoryRevisions, &search.RepositoryRevisions{Repo: searchContextRepositoryRevision.Repo, Revs: revisionSpecs})
	}
	return repositoryRevisions, nil
}

func IsAutoDefinedSearchContext(searchContext *types.SearchContext) bool {
	return searchContext.ID == 0
}

func IsInstanceLevelSearchContext(searchContext *types.SearchContext) bool {
	return searchContext.NamespaceUserID == 0 && searchContext.NamespaceOrgID == 0
}

func IsGlobalSearchContextSpec(searchContextSpec string) bool {
	// Empty search context spec resolves to global search context
	return searchContextSpec == "" || searchContextSpec == GlobalSearchContextName
}

func IsGlobalSearchContext(searchContext *types.SearchContext) bool {
	return searchContext != nil && searchContext.Name == GlobalSearchContextName
}

func GetUserSearchContext(name string, userID int32) *types.SearchContext {
	return &types.SearchContext{Name: name, Public: true, Description: "Your repositories on Sourcegraph", NamespaceUserID: userID}
}

func GetGlobalSearchContext() *types.SearchContext {
	return &types.SearchContext{Name: GlobalSearchContextName, Public: true, Description: "All repositories on Sourcegraph"}
}

func GetSearchContextSpec(searchContext *types.SearchContext) string {
	if IsInstanceLevelSearchContext(searchContext) {
		return searchContext.Name
	} else if IsAutoDefinedSearchContext(searchContext) {
		return searchContextSpecPrefix + searchContext.Name
	} else {
		var namespaceName string
		if searchContext.NamespaceUserName != "" {
			namespaceName = searchContext.NamespaceUserName
		} else {
			namespaceName = searchContext.NamespaceOrgName
		}
		return searchContextSpecPrefix + namespaceName + "/" + searchContext.Name
	}
}

func getVersionContextRepositoryRevisions(ctx context.Context, db dbutil.DB, versionContext *schema.VersionContext) ([]*types.SearchContextRepositoryRevisions, error) {
	repositoriesToRevisions := map[string][]string{}
	for _, revision := range versionContext.Revisions {
		repositoriesToRevisions[revision.Repo] = append(repositoriesToRevisions[revision.Repo], revision.Rev)
	}

	repositories := make([]string, 0, len(repositoriesToRevisions))
	for repo := range repositoriesToRevisions {
		repositories = append(repositories, repo)
	}

	repositoryNames, err := database.Repos(db).ListRepoNames(ctx, database.ReposListOptions{Names: repositories})
	if err != nil {
		return nil, err
	}

	repositoryRevisions := make([]*types.SearchContextRepositoryRevisions, len(repositoryNames))
	for idx, repositoryName := range repositoryNames {
		revisions := repositoriesToRevisions[string(repositoryName.Name)]
		repositoryRevisions[idx] = &types.SearchContextRepositoryRevisions{Repo: repositoryName, Revisions: revisions}
	}

	return repositoryRevisions, nil
}

func getSearchContextFromVersionContext(versionContext *schema.VersionContext) *types.SearchContext {
	searchContextName := regexp.MustCompile(`\s+`).ReplaceAllString(versionContext.Name, "_")
	return &types.SearchContext{Name: searchContextName, Description: versionContext.Description, Public: true}
}

func ConvertVersionContextToSearchContext(ctx context.Context, db dbutil.DB, versionContext *schema.VersionContext) (*types.SearchContext, error) {
	repositoryRevisions, err := getVersionContextRepositoryRevisions(ctx, db, versionContext)
	if err != nil {
		return nil, err
	}

	searchContext, err := CreateSearchContextWithRepositoryRevisions(
		ctx,
		db,
		getSearchContextFromVersionContext(versionContext),
		repositoryRevisions,
	)
	if err != nil {
		return nil, err
	}

	return searchContext, nil
}
