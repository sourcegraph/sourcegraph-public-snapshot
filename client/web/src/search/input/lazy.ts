import { RefObject, useMemo } from 'react'

import type { Extension } from '@codemirror/state'
import { from, of } from 'rxjs'

// The experimental search input should be shown on the search home page
// eslint-disable-next-line no-restricted-imports
import type { Source } from '@sourcegraph/branded/src/search-ui/experimental'
import type { RecentSearch } from '@sourcegraph/shared/src/settings/temporary/recentSearches'
import { useObservable } from '@sourcegraph/wildcard'

import type { SuggestionsSourceConfig } from './suggestions'

export function useLazyCreateSuggestions(enabled: boolean, parameters: SuggestionsSourceConfig): Source | undefined {
    const suggestionsModule = useObservable(
        useMemo(() => (enabled ? from(import('./suggestions')) : of(undefined)), [enabled])
    )

    return useMemo(
        () => (suggestionsModule ? suggestionsModule.createSuggestionsSource(parameters) : undefined),
        [suggestionsModule, parameters]
    )
}

export function useLazyHistoryExtension(
    enabled: boolean,
    recentSearchesRef: RefObject<RecentSearch[] | undefined>,
    submitRef: RefObject<({ query }: { query: string }) => void>
): Extension {
    const historyModule = useObservable(
        useMemo(
            () =>
                enabled
                    ? from(import('@sourcegraph/branded/src/search-ui/input/experimental/codemirror/history'))
                    : of(undefined),
            [enabled]
        )
    )

    return useMemo(
        () =>
            historyModule
                ? historyModule.searchHistoryExtension({
                      mode: {
                          name: 'History',
                          placeholder: 'Filter history',
                      },
                      source: () => recentSearchesRef.current ?? [],
                      submitQuery: query => submitRef.current?.({ query }),
                  })
                : [],
        [historyModule, recentSearchesRef, submitRef]
    )
}
