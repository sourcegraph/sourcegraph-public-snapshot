import { Extension, StateField } from '@codemirror/state'
import { mdiFilterOutline } from '@mdi/js'
import { inRange } from 'lodash'

import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { FilterKind, findFilter } from '@sourcegraph/shared/src/search/query/query'
import { Filter } from '@sourcegraph/shared/src/search/query/token'
import { isFilterType } from '@sourcegraph/shared/src/search/query/validate'

import { getQueryInformation } from '../../codemirror/parsedQuery'
import { filterValueRenderer } from '../optionRenderer'
import { suggestionSources } from '../suggestionsExtension'

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
