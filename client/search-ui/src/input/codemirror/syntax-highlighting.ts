import { RangeSetBuilder } from '@codemirror/state'
import { Decoration, EditorView } from '@codemirror/view'
import inRange from 'lodash/inRange'

import { DecoratedToken, toCSSClassName } from '@sourcegraph/shared/src/search/query/decoratedToken'

import { decoratedTokens, queryTokens } from './parsedQuery'

// Defines decorators for syntax highlighting
const tokenDecorators: { [key: string]: Decoration } = {}

// Chooses the correct decorator for the decorated token
const decoratedToDecoration = (token: DecoratedToken): Decoration => {
    const className = toCSSClassName(token)
    const decorator = tokenDecorators[className]
    return decorator || (tokenDecorators[className] = Decoration.mark({ class: className }))
}

// This provides syntax highlighting. This is a custom solution so that we an
// use our existing query parser (instead of using CodeMirror's language
// support). That's not to say that we couldn't properly integrate with
// CodeMirror's language system with more effort.
export const querySyntaxHighlighting = [
    EditorView.decorations.compute([decoratedTokens], state => {
        const tokens = state.facet(decoratedTokens)
        const builder = new RangeSetBuilder<Decoration>()
        for (const token of tokens) {
            builder.add(token.range.start, token.range.end, decoratedToDecoration(token))
        }
        return builder.finish()
    }),
]

const validFilter = Decoration.mark({ class: 'sg-filter', inclusive: false })
const invalidFilter = Decoration.mark({ class: 'sg-filter sg-invalid-filter', inclusive: false })

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
    }),
    EditorView.decorations.compute([decoratedTokens, 'selection'], state => {
        const query = state.facet(queryTokens)
        const builder = new RangeSetBuilder<Decoration>()
        for (const token of query.tokens) {
            if (token.type === 'filter') {
                const isValid =
                    token?.value?.value || // has non-empty value
                    token?.value?.quoted || // or is quoted
                    inRange(state.selection.main.head, token.range.start, token.range.end + 1) // or cursor is within field

                // +1 to include the colon (:)
                builder.add(token.range.start, token.field.range.end + 1, isValid ? validFilter : invalidFilter)
            }
        }
        return builder.finish()
    }),
]
