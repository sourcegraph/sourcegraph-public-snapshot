import * as vscode from 'vscode'

import { FileChatMessage } from './FileChatProvider'

/**
 * Stores the content of the documents that Cody is about to perform fix up for the source control diff
 */
export class CodyContentProvider implements vscode.TextDocumentContentProvider, vscode.Disposable {
    // File contents store where key = fileName
    private diffMap = new Map<string, string[]>()
    private editedFiles = new Map<string, string>()
    private _onDidChange = new vscode.EventEmitter<vscode.Uri>()
    private _subscriptions: vscode.Disposable

    constructor() {
        this._subscriptions = vscode.workspace.onDidCloseTextDocument(doc => this.deleteByFilePath(doc.uri.fsPath))
    }
    // Get content from the content store
    public provideTextDocumentContent(uri: vscode.Uri): string {
        return this.editedFiles.get(uri.fsPath) || ''
    }
    // Add to store - store origin content by fixup comment id
    public async set(uri: vscode.Uri, id: string): Promise<void> {
        this.diffMap.set(uri.fsPath, [...id])
        const activeDocument = await vscode.workspace.openTextDocument(uri)
        this.editedFiles.set(id, activeDocument.getText())
    }

    // Remove by ID
    public delete(id: string): void {
        this.editedFiles.delete(id)
    }

    // Remove by file name / fs path
    public deleteByFilePath(fileName: string): void {
        const files = this.diffMap.get(fileName)
        if (!files) {
            return
        }
        for (const id of files) {
            this.editedFiles.delete(id)
        }
    }

    // Remove all versions belong to a thread
    public deleteByThread(comments: FileChatMessage[]): void {
        comments.map(comment => this.delete(comment.id))
    }

    public get onDidChange(): vscode.Event<vscode.Uri> {
        return this._onDidChange.event
    }

    public dispose(): void {
        this._subscriptions.dispose()
        this._onDidChange.dispose()
        this.editedFiles = new Map<string, string>()
    }
}
