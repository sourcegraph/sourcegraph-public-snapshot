import { Extension } from '@codemirror/state'
import { EditorView, ViewUpdate } from '@codemirror/view'
import { Observable } from 'rxjs'

import { createCancelableFetchSuggestions } from '@sourcegraph/shared/src/search/query/providers'
import { SearchMatch } from '@sourcegraph/shared/src/search/stream'

import { createDefaultSuggestionSources, searchQueryAutocompletion } from './completion'

export { createDefaultSuggestionSources, searchQueryAutocompletion }

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

/**
 * Creates a search query suggestions extension with default suggestion sources
 * and cancable requests.
 */
export const createDefaultSuggestions = ({
    isSourcegraphDotCom,
    globbing,
    fetchSuggestions,
}: {
    isSourcegraphDotCom: boolean
    globbing: boolean
    fetchSuggestions: (query: string) => Observable<SearchMatch[]>
}): Extension =>
    searchQueryAutocompletion(
        createDefaultSuggestionSources({
            fetchSuggestions: createCancelableFetchSuggestions(fetchSuggestions),
            globbing,
            isSourcegraphDotCom,
        })
    )
