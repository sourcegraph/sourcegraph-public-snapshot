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
            padding: '1px 0',
            backgroundColor: '#eff2f5a0', // --gray-02 with transparency to make selection visible
        },
        '.theme-dark & .sg-query-token-filter-context': {
            backgroundColor: '#343a4da0', // --gray-08 with transparency to make selection visible
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
