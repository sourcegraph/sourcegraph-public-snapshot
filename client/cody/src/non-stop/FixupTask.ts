import * as vscode from 'vscode'

import { contentSanitizer } from '@sourcegraph/cody-shared/src/chat/recipes/helpers'
import { ActiveTextEditorSelection } from '@sourcegraph/cody-shared/src/editor'

import { debug } from '../log'
import { editDocByUri } from '../services/InlineAssist'

import { Diff } from './diff'
import { FixupFile } from './FixupFile'
import { CodyTaskState } from './utils'

export type taskID = string

export class FixupTask {
    public id: taskID
    private outputChannel = debug
    // TODO: Consider switching to line-based ranges like inline assist
    // In that case we probably *also* need a "point" to feed the LLM
    // because people write instructions like "replace the keys in this hash"
    // and the LLM needs to know where the cursor is.
    public selectionRange: vscode.Range
    public state: CodyTaskState = CodyTaskState.idle
    // The original text that we're working on updating
    public readonly original: string
    // The text of the streaming turn of the LLM, if any
    public inProgressReplacement: string | undefined
    // The text of the last completed turn of the LLM, if any
    public replacement: string | undefined
    // If text has been received from the LLM and a diff has been computed, it
    // is cached here. Diffs are recomputed lazily and may be stale.
    public diff: Diff | undefined

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
        if (this.state === CodyTaskState.error) {
            throw new Error('invalid transition out of error sink state')
        }
        this.state = state
    }

    public start(): void {
        this.setState(CodyTaskState.asking)
        this.output(`Task #${this.id} is currently being processed...`)
        void vscode.commands.executeCommand('setContext', 'cody.fixup.running', true)
    }

    public stop(): void {
        this.setState(CodyTaskState.ready)
        this.output(`Task #${this.id} is ready for fixup...`)
        void vscode.commands.executeCommand('setContext', 'cody.fixup.running', false)
    }

    public error(text: string = ''): void {
        this.setState(CodyTaskState.error)
        this.output(`Error for Task #${this.id} - ` + text)
        void vscode.commands.executeCommand('setContext', 'cody.fixup.running', false)
    }

    public async apply(): Promise<void> {
        this.setState(CodyTaskState.applying)
        this.output(`Task #${this.id} is being applied...`)
        await this.replaceSelection()
    }

    public queue(): void {
        this.setState(CodyTaskState.queued)
        this.output(`Task #${this.id} has been added to the queue successfully...`)
    }

    public marking(): void {
        this.setState(CodyTaskState.marking)
        this.output(`Cody is making the fixups for #${this.id}...`)
    }

    private fixed(): void {
        this.setState(CodyTaskState.fixed)
        this.output(`Task #${this.id} is fixed and completed.`)
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

    // TODO (dom) delete this once the new replacement edit method is in place
    private async replaceSelection(): Promise<void> {
        console.log('replacing start')
        const { editor, selectionRange, replacement } = this
        if (!editor || !replacement) {
            this.error()
            return
        }
        const newRange = await editDocByUri(
            editor.document.uri,
            { start: selectionRange.start.line, end: selectionRange.end.line + 1 },
            contentSanitizer(replacement)
        )
        this.selectionRange = newRange
        this.output('opening fixed doc')
        await vscode.window.showTextDocument(editor.document.uri, { selection: newRange })
        this.fixed()
    }
}
