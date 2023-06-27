import { relative } from 'path'
import url from 'url'

import { CompletionsTextEditor, LightTextDocument } from '@sourcegraph/cody-shared/src/autocomplete'
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

export class AgentEditor implements Editor, CompletionsTextEditor {
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

    // Completions
    // TODO: Unify, use file paths all over

    public getOpenDocuments(): LightTextDocument[] {
        return [...this.agent.documents.values()].map(_ => ({
            uri: url.pathToFileURL(_.filePath).toString(),
            languageId: 'TODO',
        }))
    }

    public getCurrentDocument(): LightTextDocument | null {
        const active = this.agent.activeDocumentFilePath

        if (!active) {
            return null
        }

        return {
            uri: url.pathToFileURL(active).toString(),
            languageId: 'TODO',
        }
    }

    public getDocumentTextTruncated(uri: string): Promise<string | null> {
        const doc = this.agent.documents.get(url.fileURLToPath(uri))

        if (!doc || !doc.content) {
            return Promise.resolve(null)
        }

        return Promise.resolve(doc.content.slice(0, 100_000))
    }

    public getDocumentRelativePath(uri: string): Promise<string | null> {
        const rootPath = this.agent.workspaceRootPath

        if (!rootPath) {
            return Promise.resolve(null)
        }

        return Promise.resolve(relative(rootPath, url.parse(uri).pathname!))
    }

    public getTabSize(): number {
        return 4
    }
}
