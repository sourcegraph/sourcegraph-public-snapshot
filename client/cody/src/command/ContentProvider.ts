import * as vscode from 'vscode'

/**
 * Stores the content of the documents that Cody is about to perform fix up for the source control diff
 */
export class CodyContentProvider implements vscode.TextDocumentContentProvider, vscode.Disposable {
    // File contents store
    private fileContents = new Map<string, string>()
    private _onDidChange = new vscode.EventEmitter<vscode.Uri>()

    public provideTextDocumentContent(uri: vscode.Uri): string {
        return this.fileContents.get(uri.fsPath) || ''
    }

    public set(uri: vscode.Uri, content: string): void {
        this.fileContents.set(uri.fsPath, content)
    }

    public delete(uri: vscode.Uri): void {
        this.fileContents.delete(uri.fsPath)
    }

    public get onDidChange(): vscode.Event<vscode.Uri> {
        return this._onDidChange.event
    }

    public dispose(): void {
        this._onDidChange.dispose()
    }
}
