// We want to limit the number of imported modules as much as possible

export {
    parseRepoRevision,
    buildSearchURLQuery,
    makeRepoGitURI,
    toPrettyBlobURL,
    toRepoURL,
    type AbsoluteRepoFile,
    replaceRevisionInURL,
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
    LATEST_VERSION,
    TELEMETRY_FILTER_TYPES,
    getRevision,
    aggregateStreamingSearch,
    emptyAggregateResults,
    getFileMatchUrl,
    getMatchUrl,
    getRepoMatchLabel,
    getRepoMatchUrl,
    getRepositoryUrl,
    streamComputeQuery,
    type AggregateStreamingSearchResults,
    type Alert,
    type ChunkMatch,
    type CommitMatch,
    type ContentMatch,
    type Filter,
    type LineMatch,
    type OwnerMatch,
    type PathMatch,
    type PersonMatch,
    type Progress,
    type Range,
    type RepositoryMatch,
    type SearchEvent,
    type SearchMatch,
    type Skipped,
    type StreamSearchOptions,
    type SymbolMatch,
    type TeamMatch,
} from '@sourcegraph/shared/src/search/stream'
export {
    type MatchItem,
    type MatchGroupMatch,
    type MatchGroup,
    rankPassthrough,
    rankByLine,
    truncateGroups,
} from '@sourcegraph/shared/src/components/ranking/PerFileResultRanking'
export { filterExists } from '@sourcegraph/shared/src/search/query/validate'
export {
    getRelevantTokens,
    type RelevantTokenResult,
    EMPTY_RELEVANT_TOKEN_RESULT,
} from '@sourcegraph/shared/src/search/query/analyze'
export { scanSearchQuery, scanSearchQueryAsPatterns } from '@sourcegraph/shared/src/search/query/scanner'
export { stringHuman } from '@sourcegraph/shared/src/search/query/printer'
export { KeywordKind, PatternKind, type Token } from '@sourcegraph/shared/src/search/query/token'
export { FilterType } from '@sourcegraph/shared/src/search/query/filters'
export { getGlobalSearchContextFilter, findFilter, FilterKind } from '@sourcegraph/shared/src/search/query/query'
export { isFilterOfType } from '@sourcegraph/shared/src/search/query/utils'
export { omitFilter, appendFilter, updateFilter } from '@sourcegraph/shared/src/search/query/transformer'
export type { Settings } from '@sourcegraph/shared/src/settings/settings'
export { fetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
export {
    QueryChangeSource,
    type QueryState,
    TELEMETRY_SEARCH_SOURCE_TYPE,
} from '@sourcegraph/shared/src/search/helpers'
export { migrateLocalStorageToTemporarySettings } from '@sourcegraph/shared/src/settings/temporary/migrateLocalStorageToTemporarySettings'
export type { TemporarySettings } from '@sourcegraph/shared/src/settings/temporary/TemporarySettings'
export { SyntaxKind, Occurrence } from '@sourcegraph/shared/src/codeintel/scip'
export { createCodeIntelAPI, type CodeIntelAPI } from '@sourcegraph/shared/src/codeintel/api'
export { getModeFromPath } from '@sourcegraph/shared/src/languages'
export type { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
export { repositoryInsertText } from '@sourcegraph/shared/src/search/query/completion-utils'
export { ThemeSetting, Theme } from '@sourcegraph/shared/src/theme-types'

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
