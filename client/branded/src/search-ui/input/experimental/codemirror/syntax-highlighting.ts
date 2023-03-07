import { RangeSetBuilder } from '@codemirror/state'
import { Decoration, EditorView } from '@codemirror/view'

import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { isFilterOfType } from '@sourcegraph/shared/src/search/query/utils'

import { decoratedTokens, queryTokens } from '../../codemirror/parsedQuery'

const contextFilter = Decoration.mark({ class: 'sg-query-token-filter-context', inclusiveEnd: false })

export const filterDecoration = [
    EditorView.baseTheme({
        '.sg-query-token-filter-context': {
            borderRadius: '3px',
            border: '1px solid var(--border-color)',
        },
        '&dark .sg-query-token-filter-context': {
            borderColor: 'var(--border-color-2)',
        },
    }),
    EditorView.decorations.compute([decoratedTokens, 'selection'], state => {
        const query = state.facet(queryTokens)
        const builder = new RangeSetBuilder<Decoration>()
        for (const token of query.tokens) {
            if (token.type === 'filter' && isFilterOfType(token, FilterType.context)) {
                builder.add(token.range.start, token.range.end, contextFilter)
            }
        }
        return builder.finish()
    }),
]
