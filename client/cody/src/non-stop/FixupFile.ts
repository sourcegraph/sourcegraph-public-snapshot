import * as vscode from 'vscode'

/**
 * A handle to a fixup file. FixupFileWatcher is the factory for these; do not
 * construct them directly.
 */
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
