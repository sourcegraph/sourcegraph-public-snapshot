import { Facet } from '@codemirror/state'
import { EditorView, keymap } from '@codemirror/view'

import { BlobProps } from '../Blob'

/**
 * This facet is a necessary evil to allow access to props passed to the blob
 * component from React components rendered by extensions.
 */
export const blobPropsFacet = Facet.define<BlobProps, BlobProps>({
    combine: props => props[0],
})

const noop = (): true => true

/**
 * Extension to hide/disable the caret. This is needed if CodeMirror should stay
 * focusable (EditorView.editable is true) but we still want to hide the caret.
 */
export const hideCaret = [
    // Render caret invisible
    EditorView.theme({
        '.cm-line': {
            caretColor: 'transparent !important',
        },
    }),
    // Disable basic cursor movement keys
    keymap.of([
        {
            key: 'ArrowUp',
            run: noop,
        },
        {
            key: 'ArrowDown',
            run: noop,
        },
        {
            key: 'ArrowLeft',
            run: noop,
        },
        {
            key: 'ArrowRight',
            run: noop,
        },
    ]),
]
