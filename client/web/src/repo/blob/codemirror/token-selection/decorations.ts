import type { EditorState, Extension, Range as CodeMirrorRange } from '@codemirror/state'
import { Decoration, EditorView } from '@codemirror/view'
import classNames from 'classnames'

import type { Range } from '@sourcegraph/extension-api-types'
import type { Occurrence } from '@sourcegraph/shared/src/codeintel/scip'

import { rangeToCmSelection } from '../occurrence-utils'
import { positionToOffset, sortRangeValuesByStart } from '../utils'

import { codeIntelTooltipsState } from './code-intel-tooltips'
import { definitionUrlField } from './definition'
import { documentHighlightsField, findByOccurrence } from './document-highlights'
import { isModifierKeyHeld } from './modifier-click'

interface DecorationItem {
    decoration: Decoration
    range: Range
}

function sortByFromPosition(ranges: CodeMirrorRange<Decoration>[]): CodeMirrorRange<Decoration>[] {
    return ranges.sort((a, b) => a.from - b.from)
}

function findByRange(decorations: DecorationItem[], range: Range): DecorationItem | undefined {
    return decorations.find(
        ({ range: decorationRange }) =>
            decorationRange.start.line === range.start.line &&
            decorationRange.start.character === range.start.character &&
            decorationRange.end.line === range.end.line &&
            decorationRange.end.character === range.end.character
    )
}

function addOrReplace(decorations: DecorationItem[], item: DecorationItem): DecorationItem[] {
    const existing = findByRange(decorations, item.range)
    if (existing) {
        existing.decoration = item.decoration
    } else {
        decorations.push(item)
    }
    return decorations
}

const theme = EditorView.theme({
    '.cm-token-selection-definition-ready': {
        textDecoration: 'underline',
    },
})

/**
 * Returns `true` if the editor selection is empty or is inside the occurrence range.
 */
function shouldApplyFocusStyles(state: EditorState, occurrence: Occurrence): boolean {
    if (state.selection.main.empty) {
        return true
    }

    const occurrenceRangeAsSelection = rangeToCmSelection(state, occurrence.range)
    const isEditorSelectionInsideOccurrenceRange =
        state.selection.main.from >= occurrenceRangeAsSelection.from &&
        state.selection.main.to <= occurrenceRangeAsSelection.to
    return isEditorSelectionInsideOccurrenceRange
}

/**
 * Extension providing decorations for focused, hovered, pinned occurrences, and document highlights.
 * We combine all of these into a single extension to avoid the focused element blur caused by its removal from the DOM.
 */
export function interactiveOccurrencesExtension(): Extension {
    return [
        EditorView.decorations.compute(
            [codeIntelTooltipsState, documentHighlightsField, definitionUrlField, isModifierKeyHeld, 'selection'],
            state => {
                const { focus, hover, pin } = state.field(codeIntelTooltipsState)
                let decorations: DecorationItem[] = []

                if (focus) {
                    decorations.push({
                        decoration: Decoration.mark({
                            class: classNames(
                                'interactive-occurrence', // used as interactive occurrence selector
                                shouldApplyFocusStyles(state, focus.occurrence) && 'focus-visible' // adds focus styles to the occurrence
                            ),
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

                // focused occurrence is already highlighted
                if (pin && pin.occurrence !== focus?.occurrence) {
                    // pinned decoration styles have higher precedence over the document highlights decoration
                    decorations = addOrReplace(decorations, {
                        decoration: Decoration.mark({ class: 'selection-highlight' }),
                        range: pin.occurrence.range,
                    })
                }

                // focused and pinned occurrences are already highlighted
                if (hover && hover.occurrence !== focus?.occurrence && hover.occurrence !== pin?.occurrence) {
                    // pinned decoration styles have higher precedence over the document highlights decoration
                    decorations = addOrReplace(decorations, {
                        decoration: Decoration.mark({
                            class: classNames('selection-highlight', {
                                // If the user is hovering over a selected (focused) occurrence with a definition holding the modifier key,
                                // add a class to make an occurrence to look like a link.
                                ['cm-token-selection-definition-ready']:
                                    state.field(isModifierKeyHeld) &&
                                    state.field(definitionUrlField).get(hover.occurrence).hasOccurrence,
                            }),
                        }),
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
                }, [] as CodeMirrorRange<Decoration>[])

                return Decoration.set(sortByFromPosition(ranges))
            }
        ),
        theme,
    ]
}
