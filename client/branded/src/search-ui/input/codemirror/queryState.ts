import { EditorSelection } from '@codemirror/state'
import type { EditorView } from '@codemirror/view'

import { EditorHint, QueryChangeSource, type QueryState } from '@sourcegraph/shared/src/search'

export interface UpdateFromQueryStateOptions {
    startCompletion: (view: EditorView) => void
}

/**
 * Update the input's value, selection and cursor depending on how the search query was changed.
 */
export function updateFromQueryState(
    view: EditorView,
    queryState: QueryState,
    { startCompletion }: UpdateFromQueryStateOptions
): void {
    if (queryState.changeSource === QueryChangeSource.userInput) {
        // Don't react to user input
        return
    }

    const changes =
        view.state.sliceDoc() !== queryState.query
            ? { from: 0, to: view.state.doc.length, insert: queryState.query }
            : undefined
    view.dispatch({
        // Update value if it's different
        changes,
        selection: queryState.selectionRange
            ? // Select the specified range (most of the time this will be a
              // placeholder filter value).
              EditorSelection.range(queryState.selectionRange.start, queryState.selectionRange.end)
            : // Place the cursor at the end of the query if it changed.
            changes
            ? EditorSelection.cursor(queryState.query.length)
            : undefined,
        scrollIntoView: true,
    })

    if (queryState.hint) {
        if ((queryState.hint & EditorHint.Focus) === EditorHint.Focus) {
            view.focus()
        }
        if ((queryState.hint & EditorHint.ShowSuggestions) === EditorHint.ShowSuggestions) {
            startCompletion(view)
        }
        if ((queryState.hint & EditorHint.Blur) === EditorHint.Blur) {
            view.contentDOM.blur()
        }
    }
}
