import type { EditorView } from '@codemirror/view'
import create from 'zustand'

export interface EditorStore {
    editor: null | {
        filename: string
        repo: string
        content: string
        view: EditorView
    }
}

export const useEditorStore = create<EditorStore>((): EditorStore => ({ editor: null }))
