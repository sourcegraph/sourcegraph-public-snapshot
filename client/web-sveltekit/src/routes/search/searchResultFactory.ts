import type { ComponentType, SvelteComponent } from 'svelte'

import type { SearchMatch } from '$lib/shared'

import CommitSearchResult from './CommitSearchResult.svelte'
import FileContentSearchResult from './FileContentSearchResult.svelte'
import FilePathSearchResult from './FilePathSearchResult.svelte'
import PersonSearchResult from './PersonSearchResult.svelte'
import RepoSearchResult from './RepoSearchResult.svelte'
import SymbolSearchResult from './SymbolSearchResult.svelte'
import TeamSearchResult from './TeamSearchResult.svelte'

type SearchMatchType = SearchMatch['type']

type SearchResultComponent<T extends SearchMatch> = ComponentType<SvelteComponent<{ result: Extract<SearchMatch, T> }>>

type SearchResultUIMap = {
    readonly [type in SearchMatchType]: SearchResultComponent<Extract<SearchMatch, { type: type }>>
}

const searchResultComponents: SearchResultUIMap = {
    repo: RepoSearchResult,
    symbol: SymbolSearchResult,
    content: FileContentSearchResult,
    path: FilePathSearchResult,
    person: PersonSearchResult,
    team: TeamSearchResult,
    commit: CommitSearchResult,
}

export function getSearchResultComponent<T extends SearchMatchType>(result: {
    type: T
}): SearchResultComponent<Extract<SearchMatch, { type: T }>> {
    return searchResultComponents[result.type]
}
