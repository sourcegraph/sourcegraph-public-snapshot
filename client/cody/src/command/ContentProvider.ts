import * as vscode from 'vscode'

import { FileChatMessage } from './FileChatProvider'

/**
 * Stores the content of the documents that Cody is about to perform fix up for the source control diff
 */
export class CodyContentProvider implements vscode.TextDocumentContentProvider, vscode.Disposable {
    // This stores the content of the document state when fixup is first initialized
    // the key is the task ID, and the value is the content of the document before edits
    // - Key: <string> task ID
    // - Value: <string> document content
    private fileSource = new Map<string, string>()
    // This tracks the task IDs belong toe each file path
    // - Key: <string> file name / file path
    // - Value: <string> Array of task IDs
    private diffMap = new Map<string, string[]>()
    private _onDidChange = new vscode.EventEmitter<vscode.Uri>()
    private _subscriptions: vscode.Disposable

    constructor() {
        this._subscriptions = vscode.workspace.onDidCloseTextDocument(doc => this.deleteByFilePath(doc.uri.fsPath))
    }
    // Get content from the content store
    public provideTextDocumentContent(uri: vscode.Uri): string {
        return this.fileSource.get(uri.fsPath) || ''
    }
    // Add to store - store origin content by fixup comment id
    public async set(uri: vscode.Uri, id: string): Promise<void> {
        this.diffMap.set(uri.fsPath, [...id])
        const activeDocument = await vscode.workspace.openTextDocument(uri)
        this.fileSource.set(id, activeDocument.getText())
    }

    // Remove by ID
    public delete(id: string): void {
        this.fileSource.delete(id)
    }

    // Remove by file name / fs path
    public deleteByFilePath(fileName: string): void {
        const files = this.diffMap.get(fileName)
        if (!files) {
            return
        }
        for (const id of files) {
            this.fileSource.delete(id)
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
        this.fileSource = new Map<string, string>()
    }
}
