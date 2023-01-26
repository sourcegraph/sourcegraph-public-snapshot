import { RepositoryScopeInput } from '@sourcegraph/shared/src/graphql-operations'

/**
 * Returns parsed by string repositories list.
 *
 * @param rawRepositories - string with repositories split by commas
 */
export function getSanitizedRepositories(rawRepositories: string): string[] {
    return rawRepositories
        .trim()
        .split(/\s*,\s*/)
        .filter(repo => repo)
}

/**
 * Returns parsed by string repositories list.
 *
 * @param rawRepositories - string with repositories split by commas
 */
export function getSanitizedRepositoryScope(
    rawRepositories: string,
    repoQuery: string | undefined,
    repoMode: string
): RepositoryScopeInput {
    return {
        repositories: repoMode === 'urls-list' ? getSanitizedRepositories(rawRepositories) : [],
        repositoryCriteria: repoMode === 'search-query' ? repoQuery : null,
    }
}
