import * as vscode from 'vscode'

import { FixupFile } from './FixupFile'

/**
 * Watches documents for renaming and deletion. Hands out handles for documents
 * which are durable across renames and the documents being closed and reopened.
 * (The vscode.TextDocument object is *not* durable in this way.)
 */
export class FixupFileObserver {
    private uriToFile_: Map<vscode.Uri, FixupFile> = new Map()

    private n_ = 0 // cookie for generating new ids

    // TODO: Design memory management. There's no protocol for throwing away a
    // FixupFile.
    // TODO: Consider tracking documents being closed.

    /**
     * Given a document URI, provides the corresponding FixupFile. As the
     * document is renamed or deleted the FixupFile will be updated to provide
     * the current file URI. This creates a FixupFile if one does not exist and
     * starts tracking it; see maybeForUri.
     * @param uri the URI of the document to monitor.
     * @returns a new FixupFile representing the document.
     */
    public forUri(uri: vscode.Uri): FixupFile {
        let result = this.uriToFile_.get(uri)
        if (!result) {
            result = this.newFile(uri)
            this.uriToFile_.set(uri, result)
        }
        return result
    }

    /**
     * Gets the FixupFile for a given URI, if one exists. This operation is
     * fast; vscode event sinks which are provided a URI can use this to quickly
     * check whether the file may have fixups.
     * @param uri the URI of the document of interest.
     * @returns a FixupFile representing the document, if one exists.
     */
    public maybeForUri(uri: vscode.Uri): FixupFile | undefined {
        return this.uriToFile_.get(uri)
    }

    private newFile(uri: vscode.Uri): FixupFile {
        return new FixupFile(this.n_++, uri)
    }

    public didDeleteFiles(event: vscode.FileDeleteEvent): void {
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

    public didRenameFiles(event: vscode.FileRenameEvent): void {
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
}
