import { Position, TextDocumentIdentifier, WorkspaceEdit } from 'vscode-languageserver-types'
import { RequestType } from '../jsonrpc2/messages'
import { TextDocumentRegistrationOptions } from './textDocument'

export interface RenameParams {
    /**
     * The document to rename.
     */
    textDocument: TextDocumentIdentifier

    /**
     * The position at which this request was sent.
     */
    position: Position

    /**
     * The new name of the symbol. If the given name is not valid the
     * request must return a [ResponseError](#ResponseError) with an
     * appropriate message set.
     */
    newName: string
}

/**
 * A request to rename a symbol.
 */
export namespace RenameRequest {
    export const type = new RequestType<RenameParams, WorkspaceEdit | null, void, TextDocumentRegistrationOptions>(
        'textDocument/rename'
    )
}
