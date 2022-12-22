import { ChangeSpec, EditorState, Extension } from '@codemirror/state'
import { EditorView, ViewUpdate } from '@codemirror/view'
import * as H from 'history'
import { Observable } from 'rxjs'

import { createCancelableFetchSuggestions } from '@sourcegraph/shared/src/search/query/providers-utils'
import { SearchMatch } from '@sourcegraph/shared/src/search/stream'

import {
    createDefaultSuggestionSources,
    DefaultSuggestionSourcesOptions,
    searchQueryAutocompletion,
    StandardSuggestionSource,
} from './completion'
import { loadingIndicator } from './loading-indicator'
export { tokenAt, tokens } from './parsedQuery'

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

const replacePattern = /[\n\r↵]+/g
/**
 * An extension that enforces that the input will be single line. Consecutive
 * line breaks will be replaces by a single space.
 */
export const singleLine = EditorState.transactionFilter.of(transaction => {
    if (!transaction.docChanged) {
        return transaction
    }

    const newText = transaction.newDoc.sliceString(0)
    const changes: ChangeSpec[] = []

    // new RegExp(...) creates a copy of the regular expression so that we have
    // our own stateful copy for using `exec` below.
    const lineBreakPattern = new RegExp(replacePattern)
    let match: RegExpExecArray | null = null
    while ((match = lineBreakPattern.exec(newText))) {
        // Insert space for line breaks following non-whitespace characters
        if (match.index > 0 && !/\s/.test(newText[match.index - 1])) {
            changes.push({ from: match.index, to: match.index + match[0].length, insert: ' ' })
        } else {
            // Otherwise remove it
            changes.push({ from: match.index, to: match.index + match[0].length })
        }
    }

    return changes.length > 0 ? [transaction, { changes, sequential: true }] : transaction
})

/**
 * Creates a search query suggestions extension with default suggestion sources
 * and cancable requests.
 */
export const createDefaultSuggestions = ({
    isSourcegraphDotCom,
    globbing,
    fetchSuggestions,
    disableFilterCompletion,
    disableSymbolCompletion,
    history,
    applyOnEnter,
    showWhenEmpty,
}: Omit<DefaultSuggestionSourcesOptions, 'fetchSuggestions'> & {
    fetchSuggestions: (query: string) => Observable<SearchMatch[]>
    history?: H.History
    /**
     * Whether or not to allow suggestions selection by Enter key.
     */
    applyOnEnter?: boolean
}): Extension => [
    searchQueryAutocompletion(
        createDefaultSuggestionSources({
            fetchSuggestions: createCancelableFetchSuggestions(fetchSuggestions),
            globbing,
            isSourcegraphDotCom,
            disableSymbolCompletion,
            disableFilterCompletion,
            showWhenEmpty,
            applyOnEnter,
        }),
        history,
        applyOnEnter
    ),
    loadingIndicator(),
]
