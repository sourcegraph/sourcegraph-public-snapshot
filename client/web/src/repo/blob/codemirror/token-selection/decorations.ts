import { Decoration, EditorView } from '@codemirror/view'
import { isModifierKeyHeld } from './modifier-click'
import { positionToOffset, sortRangeValuesByStart } from '../utils'
import { codeIntelTooltipsState } from './code-intel-tooltips'
import { documentHighlightsField } from './document-highlights'
import { definitionUrlField } from './definition'

export function interactiveOccurrencesExtension() {
    return [
        EditorView.decorations.compute(
            [codeIntelTooltipsState, documentHighlightsField, definitionUrlField, isModifierKeyHeld],
            state => {
                const { focus, hover, pin } = state.field(codeIntelTooltipsState)

                // TODO: add support for hovered/pinned occurrence highlights
                const highlights = state.field(documentHighlightsField)

                const ranges = []

                if (focus) {
                    const valueRangeStart = positionToOffset(state.doc, focus.occurrence.range.start)
                    const valueRangeEnd = positionToOffset(state.doc, focus.occurrence.range.end)

                    if (valueRangeStart !== null && valueRangeEnd !== null) {
                        const classes = ['interactive-occurrence', 'focus-visible', 'sourcegraph-document-highlight']
                        const attributes: { [key: string]: string } = { tabindex: '0' }

                        const { value: url, hasOccurrence: hasDefinition } = state
                            .field(definitionUrlField)
                            .get(focus.occurrence)
                        if (state.field(isModifierKeyHeld) && hasDefinition) {
                            classes.push('cm-token-selection-definition-ready')
                            if (url) {
                                attributes['data-link'] = url
                            }
                        }

                        ranges.push(
                            Decoration.mark({
                                class: classes.join(' '),
                                attributes,
                            }).range(valueRangeStart, valueRangeEnd)
                        )
                    }

                    const selected = highlights?.find(
                        ({ range }) =>
                            focus.occurrence.range.start.line === range.start.line &&
                            focus.occurrence.range.start.character === range.start.character &&
                            focus.occurrence.range.end.line === range.end.line &&
                            focus.occurrence.range.end.character === range.end.character
                    )

                    if (selected) {
                        for (const highlight of sortRangeValuesByStart(highlights)) {
                            if (highlight === selected) {
                                continue
                            }

                            const highlightRangeStart = positionToOffset(state.doc, highlight.range.start)
                            const highlightRangeEnd = positionToOffset(state.doc, highlight.range.end)

                            if (highlightRangeStart === null || highlightRangeEnd === null) {
                                continue
                            }

                            ranges.push(
                                Decoration.mark({
                                    class: 'interactive-occurrence sourcegraph-document-highlight',
                                }).range(highlightRangeStart, highlightRangeEnd)
                            )
                        }
                    }
                }

                if (pin) {
                    const valueRangeStart = positionToOffset(state.doc, pin.occurrence.range.start)
                    const valueRangeEnd = positionToOffset(state.doc, pin.occurrence.range.end)

                    if (valueRangeStart !== null && valueRangeEnd !== null) {
                        ranges.push(
                            Decoration.mark({
                                class: 'interactive-occurrence selection-highlight',
                            }).range(valueRangeStart, valueRangeEnd)
                        )
                    }
                }

                if (hover && hover.occurrence !== focus?.occurrence && hover.occurrence !== pin?.occurrence) {
                    const valueRangeStart = positionToOffset(state.doc, hover.occurrence.range.start)
                    const valueRangeEnd = positionToOffset(state.doc, hover.occurrence.range.end)

                    if (valueRangeStart !== null && valueRangeEnd !== null) {
                        ranges.push(
                            Decoration.mark({
                                class: 'interactive-occurrence selection-highlight',
                            }).range(valueRangeStart, valueRangeEnd)
                        )
                    }
                }

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
