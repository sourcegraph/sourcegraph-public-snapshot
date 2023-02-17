import { RangeSetBuilder } from '@codemirror/state'
import { Decoration, EditorView } from '@codemirror/view'
import inRange from 'lodash/inRange'

import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { isFilterOfType } from '@sourcegraph/shared/src/search/query/utils'

import { decoratedTokens, queryTokens } from '../../codemirror/parsedQuery'

const validFilter = Decoration.mark({ class: 'sg-filter' })
const invalidFilter = Decoration.mark({ class: 'sg-filter sg-invalid-filter' })
const contextFilter = Decoration.mark({ class: 'sg-context-filter', inclusiveEnd: false })

export const filterHighlight = [
    EditorView.baseTheme({
        '.sg-filter': {
            backgroundColor: 'var(--oc-blue-0)',
            borderRadius: '3px',
            padding: '0px',
        },
        '.sg-invalid-filter': {
            backgroundColor: 'var(--oc-red-1)',
            borderColor: 'var(--oc-red-2)',
        },
        '.sg-context-filter': {
            borderRadius: '3px',
            border: '1px solid var(--border-color)',
        },
        '.sg-clear-filter > button': {
            border: 'none',
            backgroundColor: 'transparent',
            padding: 0,
            width: 'var(--icon-inline-size)',
            height: 'var(--icon-inline-size)',
            color: 'var(--icon-color)',
        },
    }),
    EditorView.decorations.compute([decoratedTokens, 'selection'], state => {
        const query = state.facet(queryTokens)
        const builder = new RangeSetBuilder<Decoration>()
        for (const token of query.tokens) {
            if (token.type === 'filter') {
                const withinRange = inRange(state.selection.main.head, token.range.start, token.range.end + 1) // or cursor is within field
                const isValid =
                    token?.value?.value || // has non-empty value
                    token?.value?.quoted || // or is quoted
                    withinRange // or cursor is within field

                // context: filters are styled differnetly
                if (isFilterOfType(token, FilterType.context)) {
                    builder.add(token.range.start, token.range.end, contextFilter)
                } else {
                    // +1 to include the colon (:)
                    builder.add(token.range.start, token.field.range.end + 1, isValid ? validFilter : invalidFilter)
                }
            }
        }
        return builder.finish()
    }),
]
