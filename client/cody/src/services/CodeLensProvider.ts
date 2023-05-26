import * as vscode from 'vscode'

import { DecorationProvider } from './DecorationProvider'
import { CodyTaskState, getSingleLineRange, updateRangeOnDocChange } from './InlineAssist'

export class CodeLensProvider implements vscode.CodeLensProvider {
    private selectionRange: vscode.Range | null = null
    private static lenses: CodeLensProvider

    private status = CodyTaskState.idle
    public decorator: DecorationProvider

    private _disposables: vscode.Disposable[] = []
    private _onDidChangeCodeLenses: vscode.EventEmitter<void> = new vscode.EventEmitter<void>()
    public readonly onDidChangeCodeLenses: vscode.Event<void> = this._onDidChangeCodeLenses.event

    constructor(public id = '', private extPath = '', private fileUri: vscode.Uri | null = null) {
        this.decorator = new DecorationProvider(this.id, this.extPath)
        vscode.workspace.onDidChangeTextDocument(e => {
            if (e.document.uri.fsPath !== this.fileUri?.fsPath) {
                return
            }
            for (const change of e.contentChanges) {
                if (
                    !this.selectionRange ||
                    (change.range.end.line > this.selectionRange.start.line && this.isPending())
                ) {
                    return
                }
                if (change.range.start.line === this.selectionRange?.start.line && !this.isPending()) {
                    this.remove()
                    return
                }
                this.selectionRange = updateRangeOnDocChange(this.selectionRange, change.range, change.text)
                this.decorator.setState(this.status, this.selectionRange)
            }
            this._onDidChangeCodeLenses.fire()
        })
        vscode.workspace.onDidCloseTextDocument(e => this.removeOnFSPath(e.uri))
        vscode.workspace.onDidSaveTextDocument(e => this.removeOnFSPath(e.uri))
    }
    /**
     * Getter
     */
    public static get instance(): CodeLensProvider {
        return (this.lenses ??= new this())
    }
    /**
     * Define Current States
     */
    public updateState(state: CodyTaskState, newRange: vscode.Range): void {
        this.status = state
        this.decorator.setState(state, newRange)
        void this.decorator.decorate(newRange)
        this.selectionRange = newRange
        this._onDidChangeCodeLenses.fire()
    }
    /**
     * Remove all lenses and decorations created for task
     */
    public remove(): void {
        this.decorator.remove()
        this.selectionRange = null
        this.status = CodyTaskState.idle
        this.dispose()
        this._onDidChangeCodeLenses.fire()
    }
    /**
     * Activate code lenses
     */
    public provideCodeLenses(
        document: vscode.TextDocument,
        token: vscode.CancellationToken
    ): vscode.CodeLens[] | Thenable<vscode.CodeLens[]> {
        // Only create Code lens in known filePath
        if (!document || !token || document.uri.fsPath !== this.fileUri?.fsPath) {
            return []
        }
        this.decorator.setFileUri(this.fileUri)
        return this.createCodeLenses()
    }
    /**
     * Lenses to display above the code that Cody edited
     */
    private createCodeLenses(): vscode.CodeLens[] {
        const range = this.selectionRange
        const codeLenses: vscode.CodeLens[] = []
        if (!range) {
            return codeLenses
        }
        const codeLensRange = getSingleLineRange(range.start.line)
        const codeLensTitle = new vscode.CodeLens(codeLensRange)
        // Open Chat View
        codeLensTitle.command = {
            title: this.isPending() ? '$(sync~spin) Processing by Cody' : '✨ Edited by Cody',
            tooltip: 'Open Cody chat view',
            command: 'cody.focus',
        }
        codeLenses.push(codeLensTitle)
        // Remove decorations
        if (!this.isPending()) {
            const codeLensSave = new vscode.CodeLens(codeLensRange)
            codeLensSave.command = {
                title: 'Save',
                tooltip: 'Accept and save all changes',
                command: 'workbench.action.files.save',
            }
            codeLenses.push(codeLensSave)
        }
        return codeLenses
    }
    /**
     * Check if the file path is the same
     */
    private removeOnFSPath(uri: vscode.Uri): void {
        if (uri.fsPath === this.fileUri?.fsPath) {
            this.remove()
        }
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
