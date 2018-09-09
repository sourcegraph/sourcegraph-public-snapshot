import { RequestHandler } from '../jsonrpc2/handlers'
import { RequestType } from '../jsonrpc2/messages'
import { Definition } from '../types/location'
import { TextDocumentPositionParams, TextDocumentRegistrationOptions } from './textDocument'

export interface ImplementationClientCapabilities {
    /**
     * The text document client capabilities
     */
    textDocument?: {
        /**
         * Capabilities specific to the `textDocument/implementation`
         */
        implementation?: {
            /**
             * Whether implementation supports dynamic registration.
             */
            dynamicRegistration?: boolean
        }
    }
}

export interface ImplementationServerCapabilities {
    /**
     * The server provides Goto Implementation support.
     */
    implementationProvider?: boolean | TextDocumentRegistrationOptions
}

/**
 * A request to resolve the implementation locations of a symbol at a given text
 * document position. The request's parameter is of type [TextDocumentPositioParams]
 * (#TextDocumentPositionParams) the response is of type [Definition](#Definition) or a
 * Thenable that resolves to such.
 */
export namespace ImplementationRequest {
    export const type = new RequestType<
        TextDocumentPositionParams,
        Definition | null,
        void,
        TextDocumentRegistrationOptions
    >('textDocument/implementation')
    export type HandlerSignature = RequestHandler<TextDocumentPositionParams, Definition | null, void>
}
