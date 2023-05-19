import * as vscode from 'vscode'

import { ActiveTextEditorSelection } from '@sourcegraph/cody-shared/src/editor'

import { codyOutputChannel } from '../log'
import { editDocByUri, getFixupEditorSelection } from '../services/utils'

import { CodyTaskState, fixupTaskIcon } from './types'

// TODO(bee): Create CodeLens for each task
// TODO(bee): Create decorator for each task
// TODO(dpc): Add listener for document change to track range
export class FixupTask {
    private outputChannel = codyOutputChannel
    private readonly content: string
    private replacementContent = ''
    public iconPath: vscode.ThemeIcon | undefined = undefined
    public selectionRange: vscode.Range
    public state = CodyTaskState.idle
    public readonly documentUri: vscode.Uri

    constructor(
        public readonly id: string,
        public readonly instruction: string,
        public selection: ActiveTextEditorSelection,
        private readonly editor: vscode.TextEditor
    ) {
        this.selectionRange = editor.selection
        this.content = selection.selectedText
        this.documentUri = this.editor.document.uri

        this.set(CodyTaskState.queued)
    }
    /**
     * Set task state and update icon
     */
    private set(state: CodyTaskState): void {
        this.state = state
        const icon = fixupTaskIcon[this.state].icon
        const mode = fixupTaskIcon[this.state].id
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
        this.output(`Error for Task#${this.id}: ` + text, true)
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
            this.error('Cody did not provide any replacement for your request.')
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
}
