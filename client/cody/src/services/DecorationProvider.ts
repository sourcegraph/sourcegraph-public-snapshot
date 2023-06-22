import * as vscode from 'vscode'

import { CodyTaskState } from '../non-stop/utils'

import { getIconPath, getSingleLineRange } from './InlineAssist'

const initDecorationType = vscode.window.createTextEditorDecorationType({})

export class DecorationProvider {
    private iconPath: vscode.Uri
    private status = CodyTaskState.idle
    private range = new vscode.Range(0, 0, 0, 0)
    private editor: vscode.TextEditor | undefined

    private decorations: vscode.DecorationOptions[] = []
    private decorationsForIcon: vscode.DecorationOptions[] = []

    private decorationTypePending = this.makeDecorationType('pending')
    private decorationTypeDiff = this.makeDecorationType('diff')
    private decorationTypeError = this.makeDecorationType('error')
    private decorationTypeIcon = initDecorationType

    private _disposables: vscode.Disposable[] = []
    private _onDidChange: vscode.EventEmitter<void> = new vscode.EventEmitter<void>()
    public readonly onDidChange: vscode.Event<void> = this._onDidChange.event

    constructor(public id: string, private extPath: string, private fileUri: vscode.Uri) {
        this.editor = vscode.window.visibleTextEditors.find(editor => editor.document.uri.fsPath === fileUri.fsPath)
        // set up icon and register decoration types
        this.iconPath = getIconPath('cody', this.extPath)
        this.decorationTypeIcon = this.makeDecorationType('icon')
        this._disposables.push(
            this.decorationTypeIcon,
            this.decorationTypeDiff,
            this.decorationTypePending,
            this.decorationTypeError
        )
        vscode.window.onDidChangeActiveTextEditor(
            editor => {
                if (editor?.document.uri.fsPath === this.fileUri.fsPath) {
                    this.editor = editor
                    this.decorate()
                }
            },
            null,
            this._disposables
        )
    }

    /**
     * Highlights line where the codes updated by Cody are located.
     */
    public decorate(): void {
        const range = this.range
        const editor = this.editor
        if (!editor) {
            return
        }
        this.clear()
        this.decorations.push({ range, hoverMessage: 'Cody Assist #' + this.id })
        const rangeStartLine = getSingleLineRange(range.start.line)
        if (this.status === CodyTaskState.error) {
            this.decorationTypePending.dispose()
            this.decorationsForIcon.push({ range: rangeStartLine })
            editor.setDecorations(this.decorationTypeError, this.decorations)
            return
        }
        if (this.status === CodyTaskState.fixed) {
            this.decorationTypePending.dispose()
            this.decorationsForIcon.push({ range: rangeStartLine })
            editor.setDecorations(this.decorationTypeIcon, this.decorationsForIcon)
            editor.setDecorations(this.decorationTypeDiff, this.decorations)
            return
        }
        editor.setDecorations(this.decorationTypePending, [
            { range, hoverMessage: 'Do not make changes to the highlighted code while Cody is working on it.' },
        ])
    }

    /**
     * Clear all decorations
     */
    public clear(): void {
        this.decorations = []
        this.decorationsForIcon = []
    }

    /**
     * Define Current States
     */
    public setState(status: CodyTaskState, newRange: vscode.Range): void {
        this.status = status
        this.range = new vscode.Range(newRange.start.line, 0, newRange.end.line - 1, 0)
        this.decorate()
        this._onDidChange.fire()
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
        if (type === 'error') {
            return errorDecorationType
        }
        return vscode.window.createTextEditorDecorationType({
            isWholeLine: true,
            borderWidth: '1px',
            borderStyle: 'solid',
            overviewRulerColor: type === 'pending' ? 'rgba(161, 18, 255, 0.33)' : 'rgb(0, 203, 236, 0.22)',
            backgroundColor: type === 'pending' ? 'rgb(0, 203, 236, 0.1)' : 'rgba(161, 18, 255, 0.1)',
            overviewRulerLane: vscode.OverviewRulerLane.Right,
            light: {
                borderColor: 'rgba(161, 18, 255, 0.33)',
            },
            dark: {
                borderColor: 'rgba(161, 18, 255, 0.33)',
            },
        })
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

const errorDecorationType = vscode.window.createTextEditorDecorationType({
    isWholeLine: true,
    overviewRulerColor: 'rgba(255, 38, 86, 0.3)',
    backgroundColor: 'rgba(255, 38, 86, 0.1)',
})
