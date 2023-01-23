import { Decoration, DecorationSet, EditorView, KeyBinding, keymap, ViewPlugin, ViewUpdate } from '@codemirror/view'
import { RangeSetBuilder, StateEffect, StateField } from '@codemirror/state'
import { positionToOffset } from '../utils'
import {
    isInteractiveOccurrence,
    occurrenceAtPosition,
    positionAtCmPosition,
    rangeToCmSelection,
} from '../occurrence-utils'
import { Occurrence } from '@sourcegraph/shared/src/codeintel/scip'
import { fallbackOccurrences } from '../token-selection/selections'

const setSelectedOccurrence = StateEffect.define<Occurrence | null>()
const selectedOccurrenceField = StateField.define<Occurrence | null>({
    create() {
        return null
    },
    update(value, transaction) {
        for (const effect of transaction.effects) {
            if (effect.is(setSelectedOccurrence)) {
                return effect.value
            }
        }
        return value
    },
    provide(field) {
        return [
            EditorView.decorations.compute([field], state => {
                const value = state.field(field)
                // console.log(value)
                if (!value) return Decoration.none
                const from = positionToOffset(state.doc, value.range.start)
                const to = positionToOffset(state.doc, value.range.end)
                return Decoration.set(
                    Decoration.mark({
                        class: 'text-uppercase interactive-occurrence ',
                        attributes: {
                            tabindex: '0',
                        },
                    }).range(from, to)
                )
            }),
        ]
    },
})

class InteractiveOccurrencesManager {
    public decorations = Decoration.none

    constructor(view: EditorView) {
        this.decorations = this.computeDecorations(view)
    }

    public update(update: ViewUpdate): void {
        if (update.viewportChanged) {
            this.decorations = this.computeDecorations(update.view)
            return
        }

        for (const transaction of update.transactions) {
            for (const effect of transaction.effects) {
                if (effect.is(setSelectedOccurrence)) {
                    this.computeDecorations(update.view)
                }
            }
        }
    }

    private computeDecorations(view: EditorView): DecorationSet {
        const builder = new RangeSetBuilder<Decoration>()

        const { from, to } = view.viewport

        const selectedOccurrence = view.state.field(selectedOccurrenceField)

        for (let i = from; i < to; ) {
            const occurrence = occurrenceAtPosition(view.state, positionAtCmPosition(view, i))

            if (!occurrence) {
                i++
                continue
            }

            const from = positionToOffset(view.state.doc, occurrence.range.start)
            const to = positionToOffset(view.state.doc, occurrence.range.end)

            if (isInteractiveOccurrence(occurrence)) {
                const isSelected = occurrence === selectedOccurrence
                const decoration = Decoration.mark({
                    class: `interactive-occurrence ${isSelected ? 'selected' : ''}`,
                    attributes: {
                        // tabindex: isSelected ? '0' : '-1',
                        tabindex: '0',
                    },
                })
                builder.add(from, to, decoration)
            }

            i = to + 1
        }

        return builder.finish()
    }
}

const keybindings: KeyBinding[] = [
    // {
    //     key: 'ArrowLeft',
    //     run(view) {
    //         const pos = view.posAtDOM(document.activeElement)
    //         const range = view.state.wordAt(pos)
    //
    //         for (const { from, to } of view.visibleRanges) {
    //             const endPos = (pos < to && range && range.from - 1) || to
    //
    //             for (let i = endPos; i >= from; ) {
    //                 const word = view.state.wordAt(i)
    //                 if (word) {
    //                     const text = view.state.doc.sliceString(word.from, word.to)
    //
    //                     if (tokens.includes(text)) {
    //                         view.dispatch({
    //                             effects: setFocusTooltip.of({
    //                                 range: word,
    //                                 tooltip: null,
    //                             }),
    //                             selection: word,
    //                         })
    //
    //                         const node = view.domAtPos(i).node
    //                         const el =
    //                             node instanceof HTMLElement
    //                                 ? node
    //                                 : node.parentElement?.matches('.my-decoration')
    //                                 ? node.parentElement
    //                                 : null
    //
    //                         if (el) {
    //                             el.focus()
    //                         }
    //                         break
    //                     }
    //                     i = word.from - 1
    //                 } else {
    //                     i--
    //                 }
    //             }
    //         }
    //         return true
    //     },
    // },
    {
        key: 'ArrowRight',
        run(view) {
            const { from, to } = view.viewport

            // console.log('arrow right')

            let startPos = from
            const selectedOccurrence = view.state.field(selectedOccurrenceField)
            if (selectedOccurrence) {
                startPos = positionToOffset(view.state.doc, selectedOccurrence.range.end) ?? from
            }

            for (let i = startPos + 1; i <= to; i++) {
                const occurrence = occurrenceAtPosition(view.state, positionAtCmPosition(view, i))
                if (occurrence) {
                    // create tabindex='0'
                    view.dispatch({ effects: setSelectedOccurrence.of(occurrence) })

                    const node = view.domAtPos(i).node
                    const el = node instanceof HTMLElement ? node : node.parentElement
                    const lineEl = el?.matches('.cm-line') ? el : el?.closest('.cm-line')
                    const interactiveOccurrenceEl = lineEl?.querySelector<HTMLElement>('.interactive-occurrence')
                    if (interactiveOccurrenceEl) {
                        interactiveOccurrenceEl?.focus()
                    }

                    break
                }
            }

            return true
        },
    },
    // {
    //     key: 'Space',
    //     run(view) {
    //         const offset = view.posAtDOM(document.activeElement)
    //         const word = view.state.wordAt(offset)
    //         const text = word ? view.state.doc.sliceString(word.from, word.to) : undefined
    //
    //         if (tokens.includes(text)) {
    //             // show tooltip for focused token
    //             view.dispatch({
    //                 effects: setFocusTooltip.of({
    //                     range: word,
    //                     tooltip: {
    //                         pos: word.from,
    //                         above: true,
    //                         create() {
    //                             const dom = document.createElement('div')
    //                             dom.classList.add('tooltip')
    //                             dom.textContent = text
    //                             return { dom }
    //                         },
    //                     },
    //                 }),
    //             })
    //             return true
    //         }
    //
    //         view.dispatch({ effects: setFocusTooltip.of(null) })
    //         return true
    //
    //         // WARNING: we are not expected to modify the DOM nodes directly
    //         // event.target.classList.add("focus-highlight");
    //     },
    // },
]

export function focusDrivenCodeNavigation() {
    return [
        selectedOccurrenceField,
        fallbackOccurrences,
        keymap.of(keybindings),
        // ViewPlugin.fromClass(InteractiveOccurrencesManager, { decorations: plugin => plugin.decorations }),
    ]
}
