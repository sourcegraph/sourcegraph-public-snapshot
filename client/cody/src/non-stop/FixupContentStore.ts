import * as vscode from 'vscode'

import { taskID } from './FixupTask'

type fileName = string
type fileContent = string
/**
 * Stores the content of the documents that Cody is about to perform fix up for the source control diff
 */
export class ContentProvider implements vscode.TextDocumentContentProvider, vscode.Disposable {
    // This stores the content of the document for each task ID
    // The content is initialized by the fixup task with the original content
    // and then updated by the fixup task with the replacement content
    private contentStore = new Map<taskID, fileContent>()
    // This tracks the task IDs belong toe each file path
    private tasksByFilePath = new Map<fileName, taskID[]>()
    private _onDidChange = new vscode.EventEmitter<vscode.Uri>()
    private _disposables: vscode.Disposable

    constructor() {
        // TODO: Handle applying fixups to files which are opened and closed.
        // This is tricky because we need to re-sync the range we are tracking
        // when the file is opened.
        this._disposables = vscode.workspace.onDidCloseTextDocument(doc => this.deleteByFilePath(doc.uri.fsPath))
    }
    // Get content from the content store
    public provideTextDocumentContent(uri: vscode.Uri): string | null {
        const id = uri.fragment
        return this.contentStore.get(id) || null
    }
    // Add to store - store origin content by fixup task id
    public async set(id: string, docUri: vscode.Uri): Promise<void> {
        const doc = await vscode.workspace.openTextDocument(docUri)
        this.contentStore.set(id, doc.getText())
        this.tasksByFilePath.set(docUri.fsPath, [...(this.tasksByFilePath.get(docUri.fsPath) || []), id])
    }

    // Remove by ID
    public delete(id: string): void {
        this.contentStore.delete(id)
        // remove task from tasksByFilePath
        for (const [filePath, tasks] of this.tasksByFilePath) {
            const index = tasks.indexOf(id)
            if (index > -1) {
                tasks.splice(index, 1)
            }
            if (tasks.length === 0) {
                this.deleteByFilePath(filePath)
            }
        }
    }

    // Remove by file path
    public deleteByFilePath(fileName: string): void {
        const files = this.tasksByFilePath.get(fileName)
        if (!files) {
            return
        }
        for (const id of files) {
            this.contentStore.delete(id)
        }
    }

    public get onDidChange(): vscode.Event<vscode.Uri> {
        return this._onDidChange.event
    }

    public dispose(): void {
        this._disposables.dispose()
        this._onDidChange.dispose()
        this.contentStore = new Map<taskID, fileContent>()
        this.tasksByFilePath = new Map<fileName, taskID[]>()
    }
}
