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
        constructor(private view: EditorView) {}

        update(update: ViewUpdate) {
            const focusedRange = getSelectedToken(update.state)
            if (focusedRange !== getSelectedToken(update.startState)) {
                this.fetchHighlights(focusedRange)
            }
        }

        private fetchHighlights(range: Range | null) {
            if (range) {
                getCodeIntelAPI(this.view.state)
                    .getDocumentHighlights(this.view.state, range)
                    .then(
                        highlights => {
                            this.view.dispatch({ effects: setDocumentHighlights.of({ target: range, highlights }) })
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
    create(state) {
        return { target: getSelectedToken(state), highlights: [] }
    },

    update(value, transaction) {
        const selectedToken = getSelectedToken(transaction.state)
        if (selectedToken !== getSelectedToken(transaction.startState)) {
            // We have to clear the highlights in the same transaction the selected token
            // changes. Otherwise the (new) selected token will loose focus when selection changes
            // to a token that was previously highlighted, breaking keyboard navigation
            return { target: selectedToken, highlights: [] }
        }

        for (const effect of transaction.effects) {
            if (effect.is(setDocumentHighlights) && effect.value.target === value.target) {
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
