import * as vscode from 'vscode'

import { FileChatProvider } from './FileChatProvider'

export class CodeLensProvider implements vscode.CodeLensProvider {
    public ranges: vscode.Range[] = []
    private static lenes: CodeLensProvider
    private updatedLength = 0

    private fileUri: vscode.Uri | null = null
    private fileChatProvider: FileChatProvider | null = null

    private _disposables: vscode.Disposable[] = []
    private _onDidChangeCodeLenses: vscode.EventEmitter<void> = new vscode.EventEmitter<void>()
    public readonly onDidChangeCodeLenses: vscode.Event<void> = this._onDidChangeCodeLenses.event

    constructor(public id: string = '') {
        vscode.workspace.onDidChangeConfiguration(() => {
            this._onDidChangeCodeLenses.fire()
        })
        // vscode.workspace.onDidChangeTextDocument(() => {
        //     this._onDidChangeCodeLenses.fire()
        // })
        vscode.workspace.onDidSaveTextDocument(() => {
            // TODO: Close comment thread on save
            this.ranges = []
            this.dispose()
            this._onDidChangeCodeLenses.fire()
        })
    }

    public static get instance(): CodeLensProvider {
        return (this.lenes ??= new this())
    }

    public set(line: number, fileChatProvider: FileChatProvider, updatedLength: number): void {
        this.updatedLength = updatedLength
        this.fileChatProvider = fileChatProvider
        this.ranges = []
        this.ranges.push(new vscode.Range(line, 0, line, 1))
        this._onDidChangeCodeLenses.fire()
    }

    public remove(): void {
        this.ranges = []
    }

    public provideCodeLenses(
        document: vscode.TextDocument,
        token: vscode.CancellationToken
    ): vscode.CodeLens[] | Thenable<vscode.CodeLens[]> {
        if (!document || !token) {
            return []
        }
        this.fileUri = document.uri
        return this.fixupLenses(document)
    }

    private fixupLenses(document: vscode.TextDocument): vscode.CodeLens[] {
        const uri = vscode.Uri.parse('codyDoc:' + this.id)
        const codeLenses: vscode.CodeLens[] = []
        for (const range of this.ranges) {
            void this.decorate(range)
            const codeLensTitle = new vscode.CodeLens(range)
            // Open Chat View
            codeLensTitle.command = {
                title: 'Edited by Cody',
                tooltip: 'Open this in Cody chat view',
                command: 'cody.focus',
            }
            // Run VS Code command to show diff for current file
            const codeLensDiff = new vscode.CodeLens(range)
            codeLensDiff.command = {
                title: 'Show Diff',
                tooltip: 'Open a diff of this change',
                command: 'vscode.diff',
                arguments: [uri, document.uri, 'Cody edit diff'],
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

    /**
     * Highlights line where the codes updated by Cody are located.
     */
    public async decorate(range: vscode.Range): Promise<void> {
        if (!this.fileUri) {
            return
        }

        const editor = vscode.window.activeTextEditor
        if (editor?.document.uri.fsPath !== this.fileUri.fsPath) {
            return
        }

        const currentFile = await vscode.workspace.openTextDocument(this.fileUri)
        if (!currentFile) {
            return
        }

        const decorations: vscode.DecorationOptions[] = []
        if (range && this.fileChatProvider) {
            const start = new vscode.Position(range.start.line, 0)
            const end = new vscode.Position(range.end.line - this.updatedLength + 1, 0)
            const newRange = new vscode.Range(start, end)
            decorations.push({
                range: newRange,
                hoverMessage: this.id,
            })
            range = newRange
            this.fileChatProvider.setNewRange(newRange, this.updatedLength)
        }
        await vscode.window.showTextDocument(this.fileUri)
        vscode.window.activeTextEditor?.setDecorations(this.decorationType, decorations)
    }

    /**
     * Remove all decorations on save / accept button click
     */
    public async removeDecorate(): Promise<void> {
        if (!this.fileUri) {
            return
        }
        await vscode.workspace.openTextDocument(this.fileUri)
        const editor = vscode.window.activeTextEditor
        const decorationType = vscode.window.createTextEditorDecorationType({})
        editor?.setDecorations(decorationType, [])
        this.updatedLength = 0
        // TODO remove individually
        this.fileChatProvider?.contentProvider.delete(this.id)
        this.fileChatProvider?.contentProvider.deleteByFilePath(this.fileUri.fsPath)
    }

    /**
     * Define styles
     */
    private decorationType = vscode.window.createTextEditorDecorationType({
        isWholeLine: true,
        borderWidth: '1px',
        borderStyle: 'solid',
        before: { contentText: '✨ ' },
        backgroundColor: 'rgba(161, 18, 255, 0.33)',
        overviewRulerColor: 'rgba(161, 18, 255, 0.33)',
        overviewRulerLane: vscode.OverviewRulerLane.Right,
        light: {
            borderColor: 'rgba(161, 18, 255, 0.33)',
        },
        dark: {
            borderColor: 'rgba(161, 18, 255, 0.33)',
        },
    })

    public dispose(): void {
        for (const disposable of this._disposables) {
            disposable.dispose()
        }
        this._disposables = []
    }
}
