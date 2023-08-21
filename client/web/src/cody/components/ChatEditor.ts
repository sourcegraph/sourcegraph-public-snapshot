import { URI } from 'vscode-uri'

import {
    ActiveTextEditor,
    ActiveTextEditorDiagnostic,
    ActiveTextEditorSelection,
    ActiveTextEditorVisibleContent,
    Editor,
} from '@sourcegraph/cody-shared/dist/editor'

export interface EditorStore {
    selectedText: string
    content: string
    filename: string
    repo: string
    revision: string
}

export class ChatEditor implements Editor {
    private editor?: EditorStore | null
    constructor(editor?: EditorStore | null) {
        this.editor = editor
    }

    public get fileName(): string {
        return this.editor?.filename || ''
    }

    public get repoName(): string | undefined {
        return this.editor?.repo
    }

    public get revision(): string | undefined {
        return this.editor?.revision
    }

    public getActiveTextEditor(): ActiveTextEditor | null {
        const editor = this.editor
        if (editor) {
            return {
                content: editor.content,
                filePath: editor.filename,
                repoName: this.repoName,
                revision: this.revision,
            }
        }
        return null
    }

    public getActiveTextEditorSelection(): ActiveTextEditorSelection | null {
        const editor = this.editor

        if (!editor?.content) {
            return null
        }

        const splitText = editor.content.split(editor.selectedText)
        const precedingText = splitText[0]
        splitText.shift()
        const selectedText = editor.selectedText
        const followingText = splitText.join('')

        return {
            fileName: editor.filename,
            repoName: this.repoName,
            revision: this.revision,
            precedingText,
            selectedText,
            followingText,
        }
    }

    public getActiveTextEditorSelectionOrEntireFile(): ActiveTextEditorSelection | null {
        if (this.editor) {
            const selection = this.getActiveTextEditorSelection()
            if (selection) {
                return selection
            }

            return {
                fileName: this.editor.filename,
                repoName: this.repoName,
                revision: this.revision,
                precedingText: '',
                selectedText: this.editor.selectedText,
                followingText: '',
            }
        }

        return null
    }

    public getActiveTextEditorVisibleContent(): ActiveTextEditorVisibleContent | null {
        const editor = this.editor
        if (editor) {
            return {
                fileName: editor.filename,
                repoName: this.repoName,
                revision: this.revision,
                content: editor.selectedText,
            }
        }

        return null
    }

    /** The path of the workspace root if on the file system, otherwise `null`. */
    public getWorkspaceRootPath(): string | null {
        return null
    }

    /** Gets the active text editor's selection, or the visible content if the selected range is empty. */
    public getActiveTextEditorSelectionOrVisibleContent(): ActiveTextEditorSelection | null {
        return null
    }

    public getActiveInlineChatTextEditor(): ActiveTextEditor | null {
        return null
    }

    public getActiveInlineChatSelection(): ActiveTextEditorSelection | null {
        return null
    }

    public replaceSelection(_fileName: string, _selectedText: string, _replacement: string): Promise<void> {
        // Not implemented.
        return Promise.resolve()
    }

    public showQuickPick(labels: string[]): Promise<string | undefined> {
        // Not implemented.
        return Promise.resolve(window.prompt(`Choose between: ${labels.join(', ')}`, labels[0]) || undefined)
    }

    public async showWarningMessage(message: string): Promise<void> {
        // Not implemented.
        // eslint-disable-next-line no-console
        console.warn(message)
        return Promise.resolve()
    }

    public showInputBox(): Promise<string | undefined> {
        // Not implemented.
        return Promise.resolve(window.prompt('Enter your answer: ') || undefined)
    }

    public didReceiveFixupText(id: string, text: string, state: 'streaming' | 'complete'): Promise<void> {
        // Not implemented.
        return Promise.resolve(undefined)
    }

    /** Get diagnostics (errors, warnings, hints) for a range within the active text editor. */
    public getActiveTextEditorDiagnosticsForRange(): ActiveTextEditorDiagnostic[] | null {
        return null
    }

    /** Not implemented: The URI of the workspace root. */
    public getWorkspaceRootUri(): URI | null {
        return null
    }
}
