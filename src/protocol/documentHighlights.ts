import { DocumentHighlight } from 'vscode-languageserver-types'
import { RequestType } from '../jsonrpc2/messages'
import { TextDocumentPositionParams, TextDocumentRegistrationOptions } from './textDocument'

/**
 * Request to resolve a [DocumentHighlight](#DocumentHighlight) for a given
 * text document position. The request's parameter is of type [TextDocumentPosition]
 * (#TextDocumentPosition) the request response is of type [DocumentHighlight[]]
 * (#DocumentHighlight) or a Thenable that resolves to such.
 */
export namespace DocumentHighlightRequest {
    export const type = new RequestType<
        TextDocumentPositionParams,
        DocumentHighlight[] | null,
        void,
        TextDocumentRegistrationOptions
    >('textDocument/documentHighlight')
}
