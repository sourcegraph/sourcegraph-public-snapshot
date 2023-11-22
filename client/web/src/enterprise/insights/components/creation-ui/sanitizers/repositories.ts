import type { RepositoryScopeInput } from '@sourcegraph/shared/src/graphql-operations'

/**
 * Returns parsed by string repositories list.
 *
 * @param rawRepositories - string with repositories split by commas
 */
export function getSanitizedRepositoryScope(
    rawRepositories: string[],
    repoQuery: string | undefined,
    repoMode: string
): RepositoryScopeInput {
    return {
        repositories: repoMode === 'urls-list' ? rawRepositories : [],
        repositoryCriteria: repoMode === 'search-query' ? repoQuery : null,
    }
}
