import { Position, Range } from '@sourcegraph/extension-api-classes'
import * as clientTypes from '@sourcegraph/extension-api-types'
import * as sourcegraph from 'sourcegraph'
import { TextEdit } from './textEdit'

export interface FileOperationOptions {
    readonly overwrite?: boolean
    readonly ignoreIfExists?: boolean
    readonly ignoreIfNotExists?: boolean
    readonly recursive?: boolean
}

export enum WorkspaceEditOperationType {
    FileOperation = 0,
    FileTextEdit = 1,
}

export interface FileOperation {
    readonly type: WorkspaceEditOperationType.FileOperation
    readonly from?: URL
    readonly to?: URL
    readonly options?: FileOperationOptions
}

export interface FileTextEdit {
    readonly type: WorkspaceEditOperationType.FileTextEdit
    readonly uri: URL
    readonly edit: sourcegraph.TextEdit
}

export type WorkspaceEditOperation = FileOperation | FileTextEdit

type JSONFileOperation = Pick<FileOperation, Exclude<keyof FileOperation, 'from' | 'to'>> & {
    from?: URL | string
    to?: URL | string
}
type JSONFileTextEdit = Pick<FileTextEdit, Exclude<keyof FileTextEdit, 'uri' | 'edit'>> & {
    uri: URL | string
    edit: clientTypes.TextEdit
}

export class WorkspaceEdit implements sourcegraph.WorkspaceEdit {
    public operations: WorkspaceEditOperation[] = []

    public textEdits(): IterableIterator<[URL, sourcegraph.TextEdit[]]> {
        return this.groupedEntries().values()
    }

    private groupedEntries(): Map<string, [URL, sourcegraph.TextEdit[]]> {
        const textEdits = new Map<string, [URL, sourcegraph.TextEdit[]]>()
        for (const candidate of this.operations) {
            if (candidate.type === WorkspaceEditOperationType.FileTextEdit) {
                let textEdit = textEdits.get(candidate.uri.toString())
                if (!textEdit) {
                    textEdit = [candidate.uri, []]
                    textEdits.set(candidate.uri.toString(), textEdit)
                }
                textEdit[1].push(candidate.edit)
            }
        }
        return textEdits
    }

    public get(uri: URL): sourcegraph.TextEdit[] {
        const res: sourcegraph.TextEdit[] = []
        for (const candidate of this.operations) {
            if (
                candidate.type === WorkspaceEditOperationType.FileTextEdit &&
                candidate.uri.toString() === uri.toString()
            ) {
                res.push(candidate.edit)
            }
        }
        return res
    }

    public has(uri: URL): boolean {
        for (const edit of this.operations) {
            if (edit.type === WorkspaceEditOperationType.FileTextEdit && edit.uri.toString() === uri.toString()) {
                return true
            }
        }
        return false
    }

    public set(uri: URL, edits?: sourcegraph.TextEdit[]): void {
        if (!edits) {
            // Remove all text edits for `uri`.
            this.operations = this.operations.filter(
                edit =>
                    !(edit.type === WorkspaceEditOperationType.FileTextEdit && edit.uri.toString() === uri.toString())
            )
        } else {
            // Append edit to the end.
            for (const edit of edits) {
                if (edit) {
                    this.operations.push({ type: WorkspaceEditOperationType.FileTextEdit, uri, edit })
                }
            }
        }
    }

    public createFile(uri: URL, options?: { overwrite?: boolean; ignoreIfExists?: boolean }): void {
        this.operations.push({ type: WorkspaceEditOperationType.FileOperation, from: undefined, to: uri, options })
    }

    public deleteFile(uri: URL, options?: { recursive?: boolean; ignoreIfNotExists?: boolean }): void {
        this.operations.push({ type: WorkspaceEditOperationType.FileOperation, from: uri, to: undefined, options })
    }

    public renameFile(from: URL, to: URL, options?: { overwrite?: boolean; ignoreIfExists?: boolean }): void {
        this.operations.push({ type: WorkspaceEditOperationType.FileOperation, from, to, options })
    }

    public replace(uri: URL, range: Range, newText: string): void {
        this.operations.push({ type: WorkspaceEditOperationType.FileTextEdit, uri, edit: new TextEdit(range, newText) })
    }

    public insert(resource: URL, position: Position, newText: string): void {
        this.replace(resource, new Range(position, position), newText)
    }

    public delete(resource: URL, range: Range): void {
        this.replace(resource, range, '')
    }

    public toJSON(): { operations: (JSONFileTextEdit | JSONFileOperation)[] } {
        return {
            operations: this.operations.map(op => {
                if (op.type === WorkspaceEditOperationType.FileOperation) {
                    return {
                        ...op,
                        from: op.from && op.from.toJSON(),
                        to: op.to && op.to.toJSON(),
                    }
                }
                return {
                    ...op,
                    uri: op.uri.toJSON(),
                    edit: op.edit.toJSON(),
                }
            }),
        }
    }

    public static fromJSON(arg: ReturnType<(typeof WorkspaceEdit)['prototype']['toJSON']>): WorkspaceEdit {
        const workspaceEdit = new WorkspaceEdit()
        workspaceEdit.operations = arg.operations.map(op => {
            if (op.type === WorkspaceEditOperationType.FileOperation) {
                return {
                    ...op,
                    from: op.from !== undefined && typeof op.from === 'string' ? new URL(op.from) : op.from,
                    to: op.to !== undefined && typeof op.to === 'string' ? new URL(op.to) : op.to,
                }
            }
            return {
                ...op,
                uri: typeof op.uri === 'string' ? new URL(op.uri) : op.uri,
                edit: TextEdit.fromJSON(op.edit),
            }
        })
        return workspaceEdit
    }
}
