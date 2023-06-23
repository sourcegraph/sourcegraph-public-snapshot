import {
    ActiveTextEditor,
    ActiveTextEditorSelection,
    ActiveTextEditorViewControllers,
    ActiveTextEditorVisibleContent,
    Editor,
} from '@sourcegraph/cody-shared/src/editor'

import { Agent } from './agent'
import { DocumentOffsets } from './offsets'
import { TextDocument } from './protocol'

export class AgentEditor implements Editor {
    public controllers?: ActiveTextEditorViewControllers | undefined

    constructor(private agent: Agent) {}

    public didReceiveFixupText(): Promise<void> {
        throw new Error('Method not implemented.')
    }

    public getWorkspaceRootPath(): string | null {
        return this.agent.workspaceRootPath
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
        const offsets = new DocumentOffsets(document)
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

    public replaceSelection(): Promise<void> {
        throw new Error('Not implemented')
    }

    public showQuickPick(): Promise<string | undefined> {
        throw new Error('Not implemented')
    }

    public showWarningMessage(): Promise<void> {
        throw new Error('Not implemented')
    }

    public showInputBox(): Promise<string | undefined> {
        throw new Error('Not implemented')
    }
}
