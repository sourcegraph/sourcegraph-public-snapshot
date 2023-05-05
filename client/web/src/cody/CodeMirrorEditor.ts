import {
    ActiveTextEditor,
    ActiveTextEditorSelection,
    ActiveTextEditorVisibleContent,
    Editor,
} from '@sourcegraph/cody-shared/src/editor'

import { EditorStore } from '../stores/editor'

export class CodeMirrorEditor implements Editor {
    private editorStoreRef: React.MutableRefObject<EditorStore>
    constructor(editorStoreRef: React.MutableRefObject<EditorStore>) {
        this.editorStoreRef = editorStoreRef
    }

    public getWorkspaceRootPath(): string | null {
        return null
    }

    public getActiveTextEditor(): ActiveTextEditor | null {
        const editor = this.editorStoreRef.current.editor
        if (editor === null) {
            return null
        }

        return { content: editor.content, filePath: editor.filename }
    }

    public getActiveTextEditorSelection(): ActiveTextEditorSelection | null {
        const editor = this.editorStoreRef.current.editor

        if (!editor || editor.view.state.selection.main.empty) {
            return null
        }

        const selection = editor.view?.state.selection.main
        const { head, anchor } = selection

        if (head !== anchor) {
            const precedingText = editor.view?.state.sliceDoc(undefined, selection.from)
            const selectedText = editor.view?.state.sliceDoc(selection.from, selection.to)
            const followingText = editor.view?.state.sliceDoc(selection.to, undefined)

            return {
                fileName: editor.filename,
                precedingText,
                selectedText,
                followingText,
            }
        }

        return null
    }

    public getActiveTextEditorSelectionOrEntireFile(): ActiveTextEditorSelection | null {
        const editor = this.editorStoreRef.current.editor
        if (editor === null) {
            return null
        }

        const selection = this.getActiveTextEditorSelection()
        if (selection) {
            return selection
        }

        return {
            fileName: editor.filename,
            precedingText: '',
            selectedText: editor.content,
            followingText: '',
        }
    }

    public getActiveTextEditorVisibleContent(): ActiveTextEditorVisibleContent | null {
        const editor = this.editorStoreRef.current.editor
        if (editor === null) {
            return null
        }

        const { from, to } = editor.view.viewport

        const content = editor.view?.state.sliceDoc(from, to)
        return {
            fileName: editor.filename,
            content,
        }
    }

    public replaceSelection(_fileName: string, _selectedText: string, _replacement: string): Promise<void> {
        return Promise.resolve()
    }

    public showQuickPick(labels: string[]): Promise<string | undefined> {
        // TODO: Use a proper UI element
        return Promise.resolve(window.prompt(`Choose between: ${labels.join(', ')}`, labels[0]) || undefined)
    }

    public async showWarningMessage(message: string): Promise<void> {
        // TODO: Use a proper UI element
        // eslint-disable-next-line no-console
        console.warn(message)
        return Promise.resolve()
    }

    public showInputBox(): Promise<string | undefined> {
        // TODO: Use a proper UI element
        return Promise.resolve(window.prompt('Enter your answer: ') || undefined)
    }
}
