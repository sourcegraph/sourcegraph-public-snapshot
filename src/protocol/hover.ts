import { Hover } from 'sourcegraph'
import { RequestType } from './jsonrpc2/messages'
import { TextDocumentPositionParams, TextDocumentRegistrationOptions } from './textDocument'

/**
 * Request to request hover information at a given text document position. The request's
 * parameter is of type [TextDocumentPosition](#TextDocumentPosition) the response is of
 * type [Hover](#Hover) or a Thenable that resolves to such.
 */
export namespace HoverRequest {
    export const type = new RequestType<
        TextDocumentPositionParams,
        Hover | null,
        void,
        TextDocumentRegistrationOptions
    >('textDocument/hover')
}
