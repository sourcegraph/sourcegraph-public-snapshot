import { ChangeSpec, EditorSelection, EditorState, Extension, StateField } from '@codemirror/state'
import { mdiFilterOutline } from '@mdi/js'
import { inRange } from 'lodash'

import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import {
    FilterKind,
    findFilter,
    getGlobalSearchContextFilter,
} from '@sourcegraph/shared/src/search/query/query'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { Filter } from '@sourcegraph/shared/src/search/query/token'
import { isFilterType } from '@sourcegraph/shared/src/search/query/validate'

import { getQueryInformation, tokens } from '../../codemirror/parsedQuery'
import { filterValueRenderer } from '../optionRenderer'
import { suggestionSources } from '../suggestionsExtension'
import { EditorView } from '@codemirror/view'

/**
 * A suggestion extension which will show most recently entered context: filter if the
 * current query doesn't contain a context: filter.
 */
export function lastUsedContextSuggestion(config: { getContext: () => string | undefined }): Extension {
    return [
        lastContextField,
        suggestionSources.of({
            query: (state, position) => {
                const { token, tokens } = getQueryInformation(state, position)
                const context = state.field(lastContextField) || config.getContext()
                if (!context) {
                    return null
                }

                // Only show suggestion if the query is empty or the query does not contain a context filter and
                // the cursor is at a whitespace token
                if (
                    (token && token.type !== 'whitespace') ||
                    tokens.some(token => isFilterType(token, FilterType.context))
                ) {
                    return null
                }

                const label = `context:${context}`
                return {
                    result: [
                        {
                            title: 'Search context',
                            options: [
                                {
                                    label,
                                    icon: mdiFilterOutline,
                                    render: filterValueRenderer,
                                    kind: 'context',
                                    action: {
                                        type: 'completion',
                                        from: position,
                                        insertValue: `${label} `,
                                    },
                                },
                            ],
                        },
                    ],
                }
            },
        }),
    ]
}

function findSearchContext(query: string): Filter | undefined {
    return findFilter(query, FilterType.context, FilterKind.Global)
}

const lastContextField = StateField.define<string | undefined>({
    create(state) {
        return findSearchContext(state.sliceDoc())?.value?.value
    },
    update(value, transaction) {
        // We don't actually need to access the new query state we can just look for the first context: filter
        // in the new document
        if (transaction.docChanged) {
            const searchContextFilter = findSearchContext(transaction.newDoc.sliceString(0))
            if (
                searchContextFilter?.value?.value &&
                // Do not update the field while the user is still editing the filter value
                // (determined by the fact that the cursor is in range of the filter value)
                !inRange(
                    transaction.newSelection.main.from,
                    searchContextFilter.value.range.start - 1,
                    searchContextFilter.value.range.end + 1
                )
            ) {
                return searchContextFilter.value.value
            }
        }
        return value
    },
})

/**
 * Allows the user to overwrite the existing input value when pasting by pressing "Shift"
 */
export function shiftPasteOverwrite() {
    let shiftPressed = false
    return EditorView.domEventHandlers({
        keydown(event) {
            shiftPressed = event.shiftKey
        },
        keyup() {
            shiftPressed = false
        },
        paste(_event, view) {
            if (shiftPressed) {
                // Select the existing value to let the paste event overwrite it
                view.dispatch({
                    selection: EditorSelection.range(0, view.state.doc.length),
                })
            }
        }
    })
}

/**
 * When the user pastes a new value into the input, this extension tries to be smart about
 * using the correct context: filter.
 */
export const overrideContextOnPaste = EditorState.transactionFilter.of(transaction => {
    if (!transaction.isUserEvent('input.paste')) {
        return transaction
    }

    const currentGlobalContext = getGlobalSearchContextFilter(transaction.startState.sliceDoc(0))
    if (!currentGlobalContext) {
        return transaction
    }

    const newValue = transaction.newDoc.sliceString(0)
    const newGlobalContext = getGlobalSearchContextFilter(newValue)
    if (newGlobalContext) {
        // Only a single (global) context: filter present -> nothing to do
        return transaction
    }

    // Common situation: New query is pasted into "empty" input (only contains context: filter)
    // We assume that the pasted query is always "complete" and clear the current input
    if (tokens(transaction.startState).every(token => token.type === 'whitespace' || isFilterType(token, FilterType.context))) {
        return [{changes: {from: 0, to: transaction.startState.doc.length}}, transaction]
    }

    // Less common situation: New query pasted into _non-empty_ query input. We assume that the existing query
    // should be extended and remove the additionaly context filter.
    // CAVEAT: It's valid to have multiple context: filters in different OR branches or to negate a context:
    // filter via NOT. Detecting this properly is tricky and likely not useful in most cases. That's why
    // we bail if the new query contains any keywords.

    const scanResult = scanSearchQuery(newValue)
    if (scanResult.type !== 'success' || scanResult.term.some(token => token.type === 'keyword')) {
        return transaction
    }

    // Also unlikely: User pastes new query before the existing context: filter
    const newRangeOfCurrentContextFilter = {
        start: transaction.changes.mapPos(currentGlobalContext.filter.range.start, 0),
        end: transaction.changes.mapPos(currentGlobalContext.filter.range.end),
    }

    const changes: ChangeSpec[] = []
    for (const token of scanResult.term) {
        if (isFilterType(token, FilterType.context) && token.range.start !== newRangeOfCurrentContextFilter.start) {
             // Trim whitespace after context filter. range.end is exclusive so this is our starting point.
            const match = newValue.slice(currentGlobalContext.filter.range.end).match(/\s+/)
            changes.push({from: token.range.start, to: token.range.end + (match?.[0].length ?? 0)})
        }
    }

    return [transaction, {changes, sequential: true}]
})
