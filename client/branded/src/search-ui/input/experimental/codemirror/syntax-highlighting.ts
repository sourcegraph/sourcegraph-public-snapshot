import { RangeSetBuilder } from '@codemirror/state'
import { Decoration, EditorView } from '@codemirror/view'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { Token } from '@sourcegraph/shared/src/search/query/token'
import { isFilterOfType } from '@sourcegraph/shared/src/search/query/utils'

import { queryTokens } from '../../codemirror/parsedQuery'

const filter = Decoration.mark({ class: 'sg-query-token sg-query-token-filter' })
const pattern = Decoration.mark({ class: 'sg-query-token sg-query-token-pattern' })

function getDecorationForToken(token: Token): Decoration | null {
    switch (token.type) {
        case 'filter': {
            return filter
        }
        case 'pattern': {
            return pattern
        }
    }
    return null
}

export const filterDecoration = [
    EditorView.baseTheme({
        '.sg-query-token': {
            borderRadius: '3px',
            // We only apply little horizontal padding because it appears that
            // the padding interferes with the cursor position (CodeMirror will
            // place the cursor after the padding, not after the last character,
            // which is surprising to the user).
            padding: '1px 3px',
        },
        '.sg-query-token-pattern': {
            backgroundColor: '#E6EBF295', // --gray-03 with transparency to make text selection visible
        },
        '.sg-query-token-filter': {
            backgroundColor: '#CCEDFFa0', // --oc-blue-1 with transparency to make text selection visible
        },

        '.theme-dark & .sg-query-token-pattern': {
            backgroundColor: '#5E6E8C80', // --gray-08 with transparency to make text selection visible
        },

        '.theme-dark & .sg-query-token-filter': {
            backgroundColor: '#074884a0', // --oc-blue-9 with transparency to make text selection visible
        },
    }),
    EditorView.decorations.compute([queryTokens, 'selection'], state => {
        const query = state.facet(queryTokens)
        const builder = new RangeSetBuilder<Decoration>()
        for (const token of query.tokens) {
            let decoration: Decoration | null = null
            if (query.patternType === SearchPatternType.keyword) {
                decoration = getDecorationForToken(token)
            } else if (token.type === 'filter' && isFilterOfType(token, FilterType.context)) {
                // In non-keyword mode, the context filter is styled the same
                // way as regular patterns.
                decoration = pattern
            }
            if (decoration) {
                builder.add(token.range.start, token.range.end, decoration)
            }
        }
        return builder.finish()
    }),
]
