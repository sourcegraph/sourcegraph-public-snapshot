import type { Extension } from '@codemirror/state'
import { EditorView, type ViewUpdate } from '@codemirror/view'
import type { NavigateFunction } from 'react-router-dom'
import type { Observable } from 'rxjs'

import { createCancelableFetchSuggestions } from '@sourcegraph/shared/src/search/query/providers-utils'
import type { SearchMatch } from '@sourcegraph/shared/src/search/stream'

import {
    createDefaultSuggestionSources,
    type DefaultSuggestionSourcesOptions,
    searchQueryAutocompletion,
    type StandardSuggestionSource,
} from './completion'
import { loadingIndicator } from './loading-indicator'

export { tokenAt, tokens } from './parsedQuery'
export { placeholder } from './placeholder'

export { createDefaultSuggestionSources, searchQueryAutocompletion }
export type { StandardSuggestionSource }

/**
 * Creates an extension that calls the provided callback whenever the editor
 * content has changed.
 */
export const changeListener = (callback: (value: string) => void): Extension =>
    EditorView.updateListener.of((update: ViewUpdate) => {
        if (update.docChanged) {
            return callback(update.state.sliceDoc())
        }
    })

interface CreateDefaultSuggestionsOptions extends Omit<DefaultSuggestionSourcesOptions, 'fetchSuggestions'> {
    fetchSuggestions: (query: string) => Observable<SearchMatch[]>
    navigate?: NavigateFunction
}

/**
 * Creates a search query suggestions extension with default suggestion sources
 * and cancable requests.
 */
export const createDefaultSuggestions = ({
    isSourcegraphDotCom,
    fetchSuggestions,
    disableFilterCompletion,
    disableSymbolCompletion,
    navigate,
    showWhenEmpty,
}: CreateDefaultSuggestionsOptions): Extension => [
    searchQueryAutocompletion(
        createDefaultSuggestionSources({
            fetchSuggestions: createCancelableFetchSuggestions(fetchSuggestions),
            isSourcegraphDotCom,
            disableSymbolCompletion,
            disableFilterCompletion,
            showWhenEmpty,
        }),
        navigate
    ),
    loadingIndicator(),
]
