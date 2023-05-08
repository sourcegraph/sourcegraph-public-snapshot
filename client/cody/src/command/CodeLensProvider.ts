import * as vscode from 'vscode'

import { FileChatProvider } from './FileChatProvider'

const initDecorationType = vscode.window.createTextEditorDecorationType({})

export class CodeLensProvider implements vscode.CodeLensProvider {
    public ranges: vscode.Range | null = null
    private static lenses: CodeLensProvider

    private iconPath: vscode.Uri
    private fileUri: vscode.Uri | null = null
    private fileChatProvider: FileChatProvider | null = null

    public isPending = false
    public status = 'none'

    private decorationTypePending = this.makeDecorationType('pending')
    private decorationTypeDiff = this.makeDecorationType('diff')
    private decorationTypeIcon = initDecorationType

    private _disposables: vscode.Disposable[] = []
    private _onDidChangeCodeLenses: vscode.EventEmitter<void> = new vscode.EventEmitter<void>()
    public readonly onDidChangeCodeLenses: vscode.Event<void> = this._onDidChangeCodeLenses.event

    constructor(public id = '', private extPath = '') {
        // set up icon and register decoration types
        const extensionPath = vscode.Uri.file(this.extPath)
        const webviewPath = vscode.Uri.joinPath(extensionPath, 'dist')
        this.iconPath = vscode.Uri.joinPath(webviewPath, 'cody.png')
        this.decorationTypeIcon = this.makeDecorationType('icon')
        this._disposables.push(this.decorationTypeIcon, this.decorationTypeDiff, this.decorationTypePending)

        vscode.workspace.onDidChangeTextDocument(e => {
            if (e.document.uri.fsPath === this.fileUri?.fsPath) {
                this._onDidChangeCodeLenses.fire()
            }
        })
        vscode.workspace.onDidCloseTextDocument(e => {
            if (e.uri.fsPath === this.fileUri?.fsPath) {
                this.remove()
                this._onDidChangeCodeLenses.fire()
            }
        })
        vscode.workspace.onDidSaveTextDocument(e => {
            if (e.uri.fsPath === this.fileUri?.fsPath) {
                this.remove()
                this._onDidChangeCodeLenses.fire()
            }
        })
    }

    public static get instance(): CodeLensProvider {
        return (this.lenses ??= new this())
    }

    public updatePendingStatus(state: boolean, newRange?: vscode.Range): void {
        this.isPending = state
        if (newRange) {
            this.ranges = newRange
        }
        this._onDidChangeCodeLenses.fire()
    }

    public rangeUpdate(newRange: vscode.Range): void {
        this.ranges = newRange
        this._onDidChangeCodeLenses.fire()
    }

    public addRange(range: vscode.Range): void {
        this.ranges = range
        this._onDidChangeCodeLenses.fire()
    }

    public getNewRange(line: number): vscode.Range {
        return new vscode.Range(line, 0, line, 0)
    }

    /**
     * Remove all lenses created for task
     */
    public remove(): void {
        this.ranges = null
        this.status = 'none'
        this.dispose()
    }

    public provideCodeLenses(
        document: vscode.TextDocument,
        token: vscode.CancellationToken
    ): vscode.CodeLens[] | Thenable<vscode.CodeLens[]> {
        if (!document || !token) {
            return []
        }
        this.fileUri = document.uri
        return this.makeLenses(document)
    }

    /**
     * Lenses to display above the code that Cody edited
     */
    private makeLenses(document: vscode.TextDocument): vscode.CodeLens[] {
        const uri = vscode.Uri.parse('codyDoc:' + this.id)
        const codeLenses: vscode.CodeLens[] = []
        if (!this.ranges) {
            return []
        }

        const range = this.ranges
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
            // Run VS Code command to show diff for current file
            const codeLensDiff = new vscode.CodeLens(codeLensRange)
            codeLensDiff.command = {
                title: '$(git-commit) Show Diff',
                tooltip: 'Open a diff of this change',
                command: 'vscode.diff',
                arguments: [uri, document.uri, 'Cody edit diff'],
            }

            // Remove decorations
            const codeLensSave = new vscode.CodeLens(codeLensRange)
            codeLensSave.command = {
                title: 'Accept',
                tooltip: 'Accept and save all changes',
                command: 'workbench.action.files.save',
            }

            codeLenses.push(codeLensDiff, codeLensSave)
        }
        void this.decorate(this.ranges)
        return codeLenses
    }

    /**
     * Highlights line where the codes updated by Cody are located.
     */
    public async decorate(range: vscode.Range): Promise<void> {
        if (!this.fileUri) {
            console.error('cant find file')
            return
        }
        const currentFile = await vscode.workspace.openTextDocument(this.fileUri)
        if (!currentFile || !range) {
            return
        }
        const decorations: vscode.DecorationOptions[] = []
        const decorationsForIcon: vscode.DecorationOptions[] = []

        await vscode.window.showTextDocument(this.fileUri)

        if (!this.isPending && this.status !== 'done') {
            this.decorationTypePending.dispose()
            decorations.push({
                range,
                hoverMessage: 'Cody Task#' + this.id,
            })
            decorationsForIcon.push({ range: this.getNewRange(range.start.line) })
            // Add Cody icon to gutter
            vscode.window.activeTextEditor?.setDecorations(this.decorationTypeIcon, decorationsForIcon)
            vscode.window.activeTextEditor?.setDecorations(this.decorationTypeDiff, decorations)
        } else {
            vscode.window.activeTextEditor?.setDecorations(this.decorationTypePending, [range])
        }
    }

    /**
     * Remove all decorations on save / accept button click
     */
    public async removeDecorate(): Promise<void> {
        if (!this.fileUri) {
            return
        }
        await vscode.workspace.openTextDocument(this.fileUri)
        this.decorationTypePending.dispose()
        this.decorationTypeIcon.dispose()
        this.decorationTypeDiff.dispose()
        // TODO remove individually
        this.fileChatProvider?.contentProvider.delete(this.id)
        this.fileChatProvider?.contentProvider.deleteByFilePath(this.fileUri.fsPath)
    }

    /**
     * Define styles
     */
    private makeDecorationType(type?: string): vscode.TextEditorDecorationType {
        if (type === 'icon') {
            return vscode.window.createTextEditorDecorationType({
                gutterIconPath: this.iconPath,
                gutterIconSize: 'contain',
            })
        }
        return vscode.window.createTextEditorDecorationType({
            isWholeLine: true,
            borderWidth: '1px',
            borderStyle: 'solid',
            overviewRulerColor: type === 'pending' ? 'rgba(161, 18, 255, 0.33)' : 'rgb(0, 203, 236, 0.22)',
            backgroundColor: type === 'pending' ? 'rgb(0, 203, 236, 0.22)' : 'rgba(161, 18, 255, 0.33)',
            overviewRulerLane: vscode.OverviewRulerLane.Right,
            light: {
                borderColor: 'rgba(161, 18, 255, 0.33)',
            },
            dark: {
                borderColor: 'rgba(161, 18, 255, 0.33)',
            },
        })
    }

    public dispose(): void {
        for (const disposable of this._disposables) {
            disposable.dispose()
        }
        this._disposables = []
    }
}
