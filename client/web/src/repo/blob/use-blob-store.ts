import type { EditorView } from '@codemirror/view'
import create from 'zustand'

import { lineScrollEnforcing, type SelectedLineRange, setSelectedLines } from './codemirror/linenumbers'

interface BlobStore {
    editorView: EditorView | null
}

const useBlobUIStore = create<BlobStore>(() => ({ editorView: null }))

/**
 * [PRIVATE/INTERNAL] blob UI API, it's supposed to be used only for set
 * edit view object and should not be used anywhere outside of blob UI component.
 */
export const setBlobEditView = (editView: EditorView | null): void => {
    useBlobUIStore.setState({ editorView: editView })
}

export const getBlobEditView = (): EditorView | null => useBlobUIStore.getState()?.editorView

// Public blob UI API
export const scrollIntoView = (target: SelectedLineRange): void => {
    const { editorView } = useBlobUIStore.getState()

    if (editorView) {
        editorView.dispatch({
            effects: setSelectedLines.of(target),
            annotations: lineScrollEnforcing.of('scroll-enforcing'),
        })
    }
}
