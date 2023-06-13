import * as vscode from 'vscode'

import { CodyTaskState } from '../non-stop/utils'

import { DecorationProvider } from './DecorationProvider'
import { editDocByUri, getSingleLineRange, updateRangeOnDocChange } from './InlineAssist'

export class CodeLensProvider implements vscode.CodeLensProvider {
    private selectionRange: vscode.Range | null = null
    private contextStore = new Map<string, { docUri: vscode.Uri; original: string; replacement: string }>()

    private status = CodyTaskState.idle
    public decorator: DecorationProvider

    private _disposables: vscode.Disposable[] = []
    private _onDidChangeCodeLenses: vscode.EventEmitter<void> = new vscode.EventEmitter<void>()
    public readonly onDidChangeCodeLenses: vscode.Event<void> = this._onDidChangeCodeLenses.event

    constructor(public id: string, private extPath: string, private thread: vscode.CommentThread) {
        this.provideCodeLenses = this.provideCodeLenses.bind(this)
        this.decorator = new DecorationProvider(this.id, this.extPath)
        vscode.workspace.onDidChangeTextDocument(e => {
            if (e.document.uri.fsPath !== this.thread?.uri.fsPath) {
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
     * Define Current States
     */
    public updateState(state: CodyTaskState, newRange: vscode.Range): void {
        this.status = state
        this.decorator.setState(state, newRange)
        void this.decorator.decorate(newRange)
        this.selectionRange = newRange
        this._onDidChangeCodeLenses.fire()
    }

    public storeContext(id: string, docUri: vscode.Uri, original: string, replacement: string): void {
        this.contextStore.set(id, { docUri, original, replacement })
    }

    public async undo(id: string): Promise<void> {
        const context = this.contextStore.get(id)
        const chatSelection = this.selectionRange
        if (!context || !chatSelection) {
            return
        }
        const range = new vscode.Selection(chatSelection.start, new vscode.Position(chatSelection.end.line + 1, 0))
        await editDocByUri(context.docUri, { start: range.start.line, end: range.end.line }, context.original + '\n')
        this.remove()
    }
    /**
     * Remove all lenses and decorations created for task
     */
    public remove(): void {
        this.decorator.remove()
        this.selectionRange = null
        this.status = CodyTaskState.idle
        this.thread.dispose()
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
        if (!document || !token || document.uri.fsPath !== this.thread.uri.fsPath) {
            return []
        }
        this.decorator.setFileUri(this.thread.uri)
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
            ? getErrorLenses(this.id, codeLensRange)
            : getLenses(this.id, codeLensRange, this.isPending())
    }
    /**
     * Check if the file path is the same
     */
    private removeOnFSPath(uri: vscode.Uri): void {
        if (this.status === CodyTaskState.asking) {
            return
        }
        if (uri.fsPath === this.thread.uri.fsPath) {
            this.remove()
        }
    }
    /**
     * Check if it is in pending state
     */
    public isPending(): boolean {
        return this.status === CodyTaskState.asking
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

function getLenses(id: string, codeLensRange: vscode.Range, isPending: boolean): vscode.CodeLens[] {
    const title = new vscode.CodeLens(codeLensRange)
    // Open Chat View
    title.command = {
        title: isPending ? '$(sync~spin) Asking Cody...' : '✨ Edited by Cody',
        tooltip: 'Open Cody chat view',
        command: 'cody.focus',
    }
    const undo = getInlineUndoLens(id, codeLensRange)
    const close = getInlineCloseLens(id, codeLensRange)
    return isPending ? [title, close] : [title, undo, close]
}

function getErrorLenses(id: string, codeLensRange: vscode.Range): vscode.CodeLens[] {
    const title = new vscode.CodeLens(codeLensRange)
    title.command = {
        title: '⛔️ Not Edited by Cody',
        tooltip: 'Open Cody chat view',
        command: 'cody.focus',
    }
    const close = getInlineCloseLens(id, codeLensRange)
    return [title, close]
}

function getInlineCloseLens(id: string, codeLensRange: vscode.Range): vscode.CodeLens {
    const lens = new vscode.CodeLens(codeLensRange)
    lens.command = {
        title: 'Close',
        tooltip: 'Click to remove decorations',
        command: 'cody.inline.decorations.remove',
        arguments: [id],
    }
    return lens
}

function getInlineUndoLens(id: string, codeLensRange: vscode.Range): vscode.CodeLens {
    const lens = new vscode.CodeLens(codeLensRange)
    lens.command = {
        title: '$(undo) Undo',
        tooltip: 'Undo this change',
        command: 'cody.inline.fix.undo',
        arguments: [id],
    }
    return lens
}
