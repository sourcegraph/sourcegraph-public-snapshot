import {
    selectAll,
    selectCharLeft,
    selectCharRight,
    selectDocEnd,
    selectDocStart,
    selectGroupLeft,
    selectGroupRight,
    selectLineDown,
    selectLineUp,
    selectPageDown,
    selectPageUp,
} from '@codemirror/commands'
import { type Extension, Prec } from '@codemirror/state'
import { EditorView, type KeyBinding, keymap, layer, RectangleMarker } from '@codemirror/view'

import { blobPropsFacet } from '..'
import { syntaxHighlight } from '../highlight'
import { positionAtCmPosition, closestOccurrenceByCharacter } from '../occurrence-utils'
import { CodeIntelTooltip } from '../tooltips/CodeIntelTooltip'
import { LoadingTooltip } from '../tooltips/LoadingTooltip'
import { positionToOffset } from '../utils'

import {
    getCodeIntelTooltipState,
    setFocusedOccurrenceTooltip,
    selectOccurrence,
    getHoverTooltip,
} from './code-intel-tooltips'
import { goToDefinitionAtOccurrence } from './definition'
import { isModifierKeyHeld } from './modifier-click'

const keybindings: KeyBinding[] = [
    {
        key: 'Space',
        run(view) {
            const selected = getCodeIntelTooltipState(view, 'focus')
            if (!selected) {
                return true
            }

            if (selected.tooltip instanceof CodeIntelTooltip) {
                view.dispatch({ effects: setFocusedOccurrenceTooltip.of(null) })
                return true
            }

            const offset = positionToOffset(view.state.doc, selected.occurrence.range.start)
            if (offset === null) {
                return true
            }

            // show loading tooltip
            view.dispatch({ effects: setFocusedOccurrenceTooltip.of(new LoadingTooltip(offset)) })

            getHoverTooltip(view, offset)
                .then(value => view.dispatch({ effects: setFocusedOccurrenceTooltip.of(value) }))
                .finally(() => {
                    // close loading tooltip if any
                    const current = getCodeIntelTooltipState(view, 'focus')
                    if (current?.tooltip instanceof LoadingTooltip && current?.occurrence === selected.occurrence) {
                        view.dispatch({ effects: setFocusedOccurrenceTooltip.of(null) })
                    }
                })

            return true
        },
    },
    {
        key: 'Escape',
        run(view) {
            const current = getCodeIntelTooltipState(view, 'focus')
            if (current?.tooltip instanceof CodeIntelTooltip) {
                view.dispatch({ effects: setFocusedOccurrenceTooltip.of(null) })
            }

            return true
        },
    },

    {
        key: 'Enter',
        run(view) {
            const selected = getCodeIntelTooltipState(view, 'focus')
            if (!selected?.occurrence) {
                return false
            }

            const offset = positionToOffset(view.state.doc, selected.occurrence.range.start)
            if (offset === null) {
                return true
            }

            // show loading tooltip
            view.dispatch({ effects: setFocusedOccurrenceTooltip.of(new LoadingTooltip(offset)) })

            const startTime = Date.now()
            goToDefinitionAtOccurrence(view, selected.occurrence)
                .then(
                    ({ handler, url }) => {
                        if (view.state.field(isModifierKeyHeld) && url) {
                            window.open(url, '_blank')
                        } else {
                            handler(selected.occurrence.range.start, startTime)
                        }
                    },
                    () => {}
                )
                .finally(() => {
                    // close loading tooltip if any
                    const current = getCodeIntelTooltipState(view, 'focus')
                    if (current?.tooltip instanceof LoadingTooltip && current?.occurrence === selected.occurrence) {
                        view.dispatch({ effects: setFocusedOccurrenceTooltip.of(null) })
                    }
                })

            return true
        },
    },

    {
        key: 'Mod-ArrowRight',
        run(view) {
            view.state.facet(blobPropsFacet).navigate(1)
            return true
        },
    },
    {
        key: 'Mod-ArrowLeft',
        run(view) {
            view.state.facet(blobPropsFacet).navigate(-1)
            return true
        },
    },
]

/**
 * Keybindings to support text selection.
 * Modified version of a [standard CodeMirror keymap](https://sourcegraph.com/github.com/codemirror/commands@ca4171b381dc487ec0313cb0ea4dd29151c7a792/-/blob/src/commands.ts?L791-858).
 */
const textSelectionKeybindings: KeyBinding[] = [
    {
        key: 'ArrowLeft',
        shift: selectCharLeft,
    },

    {
        key: 'ArrowRight',
        shift: selectCharRight,
    },

    {
        key: 'Mod-ArrowLeft',
        mac: 'Alt-ArrowLeft',
        shift: selectGroupLeft,
    },

    {
        key: 'Mod-ArrowRight',
        mac: 'Alt-ArrowRight',
        shift: selectGroupRight,
    },

    {
        key: 'ArrowUp',
        shift: selectLineUp,
    },

    {
        key: 'ArrowDown',
        shift: selectLineDown,
    },

    {
        key: 'Mod-Home',
        mac: 'Cmd-ArrowUp',
        shift: selectDocStart,
    },

    {
        key: 'Mod-End',
        mac: 'Cmd-ArrowDown',
        shift: selectDocEnd,
    },

    {
        key: 'PageUp',
        mac: 'Ctrl-ArrowUp',
        shift: selectPageUp,
    },

    {
        key: 'PageDown',
        mac: 'Ctrl-ArrowDown',
        shift: selectPageDown,
    },

    {
        key: 'Mod-a',
        run: selectAll,
    },
]

function keyDownHandler(event: KeyboardEvent, view: EditorView): boolean {
    switch (event.key) {
        case 'ArrowLeft': {
            const selectedOccurrence = getCodeIntelTooltipState(view, 'focus')?.occurrence
            const position = selectedOccurrence?.range.start || positionAtCmPosition(view, view.viewport.from)
            const table = view.state.facet(syntaxHighlight)
            const occurrence = closestOccurrenceByCharacter(position.line, table, position, occurrence =>
                occurrence.range.start.isSmaller(position)
            )
            if (occurrence) {
                selectOccurrence(view, occurrence)
            }

            return true
        }
        case 'ArrowRight': {
            const selectedOccurrence = getCodeIntelTooltipState(view, 'focus')?.occurrence
            const position = selectedOccurrence?.range.start || positionAtCmPosition(view, view.viewport.from)
            const table = view.state.facet(syntaxHighlight)
            const occurrence = closestOccurrenceByCharacter(position.line, table, position, occurrence =>
                occurrence.range.start.isGreater(position)
            )
            if (occurrence) {
                selectOccurrence(view, occurrence)
            }

            return true
        }
        case 'ArrowUp': {
            const selectedOccurrence = getCodeIntelTooltipState(view, 'focus')?.occurrence
            const position = selectedOccurrence?.range.start || positionAtCmPosition(view, view.viewport.from)
            const table = view.state.facet(syntaxHighlight)
            for (let line = position.line - 1; line >= 0; line--) {
                const occurrence = closestOccurrenceByCharacter(line, table, position)
                if (occurrence) {
                    selectOccurrence(view, occurrence)
                    return true
                }
            }
            return true
        }
        case 'ArrowDown': {
            const selectedOccurrence = getCodeIntelTooltipState(view, 'focus')?.occurrence
            const position = selectedOccurrence?.range.start || positionAtCmPosition(view, view.viewport.from)
            const table = view.state.facet(syntaxHighlight)
            for (let line = position.line + 1; line < table.lineIndex.length; line++) {
                const occurrence = closestOccurrenceByCharacter(line, table, position)
                if (occurrence) {
                    selectOccurrence(view, occurrence)
                    return true
                }
            }
            return true
        }

        default: {
            return false
        }
    }
}

/**
 * Keyboard event handlers defined via {@link keymap} facet do not work with the screen reader enabled while
 * keypress handlers defined via {@link EditorView.domEventHandlers} still work.
 */
function occurrenceKeyboardNavigation(): Extension {
    return [
        EditorView.domEventHandlers({
            keydown: keyDownHandler,
        }),
    ]
}

/**
 * For some reason, editor selection updates made by {@link textSelectionKeybindings} handlers are not synced with the
 * [browser selection](https://developer.mozilla.org/en-US/docs/Web/API/Selection). This function is a workaround to ensure
 * that the browser selection is updated when the editor selection is updated.
 *
 * This layer has explicitly set high precedence to be always shown above other layers (e.g. selected lines layer).
 *
 * @see https://codemirror.net/docs/ref/#view.drawSelection
 * @see https://sourcegraph.com/github.com/codemirror/view@84f483ae4097a71d04374cdb24c5edc09d211105/-/blob/src/draw-selection.ts?L92-102
 */
const selectionLayer = Prec.high(
    layer({
        above: false,
        markers(view) {
            return view.state.selection.ranges
                .map(range => (range.empty ? [] : RectangleMarker.forRange(view, 'cm-selectionBackground', range)))
                .reduce((a, b) => a.concat(b))
        },
        update(update) {
            return update.docChanged || update.selectionSet || update.viewportChanged
        },
        class: 'cm-selectionLayer',
    })
)

const theme = EditorView.theme({
    '.cm-line': {
        '& ::selection': {
            backgroundColor: 'transparent !important',
        },
        '&::selection': {
            backgroundColor: 'transparent !important',
        },
    },
    '.cm-selectionLayer .cm-selectionBackground': {
        background: 'var(--code-selection-bg-2)',
    },
})

/**
 * Extension that adds support for the text selection with keyboard.
 */
function textSelectionExtension(): Extension {
    return [keymap.of(textSelectionKeybindings), selectionLayer, theme]
}

export function keyboardShortcutsExtension(): Extension {
    return [textSelectionExtension(), keymap.of(keybindings), occurrenceKeyboardNavigation()]
}
