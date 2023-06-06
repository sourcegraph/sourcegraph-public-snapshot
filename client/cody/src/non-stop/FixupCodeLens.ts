import * as vscode from 'vscode'

import { getSingleLineRange } from '../services/InlineAssist'

import { CodyTaskState } from './utils'

export class FixupCodeLens implements vscode.CodeLensProvider {
    private selectionRange: vscode.Range | null = null
    private static lenses: FixupCodeLens

    private status = CodyTaskState.idle

    private _disposables: vscode.Disposable[] = []
    private _onDidChangeCodeLenses: vscode.EventEmitter<void> = new vscode.EventEmitter<void>()
    public readonly onDidChangeCodeLenses: vscode.Event<void> = this._onDidChangeCodeLenses.event

    constructor(public id = '') {
        this._disposables.push(vscode.languages.registerCodeLensProvider({ scheme: 'file' }, this))
        // register a command to dispose the code lens
        this._disposables.push(
            vscode.commands.registerCommand('cody.fixup.codelens.save', (id: string) => this.save(id)),
            vscode.commands.registerCommand('cody.fixup.codelens.remove', (id: string) => this.remove(id))
        )
        // create a code lens and provider
        this.provideCodeLenses = this.provideCodeLenses.bind(this)
    }
    /**
     * Getter
     */
    public static get instance(): FixupCodeLens {
        return (this.lenses ??= new this())
    }
    /**
     * Define Current States
     */
    public updateState(state: CodyTaskState, startLine: number): void {
        this.status = state
        const newRange = getSingleLineRange(startLine)
        this.selectionRange = newRange
        this._onDidChangeCodeLenses.fire()
    }
    private save(id: string): void {
        if (id && id !== this.id) {
            return
        }
        // execute command to save current file
        void vscode.commands.executeCommand('workbench.action.files.save')
        this.remove(id)
    }
    /**
     * Remove all lenses and decorations created for task
     */
    public remove(id?: string): void {
        if (id && id !== this.id) {
            return
        }
        this.selectionRange = null
        this.status = CodyTaskState.idle
        this.dispose()
        this._onDidChangeCodeLenses.fire()
    }
    /**
     * Activate code lenses
     */
    public provideCodeLenses(): vscode.CodeLens[] | Thenable<vscode.CodeLens[]> {
        return this.createCodeLenses()
    }
    /**
     * Lenses to display above the code that Cody edited
     */
    private createCodeLenses(): vscode.CodeLens[] {
        const range = this.selectionRange
        if (!range) {
            return []
        }
        const codeLensRange = getSingleLineRange(range.start.line)
        return this.status === CodyTaskState.error
            ? getErrorLenses(codeLensRange, this.id)
            : getLenses(codeLensRange, this.status, this.isPending())
    }
    /**
     * Check if it is in pending state
     */
    public isPending(): boolean {
        return this.status === CodyTaskState.pending
    }
    /**
     * Dispose the disposables
     */
    public dispose(): void {
        for (const disposable of this._disposables) {
            disposable.dispose()
        }
        this._disposables = []
    }
}

function getLenses(codeLensRange: vscode.Range, state: CodyTaskState, isPending: boolean): vscode.CodeLens[] {
    const titleLens = new vscode.CodeLens(codeLensRange)
    const codeLensTitle =
        state === CodyTaskState.applying
            ? '$(sync~spin) Applying by Cody'
            : isPending
            ? '$(sync~spin) Processing by Cody'
            : '✨ Edited by Cody'
    // Open Chat View
    titleLens.command = {
        title: codeLensTitle,
        tooltip: 'Open Cody chat view',
        command: 'cody.focus',
    }
    const saveLens = new vscode.CodeLens(codeLensRange)
    saveLens.command = {
        title: 'Save',
        tooltip: 'Accept and save all changes',
        command: 'cody.fixup.codelens.save',
    }

    return isPending ? [titleLens] : [titleLens, saveLens]
}

function getErrorLenses(codeLensRange: vscode.Range, id: string): vscode.CodeLens[] {
    const codeLensError = new vscode.CodeLens(codeLensRange)
    codeLensError.command = {
        title: '⛔️ Not Edited by Cody',
        tooltip: 'Open Cody chat view',
        command: 'cody.focus',
    }
    const codeLensClose = new vscode.CodeLens(codeLensRange)
    codeLensClose.command = {
        title: 'Close',
        tooltip: 'Click to remove decorations',
        command: 'cody.fixup.codelens.remove',
        arguments: [id],
    }
    return [codeLensError, codeLensClose]
}
