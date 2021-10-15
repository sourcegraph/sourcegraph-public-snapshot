import * as H from 'history'

import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { CaseSensitivityProps, PatternTypeProps, SearchContextProps } from '@sourcegraph/shared/src/search'
import { CharacterRange } from '@sourcegraph/shared/src/search/query/token'

export enum QueryChangeSource {
    /**
     * When the user has typed in the query or selected a suggestion.
     * Prevents fetching/showing suggestions on every component update.
     */
    userInput,
    searchReference,
    searchTypes,
}

/**
 * The search query and additional information depending on how the query was
 * changed. See MonacoQueryInput for how this data is applied to the editor.
 */
export type QueryState =
    | {
          /** Used to know how a change comes to be. This needs to be defined as
           * optional so that unknown sources can make changes. */
          changeSource?: QueryChangeSource.userInput
          query: string
      }
    | {
          /** Changes from the search side bar */
          changeSource: QueryChangeSource.searchReference | QueryChangeSource.searchTypes
          query: string
          /** The query input will apply this selection */
          selectionRange: CharacterRange
          /** Ensure that newly added or updated filters are completely visible in
           * the query input. */
          revealRange: CharacterRange
          /** Whether or not to trigger the completion popover. The popover is
           * triggered at the end of the selection. */
          showSuggestions?: boolean
      }

export interface SubmitSearchParameters
    extends Partial<Pick<ActivationProps, 'activation'>>,
        Pick<PatternTypeProps, 'patternType'>,
        Pick<CaseSensitivityProps, 'caseSensitive'>,
        Pick<SearchContextProps, 'selectedSearchContextSpec'> {
    history: H.History
    query: string
    source:
        | 'home'
        | 'nav'
        | 'repo'
        | 'tree'
        | 'filter'
        | 'type'
        | 'scopePage'
        | 'communitySearchContextPage'
        | 'excludedResults'
    searchParameters?: { key: string; value: string }[]
}

export interface SubmitSearchProps {
    submitSearch: (parameters: Partial<Omit<SubmitSearchParameters, 'query'>>) => void
}

/**
 * Toggles the given search scope by adding or removing it from the current
 * user query string.
 *
 * @param query The current user query.
 * @param searchFilter The search scope (sub query) or dynamic filter to toggle (add/remove) from the current user query.
 * @returns The new query.
 */
export function toggleSubquery(query: string, searchFilter: string): string {
    const index = queryIndexOfScope(query, searchFilter)
    if (index === -1) {
        // Scope doesn't exist in search query, so add it now.
        return [query.trim(), searchFilter].filter(string => string).join(' ') + ' '
    }

    // Scope exists in the search query, so remove it now.
    return (query.slice(0, index).trim() + ' ' + query.slice(index + searchFilter.length).trim()).trim()
}

/**
 * Returns the index that a given search scope occurs in a given search query.
 * Attempts to not match a scope that is a substring of another scope.
 *
 * @param query The full query string
 * @param scope A scope (sub query) that is searched for in `query`
 * @returns The index in `query`, or `-1` if not found
 */
export function queryIndexOfScope(query: string, scope: string): number {
    let index = 0
    while (true) {
        index = query.indexOf(scope, index)
        if (index === -1) {
            break
        }

        const boundAtStart = index === 0 || query[index - 1] === ' '
        const boundAtEnd = index + scope.length === query.length || query[index + scope.length] === ' '

        // prevent matching scopes that are substrings of other scopes
        if (!boundAtStart || !boundAtEnd) {
            index = index + 1
        } else {
            break
        }
    }
    return index
}
