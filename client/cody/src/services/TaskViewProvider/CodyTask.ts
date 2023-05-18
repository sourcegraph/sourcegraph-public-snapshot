import * as vscode from 'vscode'

import { ActiveTextEditorSelection } from '@sourcegraph/cody-shared/src/editor'

import { codyOutputChannel } from '../../log'
import { editDocByUri, getFixupEditorSelection } from '../utils'

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

// TODO(bee): Create CodeLens for each task
// TODO(bee): Create decorator for each task
// TODO(dpc): Add listener for document change to track range
export class CodyTask extends vscode.TreeItem {
    private outputChannel = codyOutputChannel
    public contextValue: string
    private replacementContent = ''
    public readonly label: string
    public selectionRange: vscode.Range | vscode.Selection = new vscode.Range(0, 0, 0, 0)
    public state: CodyTaskState = CodyTaskState.queued
    private readonly content: string
    public readonly documentUri: vscode.Uri

    constructor(
        public readonly id: string,
        private input: string,
        private selection: ActiveTextEditorSelection,
        private editor: vscode.TextEditor
    ) {
        super(input)
        this.label = this.input
        this.tooltip = `${this.label}-${selection.fileName}`
        this.description = selection.fileName
        this.content = selection.selectedText
        this.contextValue = 'cody.tasks'
        this.documentUri = this.editor.document.uri

        const icon = this.taskIcons[this.state].icon
        const mode = this.taskIcons[this.state].id
        this.iconPath = new vscode.ThemeIcon(icon, new vscode.ThemeColor(mode))
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
        this.output(`Starting Task#${this.id}...`)
    }

    public stop(): void {
        this.set(CodyTaskState.done)
        this.output(`Task#${this.id}... has been completed...`)
    }

    public error(text: string = ''): void {
        this.set(CodyTaskState.error)
        this.output(`Error from Task#${this.id}: ` + text, true)
    }

    public queue(): void {
        this.set(CodyTaskState.queued)
        this.output(`Added to Fixup queue: Task#${this.id}`)
    }
    /**
     * Print output to the VS Code Output Channel under Cody AI by Sourcegraph
     */
    private output(text: string, show = false): void {
        this.outputChannel.appendLine('Non-Stop Cody: ' + text)
        if (show) {
            this.outputChannel.show()
        }
    }
    /**
     * Do replacement in doc and update code lens and decorator
     */
    public async replace(newContent: string, newRange: vscode.Range): Promise<void> {
        this.replacementContent = newContent
        this.selectionRange = newRange

        if (!newContent.trim() || newContent.trim() === this.selection.selectedText.trim()) {
            this.error('New content is empty')
            return
        }

        await editDocByUri(
            this.documentUri,
            { start: newRange.start.line, end: newRange.end.line },
            this.replacementContent
        )

        this.stop()
    }
    /**
     * Get current selection info from doc uri
     */
    public async makeSelection(): Promise<ActiveTextEditorSelection | null> {
        const { selection, selectionRange } = await getFixupEditorSelection(this.documentUri, this.selectionRange)
        this.selectionRange = selectionRange
        this.selection = selection
        return selection
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
