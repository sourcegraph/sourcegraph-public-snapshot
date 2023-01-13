import { RangeSetBuilder } from '@codemirror/state'
import { Decoration, EditorView, WidgetType } from '@codemirror/view'
import { mdiClose } from '@mdi/js'
import inRange from 'lodash/inRange'

import { DecoratedToken, toCSSClassName } from '@sourcegraph/shared/src/search/query/decoratedToken'
import { Token } from '@sourcegraph/shared/src/search/query/token'
import { createSVGIcon } from '@sourcegraph/shared/src/util/dom'

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
const contextFilter = Decoration.mark({ class: 'sg-context-filter', inclusive: true })
const replaceContext = Decoration.replace({})
class ClearTokenWidget extends WidgetType {
    constructor(private token: Token) {
        super()
    }

    public toDOM(view: EditorView): HTMLElement {
        const wrapper = document.createElement('span')
        wrapper.setAttribute('aria-hidden', 'true')
        wrapper.className = 'sg-clear-filter'

        const button = document.createElement('button')
        button.type = 'button'
        button.addEventListener('click', () => {
            view.dispatch({
                // -1/+1 to include possible leading and trailing whitespace
                changes: {
                    from: Math.max(this.token.range.start - 1, 0),
                    to: Math.min(this.token.range.end + 1, view.state.doc.length),
                },
            })
            if (!view.hasFocus) {
                view.focus()
            }
        })
        button.append(createSVGIcon(mdiClose))
        wrapper.append(button)

        return wrapper
    }
}

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
            padding: '0.125rem 0',
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

                if (token.field.value === 'context') {
                    builder.add(token.range.start, token.range.end, contextFilter)
                    if (!withinRange && token.value?.value) {
                        // hide context: field name and show remove button
                        builder.add(token.range.start, token.field.range.end + 1, replaceContext)
                        builder.add(
                            token.range.end,
                            token.range.end,
                            Decoration.widget({ widget: new ClearTokenWidget(token) })
                        )
                    }
                } else {
                    // +1 to include the colon (:)
                    builder.add(token.range.start, token.field.range.end + 1, isValid ? validFilter : invalidFilter)
                }
            }
        }
        return builder.finish()
    }),
]
