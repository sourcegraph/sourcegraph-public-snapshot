import { type Extension, Prec } from '@codemirror/state'
import { type EditorView, keymap } from '@codemirror/view'

/**
 * Focuses the editor and positions the cursor at the end of the document.
 * This only happens when the editor doesn't already have focus.
 */
export function focusEditor(editor: EditorView): void {
    if (!editor.hasFocus) {
        editor.focus()
        editor.dispatch({
            selection: { anchor: editor.state.doc.length },
            scrollIntoView: true,
        })
    }
}

/**
 * Defines default keybindings for input elements inside blocks:
 * - Mod+Enter will call 'runBlock'
 * - Escape will remove focus from the editor
 */
export function blockKeymap({ runBlock }: { runBlock: () => void }): Extension {
    return Prec.high(
        keymap.of([
            {
                key: 'Mod-Enter',
                run: () => {
                    runBlock()
                    return true
                },
            },
            {
                key: 'Escape',
                run: view => {
                    view.contentDOM.blur()
                    return true
                },
            },
        ])
    )
}
