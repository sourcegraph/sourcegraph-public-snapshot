import { EditorView } from '@codemirror/view'
import create from 'zustand'

import { lineScrollEnforcing, SelectedLineRange, setSelectedLines } from './codemirror/linenumbers'

interface BlobStore {
    editView: EditorView | null
}

export const useBlobUIStore = create<BlobStore>(() => ({ editView: null }))

/**
 * [PRIVATE/INTERNAL] blob UI API, it's supposed to be used only for set
 * edit view object and should not be used anywhere outside of blob UI component.
 */
export const setBlobEditView = (editView: EditorView | null): void => {
    useBlobUIStore.setState({ editView })
}

// Public blob UI API
export const scrollIntoView = (target: SelectedLineRange): void => {
    const { editView } = useBlobUIStore.getState()

    if (editView) {
        editView.dispatch({
            effects: setSelectedLines.of(target),
            annotations: lineScrollEnforcing.of('scroll-enforcing'),
        })
    }
}
