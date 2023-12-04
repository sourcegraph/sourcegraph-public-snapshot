// We want to limit the number of imported modules as much as possible

export type { AbsoluteRepoFile } from '@sourcegraph/shared/src/util/url'

export {
    parseRepoRevision,
    parseQueryAndHash,
    buildSearchURLQuery,
    makeRepoURI,
    type RevisionSpec,
    type ResolvedRevisionSpec,
    type RepoSpec,
} from '@sourcegraph/shared/src/util/url'
export {
    isCloneInProgressErrorLike,
    isRepoSeeOtherErrorLike,
    isRepoNotFoundErrorLike,
    isRevisionNotFoundErrorLike,
    CloneInProgressError,
    RepoNotFoundError,
    RepoSeeOtherError,
    RevisionNotFoundError,
} from '@sourcegraph/shared/src/backend/errors'
export { viewerSettingsQuery } from '@sourcegraph/shared/src/backend/settings'
export { SectionID as SearchSidebarSectionID } from '@sourcegraph/shared/src/settings/temporary/searchSidebar'
export { TemporarySettingsStorage } from '@sourcegraph/shared/src/settings/temporary/TemporarySettingsStorage'
export {
    type Skipped,
    getFileMatchUrl,
    getRepositoryUrl,
    aggregateStreamingSearch,
    LATEST_VERSION,
    type AggregateStreamingSearchResults,
    type StreamSearchOptions,
    getRepoMatchLabel,
    getRepoMatchUrl,
    getMatchUrl,
    type RepositoryMatch,
    type SymbolMatch,
    type PathMatch,
    type ContentMatch,
    type SearchMatch,
    type OwnerMatch,
    type TeamMatch,
    type PersonMatch,
    type CommitMatch,
    type Progress,
    type Range,
} from '@sourcegraph/shared/src/search/stream'
export type {
    MatchItem,
    MatchGroupMatch,
    MatchGroup,
    PerFileResultRanking,
    RankingResult,
} from '@sourcegraph/shared/src/components/ranking/PerFileResultRanking'
export { ZoektRanking } from '@sourcegraph/shared/src/components/ranking/ZoektRanking'
export { LineRanking } from '@sourcegraph/shared/src/components/ranking/LineRanking'
export { type AuthenticatedUser, currentAuthStateQuery } from '@sourcegraph/shared/src/auth'
export { filterExists } from '@sourcegraph/shared/src/search/query/validate'
export { FilterType } from '@sourcegraph/shared/src/search/query/filters'
export { getGlobalSearchContextFilter, findFilter, FilterKind } from '@sourcegraph/shared/src/search/query/query'
export { omitFilter, appendFilter } from '@sourcegraph/shared/src/search/query/transformer'
export {
    type SettingsCascade,
    type SettingsSubject,
    type SettingsCascadeOrError,
    SettingsProvider,
    gqlToCascade,
} from '@sourcegraph/shared/src/settings/settings'
export { fetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
export { QueryChangeSource, type QueryState } from '@sourcegraph/shared/src/search/helpers'
export { migrateLocalStorageToTemporarySettings } from '@sourcegraph/shared/src/settings/temporary/migrateLocalStorageToTemporarySettings'
export type { TemporarySettings } from '@sourcegraph/shared/src/settings/temporary/TemporarySettings'
export { SyntaxKind } from '@sourcegraph/shared/src/codeintel/scip'
export { shortcutDisplayName } from '@sourcegraph/shared/src/keyboardShortcuts'
export { createCodeIntelAPI, type CodeIntelAPI } from '@sourcegraph/shared/src/codeintel/api'
export { getModeFromPath } from '@sourcegraph/shared/src/languages'
export type { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'

// Copies of non-reusable code

// Currently defined in client/shared/src/components/RepoLink.tsx

/**
 * Returns the friendly display form of the repository name (e.g., removing "github.com/").
 */
export function displayRepoName(repoName: string): string {
    let parts = repoName.split('/')
    if (parts.length > 0 && parts[0].includes('.')) {
        parts = parts.slice(1) // remove hostname from repo name (reduce visual noise)
    }
    return parts.join('/')
}

/**
 * Splits the repository name into the dir and base components.
 */
export function splitPath(path: string): [string, string] {
    const components = path.split('/')
    return [components.slice(0, -1).join('/'), components.at(-1) ?? '']
}
