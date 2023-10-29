import { StateField, StateEffect, Facet, Transaction } from '@codemirror/state'
import { Decoration, EditorView, PluginValue, ViewPlugin, ViewUpdate } from '@codemirror/view'

import { getCodeIntelAPI } from './api'
import { codeIntelDecorations } from './decorations'

interface Range {
    from: number
    to: number
}

const DocumentHighlightLoader = ViewPlugin.fromClass(
    class implements PluginValue {
        private loading = new Set<Range>()
        constructor(private view: EditorView) {}

        update(update: ViewUpdate) {
            const highlights = update.state.field(documentHighlights)
            if (highlights !== update.startState.field(documentHighlights)) {
                for (const highlight of highlights) {
                    if (highlight.loading) {
                        this.fetchHighlights(highlight.range)
                    }
                }
            }
        }

        private fetchHighlights(range: Range) {
            if (!this.loading.has(range)) {
                getCodeIntelAPI(this.view.state)
                    .getDocumentHighlights(this.view.state, range)
                    .then(
                        highlights => {
                            this.view.dispatch({ effects: setDocumentHighlights.of({ range: range, highlights }) })
                            this.loading.delete(range)
                        },
                        () => {}
                    )
            }
        }
    }
)

const documentHighlightDecoration = Decoration.mark({ class: 'sourcegraph-document-highlight' })
const setDocumentHighlights = StateEffect.define<{ range: Range; highlights: Range[] }>()

class Highlights {
    constructor(public range: Range, public highlights: Range[], public loading: boolean) {}

    update(transaction: Transaction) {
        for (const effect of transaction.effects) {
            if (effect.is(setDocumentHighlights) && effect.value.range === this.range) {
                return new Highlights(this.range, effect.value.highlights, false)
            }
        }
        return this
    }
}

const documentHighlights = StateField.define<Highlights[]>({
    create(state) {
        return state.facet(showDocumentHighlights).map(range => new Highlights(range, [], true))
    },

    update(value, transaction) {
        let newValue = value
        const newRanges = transaction.state.facet(showDocumentHighlights)
        if (newRanges !== transaction.startState.facet(showDocumentHighlights)) {
            newValue = newRanges.map(range => {
                let seenAt = -1
                for (let i = 0; i < value.length; i++) {
                    if (range === value[i].range) {
                        seenAt = i
                    }
                }
                return seenAt > -1 ? value[seenAt] : new Highlights(range, [], true)
            })
        }

        newValue = newValue.map(value => value.update(transaction))

        return value.length === newValue.length && value.every((highlights, i) => highlights === newValue[i])
            ? value
            : newValue
    },

    provide(field) {
        return codeIntelDecorations.compute([field], state =>
            Decoration.set(
                state
                    .field(field)
                    .flatMap(entry =>
                        entry.highlights.map(highlight =>
                            documentHighlightDecoration.range(highlight.from, highlight.to)
                        )
                    ),
                true
            )
        )
    },
})

export const showDocumentHighlights: Facet<Range> = Facet.define<Range>({
    enables: [documentHighlights, DocumentHighlightLoader],
})
