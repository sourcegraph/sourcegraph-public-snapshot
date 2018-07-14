import { DocumentSymbolParams, SymbolInformation, WorkspaceSymbolParams } from 'vscode-languageserver-types'
import { RequestType } from '../jsonrpc2/messages'
import { TextDocumentRegistrationOptions } from './textDocument'

/**
 * A request to list all symbols found in a given text document. The request's
 * parameter is of type [TextDocumentIdentifier](#TextDocumentIdentifier) the
 * response is of type [SymbolInformation[]](#SymbolInformation) or a Thenable
 * that resolves to such.
 */
export namespace DocumentSymbolRequest {
    export const type = new RequestType<
        DocumentSymbolParams,
        SymbolInformation[] | null,
        void,
        TextDocumentRegistrationOptions
    >('textDocument/documentSymbol')
}

/**
 * A request to list project-wide symbols matching the query string given
 * by the [WorkspaceSymbolParams](#WorkspaceSymbolParams). The response is
 * of type [SymbolInformation[]](#SymbolInformation) or a Thenable that
 * resolves to such.
 */
export namespace WorkspaceSymbolRequest {
    export const type = new RequestType<WorkspaceSymbolParams, SymbolInformation[] | null, void, void>(
        'workspace/symbol'
    )
}
