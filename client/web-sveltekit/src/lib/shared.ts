// We want to limit the number of imported modules as much as possible

export { parseRepoRevision, buildSearchURLQuery, makeRepoURI } from '@sourcegraph/shared/src/util/url'
export {
    isCloneInProgressErrorLike,
    isRepoSeeOtherErrorLike,
    isRepoNotFoundErrorLike,
} from '@sourcegraph/shared/src/backend/errors'
export { SectionID as SearchSidebarSectionID } from '@sourcegraph/shared/src/settings/temporary/searchSidebar'
export { TemporarySettingsStorage } from '@sourcegraph/shared/src/settings/temporary/TemporarySettingsStorage'
export {
    type ContentMatch,
    type Skipped,
    getFileMatchUrl,
    getRepositoryUrl,
    aggregateStreamingSearch,
    LATEST_VERSION,
    type AggregateStreamingSearchResults,
    type StreamSearchOptions,
    type SearchMatch,
    getRepoMatchLabel,
    getRepoMatchUrl,
    type RepositoryMatch,
    type SymbolMatch,
    type Progress,
} from '@sourcegraph/shared/src/search/stream'
export type {
    MatchItem,
    MatchGroupMatch,
    MatchGroup,
} from '@sourcegraph/shared/src/components/ranking/PerFileResultRanking'
export { ZoektRanking } from '@sourcegraph/shared/src/components/ranking/ZoektRanking'
export type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
export { filterExists } from '@sourcegraph/shared/src/search/query/validate'
export { FilterType } from '@sourcegraph/shared/src/search/query/filters'
export { getGlobalSearchContextFilter } from '@sourcegraph/shared/src/search/query/query'
export { omitFilter } from '@sourcegraph/shared/src/search/query/transformer'
export { observeSystemIsLightTheme } from '@sourcegraph/shared/src/theme'
export type { PlatformContext } from '@sourcegraph/shared/src/platform/context'
export { type SettingsCascade, type SettingsSubject, gqlToCascade } from '@sourcegraph/shared/src/settings/settings'
export { fetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
export { QueryChangeSource, type QueryState } from '@sourcegraph/shared/src/search/helpers'
export { migrateLocalStorageToTemporarySettings } from '@sourcegraph/shared/src/settings/temporary/migrateLocalStorageToTemporarySettings'
export type { TemporarySettings } from '@sourcegraph/shared/src/settings/temporary/TemporarySettings'
export { SyntaxKind } from '@sourcegraph/shared/src/codeintel/scip'

// Copies of non-reusable code

// Currently defined in client/shared/src/components/RepoLink.tsx

/**
 * Returns the friendly display form of the repository name (e.g., removing "github.com/").
 */
export function displayRepoName(repoName: string): string {
    let parts = repoName.split('/')
    if (parts.length >= 3 && parts[0].includes('.')) {
        parts = parts.slice(1) // remove hostname from repo name (reduce visual noise)
    }
    return parts.join('/')
}

/**
 * Splits the repository name into the dir and base components.
 */
export function splitPath(path: string): [string, string] {
    const components = path.split('/')
    return [components.slice(0, -1).join('/'), components[components.length - 1]]
}
