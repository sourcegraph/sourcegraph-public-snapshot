import * as vscode from 'vscode'

import { ActiveTextEditorSelection } from '@sourcegraph/cody-shared/src/editor'
import { SURROUNDING_LINES } from '@sourcegraph/cody-shared/src/prompt/constants'

export enum CodyTaskState {
    'idle' = 0,
    'pending' = 1,
    'done' = 2,
    'error' = 3,
    'stopped' = 4,
    'queued' = 5,
}

type CodyTaskIcon = {
    [key in CodyTaskState]: {
        id: string
        icon: string
    }
}

// TODO: Create CodeLens for each task
export class CodyTask extends vscode.TreeItem {
    public contextValue: string
    private replacementContent = ''
    public errorMsg = ''
    public readonly label: string
    public selectionRange: vscode.Range | vscode.Selection = new vscode.Range(0, 0, 0, 0)
    public state: CodyTaskState = CodyTaskState.queued
    private readonly content: string
    public readonly uri: string
    constructor(public readonly id: string, private input: string, private selection: ActiveTextEditorSelection) {
        super(input)
        this.label = this.input
        this.uri = selection.fileName
        this.tooltip = `${this.label}-${this.uri}`
        this.description = selection.fileName
        this.content = selection.selectedText
        this.contextValue = 'cody.tasks'
        const icon = this.taskIcons[this.state].icon
        const mode = this.taskIcons[this.state].id
        this.iconPath = new vscode.ThemeIcon(icon, new vscode.ThemeColor(mode))

        // Get selection range and then create code lens
        this.getEditor()
            .then(editor => {
                this.selectionRange = editor?.selection || new vscode.Range(0, 0, 0, 0)
            })
            .catch(() => {})
    }
    /**
     * Set task state and update icon
     */
    private set(state: CodyTaskState): void {
        this.state = state
        const icon = this.taskIcons[this.state].icon
        const mode = this.taskIcons[this.state].id
        this.iconPath = new vscode.ThemeIcon(icon, new vscode.ThemeColor(mode))
    }

    public start(): void {
        this.set(CodyTaskState.pending)
    }

    public stop(): void {
        this.set(CodyTaskState.done)
    }

    public queue(): void {
        this.set(CodyTaskState.queued)
    }
    /**
     * Do replacement in doc and update code lens and decorator
     */
    public async replace(newContent: string, newRange: vscode.Range): Promise<void> {
        const activeEditor = await this.getEditor()
        if (!activeEditor) {
            this.set(CodyTaskState.error)
            return
        }
        await activeEditor.edit(edit => {
            edit.replace(this.selectionRange, newContent)
        })
        this.replacementContent = newContent
        this.selectionRange = newRange
        this.set(CodyTaskState.done)
    }
    /**
     * Get current selected lines from the comment thread.
     * Add an extra line to the end line to prevent empty selection on single line selection
     */
    public async makeSelection(): Promise<ActiveTextEditorSelection | null> {
        const activeDocument = await vscode.workspace.openTextDocument(this.uri)
        const lineLength = activeDocument.lineAt(this.selectionRange.end.line).text.length
        const startPost = new vscode.Position(this.selectionRange.start.line, 0)
        const endPost = new vscode.Position(this.selectionRange.end.line, lineLength)
        const selectionRange: vscode.Range | vscode.Selection = new vscode.Range(startPost, endPost)
        const precedingText = activeDocument.getText(
            new vscode.Range(
                new vscode.Position(Math.max(0, this.selectionRange.start.line - SURROUNDING_LINES), 0),
                this.selectionRange.start
            )
        )
        const followingText = activeDocument.getText(
            new vscode.Range(
                this.selectionRange.end,
                new vscode.Position(this.selectionRange.end.line + 1 + SURROUNDING_LINES, 0)
            )
        )
        const selection = {
            fileName: vscode.workspace.asRelativePath(this.uri),
            selectedText: activeDocument.getText(selectionRange),
            precedingText,
            followingText,
        }

        this.selectionRange = selectionRange
        this.selection = selection
        return selection
    }
    /**
     * Get Editor by URI
     */
    public async getEditor(): Promise<vscode.TextEditor | null> {
        const activeEditor = vscode.window.activeTextEditor
        if (activeEditor) {
            return activeEditor
        }
        const doc = vscode.Uri.file(this.uri)
        await vscode.window.showTextDocument(doc)
        return vscode.window.activeTextEditor || null
    }
    /**
     * Return latest selection
     */
    public getSelection(): ActiveTextEditorSelection | null {
        return this.selection
    }
    /**
     * Return latest selection range
     */
    public getSelectionRange(): vscode.Range | vscode.Selection {
        return this.selectionRange
    }
    /**
     * Return context returned by Cody
     */
    public getContext(): { original: string; replacement: string } {
        return {
            original: this.content,
            replacement: this.replacementContent,
        }
    }
    /**
     * Task Info
     */
    private taskIcons: CodyTaskIcon = {
        [CodyTaskState.idle]: {
            id: 'idle',
            icon: 'smiley',
        },
        [CodyTaskState.pending]: {
            id: 'pending',
            icon: 'sync~spin',
        },
        [CodyTaskState.done]: {
            id: 'done',
            icon: 'issue-closed',
        },
        [CodyTaskState.error]: {
            id: 'error',
            icon: 'stop',
        },
        [CodyTaskState.queued]: {
            id: 'queue',
            icon: 'clock',
        },
        [CodyTaskState.stopped]: {
            id: 'removed',
            icon: 'circle-slash',
        },
    }
}
