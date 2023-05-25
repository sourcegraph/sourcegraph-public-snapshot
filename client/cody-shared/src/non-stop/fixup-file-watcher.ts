import * as vscode from 'vscode'

// A handle to a fixup file. FixupFileWatcher is the factory for these; do not
// construct them directly.
export class FixupFile {
    constructor(private id_: number, public uri_: vscode.Uri) {}

    public deleted_ = false

    public get isDeleted(): boolean {
        return this.deleted_
    }

    public get uri(): vscode.Uri {
        return this.uri_
    }

    public toString(): string {
        return `FixupFile${this.id_}(${this.uri_})`
    }

    // TODO: Add convenience properties for the file name, type and a change
    // notification so the tree view can track file renames and deletions
}

// Watches documents for renaming and deletion. Hands out handles for documents
// which are durable across renames and the documents being closed and reopened.
// (The vscode.TextDocument object is *not* durable in this way.)
export class FixupFileWatcher implements vscode.Disposable {
    private subscriptions_: vscode.Disposable[] = []
    private uriToFile_: Map<vscode.Uri, FixupFile> = new Map()

    private n_ = 0 // cookie for generating new ids

    constructor(workspace: typeof vscode.workspace) {
        this.subscriptions_.push(workspace.onDidRenameFiles(this.didRenameFiles.bind(this)))
        this.subscriptions_.push(workspace.onDidDeleteFiles(this.didDeleteFiles.bind(this)))
    }

    public toFixupFileId(uri: vscode.Uri): FixupFile {
        let result = this.uriToFile_.get(uri)
        if (!result) {
            result = this.newFile(uri)
            this.uriToFile_.set(uri, result)
        }
        return result
    }

    private newFile(uri: vscode.Uri): FixupFile {
        return new FixupFile(this.n_++, uri)
    }

    private didDeleteFiles(event: vscode.FileDeleteEvent): void {
        // TODO: There is only one delete event for a folder. Scan all of the
        // Uris to find sub-files and compute their new name.
        for (const uri of event.files) {
            const file = this.uriToFile_.get(uri)
            if (file) {
                file.deleted_ = true
                this.uriToFile_.delete(uri)
            }
        }
    }

    private didRenameFiles(event: vscode.FileRenameEvent): void {
        // TODO: There is only one rename event for a folder. Scan all of the
        // Uris to find sub-files and compute their new name.
        for (const { oldUri, newUri } of event.files) {
            const file = this.uriToFile_.get(oldUri)
            if (file) {
                this.uriToFile_.delete(oldUri)
                this.uriToFile_.set(newUri, file)
                file.uri_ = newUri
            }
        }
    }

    public dispose(): void {
        for (const subscription of this.subscriptions_) {
            subscription.dispose()
        }
    }
}
