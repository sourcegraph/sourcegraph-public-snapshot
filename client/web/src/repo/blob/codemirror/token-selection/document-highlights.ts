import { RangeSetBuilder, type Extension, StateField, StateEffect } from '@codemirror/state'
import { Decoration, EditorView, PluginValue, ViewPlugin, ViewUpdate } from '@codemirror/view'

import { getCodeIntelAPI } from './api'
import { getSelectedToken } from './token-selection'

interface Range {
    from: number
    to: number
}

const DocumentHighlightLoader = ViewPlugin.fromClass(
    class implements PluginValue {
        private focusedRange: Range | null = null

        constructor(private view: EditorView) {}

        update(update: ViewUpdate) {
            const focusedRange = getSelectedToken(update.state)
            if (focusedRange !== getSelectedToken(update.startState)) {
                this.fetchHighlights(focusedRange)
            }
        }

        private fetchHighlights(range: Range | null) {
            this.focusedRange = range

            if (range) {
                getCodeIntelAPI(this.view.state)
                    .getDocumentHighlights(this.view.state, range)
                    .then(
                        highlights => {
                            if (this.focusedRange === range) {
                                this.view.dispatch({ effects: setDocumentHighlights.of({ target: range, highlights }) })
                            }
                        },
                        () => {}
                    )
            }
        }
    }
)

const documentHighlightDecoration = Decoration.mark({ class: 'sourcegraph-document-highlight' })
const setDocumentHighlights = StateEffect.define<{ target: Range; highlights: Range[] }>()

const documentHighlights = StateField.define<{ target: Range | null; highlights: Range[] }>({
    create() {
        return { target: null, highlights: [] }
    },

    update(value, transaction) {
        const selectedToken = getSelectedToken(transaction.state)
        if (selectedToken !== getSelectedToken(transaction.startState)) {
            return { target: selectedToken, highlights: [] }
        }

        for (const effect of transaction.effects) {
            // We have to clear the highlights in the same transaction the selected token
            // changes. Otherwise the (new) selected token will loose focus when selection changes
            // to a token that was previously highlighted, breaking keyboard navigation
            if (
                effect.is(setDocumentHighlights) &&
                effect.value.target.from === value.target?.from &&
                effect.value.target.to === value.target.to
            ) {
                return effect.value
            }
        }
        return value
    },

    provide(field) {
        return EditorView.decorations.compute([field], state => {
            const builder = new RangeSetBuilder<Decoration>()
            const { target, highlights } = state.field(field) ?? { target: null, highlights: [] }
            if (target) {
                for (const highlight of highlights.sort((a, b) => a.from - b.from)) {
                    if (highlight.from !== target.from || highlight.to !== target.to) {
                        // Focused occurrence is already highlighted.
                        builder.add(highlight.from, highlight.to, documentHighlightDecoration)
                    }
                }
            }
            return builder.finish()
        })
    },
})

export const documentHighlightsExtension: Extension = [DocumentHighlightLoader, documentHighlights]
