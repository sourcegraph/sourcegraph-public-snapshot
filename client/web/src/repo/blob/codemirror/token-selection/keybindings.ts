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
import { NavigateFunction } from 'react-router-dom'

import { syntaxHighlight } from '../highlight'
import { positionAtCmPosition, closestOccurrenceByCharacter } from '../occurrence-utils'
import { positionToOffset, preciseOffsetAtCoords } from '../utils'

import { getCodeIntelAPI } from './api'
import { getSelectedToken, setSelection } from './token-selection'
import { getTooltipState, hideTooltipForKey, showCodeIntelTooltipAtRange } from './tooltips'

const keybindings: KeyBinding[] = [
    {
        key: 'Space',
        run(view) {
            const selected = getSelectedToken(view.state)
            if (!selected) {
                return true
            }

            if (getTooltipState(view.state, 'focus')?.visible) {
                view.dispatch(hideTooltipForKey('focus'))
                return true
            }

            // show loading tooltip
            showCodeIntelTooltipAtRange(view, selected, 'focus', true)
            return true
        },
    },
    {
        key: 'Escape',
        run(view) {
            view.dispatch(hideTooltipForKey('focus'))
            return true
        },
    },

    {
        key: 'Enter',
        run(view) {
            const selected = getSelectedToken(view.state)
            if (!selected) {
                return false
            }

            // TODO: show loading tooltip
            // TODO ? : const startTime = Date.now()
            getCodeIntelAPI(view.state).goToDefinitionAt(view, selected.from)
            return true
        },
    },
    {
        key: 'Mod-Enter',
        run(view) {
            const selected = getSelectedToken(view.state)
            if (!selected) {
                return false
            }

            // TODO: show loading tooltip
            // TODO ? : const startTime = Date.now()
            getCodeIntelAPI(view.state).goToDefinitionAt(view, selected.from, { newWindow: true })
            return true
        },
    },
]

function navigationKeybindings(navigate: NavigateFunction): KeyBinding[] {
    return [
        {
            key: 'Mod-ArrowRight',
            run() {
                navigate(1)
                return true
            },
        },
        {
            key: 'Mod-ArrowLeft',
            run() {
                navigate(-1)
                return true
            },
        },
    ]
}

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
            const range = getSelectedToken(view.state)?.from || view.viewport.from
            const position = positionAtCmPosition(view.state.doc, range)
            const table = view.state.facet(syntaxHighlight)
            const occurrence = closestOccurrenceByCharacter(position.line, table, position, occurrence =>
                occurrence.range.start.isSmaller(position)
            )
            const anchor = occurrence ? positionToOffset(view.state.doc, occurrence.range.start) : null
            if (anchor !== null) {
                view.dispatch(setSelection(anchor), hideTooltipForKey('focus'))
            }

            return true
        }
        case 'ArrowRight': {
            const range = getSelectedToken(view.state)?.from || view.viewport.from
            const position = positionAtCmPosition(view.state.doc, range)
            const table = view.state.facet(syntaxHighlight)
            const occurrence = closestOccurrenceByCharacter(position.line, table, position, occurrence =>
                occurrence.range.start.isGreater(position)
            )
            const anchor = occurrence ? positionToOffset(view.state.doc, occurrence.range.start) : null
            if (anchor !== null) {
                view.dispatch(setSelection(anchor), hideTooltipForKey('focus'))
            }

            return true
        }
        case 'ArrowUp': {
            const range = getSelectedToken(view.state)?.from || view.viewport.from
            const position = positionAtCmPosition(view.state.doc, range)
            const table = view.state.facet(syntaxHighlight)
            for (let line = position.line - 1; line >= 0; line--) {
                const occurrence = closestOccurrenceByCharacter(line, table, position)
                const anchor = occurrence ? positionToOffset(view.state.doc, occurrence.range.start) : null
                if (anchor !== null) {
                    view.dispatch(setSelection(anchor), hideTooltipForKey('focus'))
                    return true
                }
            }
            return true
        }
        case 'ArrowDown': {
            const range = getSelectedToken(view.state)?.from || view.viewport.from
            const position = positionAtCmPosition(view.state.doc, range)
            const table = view.state.facet(syntaxHighlight)
            for (let line = position.line + 1; line < table.lineIndex.length; line++) {
                const occurrence = closestOccurrenceByCharacter(line, table, position)
                const anchor = occurrence ? positionToOffset(view.state.doc, occurrence.range.start) : null
                if (anchor !== null) {
                    view.dispatch(setSelection(anchor), hideTooltipForKey('focus'))
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

/**
 * Not a keybinding but related to it. This extension closes
 * focus-related tooltips when selection moves elsewhere.
 */
const closeTooltipOnClickOutside = EditorView.domEventHandlers({
    click(event, view) {
        const offset = preciseOffsetAtCoords(view, event)
        if (offset) {
            const range = getCodeIntelAPI(view.state).findOccurrenceRangeAt(offset, view.state)
            if (range && range?.from === getTooltipState(view.state, 'focus')?.position.range.from) {
                return
            }
        }
        view.dispatch(hideTooltipForKey('focus'))
    },
})

export function keyboardShortcutsExtension(navigate: NavigateFunction): Extension {
    return [
        textSelectionExtension(),
        keymap.of([...keybindings, ...navigationKeybindings(navigate)]),
        occurrenceKeyboardNavigation(),
        closeTooltipOnClickOutside,
    ]
}
