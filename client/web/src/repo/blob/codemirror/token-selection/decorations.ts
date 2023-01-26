import { Extension, Range } from '@codemirror/state'
import { Decoration, EditorView } from '@codemirror/view'

import { positionToOffset, sortRangeValuesByStart } from '../utils'

import { codeIntelTooltipsState } from './code-intel-tooltips'
import { definitionUrlField } from './definition'
import { documentHighlightsField, findByOccurrence } from './document-highlights'
import { isModifierKeyHeld } from './modifier-click'

/**
 * Extension providing decorations for focused, hovered, pinned occurrences, and document highlights.
 */
export function interactiveOccurrencesExtension(): Extension {
    return [
        EditorView.decorations.compute(
            [codeIntelTooltipsState, documentHighlightsField, definitionUrlField, isModifierKeyHeld],
            state => {
                const { focus, hover, pin } = state.field(codeIntelTooltipsState)
                const decorations = []

                if (focus) {
                    const classes = [
                        'interactive-occurrence', // used as interactive occurrence selector
                        'focus-visible', // prevents code editor from blur when focused element inside it changes
                        'sourcegraph-document-highlight', // highlights the selected (focused) occurrence
                    ]

                    // If the user is hovering over an occurrence with a definition holding the modifier key,
                    // add a class to make an occurrence to look like a link.
                    const { hasOccurrence: hasDefinition } = state.field(definitionUrlField).get(focus.occurrence)
                    if (state.field(isModifierKeyHeld) && hasDefinition) {
                        classes.push('cm-token-selection-definition-ready')
                    }

                    decorations.push({
                        decoration: Decoration.mark({
                            class: classes.join(' '),
                            attributes: {
                                // Selected (focused) occurrence is the only focusable element in the editor.
                                // This helps to maintain the focus position when editor is blurred and then focused again.
                                tabindex: '0',
                            },
                        }),
                        range: focus.occurrence.range,
                    })

                    // Decorate selected (focused) occurrence document highlights.
                    const highlights = state.field(documentHighlightsField)
                    const focusedOccurrenceHighlight = findByOccurrence(highlights, focus.occurrence)
                    if (focusedOccurrenceHighlight) {
                        for (const highlight of sortRangeValuesByStart(highlights)) {
                            if (highlight === focusedOccurrenceHighlight) {
                                // Focused occurrence is already highlighted.
                                continue
                            }

                            decorations.push({
                                decoration: Decoration.mark({
                                    class: 'sourcegraph-document-highlight',
                                }),
                                range: highlight.range,
                            })
                        }
                    }
                }

                if (pin) {
                    decorations.push({
                        decoration: Decoration.mark({ class: 'selection-highlight' }),
                        range: pin.occurrence.range,
                    })
                }

                if (hover) {
                    decorations.push({
                        decoration: Decoration.mark({ class: 'selection-highlight' }),
                        range: hover.occurrence.range,
                    })
                }

                const ranges = decorations.reduce((acc, { decoration, range }) => {
                    const from = positionToOffset(state.doc, range.start)
                    const to = positionToOffset(state.doc, range.end)

                    if (from !== null && to !== null) {
                        acc.push(decoration.range(from, to))
                    }

                    return acc
                }, [] as Range<Decoration>[])

                return Decoration.set(ranges.sort((a, b) => a.from - b.from))
            }
        ),
        EditorView.theme({
            '.cm-token-selection-definition-ready': {
                textDecoration: 'underline',
            },
        }),
    ]
}
