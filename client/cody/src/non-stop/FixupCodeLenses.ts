import * as vscode from 'vscode'

import { getLensesForTask } from './codelenses'
import { FixupTask } from './FixupTask'
import { FixupFileCollection } from './roles'
import { CodyTaskState } from './utils'

export class FixupCodeLenses implements vscode.CodeLensProvider {
    private taskLenses = new Map<FixupTask, vscode.CodeLens[]>()

    private _disposables: vscode.Disposable[] = []
    private _onDidChangeCodeLenses: vscode.EventEmitter<void> = new vscode.EventEmitter<void>()
    public readonly onDidChangeCodeLenses: vscode.Event<void> = this._onDidChangeCodeLenses.event

    /**
     * Create a code lens provider
     */
    constructor(private readonly files: FixupFileCollection) {
        this.provideCodeLenses = this.provideCodeLenses.bind(this)
        this._disposables.push(vscode.languages.registerCodeLensProvider('*', this))
    }

    /**
     * Gets the code lenses for the specified document.
     */
    public provideCodeLenses(
        document: vscode.TextDocument,
        token: vscode.CancellationToken
    ): vscode.CodeLens[] | Thenable<vscode.CodeLens[]> {
        const file = this.files.maybeFileForUri(document.uri)
        if (!file) {
            return []
        }
        const lenses = []
        for (const task of this.files.tasksForFile(file)) {
            lenses.push(...(this.taskLenses.get(task) || []))
        }
        console.log(`lenses for: ${document.uri} (${lenses.length})`)
        return lenses
    }

    public didUpdateTask(task: FixupTask): void {
        if (task.state === CodyTaskState.fixed || task.state === CodyTaskState.error) {
            this.removeLensesFor(task)
            return
        }
        this.taskLenses.set(task, getLensesForTask(task))
        this.notifyCodeLensesChanged()
    }

    public didDeleteTask(task: FixupTask): void {
        this.removeLensesFor(task)
    }

    private removeLensesFor(task: FixupTask): void {
        if (this.taskLenses.delete(task)) {
            // TODO: Clean up the fixup file when there are no remaining code lenses
            this.notifyCodeLensesChanged()
        }
    }

    private notifyCodeLensesChanged(): void {
        this._onDidChangeCodeLenses.fire()
    }

    /**
     * Dispose the disposables
     */
    public dispose(): void {
        this.taskLenses.clear()
        for (const disposable of this._disposables) {
            disposable.dispose()
        }
        this._disposables = []
    }
}
