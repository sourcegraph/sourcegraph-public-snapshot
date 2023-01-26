import { Extension, Range } from '@codemirror/state'
import { Decoration, EditorView } from '@codemirror/view'

import { positionToOffset, sortRangeValuesByStart } from '../utils'

import { codeIntelTooltipsState } from './code-intel-tooltips'
import { definitionUrlField } from './definition'
import { documentHighlightsField, findByOccurrence } from './document-highlights'
import { isModifierKeyHeld } from './modifier-click'

export function interactiveOccurrencesExtension(): Extension {
    return [
        EditorView.decorations.compute(
            [codeIntelTooltipsState, documentHighlightsField, definitionUrlField, isModifierKeyHeld],
            state => {
                const { focus, hover, pin } = state.field(codeIntelTooltipsState)
                const decorations = []

                if (focus) {
                    const classes = ['interactive-occurrence', 'focus-visible', 'sourcegraph-document-highlight']
                    const attributes: { [key: string]: string } = { tabindex: '0' }

                    // If the user is hovering over an occurrence with a definition holding the modifier key,
                    // add a class to make an occurrence to look like a link.
                    const { hasOccurrence: hasDefinition } = state.field(definitionUrlField).get(focus.occurrence)
                    if (state.field(isModifierKeyHeld) && hasDefinition) {
                        classes.push('cm-token-selection-definition-ready')
                    }

                    decorations.push({
                        decoration: Decoration.mark({ class: classes.join(' '), attributes }),
                        range: focus.occurrence.range,
                    })

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
