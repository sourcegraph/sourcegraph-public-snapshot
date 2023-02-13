import { RangeSetBuilder } from '@codemirror/state'
import {
    Decoration,
    DecorationSet,
    EditorView,
    PluginValue,
    ViewPlugin,
    ViewUpdate,
    WidgetType,
} from '@codemirror/view'
import { mdiClose } from '@mdi/js'
import inRange from 'lodash/inRange'

import { Token } from '@sourcegraph/shared/src/search/query/token'
import { createSVGIcon } from '@sourcegraph/shared/src/util/dom'

import { decoratedTokens, queryTokens } from '../../codemirror/parsedQuery'

const validFilter = Decoration.mark({ class: 'sg-filter' })
const invalidFilter = Decoration.mark({ class: 'sg-filter sg-invalid-filter' })
const contextFilter = Decoration.mark({ class: 'sg-context-filter', inclusiveEnd: true })
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

                // context: filters are handled by the view plugin defined below
                if (token.field.value !== 'context') {
                    // +1 to include the colon (:)
                    builder.add(token.range.start, token.field.range.end + 1, isValid ? validFilter : invalidFilter)
                }
            }
        }
        return builder.finish()
    }),
    // ViewPlugin handling decorating context: filters
    ViewPlugin.fromClass(
        class implements PluginValue {
            public decorations: DecorationSet

            constructor(view: EditorView) {
                this.decorations = this.createDecorations(view)
            }

            public update(update: ViewUpdate): void {
                if (update.focusChanged || update.selectionSet || update.docChanged) {
                    this.decorations = this.createDecorations(update.view)
                }
            }

            private createDecorations(view: EditorView): DecorationSet {
                const query = view.state.facet(queryTokens)
                const builder = new RangeSetBuilder<Decoration>()
                for (const token of query.tokens) {
                    if (token.type === 'filter' && token.field.value === 'context') {
                        const withinRange = inRange(
                            view.state.selection.main.head,
                            token.range.start,
                            token.range.end + 1
                        ) // or cursor is within field
                        if (token.value?.value && (!withinRange || !view.hasFocus)) {
                            // hide context: field name and show remove button
                            builder.add(token.range.start, token.field.range.end + 1, replaceContext)
                            builder.add(token.range.start, token.range.end, contextFilter)
                            builder.add(
                                token.range.end,
                                token.range.end,
                                Decoration.widget({ widget: new ClearTokenWidget(token), side: -1 })
                            )
                        } else {
                            builder.add(token.range.start, token.range.end, contextFilter)
                        }
                    }
                }
                return builder.finish()
            }
        },
        { decorations: plugin => plugin.decorations }
    ),
]
