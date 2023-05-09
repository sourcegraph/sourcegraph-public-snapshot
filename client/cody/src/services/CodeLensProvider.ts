import * as vscode from 'vscode'

import { DecorationProvider } from './DecorationProvider'

export class CodeLensProvider implements vscode.CodeLensProvider {
    public ranges: vscode.Range | null = null
    private static lenses: CodeLensProvider

    private fileUri: vscode.Uri | null = null

    public isPending = false
    public status = 'none'

    public decorator: DecorationProvider

    private _disposables: vscode.Disposable[] = []
    private _onDidChangeCodeLenses: vscode.EventEmitter<void> = new vscode.EventEmitter<void>()
    public readonly onDidChangeCodeLenses: vscode.Event<void> = this._onDidChangeCodeLenses.event

    constructor(public id = '', private extPath = '') {
        this.decorator = new DecorationProvider(this.id, this.extPath)

        vscode.workspace.onDidCloseTextDocument(e => {
            if (e.uri.fsPath === this.fileUri?.fsPath) {
                this.remove()
            }
        })

        vscode.workspace.onDidSaveTextDocument(e => {
            if (e.uri.fsPath === this.fileUri?.fsPath) {
                this.remove()
            }
        })

        vscode.workspace.onDidChangeTextDocument(e => {
            if (e.document.uri.fsPath !== this.fileUri?.fsPath) {
                return
            }
            for (const change of e.contentChanges) {
                if (!this.ranges) {
                    return
                }
                if (change.range.start.line + 1 >= this.ranges?.start.line && !this.isPending) {
                    this.remove()
                    return
                }
                if (change.range.end.line > this.ranges.start.line) {
                    return
                }
                let addedLines = 0
                if (change.text.includes('\n')) {
                    addedLines = change.text.split('\n').length - 1
                } else if (change.range.end.line - change.range.start.line > 0) {
                    addedLines -= change.range.end.line - change.range.start.line
                }
                const newRange = new vscode.Range(
                    new vscode.Position(this.ranges.start.line + addedLines, 0),
                    new vscode.Position(this.ranges.end.line + addedLines, 0)
                )
                this.ranges = newRange
                this.decorator.rangeUpdate(newRange)
            }
            this._onDidChangeCodeLenses.fire()
        })
    }

    public static get instance(): CodeLensProvider {
        return (this.lenses ??= new this())
    }

    public updatePendingStatus(pending: boolean, newRange: vscode.Range): void {
        this.decorator.setStatus(pending ? 'pending' : 'done')
        void this.decorator.decorate(newRange)
        this.isPending = pending
        this.ranges = newRange
        this._onDidChangeCodeLenses.fire()
    }

    public newRange(range: vscode.Range): void {
        this.ranges = range
        this._onDidChangeCodeLenses.fire()
    }

    public getNewRange(line: number): vscode.Range {
        return new vscode.Range(line, 0, line, 0)
    }

    /**
     * Remove all lenses and decorations created for task
     */
    public remove(): void {
        this.decorator.remove()
        this.ranges = null
        this.status = 'none'
        this.dispose()
        this._onDidChangeCodeLenses.fire()
    }

    public provideCodeLenses(
        document: vscode.TextDocument,
        token: vscode.CancellationToken
    ): vscode.CodeLens[] | Thenable<vscode.CodeLens[]> {
        if (!document || !token) {
            return []
        }
        this.fileUri = document.uri
        this.decorator.fileUri = document.uri
        return this.makeLenses()
    }

    /**
     * Lenses to display above the code that Cody edited
     */
    private makeLenses(): vscode.CodeLens[] {
        const range = this.ranges
        const codeLenses: vscode.CodeLens[] = []
        if (!range) {
            return []
        }
        const codeLensRange = this.getNewRange(range.start.line)
        const codeLensTitle = new vscode.CodeLens(codeLensRange)
        // Open Chat View
        codeLensTitle.command = {
            title: this.isPending ? '$(sync~spin) Processing by Cody' : '✨ Edited by Cody',
            tooltip: 'Open this in Cody chat view',
            command: 'cody.focus',
        }
        codeLenses.push(codeLensTitle)

        if (!this.isPending) {
            // Remove decorations
            const codeLensSave = new vscode.CodeLens(codeLensRange)
            codeLensSave.command = {
                title: 'Done',
                tooltip: 'Accept and save all changes',
                command: 'workbench.action.files.save',
            }

            codeLenses.push(codeLensSave)
        }
        return codeLenses
    }

    public dispose(): void {
        for (const disposable of this._disposables) {
            disposable.dispose()
        }
        this._disposables = []
    }
}
