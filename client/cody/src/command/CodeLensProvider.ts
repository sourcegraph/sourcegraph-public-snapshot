import * as vscode from 'vscode'

import { FileChatProvider } from '../chat/FileChatProvider'

export class CodeLensProvider implements vscode.CodeLensProvider {
    public ranges: vscode.Range[] = []
    private _onDidChangeCodeLenses: vscode.EventEmitter<void> = new vscode.EventEmitter<void>()
    public readonly onDidChangeCodeLenses: vscode.Event<void> = this._onDidChangeCodeLenses.event
    private static lenes: CodeLensProvider
    private fileChatProvider: FileChatProvider | null = null

    constructor() {
        vscode.workspace.onDidChangeConfiguration(() => {
            this._onDidChangeCodeLenses.fire()
        })
        vscode.workspace.onDidSaveTextDocument(() => {
            this.ranges = []
            this.fileChatProvider?.removeDecorate()
            this._onDidChangeCodeLenses.fire()
        })
    }

    public static get instance(): CodeLensProvider {
        return (this.lenes ??= new this())
    }

    public set(line: number, fileChatProvider: FileChatProvider): void {
        this.fileChatProvider = fileChatProvider
        this.ranges = []
        this.ranges.push(new vscode.Range(line, 0, line, 1))
        this._onDidChangeCodeLenses.fire()
    }

    public remove(): void {
        this.ranges = []
        this._onDidChangeCodeLenses.fire()
    }

    public provideCodeLenses(
        document: vscode.TextDocument,
        token: vscode.CancellationToken
    ): vscode.CodeLens[] | Thenable<vscode.CodeLens[]> {
        const codeLenses: vscode.CodeLens[] = []
        if (!document || !token) {
            return codeLenses
        }
        for (const range of this.ranges) {
            const codeLens = new vscode.CodeLens(range)
            codeLens.command = {
                title: 'Edited by Cody',
                tooltip: 'Click here to open chat view for details.',
                command: 'cody.focus',
            }
            const codeLens1 = new vscode.CodeLens(range)
            codeLens1.command = {
                title: 'Show Diff',
                tooltip: 'Open diff view.',
                command: 'workbench.files.action.compareWithSaved',
            }
            const codeLens2 = new vscode.CodeLens(range)
            codeLens2.command = {
                title: 'Accept',
                tooltip: 'Accept and save all changes',
                command: 'workbench.action.files.save',
            }
            codeLenses.push(codeLens, codeLens1, codeLens2)
        }
        return codeLenses
    }
}
