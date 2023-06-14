import {
    ActiveTextEditor,
    ActiveTextEditorSelection,
    ActiveTextEditorViewControllers,
    ActiveTextEditorVisibleContent,
    Editor,
} from '@sourcegraph/cody-shared/src/editor'

import { Agent } from './agent'
import { Offsets } from './offsets'
import { TextDocument } from './protocol'

export class AgentEditor implements Editor {
    public controllers?: ActiveTextEditorViewControllers | undefined

    constructor(private agent: Agent) {}

    public didReceiveFixupText(id: string, text: string, state: 'streaming' | 'complete'): Promise<void> {
        throw new Error('Method not implemented.')
    }

    public getWorkspaceRootPath(): string | null {
        return this.agent.workspaceRootFilePath
    }

    private activeDocument(): TextDocument | undefined {
        if (this.agent.activeDocumentFilePath === null) {
            return undefined
        }
        return this.agent.documents.get(this.agent.activeDocumentFilePath)
    }

    public getActiveTextEditor(): ActiveTextEditor | null {
        const document = this.activeDocument()
        if (document === undefined) {
            return null
        }
        return {
            filePath: document.filePath,
            content: document.content || '',
        }
    }

    public getActiveTextEditorSelection(): ActiveTextEditorSelection | null {
        const document = this.activeDocument()
        if (document === undefined || document.content === undefined || document.selection === undefined) {
            return null
        }
        const offsets = new Offsets(document)
        const from = offsets.offset(document.selection.start)
        const to = offsets.offset(document.selection.end)
        return {
            fileName: document.filePath || '',
            precedingText: document.content.slice(0, from),
            selectedText: document.content.slice(from, to),
            followingText: document.content.slice(to, document.content.length),
        }
    }

    public getActiveTextEditorSelectionOrEntireFile(): ActiveTextEditorSelection | null {
        const document = this.activeDocument()
        if (document !== undefined && document.selection === undefined) {
            return {
                fileName: document.filePath || '',
                precedingText: '',
                selectedText: document.content || '',
                followingText: '',
            }
        }
        return this.getActiveTextEditorSelection()
    }

    public getActiveTextEditorVisibleContent(): ActiveTextEditorVisibleContent | null {
        const document = this.activeDocument()
        if (document === undefined) {
            return null
        }
        return {
            content: document.content || '',
            fileName: document.filePath,
        }
    }

    public async replaceSelection(fileName: string, selectedText: string, replacement: string): Promise<void> {
        // Handle possible failure
        await this.agent.request('editor/replaceSelection', {
            fileName,
            selectedText,
            replacement,
        })
    }

    public async showQuickPick(labels: string[]): Promise<string | undefined> {
        const result = await this.agent.request('editor/quickPick', labels)
        return result || undefined
    }

    public showWarningMessage(message: string): Promise<void> {
        this.agent.notify('editor/warning', message)
        return Promise.resolve()
    }

    public async showInputBox(prompt?: string | undefined): Promise<string | undefined> {
        return (await this.agent.request('editor/prompt', prompt || '')) || undefined
    }
}
