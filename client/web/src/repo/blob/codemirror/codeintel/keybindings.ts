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
import type { NavigateFunction } from 'react-router-dom'

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

/**
 * For some reason, editor selection updates made by {@link textSelectionKeybindings} handlers are not synced with the
 * [browser selection](https://developer.mozilla.org/en-US/docs/Web/API/Selection). This function is a workaround to ensure
 * that the browser selection is updated when the editor selection is updated.
 *
 * This layer has explicitly set high precedence to be always shown above other layers (e.g. selected lines layer).
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

export function keyboardShortcutsExtension(navigate: NavigateFunction): Extension {
    return [textSelectionExtension(), keymap.of([...navigationKeybindings(navigate)])]
}
