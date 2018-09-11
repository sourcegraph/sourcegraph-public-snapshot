import { RequestType } from './jsonrpc2/messages'
import { Definition } from './plainTypes'
import { TextDocumentPositionParams, TextDocumentRegistrationOptions } from './textDocument'

/**
 * A request to resolve the definition location of a symbol at a given text
 * document position. The request's parameter is of type [TextDocumentPosition]
 * (#TextDocumentPosition) the response is of type [Definition](#Definition) or a
 * Thenable that resolves to such.
 */
export namespace DefinitionRequest {
    export const type = new RequestType<
        TextDocumentPositionParams,
        Definition | null,
        void,
        TextDocumentRegistrationOptions
    >('textDocument/definition')
}
