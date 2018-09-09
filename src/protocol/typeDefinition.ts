import { RequestHandler } from '../jsonrpc2/handlers'
import { RequestType } from '../jsonrpc2/messages'
import { Definition } from '../types/location'
import { TextDocumentPositionParams, TextDocumentRegistrationOptions } from './textDocument'

export interface TypeDefinitionClientCapabilities {
    /**
     * The text document client capabilities
     */
    textDocument?: {
        /**
         * Capabilities specific to the `textDocument/typeDefinition`
         */
        typeDefinition?: {
            /**
             * Whether implementation supports dynamic registration.
             */
            dynamicRegistration?: boolean
        }
    }
}

export interface TypeDefinitionServerCapabilities {
    /**
     * The server provides Goto Type Definition support.
     */
    typeDefinitionProvider?: boolean | TextDocumentRegistrationOptions
}

/**
 * A request to resolve the type definition locations of a symbol at a given text
 * document position. The request's parameter is of type [TextDocumentPositioParams]
 * (#TextDocumentPositionParams) the response is of type [Definition](#Definition) or a
 * Thenable that resolves to such.
 */
export namespace TypeDefinitionRequest {
    export const type = new RequestType<
        TextDocumentPositionParams,
        Definition | null,
        void,
        TextDocumentRegistrationOptions
    >('textDocument/typeDefinition')
    export type HandlerSignature = RequestHandler<TextDocumentPositionParams, Definition | null, void>
}
