import { Completion, insertCompletionText } from '@codemirror/autocomplete'
import { EditorView } from '@codemirror/view'

import { StandardSuggestionSource } from '@sourcegraph/search-ui'
import { stringHuman } from '@sourcegraph/shared/src/search/query/printer'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { RecentSearch } from '@sourcegraph/shared/src/settings/temporary/recentSearches'

export function searchHistorySource({
    recentSearches,
    selectedSearchContext,
    onSelection,
}: {
    recentSearches: RecentSearch[] | undefined
    selectedSearchContext?: string
    onSelection: (index: number) => void
}): StandardSuggestionSource {
    return (_context, tokens) => {
        if (tokens.length > 0) {
            return null
        }

        // If there are no tokens we must be at position 0
        try {
            if (!recentSearches || recentSearches.length === 0) {
                return null
            }

            const createApplyCompletion = (index: number) => (
                view: EditorView,
                completion: Completion,
                from: number,
                to: number
            ) => {
                onSelection(index)
                view.dispatch(insertCompletionText(view.state, completion.label, from, to))
            }

            return {
                from: 0,
                filter: false,
                options: recentSearches
                    .map(
                        (search): Completion => {
                            let query = search.query

                            {
                                const result = scanSearchQuery(search.query)
                                if (result.type === 'success') {
                                    query = stringHuman(
                                        result.term.filter(term => {
                                            switch (term.type) {
                                                case 'filter':
                                                    if (
                                                        term.field.value === 'context' &&
                                                        term.value?.value === selectedSearchContext
                                                    ) {
                                                        return false
                                                    }
                                                    return true
                                                default:
                                                    return true
                                            }
                                        })
                                    )
                                }
                                // TODO: filter out invalid searches
                            }

                            return {
                                label: query.trim(),
                                type: 'searchhistory',
                            }
                        }
                    )
                    .filter(completion => completion.label.trim() !== '')
                    .map((completion, index) => {
                        // This is here not in the .map call above so we can use
                        // the correct index after filtering out empty entries
                        completion.apply = createApplyCompletion(index)
                        return completion
                    }),
            }
        } catch {
            return null
        }
    }
}
