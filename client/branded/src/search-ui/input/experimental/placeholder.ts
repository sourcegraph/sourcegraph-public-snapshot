// We want to show a placeholder also if the query only contains a context

import type { EditorState } from '@codemirror/state'
import { inRange } from 'lodash'

import { getTokenLength } from '@sourcegraph/shared/src/search/query/utils'

import { tokens } from '../codemirror/parsedQuery'

// filter.
export function showWhenEmptyWithoutContext(state: EditorState): boolean {
    // Show placeholder when empty
    if (state.doc.length === 0) {
        return true
    }

    const queryTokens = tokens(state)

    if (queryTokens.length > 2) {
        return false
    }
    // Only show the placeholder if the cursor is at the end of the content
    if (state.selection.main.from !== state.doc.length) {
        return false
    }

    // If there are two tokens, only show the placeholder if the second one is a
    // whitespace of length 1
    if (queryTokens.length === 2 && (queryTokens[1].type !== 'whitespace' || getTokenLength(queryTokens[1]) !== 1)) {
        return false
    }

    return (
        queryTokens.length > 0 &&
        queryTokens[0].type === 'filter' &&
        queryTokens[0].field.value === 'context' &&
        !inRange(state.selection.main.from, queryTokens[0].range.start, queryTokens[0].range.end + 1)
    )
}
