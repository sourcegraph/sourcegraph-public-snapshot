import * as vscode from 'vscode'

import { FileChatProvider } from './FileChatProvider'

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
        if (!document || !token) {
            return []
        }
        return this.fixupLenses()
    }

    private fixupLenses(): vscode.CodeLens[] {
        const codeLenses: vscode.CodeLens[] = []
        for (const range of this.ranges) {
            const codeLensTitle = new vscode.CodeLens(range)
            // Open Chat View
            codeLensTitle.command = {
                title: 'Edited by Cody',
                tooltip: 'Click here to open chat view for details.',
                command: 'cody.focus',
            }
            // Run VS Code command to show diff for current file
            const codeLensDiff = new vscode.CodeLens(range)
            codeLensDiff.command = {
                title: 'Show Diff',
                tooltip: 'Open diff view.',
                command: 'workbench.files.action.compareWithSaved',
            }
            // Run VS Code command to save all files
            const codeLensSave = new vscode.CodeLens(range)
            codeLensSave.command = {
                title: 'Accept',
                tooltip: 'Accept and save all changes',
                command: 'workbench.action.files.save',
            }
            codeLenses.push(codeLensTitle, codeLensDiff, codeLensSave)
        }
        return codeLenses
    }
}
