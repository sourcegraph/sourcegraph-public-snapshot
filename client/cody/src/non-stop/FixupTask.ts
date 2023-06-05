import * as vscode from 'vscode'

import { ActiveTextEditorSelection } from '@sourcegraph/cody-shared/src/editor'

import { debug } from '../log'

import { FixupFile } from './FixupFile'
import { CodyTaskState } from './utils'

// TODO(bee): Create CodeLens for each task
// TODO(bee): Create decorator for each task
export class FixupTask {
    public id: string
    private outputChannel = debug
    // TODO: Consider switching to line-based ranges like inline assist
    // In that case we probably *also* need a "point" to feed the LLM
    // because people write instructions like "replace the keys in this hash"
    // and the LLM needs to know where the cursor is.
    public selectionRange: vscode.Range
    public state = CodyTaskState.idle
    // The original text that we're working on updating
    public readonly original: string
    // The text of the streaming turn of the LLM, if any
    public inProgressReplacement: string | undefined
    // The text of the last completed turn of the LLM, if any
    public replacement: string | undefined

    constructor(
        public readonly fixupFile: FixupFile,
        public readonly instruction: string,
        public selection: ActiveTextEditorSelection,
        public readonly editor: vscode.TextEditor
    ) {
        this.id = Date.now().toString(36).replace(/\d+/g, '')
        this.selectionRange = editor.selection
        this.original = selection.selectedText
        this.queue()
    }
    /**
     * Set latest state for task and then update icon accordingly
     */
    private setState(state: CodyTaskState): void {
        if (this.state !== CodyTaskState.error) {
            this.state = state
        }
    }

    public start(): void {
        this.setState(CodyTaskState.pending)
        this.output(`Task #${this.id} is currently being processed...`)
    }

    public stop(): void {
        this.setState(CodyTaskState.done)
        this.output(`Task #${this.id} has been completed...`)
    }

    public error(text: string = ''): void {
        this.setState(CodyTaskState.error)
        this.output(`Error for Task #${this.id} - ` + text)
    }

    public apply(): void {
        this.setState(CodyTaskState.applying)
        this.output(`Task #${this.id} is being applied...`)
    }

    private queue(): void {
        this.setState(CodyTaskState.queued)
        this.output(`Task #${this.id} has been added to the queue successfully...`)
    }
    /**
     * Print output to the VS Code Output Channel under Cody AI by Sourcegraph
     */
    private output(text: string): void {
        this.outputChannel('Cody Fixups:', text)
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
}
