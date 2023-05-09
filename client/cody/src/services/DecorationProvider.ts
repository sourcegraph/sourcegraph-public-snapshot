import * as vscode from 'vscode'

const initDecorationType = vscode.window.createTextEditorDecorationType({})

export class DecorationProvider {
    private iconPath: vscode.Uri
    public ranges: vscode.Range | null = null
    public fileUri: vscode.Uri | null = null
    private status = 'none'

    private decorations: vscode.DecorationOptions[] = []
    private decorationsForIcon: vscode.DecorationOptions[] = []

    private decorationTypePending = this.makeDecorationType('pending')
    private decorationTypeDiff = this.makeDecorationType('diff')
    private decorationTypeIcon = initDecorationType

    private static decorates: DecorationProvider
    private _disposables: vscode.Disposable[] = []
    private _onDidChange: vscode.EventEmitter<void> = new vscode.EventEmitter<void>()
    public readonly onDidChange: vscode.Event<void> = this._onDidChange.event

    constructor(public id = '', private extPath = '') {
        // set up icon and register decoration types
        const extensionPath = vscode.Uri.file(this.extPath)
        const webviewPath = vscode.Uri.joinPath(extensionPath, 'dist')
        this.iconPath = vscode.Uri.joinPath(webviewPath, 'cody.png')
        this.decorationTypeIcon = this.makeDecorationType('icon')
        this._disposables.push(this.decorationTypeIcon, this.decorationTypeDiff, this.decorationTypePending)
    }

    public static get instance(): DecorationProvider {
        return (this.decorates ??= new this())
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

        await vscode.window.showTextDocument(this.fileUri)
        if (this.status === 'done') {
            this.decorationTypePending.dispose()
            this.decorations.push({
                range,
                hoverMessage: 'Cody Task#' + this.id,
            })
            this.decorationsForIcon.push({ range: this.singleLineRange(range.start.line) })
            // Add Cody icon to gutter
            vscode.window.activeTextEditor?.setDecorations(this.decorationTypeIcon, this.decorationsForIcon)
            vscode.window.activeTextEditor?.setDecorations(this.decorationTypeDiff, this.decorations)
        } else {
            vscode.window.activeTextEditor?.setDecorations(this.decorationTypePending, [range])
        }
    }

    public rangeUpdate(newRange: vscode.Range): void {
        this.ranges = newRange
        this._onDidChange.fire()
    }

    public singleLineRange(line: number): vscode.Range {
        return new vscode.Range(line, 0, line, 0)
    }

    /**
     * Clear all decorations
     */
    public async clear(): Promise<void> {
        if (!this.fileUri) {
            return
        }
        await vscode.workspace.openTextDocument(this.fileUri)
        this.decorationTypePending.dispose()
        this.decorationTypeIcon.dispose()
        this.decorationTypeDiff.dispose()
    }

    /**
     * Remove all lenses created for task
     */
    public remove(): void {
        this.ranges = null
        this.dispose()
    }

    public setStatus(status: string): void {
        this.status = status
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
